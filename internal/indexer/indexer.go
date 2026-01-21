package indexer

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/seth/obsidx/internal/ann"
	"github.com/seth/obsidx/internal/chunker"
	"github.com/seth/obsidx/internal/embed"
	"github.com/seth/obsidx/internal/store"
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

	chunks := chunker.ChunkMarkdown(string(content))
	if len(chunks) == 0 {
		return nil // Empty file
	}

	// Embed all chunks
	vectors := make([][]float32, len(chunks))
	for i, chunk := range chunks {
		vec, err := idx.embedder.Embed(ctx, chunk.Content)
		if err != nil {
			return fmt.Errorf("embed chunk %d: %w", i, err)
		}
		vectors[i] = vec
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
	for i, chunk := range chunks {
		storeChunk := &store.Chunk{
			Path:           path,
			HeadingPath:    chunk.HeadingPath,
			ChunkIndex:     chunk.ChunkIndex,
			Content:        chunk.Content,
			ContentSHA256:  chunker.ComputeContentHash(chunk.Content),
			StartLine:      chunk.StartLine,
			EndLine:        chunk.EndLine,
			Category:       chunk.Category,
			Status:         chunk.Status,
			Scope:          chunk.Scope,
			NoteType:       chunk.NoteType,
			CategoryWeight: chunk.CategoryWeight,
		}

		chunkID, err := idx.store.InsertChunk(ctx, tx, storeChunk)
		if err != nil {
			return fmt.Errorf("insert chunk %d: %w", i, err)
		}

		embedding := &store.Embedding{
			ChunkID: chunkID,
			Dim:     len(vectors[i]),
			Vec:     vectors[i],
		}

		if err := idx.store.InsertEmbedding(ctx, tx, embedding); err != nil {
			return fmt.Errorf("insert embedding %d: %w", i, err)
		}

		// Add to ANN index
		if err := idx.annIndex.Add(uint64(chunkID), vectors[i]); err != nil {
			return fmt.Errorf("add to ann index: %w", err)
		}
	}

	// Update file info
	fileInfo := &store.FileInfo{
		Path:          path,
		SHA256:        fileHash,
		MtimeUnix:     mtime,
		IndexedAtUnix: time.Now().Unix(),
	}
	if err := idx.store.UpsertFileInfo(ctx, fileInfo); err != nil {
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
