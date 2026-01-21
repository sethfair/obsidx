package embed

import (
	"context"
)

// Embedder is the interface for embedding text into vectors
type Embedder interface {
	// Embed converts text into a vector
	Embed(ctx context.Context, text string) ([]float32, error)

	// Dimension returns the embedding dimension
	Dimension() int

	// ModelName returns the model identifier
	ModelName() string
}

// BatchEmbedder can embed multiple texts in one call (optional optimization)
type BatchEmbedder interface {
	Embedder

	// EmbedBatch embeds multiple texts efficiently
	EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)
}
