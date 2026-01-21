package embed

import (
	"context"
	"math"
	"strings"
)

// LocalEmbedder provides simple TF-IDF based embeddings as a fallback
// This is a basic implementation that doesn't require external services
// For production use, consider using proper embeddings (Ollama, OpenAI, etc.)
type LocalEmbedder struct {
	dimension int
	modelName string
	vocab     map[string]int
	idf       map[string]float32
}

// NewLocalEmbedder creates a simple local embedder
func NewLocalEmbedder(dimension int) *LocalEmbedder {
	return &LocalEmbedder{
		dimension: dimension,
		modelName: "local-tfidf",
		vocab:     make(map[string]int),
		idf:       make(map[string]float32),
	}
}

// Embed converts text to a simple vector using TF-IDF approach
// This is NOT as good as proper neural embeddings but works without external deps
func (l *LocalEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	// Tokenize
	tokens := tokenize(text)
	if len(tokens) == 0 {
		return make([]float32, l.dimension), nil
	}

	// Count term frequencies
	tf := make(map[string]int)
	for _, token := range tokens {
		tf[token]++
	}

	// Build a simple hash-based vector
	vec := make([]float32, l.dimension)
	for token, count := range tf {
		// Use simple hash to map token to dimensions
		h := hash(token)
		for i := 0; i < 3; i++ { // Spread each token across 3 dimensions
			idx := (h + uint32(i)) % uint32(l.dimension)
			vec[idx] += float32(count)
		}
	}

	// Normalize
	var norm float32
	for _, v := range vec {
		norm += v * v
	}
	norm = float32(math.Sqrt(float64(norm)))
	if norm > 0 {
		for i := range vec {
			vec[i] /= norm
		}
	}

	return vec, nil
}

// Dimension returns the embedding dimension
func (l *LocalEmbedder) Dimension() int {
	return l.dimension
}

// ModelName returns the model identifier
func (l *LocalEmbedder) ModelName() string {
	return l.modelName
}

// Simple tokenization
func tokenize(text string) []string {
	text = strings.ToLower(text)
	// Split on non-alphanumeric
	var tokens []string
	var current strings.Builder

	for _, r := range text {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			current.WriteRune(r)
		} else if current.Len() > 0 {
			if current.Len() > 2 { // Skip very short tokens
				tokens = append(tokens, current.String())
			}
			current.Reset()
		}
	}
	if current.Len() > 2 {
		tokens = append(tokens, current.String())
	}

	return tokens
}

// Simple hash function (FNV-1a)
func hash(s string) uint32 {
	h := uint32(2166136261)
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= 16777619
	}
	return h
}
