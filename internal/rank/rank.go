package rank

import (
	"container/heap"
	"math"

	"github.com/seth/obsidx/internal/store"
)

// Result represents a ranked search result
type Result struct {
	Chunk store.ChunkWithEmbedding
	Score float32 // higher is better
}

// CosineSimilarity computes cosine similarity between two vectors
// Returns value in [0, 1] where 1 is most similar
func CosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float32
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB))))
}

// RerankCosine reranks chunks by exact cosine similarity and returns top N
// Applies category weights for tiered retrieval
func RerankCosine(queryVec []float32, chunks []store.ChunkWithEmbedding, topN int) []Result {
	if len(chunks) == 0 {
		return nil
	}

	// Compute all scores with category weighting
	scores := make([]Result, len(chunks))
	for i, chunk := range chunks {
		baseSimilarity := CosineSimilarity(queryVec, chunk.Vec)

		// Apply category weight (canon gets 1.20x, workbench gets 0.90x, etc.)
		weightedScore := baseSimilarity * chunk.CategoryWeight

		scores[i] = Result{
			Chunk: chunk,
			Score: weightedScore,
		}
	}

	// Use max-heap to find top N
	if topN > len(scores) {
		topN = len(scores)
	}

	h := &resultHeap{}
	heap.Init(h)

	for _, r := range scores {
		if h.Len() < topN {
			heap.Push(h, r)
		} else if r.Score > (*h)[0].Score {
			heap.Pop(h)
			heap.Push(h, r)
		}
	}

	// Extract results in descending order
	results := make([]Result, h.Len())
	for i := len(results) - 1; i >= 0; i-- {
		results[i] = heap.Pop(h).(Result)
	}

	return results
}

// resultHeap is a min-heap of Results by score
type resultHeap []Result

func (h resultHeap) Len() int           { return len(h) }
func (h resultHeap) Less(i, j int) bool { return h[i].Score < h[j].Score }
func (h resultHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *resultHeap) Push(x interface{}) {
	*h = append(*h, x.(Result))
}

func (h *resultHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
