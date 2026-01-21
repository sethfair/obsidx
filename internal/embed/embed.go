package embed

import (
	"context"
)

// Embedder converts text into vector embeddings using Ollama
type Embedder interface {
	// Embed converts text into a vector
	Embed(ctx context.Context, text string) ([]float32, error)

	// Dimension returns the embedding dimension
	Dimension() int

	// ModelName returns the model identifier
	ModelName() string

	// Ping checks if the embedding service is available
	Ping(ctx context.Context) error
}
