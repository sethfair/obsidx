package ann

import (
	"container/heap"
	"fmt"
	"math"
	"runtime"
	"sort"
	"sync"
)

// BruteForce is an exact nearest-neighbor index: flat storage, full-scan
// cosine search parallelized across CPU cores. At vault scale (~100k
// vectors × 768 dims) a full scan completes in tens of milliseconds, and
// unlike graph-based ANN it cannot lose recall to degenerate data — the
// coder/hnsw graph shipped previously returned near-zero recall on this
// vault's real embeddings even after duplicate-vector clusters were purged
// (2026-07-23), while an exact scan is immune by construction.
type BruteForce struct {
	dim  int
	mu   sync.RWMutex
	ids  []uint64
	vecs [][]float32 // stored L2-normalized so search is a pure dot product
}

// NewBruteForce creates an exact-search index for vectors of the given dimension.
func NewBruteForce(dim int) *BruteForce {
	return &BruteForce{dim: dim}
}

// Add inserts a vector with the given ID. The vector is copied and normalized.
// Zero-norm vectors are rejected: they are unembeddable junk (e.g. the
// embedding of an empty string) and would rank at sim 0 above genuinely
// anti-correlated results.
func (b *BruteForce) Add(id uint64, vec []float32) error {
	if len(vec) != b.dim {
		return fmt.Errorf("vector dimension mismatch: got %d, expected %d", len(vec), b.dim)
	}
	norm := float32(0)
	for _, x := range vec {
		norm += x * x
	}
	norm = float32(math.Sqrt(float64(norm)))
	if norm == 0 {
		return fmt.Errorf("zero-norm vector for id %d", id)
	}
	stored := make([]float32, len(vec))
	for i, x := range vec {
		stored[i] = x / norm
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	b.ids = append(b.ids, id)
	b.vecs = append(b.vecs, stored)
	return nil
}

type scored struct {
	id  uint64
	sim float32
}

// worse reports whether a ranks strictly below b under the total order
// (sim descending, id ascending). A total order makes tied-score results
// deterministic regardless of worker count or stripe boundaries.
func worse(a, b scored) bool {
	if a.sim != b.sim {
		return a.sim < b.sim
	}
	return a.id > b.id
}

// minHeap keeps the k best candidates seen so far under the total order;
// the root is the worst of them, so it can be evicted cheaply.
type minHeap []scored

func (h minHeap) Len() int            { return len(h) }
func (h minHeap) Less(i, j int) bool  { return worse(h[i], h[j]) }
func (h minHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *minHeap) Push(x interface{}) { *h = append(*h, x.(scored)) }
func (h *minHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[:n-1]
	return item
}

// Search returns the k exact nearest neighbors by cosine similarity.
func (b *BruteForce) Search(vec []float32, k int) ([]uint64, error) {
	if len(vec) != b.dim {
		return nil, fmt.Errorf("vector dimension mismatch: got %d, expected %d", len(vec), b.dim)
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	n := len(b.ids)
	if n == 0 || k <= 0 {
		return nil, nil
	}
	if k > n {
		k = n
	}

	// Normalize the query once; stored vectors are pre-normalized.
	// A zero-norm query means the embedding failed — surfacing an error is
	// better than confidently returning k arbitrary IDs at sim 0.
	norm := float32(0)
	for _, x := range vec {
		norm += x * x
	}
	norm = float32(math.Sqrt(float64(norm)))
	if norm == 0 {
		return nil, fmt.Errorf("zero-norm query vector")
	}
	q := make([]float32, len(vec))
	for i, x := range vec {
		q[i] = x / norm
	}

	// Parallel scan: each worker keeps its own top-k heap over a stripe.
	workers := runtime.GOMAXPROCS(0)
	if workers > n {
		workers = 1
	}
	heaps := make([]minHeap, workers)
	var wg sync.WaitGroup
	stripe := (n + workers - 1) / workers
	for w := 0; w < workers; w++ {
		start := w * stripe
		end := start + stripe
		if end > n {
			end = n
		}
		if start >= end {
			continue
		}
		wg.Add(1)
		go func(w, start, end int) {
			defer wg.Done()
			h := make(minHeap, 0, k)
			for i := start; i < end; i++ {
				v := b.vecs[i]
				var dot float32
				for j := range q {
					dot += q[j] * v[j]
				}
				cand := scored{b.ids[i], dot}
				if len(h) < k {
					heap.Push(&h, cand)
				} else if worse(h[0], cand) {
					h[0] = cand
					heap.Fix(&h, 0)
				}
			}
			heaps[w] = h
		}(w, start, end)
	}
	wg.Wait()

	// Merge worker heaps and take the global top-k.
	var all []scored
	for _, h := range heaps {
		all = append(all, h...)
	}
	sort.Slice(all, func(i, j int) bool { return worse(all[j], all[i]) })
	if len(all) > k {
		all = all[:k]
	}
	ids := make([]uint64, len(all))
	for i, s := range all {
		ids[i] = s.id
	}
	return ids, nil
}

// Close releases resources (none held).
func (b *BruteForce) Close() error { return nil }

// Size returns the number of stored vectors.
func (b *BruteForce) Size() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.ids)
}
