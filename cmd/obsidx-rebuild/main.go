package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/sethfair/obsidx/internal/ann"
	"github.com/sethfair/obsidx/internal/store"
)

var (
	dbPath    = flag.String("db", ".obsidian-index/obsidx.db", "Path to SQLite database")
	dimension = flag.Int("dim", 768, "Embedding dimension")
	modelName = flag.String("model", "default", "Embedding model name")
)

func main() {
	flag.Parse()

	ctx := context.Background()

	// Open store
	st, err := store.Open(*dbPath, *dimension)
	if err != nil {
		log.Fatalf("Open store: %v", err)
	}
	defer st.Close()

	// Initialize ANN index
	annCfg := ann.DefaultHNSWConfig(*dimension)
	annIndex, err := ann.NewHNSW(annCfg)
	if err != nil {
		log.Fatalf("Create ANN index: %v", err)
	}
	defer annIndex.Close()

	// Rebuild
	if err := rebuild(ctx, st, annIndex); err != nil {
		log.Fatalf("Rebuild failed: %v", err)
	}

	log.Println("Rebuild complete!")
}

func rebuild(ctx context.Context, st *store.SQLite, annIndex ann.Index) error {
	log.Println("Starting rebuild from SQLite...")

	// Get current active chunk count
	activeCount, err := st.GetActiveChunkCount(ctx)
	if err != nil {
		return fmt.Errorf("get active count: %w", err)
	}

	log.Printf("Active chunks to rebuild: %d\n", activeCount)

	// Stream and add all active embeddings
	rows, err := st.StreamActiveEmbeddings(ctx)
	if err != nil {
		return fmt.Errorf("stream embeddings: %w", err)
	}
	defer rows.Close()

	count := 0
	startTime := time.Now()

	for rows.Next() {
		var id uint64
		var vecBlob []byte
		if err := rows.Scan(&id, &vecBlob); err != nil {
			return fmt.Errorf("scan row: %w", err)
		}

		vec, err := store.BytesToFloat32(vecBlob)
		if err != nil {
			return fmt.Errorf("decode vec for chunk %d: %w", id, err)
		}

		if err := annIndex.Add(id, vec); err != nil {
			return fmt.Errorf("add chunk %d to index: %w", id, err)
		}

		count++
		if count%1000 == 0 {
			elapsed := time.Since(startTime)
			rate := float64(count) / elapsed.Seconds()
			log.Printf("Progress: %d/%d vectors (%.1f vec/sec)\n", count, activeCount, rate)
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("iteration error: %w", err)
	}

	elapsed := time.Since(startTime)
	log.Printf("Indexed %d vectors in %v (%.1f vec/sec)\n",
		count, elapsed, float64(count)/elapsed.Seconds())

	// Update index metadata
	meta := map[string]string{
		"dim":                         fmt.Sprintf("%d", st.Dim()),
		"embedding_model_name":        *modelName,
		"built_at_unix":               fmt.Sprintf("%d", time.Now().Unix()),
		"active_chunk_count_at_build": fmt.Sprintf("%d", activeCount),
	}

	if err := st.SetIndexMeta(ctx, meta); err != nil {
		return fmt.Errorf("update index metadata: %w", err)
	}

	log.Println("Index metadata updated")
	return nil
}
