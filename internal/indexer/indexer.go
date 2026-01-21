package indexer

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sethfair/obsidx/internal/ann"
	"github.com/sethfair/obsidx/internal/chunker"
	"github.com/sethfair/obsidx/internal/embed"
	"github.com/sethfair/obsidx/internal/metadata"
	"github.com/sethfair/obsidx/internal/store"
)

// Indexer manages the indexing process
type Indexer struct {
	store    *store.SQLite
	embedder embed.Embedder
	annIndex ann.Index
	vaultDir string
}

// New creates a new indexer
func New(st *store.SQLite, embedder embed.Embedder, annIndex ann.Index, vaultDir string) *Indexer {
	return &Indexer{
		store:    st,
		embedder: embedder,
		annIndex: annIndex,
		vaultDir: vaultDir,
	}
}

// IndexFile processes a single file
func (idx *Indexer) IndexFile(ctx context.Context, path string) error {
	// Compute file hash
	fileHash, mtime, err := computeFileHash(path)
	if err != nil {
		return fmt.Errorf("hash file: %w", err)
	}

	// Check if file has changed
	existing, err := idx.store.GetFileInfo(ctx, path)
	if err != nil {
		return fmt.Errorf("get file info: %w", err)
	}

	if existing != nil && existing.SHA256 == fileHash {
		// File unchanged, skip
		return nil
	}

	// Read and chunk file
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	contentStr := string(content)

	// Extract metadata from front matter
	noteMeta := metadata.ParseFrontMatter(contentStr)

	// Infer category from path if not explicit
	if noteMeta.Category == "" {
		noteMeta.InferredCategory = metadata.InferCategoryFromPath(path)
	}

	// Get effective category and weight
	effectiveCategory := noteMeta.EffectiveCategory()
	categoryWeight := noteMeta.CombinedWeight()

	chunks := chunker.ChunkMarkdown(contentStr)
	if len(chunks) == 0 {
		return nil // Empty file
	}

	// Apply metadata to all chunks
	for i := range chunks {
		chunks[i].Category = effectiveCategory
		chunks[i].Status = noteMeta.Status
		chunks[i].Scope = noteMeta.Scope
		chunks[i].NoteType = noteMeta.Type
		chunks[i].CategoryWeight = categoryWeight
	}

	// Embed chunks, skipping empty ones
	type chunkWithVector struct {
		chunk  chunker.Chunk
		vector []float32
		index  int
	}

	validChunks := make([]chunkWithVector, 0, len(chunks))

	for i, chunk := range chunks {
		// Skip chunks that are too short
		trimmed := strings.TrimSpace(chunk.Content)
		if len(trimmed) < 10 {
			continue
		}

		vec, err := idx.embedder.Embed(ctx, chunk.Content)
		if err != nil {
			// Log but continue with other chunks
			fmt.Printf("  Warning: embed chunk %d failed: %v\n", i, err)
			continue
		}

		// Skip empty embeddings
		if len(vec) == 0 {
			fmt.Printf("  Warning: empty embedding for chunk %d, skipping\n", i)
			continue
		}

		validChunks = append(validChunks, chunkWithVector{
			chunk:  chunk,
			vector: vec,
			index:  i,
		})
	}

	if len(validChunks) == 0 {
		return nil // No valid chunks to index
	}

	// Store in SQLite with transaction
	tx, err := idx.store.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Mark existing chunks inactive
	if err := idx.store.MarkChunksInactive(ctx, tx, path); err != nil {
		return fmt.Errorf("mark chunks inactive: %w", err)
	}

	// Insert new chunks and embeddings
	for _, cwv := range validChunks {
		storeChunk := &store.Chunk{
			Path:           path,
			HeadingPath:    cwv.chunk.HeadingPath,
			ChunkIndex:     cwv.chunk.ChunkIndex,
			Content:        cwv.chunk.Content,
			ContentSHA256:  chunker.ComputeContentHash(cwv.chunk.Content),
			StartLine:      cwv.chunk.StartLine,
			EndLine:        cwv.chunk.EndLine,
			Category:       cwv.chunk.Category,
			Status:         cwv.chunk.Status,
			Scope:          cwv.chunk.Scope,
			NoteType:       cwv.chunk.NoteType,
			CategoryWeight: cwv.chunk.CategoryWeight,
		}

		chunkID, err := idx.store.InsertChunk(ctx, tx, storeChunk)
		if err != nil {
			return fmt.Errorf("insert chunk %d: %w", cwv.index, err)
		}

		embedding := &store.Embedding{
			ChunkID: chunkID,
			Dim:     len(cwv.vector),
			Vec:     cwv.vector,
		}

		if err := idx.store.InsertEmbedding(ctx, tx, embedding); err != nil {
			return fmt.Errorf("insert embedding %d: %w", cwv.index, err)
		}

		// Add to ANN index
		if err := idx.annIndex.Add(uint64(chunkID), cwv.vector); err != nil {
			return fmt.Errorf("add to ann index: %w", err)
		}
	}

	// Update file info (within the same transaction)
	fileInfo := &store.FileInfo{
		Path:          path,
		SHA256:        fileHash,
		MtimeUnix:     mtime,
		IndexedAtUnix: time.Now().Unix(),
	}
	if err := idx.store.UpsertFileInfoTx(ctx, tx, fileInfo); err != nil {
		return fmt.Errorf("upsert file info: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

// IndexVault processes all markdown files in the vault
func (idx *Indexer) IndexVault(ctx context.Context) error {
	return filepath.Walk(idx.vaultDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// Skip hidden directories
			if info.Name() != "." && info.Name()[0] == '.' {
				return filepath.SkipDir
			}
			return nil
		}

		// Only index markdown files
		if filepath.Ext(path) != ".md" {
			return nil
		}

		fmt.Printf("Indexing: %s\n", path)
		if err := idx.IndexFile(ctx, path); err != nil {
			fmt.Printf("Error indexing %s: %v\n", path, err)
			// Continue with other files
		}

		return nil
	})
}

// computeFileHash returns SHA256 hash and mtime of a file
func computeFileHash(path string) (string, int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return "", 0, err
	}

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", 0, err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), stat.ModTime().Unix(), nil
}
