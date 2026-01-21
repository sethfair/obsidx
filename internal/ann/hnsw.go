package ann

import (
	"fmt"
	"math"
	"sort"
	"sync"
)

// HNSWConfig contains HNSW index parameters
type HNSWConfig struct {
	Dim            int
	M              int    // default: 16
	EfConstruction int    // default: 128
	EfSearch       int    // default: 64
	DistanceMetric string // "cosine" or "dot"
}

// DefaultHNSWConfig returns sensible defaults
func DefaultHNSWConfig(dim int) HNSWConfig {
	return HNSWConfig{
		Dim:            dim,
		M:              16,
		EfConstruction: 128,
		EfSearch:       64,
		DistanceMetric: "cosine",
	}
}

// HNSWIndex is a simple in-memory vector index
// Uses brute-force search which is efficient for <100k vectors
// Can be replaced with true HNSW later if needed
type HNSWIndex struct {
	cfg     HNSWConfig
	vectors map[uint64][]float32
	mu      sync.RWMutex
}

// NewHNSW creates a new HNSW index
func NewHNSW(cfg HNSWConfig) (*HNSWIndex, error) {
	if cfg.M == 0 {
		cfg = DefaultHNSWConfig(cfg.Dim)
	}

	return &HNSWIndex{
		cfg:     cfg,
		vectors: make(map[uint64][]float32),
	}, nil
}

// Add inserts a vector
func (h *HNSWIndex) Add(id uint64, vec []float32) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(vec) != h.cfg.Dim {
		return fmt.Errorf("vector dimension mismatch: got %d, expected %d", len(vec), h.cfg.Dim)
	}

	// Make a copy to avoid external mutation
	vecCopy := make([]float32, len(vec))
	copy(vecCopy, vec)
	h.vectors[id] = vecCopy

	return nil
}

// Search returns nearest neighbors using brute-force search
func (h *HNSWIndex) Search(vec []float32, k int) ([]uint64, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if len(vec) != h.cfg.Dim {
		return nil, fmt.Errorf("vector dimension mismatch: got %d, expected %d", len(vec), h.cfg.Dim)
	}

	if len(h.vectors) == 0 {
		return nil, nil
	}

	// Compute distances to all vectors
	type result struct {
		id   uint64
		dist float32
	}

	results := make([]result, 0, len(h.vectors))
	for id, storedVec := range h.vectors {
		var dist float32
		switch h.cfg.DistanceMetric {
		case "cosine":
			// Cosine distance = 1 - cosine similarity
			dist = 1.0 - cosineSimilarity(vec, storedVec)
		case "dot":
			// For dot product, negate to make it a "distance" (lower is better)
			dist = -dotProduct(vec, storedVec)
		default:
			// L2 distance
			dist = l2Distance(vec, storedVec)
		}
		results = append(results, result{id: id, dist: dist})
	}

	// Sort by distance (ascending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].dist < results[j].dist
	})

	// Return top k IDs
	if k > len(results) {
		k = len(results)
	}

	ids := make([]uint64, k)
	for i := 0; i < k; i++ {
		ids[i] = results[i].id
	}

	return ids, nil
}

// Save persists the index (no-op for in-memory version)
func (h *HNSWIndex) Save(_ string) error {
	return nil
}

// Load restores the index (no-op - we rebuild from SQLite)
func (h *HNSWIndex) Load(_ string) error {
	return fmt.Errorf("load not implemented: use rebuild instead")
}

// Close releases resources
func (h *HNSWIndex) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.vectors = nil
	return nil
}

// Size returns the number of vectors
func (h *HNSWIndex) Size() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.vectors)
}

// Distance calculation helpers

func cosineSimilarity(a, b []float32) float32 {
	var dot, normA, normB float32
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB))))
}

func dotProduct(a, b []float32) float32 {
	var sum float32
	for i := range a {
		sum += a[i] * b[i]
	}
	return sum
}

func l2Distance(a, b []float32) float32 {
	var sum float32
	for i := range a {
		diff := a[i] - b[i]
		sum += diff * diff
	}
	return float32(math.Sqrt(float64(sum)))
}
