package ann

// Index is the nearest-neighbor search interface. It abstracts the search
// implementation (currently BruteForce exact scan — a coder/hnsw graph was
// used before 2026-07-23 but showed near-zero recall on real vault
// embeddings; see BruteForce) so it can be swapped if needed. The index is
// in-memory only, rebuilt from SQLite at startup.
type Index interface {
	// Add inserts a vector with given ID
	Add(id uint64, vec []float32) error

	// Search returns the k nearest neighbor IDs for the query vector
	Search(vec []float32, k int) ([]uint64, error)

	// Close releases resources
	Close() error

	// Size returns the number of vectors in the index
	Size() int
}
