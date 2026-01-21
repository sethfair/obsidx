package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/sethfair/obsidx/internal/ann"
	"github.com/sethfair/obsidx/internal/embed"
	"github.com/sethfair/obsidx/internal/indexer"
	"github.com/sethfair/obsidx/internal/store"
	"github.com/sethfair/obsidx/internal/watcher"
)

var (
	vaultDir   = flag.String("vault", "", "Path to Obsidian vault (required)")
	dbPath     = flag.String("db", ".obsidian-index/obsidx.db", "Path to SQLite database")
	indexDir   = flag.String("index", ".obsidian-index/hnsw", "Path to HNSW index directory")
	ollamaURL  = flag.String("ollama-url", "http://localhost:11434", "Ollama API endpoint")
	embedModel = flag.String("model", "nomic-embed-text", "Ollama embedding model (nomic-embed-text, all-minilm, etc)")
	watchMode  = flag.Bool("watch", false, "Watch mode: continuously monitor for changes")
	debounceMs = flag.Int("debounce", 500, "Debounce time in milliseconds for watch mode")
)

func main() {
	flag.Parse()

	if *vaultDir == "" {
		log.Fatal("--vault is required")
	}

	// Ensure index directory exists
	if err := os.MkdirAll(filepath.Dir(*dbPath), 0755); err != nil {
		log.Fatalf("Create db directory: %v", err)
	}
	if err := os.MkdirAll(*indexDir, 0755); err != nil {
		log.Fatalf("Create index directory: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("Shutting down...")
		cancel()
	}()

	// Initialize Ollama embedder
	embedder := embed.NewOllamaEmbedder(*ollamaURL, *embedModel, 0)

	// Test connection and get dimension
	if err := embedder.Ping(ctx); err != nil {
		log.Fatalf("Cannot connect to Ollama at %s: %v\n"+
			"Make sure Ollama is running and the model is installed:\n"+
			"  ollama serve\n"+
			"  ollama pull %s",
			*ollamaURL, err, *embedModel)
	}

	// Get actual dimension from a test embedding
	testVec, err := embedder.Embed(ctx, "test")
	if err != nil {
		log.Fatalf("Failed to generate test embedding: %v", err)
	}
	actualDim := len(testVec)
	log.Printf("Connected to Ollama - model: %s, dimension: %d\n", *embedModel, actualDim)

	// Initialize store
	st, err := store.Open(*dbPath, actualDim)
	if err != nil {
		log.Fatalf("Open store: %v", err)
	}
	defer st.Close()

	// Initialize ANN index
	annCfg := ann.DefaultHNSWConfig(actualDim)
	annIndex, err := ann.NewHNSW(annCfg)
	if err != nil {
		log.Fatalf("Create ANN index: %v", err)
	}
	defer annIndex.Close()

	// Check if we need to rebuild index
	if err := checkAndRebuild(ctx, st, annIndex, actualDim, *embedModel); err != nil {
		log.Fatalf("Check/rebuild index: %v", err)
	}

	// Create indexer
	idx := indexer.New(st, embedder, annIndex, *vaultDir)

	if *watchMode {
		// Watch mode: monitor for changes
		log.Printf("Starting watcher on %s\n", *vaultDir)

		changeCount := 0
		w, err := watcher.New(func(path string) {
			changeCount++
			relPath, _ := filepath.Rel(*vaultDir, path)
			log.Printf("📝 [%d] Detected change: %s", changeCount, relPath)
			if err := idx.IndexFile(ctx, path); err != nil {
				log.Printf("❌ Error indexing %s: %v\n", relPath, err)
			} else {
				log.Printf("✓ Re-indexed: %s\n", relPath)
			}
		}, time.Duration(*debounceMs)*time.Millisecond)
		if err != nil {
			log.Fatalf("Create watcher: %v", err)
		}
		defer w.Close()

		// Do initial full index
		log.Println("Performing initial full index...")
		if err := idx.IndexVault(ctx); err != nil {
			log.Printf("Initial index error: %v\n", err)
		}

		// Get stats after initial index
		activeCount, _ := st.GetActiveChunkCount(ctx)
		log.Printf("✓ Initial index complete - %d active chunks indexed\n", activeCount)
		log.Println("👀 Watching for changes... (Press Ctrl+C to stop)")
		log.Println("")

		// Start periodic heartbeat in background
		heartbeatTicker := time.NewTicker(5 * time.Minute)
		defer heartbeatTicker.Stop()
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case <-heartbeatTicker.C:
					count, _ := st.GetActiveChunkCount(ctx)
					log.Printf("💓 Still watching... (chunks: %d, changes processed: %d)\n", count, changeCount)
				}
			}
		}()

		// Start watching
		if err := w.Watch(ctx, *vaultDir); err != nil && err != context.Canceled {
			log.Fatalf("Watch error: %v", err)
		}
	} else {
		// One-shot mode: index and exit
		log.Printf("Indexing vault: %s\n", *vaultDir)
		if err := idx.IndexVault(ctx); err != nil {
			log.Fatalf("Index vault: %v", err)
		}
		log.Println("Indexing complete")
	}
}

// checkAndRebuild checks if index needs rebuilding and rebuilds if necessary
func checkAndRebuild(ctx context.Context, st *store.SQLite, annIndex ann.Index, dim int, model string) error {
	// Check index metadata
	storedDim, _ := st.GetIndexMetaInt(ctx, "dim")
	storedModel, _ := st.GetIndexMeta(ctx, "embedding_model_name")

	needsRebuild := false
	if storedDim != dim {
		log.Printf("Dimension changed: %d -> %d, rebuilding index\n", storedDim, dim)
		needsRebuild = true
	}
	if storedModel != model {
		log.Printf("Model changed: %s -> %s, rebuilding index\n", storedModel, model)
		needsRebuild = true
	}

	if needsRebuild {
		return rebuildIndex(ctx, st, annIndex, dim, model)
	}

	// Load existing vectors into HNSW
	log.Println("Loading existing embeddings into HNSW...")
	rows, err := st.StreamActiveEmbeddings(ctx)
	if err != nil {
		return fmt.Errorf("stream embeddings: %w", err)
	}
	defer rows.Close()

	count := 0
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

		if count%1000 == 0 {
			log.Printf("Loaded %d vectors...\n", count)
		}
	}

	log.Printf("Loaded %d vectors into HNSW\n", count)
	return nil
}

// rebuildIndex rebuilds the HNSW index from SQLite
func rebuildIndex(ctx context.Context, st *store.SQLite, annIndex ann.Index, dim int, model string) error {
	log.Println("Rebuilding HNSW index from SQLite...")

	rows, err := st.StreamActiveEmbeddings(ctx)
	if err != nil {
		return fmt.Errorf("stream embeddings: %w", err)
	}
	defer rows.Close()

	count := 0
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

		if count%1000 == 0 {
			log.Printf("Indexed %d vectors...\n", count)
		}
	}

	// Update metadata
	activeCount, err := st.GetActiveChunkCount(ctx)
	if err != nil {
		return fmt.Errorf("get active count: %w", err)
	}

	meta := map[string]string{
		"dim":                         fmt.Sprintf("%d", dim),
		"embedding_model_name":        model,
		"built_at_unix":               fmt.Sprintf("%d", time.Now().Unix()),
		"active_chunk_count_at_build": fmt.Sprintf("%d", activeCount),
	}

	if err := st.SetIndexMeta(ctx, meta); err != nil {
		return fmt.Errorf("set index meta: %w", err)
	}

	log.Printf("Rebuild complete: %d vectors\n", count)
	return nil
}
