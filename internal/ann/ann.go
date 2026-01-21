package ann

// Index is the ANN (approximate nearest neighbor) interface
// This abstracts the HNSW implementation so we can swap it if needed
type Index interface {
	// Add inserts a vector with given ID
	Add(id uint64, vec []float32) error

	// Search returns the k nearest neighbor IDs for the query vector
	Search(vec []float32, k int) ([]uint64, error)

	// Save persists the index to disk
	Save(dir string) error

	// Load restores the index from disk
	Load(dir string) error

	// Close releases resources
	Close() error

	// Size returns the number of vectors in the index
	Size() int
}
