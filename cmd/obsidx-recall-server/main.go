package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sethfair/obsidx/internal/ann"
	"github.com/sethfair/obsidx/internal/embed"
	"github.com/sethfair/obsidx/internal/rank"
	"github.com/sethfair/obsidx/internal/store"
)

var (
	dbPath     = flag.String("db", ".obsidian-index/obsidx.db", "Path to SQLite database")
	port       = flag.Int("port", 8765, "HTTP server port")
	ollamaURL  = flag.String("ollama-url", "http://localhost:11434", "Ollama API endpoint")
	embedModel = flag.String("model", "nomic-embed-text", "Ollama embedding model")
)

type Server struct {
	store    *store.SQLite
	embedder embed.Embedder
	annIndex ann.Index
	ctx      context.Context
}

type SearchRequest struct {
	Query      string `json:"query"`
	TopN       int    `json:"top_n"`
	CandidateK int    `json:"candidate_k"`
}

type SearchResponse struct {
	Results []ResultItem `json:"results"`
	Timing  TimingInfo   `json:"timing"`
	Error   string       `json:"error,omitempty"`
}

type ResultItem struct {
	Score          float32  `json:"score"`
	Path           string   `json:"path"`
	HeadingPath    string   `json:"heading_path"`
	Status         string   `json:"status"`
	Scope          string   `json:"scope"`
	StartLine      int      `json:"start_line"`
	EndLine        int      `json:"end_line"`
	Content        string   `json:"content"`
	CategoryWeight float32  `json:"category_weight"`
	Tags           []string `json:"tags"`
}

type TimingInfo struct {
	EmbedMs  int64 `json:"embed_ms"`
	SearchMs int64 `json:"search_ms"`
	FetchMs  int64 `json:"fetch_ms"`
	RerankMs int64 `json:"rerank_ms"`
	TotalMs  int64 `json:"total_ms"`
}

func main() {
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("Shutting down server...")
		cancel()
	}()

	log.Printf("🚀 Starting obsidx recall server on port %d", *port)
	log.Printf("📂 Database: %s", *dbPath)

	// Open store
	st, err := store.Open(*dbPath, 0)
	if err != nil {
		log.Fatalf("Failed to open store: %v", err)
	}
	defer st.Close()

	// Get stored dimension and model
	storedDim, err := st.GetIndexMetaInt(ctx, "dim")
	if err != nil || storedDim == 0 {
		log.Fatal("No indexed data found. Run obsidx-indexer first.")
	}

	storedModel, _ := st.GetIndexMeta(ctx, "embedding_model_name")
	log.Printf("📊 Index: dim=%d, model=%s", storedDim, storedModel)

	// Initialize embedder
	log.Printf("🔌 Connecting to Ollama at %s...", *ollamaURL)
	embedder := embed.NewOllamaEmbedder(*ollamaURL, *embedModel, storedDim)
	if err := embedder.Ping(ctx); err != nil {
		log.Fatalf("Cannot connect to Ollama: %v", err)
	}
	log.Printf("✓ Connected to Ollama")

	// Build HNSW index (one time!)
	log.Printf("🏗️  Building HNSW index...")
	annCfg := ann.DefaultHNSWConfig(storedDim)
	annIndex, err := ann.NewHNSW(annCfg)
	if err != nil {
		log.Fatalf("Failed to create HNSW index: %v", err)
	}
	defer annIndex.Close()

	// Load vectors from SQLite
	if err := loadIndex(ctx, st, annIndex); err != nil {
		log.Fatalf("Failed to load index: %v", err)
	}

	log.Printf("✅ Server ready - index loaded and cached in memory")
	log.Printf("   Searches will be <100ms (no index rebuild!)")
	log.Printf("")

	// Create server
	srv := &Server{
		store:    st,
		embedder: embedder,
		annIndex: annIndex,
		ctx:      ctx,
	}

	// Setup HTTP handlers
	http.HandleFunc("/search", srv.handleSearch)
	http.HandleFunc("/health", srv.handleHealth)
	http.HandleFunc("/stats", srv.handleStats)

	// Start server
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: http.DefaultServeMux,
	}

	go func() {
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("✓ Server stopped")
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	startTime := time.Now()
	var req SearchRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Defaults
	if req.TopN == 0 {
		req.TopN = 12
	}
	if req.CandidateK == 0 {
		req.CandidateK = 200
	}

	var timing TimingInfo

	// 1. Embed query
	embedStart := time.Now()
	queryVec, err := s.embedder.Embed(s.ctx, req.Query)
	if err != nil {
		s.sendError(w, fmt.Sprintf("Failed to embed query: %v", err), http.StatusInternalServerError)
		return
	}
	timing.EmbedMs = time.Since(embedStart).Milliseconds()

	// 2. HNSW search
	searchStart := time.Now()
	candidateIDs, err := s.annIndex.Search(queryVec, req.CandidateK)
	if err != nil {
		s.sendError(w, fmt.Sprintf("Search failed: %v", err), http.StatusInternalServerError)
		return
	}
	timing.SearchMs = time.Since(searchStart).Milliseconds()

	if len(candidateIDs) == 0 {
		s.sendResponse(w, &SearchResponse{
			Results: []ResultItem{},
			Timing:  timing,
		})
		return
	}

	// 3. Fetch chunks
	fetchStart := time.Now()
	chunks, err := s.store.GetChunksByIDs(s.ctx, candidateIDs)
	if err != nil {
		s.sendError(w, fmt.Sprintf("Failed to fetch chunks: %v", err), http.StatusInternalServerError)
		return
	}
	timing.FetchMs = time.Since(fetchStart).Milliseconds()

	// 4. Rerank
	rerankStart := time.Now()
	results := rank.RerankCosine(queryVec, chunks, req.TopN)
	timing.RerankMs = time.Since(rerankStart).Milliseconds()

	timing.TotalMs = time.Since(startTime).Milliseconds()

	// Convert to response format
	items := make([]ResultItem, len(results))
	for i, r := range results {
		items[i] = ResultItem{
			Score:          r.Score,
			Path:           r.Chunk.Path,
			HeadingPath:    r.Chunk.HeadingPath,
			Status:         r.Chunk.Status,
			Scope:          r.Chunk.Scope,
			StartLine:      r.Chunk.StartLine,
			EndLine:        r.Chunk.EndLine,
			Content:        r.Chunk.Content,
			CategoryWeight: r.Chunk.CategoryWeight,
			Tags:           r.Chunk.Tags,
		}
	}

	s.sendResponse(w, &SearchResponse{
		Results: items,
		Timing:  timing,
	})

	log.Printf("✓ Search: \"%s\" → %d results in %dms (embed:%dms, search:%dms, fetch:%dms, rerank:%dms)",
		req.Query, len(items), timing.TotalMs, timing.EmbedMs, timing.SearchMs, timing.FetchMs, timing.RerankMs)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      "ok",
		"index_size":  s.annIndex.Size(),
		"server_time": time.Now().Unix(),
	})
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	activeCount, _ := s.store.GetActiveChunkCount(s.ctx)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"index_vectors": s.annIndex.Size(),
		"active_chunks": activeCount,
		"db_path":       *dbPath,
	})
}

func (s *Server) sendResponse(w http.ResponseWriter, resp *SearchResponse) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) sendError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(&SearchResponse{
		Error: msg,
	})
}

func loadIndex(ctx context.Context, st *store.SQLite, annIndex ann.Index) error {
	rows, err := st.StreamActiveEmbeddings(ctx)
	if err != nil {
		return fmt.Errorf("stream embeddings: %w", err)
	}
	defer rows.Close()

	count := 0
	lastLog := time.Now()

	for rows.Next() {
		var id uint64
		var vecBlob []byte
		if err := rows.Scan(&id, &vecBlob); err != nil {
			return fmt.Errorf("scan row: %w", err)
		}

		vec, err := store.BytesToFloat32(vecBlob)
		if err != nil {
			return fmt.Errorf("decode vec: %w", err)
		}

		if err := annIndex.Add(id, vec); err != nil {
			return fmt.Errorf("add to index: %w", err)
		}
		count++

		if count%1000 == 0 || time.Since(lastLog) > 2*time.Second {
			log.Printf("   Loading vectors: %d...", count)
			lastLog = time.Now()
		}
	}

	log.Printf("✓ Loaded %d vectors into HNSW index", count)
	return nil
}
