package ann

import (
	"fmt"
	"math"
	"os"
	"sync"

	"github.com/coder/hnsw"
)

// HNSWConfig contains HNSW index parameters
type HNSWConfig struct {
	Dim            int
	M              int // default: 16 - number of connections per layer
	EfConstruction int // default: 200 - size of dynamic candidate list during construction
	EfSearch       int // default: 100 - size of dynamic candidate list during search
}

// DefaultHNSWConfig returns sensible defaults for production use
func DefaultHNSWConfig(dim int) HNSWConfig {
	return HNSWConfig{
		Dim:            dim,
		M:              16,  // Good balance between speed and recall
		EfConstruction: 200, // Higher = better index quality, slower indexing
		EfSearch:       100, // Higher = better recall, slower search
	}
}

// HNSWIndex wraps the coder/hnsw library with our interface
type HNSWIndex struct {
	cfg   HNSWConfig
	graph *hnsw.Graph[uint64]
	mu    sync.RWMutex
	count int
}

// NewHNSW creates a new HNSW index
func NewHNSW(cfg HNSWConfig) (*HNSWIndex, error) {
	if cfg.M == 0 {
		cfg = DefaultHNSWConfig(cfg.Dim)
	}

	graph := hnsw.NewGraph[uint64]()
	graph.M = cfg.M
	graph.EfSearch = cfg.EfSearch
	graph.Distance = cosineDistance // Use cosine distance for text embeddings

	return &HNSWIndex{
		cfg:   cfg,
		graph: graph,
		count: 0,
	}, nil
}

// Add inserts a vector into the HNSW index
func (h *HNSWIndex) Add(id uint64, vec []float32) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(vec) != h.cfg.Dim {
		return fmt.Errorf("vector dimension mismatch: got %d, expected %d", len(vec), h.cfg.Dim)
	}

	// Create node and add to graph
	node := hnsw.MakeNode(id, vec)
	h.graph.Add(node)
	h.count++

	return nil
}

// Search returns k nearest neighbors using HNSW
func (h *HNSWIndex) Search(vec []float32, k int) ([]uint64, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if len(vec) != h.cfg.Dim {
		return nil, fmt.Errorf("vector dimension mismatch: got %d, expected %d", len(vec), h.cfg.Dim)
	}

	if h.count == 0 {
		return nil, nil
	}

	// Search HNSW - returns list of nodes (EfSearch is set on graph config)
	results := h.graph.Search(vec, k)

	// Extract IDs from results
	ids := make([]uint64, len(results))
	for i, node := range results {
		ids[i] = node.Key
	}

	return ids, nil
}

// Save persists the index to disk
func (h *HNSWIndex) Save(path string) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return h.graph.Export(f)
}

// Load restores the index from disk
func (h *HNSWIndex) Load(path string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	h.graph = hnsw.NewGraph[uint64]()
	h.graph.M = h.cfg.M
	h.graph.Distance = cosineDistance

	if err := h.graph.Import(f); err != nil {
		return err
	}

	h.count = h.graph.Len()
	return nil
}

// Close releases resources
func (h *HNSWIndex) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.graph = nil
	h.count = 0
	return nil
}

// Size returns the number of vectors in the index
func (h *HNSWIndex) Size() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.count
}

// cosineDistance computes 1 - cosine_similarity
// Lower distance = more similar
func cosineDistance(a, b []float32) float32 {
	var dot, normA, normB float32
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 1.0 // Maximum distance
	}
	similarity := dot / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB))))
	return 1.0 - similarity
}
