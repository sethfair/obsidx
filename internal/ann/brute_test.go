package ann

import (
	"math"
	"math/rand"
	"sort"
	"sync"
	"testing"
)

func randVec(rng *rand.Rand, dim int) []float32 {
	v := make([]float32, dim)
	for i := range v {
		v[i] = rng.Float32()*2 - 1
	}
	return v
}

// TestBruteForceFindsPlantedNeighbor verifies exact recall: a vector nearly
// identical to the query must always be the first result, regardless of how
// many other vectors are in the index.
func TestBruteForceFindsPlantedNeighbor(t *testing.T) {
	const dim = 64
	rng := rand.New(rand.NewSource(42))

	idx := NewBruteForce(dim)

	query := randVec(rng, dim)
	planted := make([]float32, dim)
	copy(planted, query)
	planted[0] += 0.001 // near-identical, not equal

	const plantedID = 999_999
	if err := idx.Add(plantedID, planted); err != nil {
		t.Fatalf("add planted: %v", err)
	}
	for i := 0; i < 5000; i++ {
		if err := idx.Add(uint64(i), randVec(rng, dim)); err != nil {
			t.Fatalf("add %d: %v", i, err)
		}
	}

	ids, err := idx.Search(query, 10)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(ids) != 10 {
		t.Fatalf("expected 10 results, got %d", len(ids))
	}
	if ids[0] != plantedID {
		t.Errorf("planted near-duplicate not ranked first; got id %d", ids[0])
	}
}

// TestBruteForceSurvivesDuplicateClusters reproduces the HNSW failure mode
// that motivated this implementation: thousands of identical vectors must not
// prevent the true nearest neighbor from being returned.
func TestBruteForceSurvivesDuplicateClusters(t *testing.T) {
	const dim = 64
	rng := rand.New(rand.NewSource(7))

	idx := NewBruteForce(dim)

	// One duplicate vector inserted 2000 times.
	dup := randVec(rng, dim)
	for i := 0; i < 2000; i++ {
		if err := idx.Add(uint64(i), dup); err != nil {
			t.Fatalf("add dup %d: %v", i, err)
		}
	}
	// Background noise.
	for i := 2000; i < 4000; i++ {
		if err := idx.Add(uint64(i), randVec(rng, dim)); err != nil {
			t.Fatalf("add noise %d: %v", i, err)
		}
	}

	// Target orthogonal-ish to the dup cluster.
	query := randVec(rng, dim)
	target := make([]float32, dim)
	copy(target, query)
	const targetID = 555_555
	if err := idx.Add(targetID, target); err != nil {
		t.Fatalf("add target: %v", err)
	}

	ids, err := idx.Search(query, 5)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(ids) == 0 || ids[0] != targetID {
		t.Errorf("true nearest neighbor lost among duplicate cluster; got %v", ids)
	}
}

func TestBruteForceSmallerKThanIndex(t *testing.T) {
	idx := NewBruteForce(4)
	if err := idx.Add(1, []float32{1, 0, 0, 0}); err != nil {
		t.Fatal(err)
	}
	if err := idx.Add(2, []float32{0, 1, 0, 0}); err != nil {
		t.Fatal(err)
	}
	ids, err := idx.Search([]float32{1, 0, 0, 0}, 10)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("expected 2 results when k > index size, got %d", len(ids))
	}
	if ids[0] != 1 {
		t.Errorf("expected id 1 first, got %d", ids[0])
	}
	if idx.Size() != 2 {
		t.Errorf("Size() = %d, want 2", idx.Size())
	}
}

func TestBruteForceDimensionMismatch(t *testing.T) {
	idx := NewBruteForce(4)
	if err := idx.Add(1, []float32{1, 0}); err == nil {
		t.Error("expected error adding wrong-dimension vector")
	}
	if _, err := idx.Search([]float32{1, 0}, 1); err == nil {
		t.Error("expected error searching wrong-dimension vector")
	}
}

// Zero-norm vectors are unembeddable junk (e.g. the embedding of an empty
// string): storing them would rank sim-0 junk above genuinely anti-correlated
// chunks, and a zero-norm query would return arbitrary IDs with confident
// scores. Both must be rejected loudly.
func TestBruteForceRejectsZeroNormAdd(t *testing.T) {
	idx := NewBruteForce(4)
	if err := idx.Add(1, []float32{0, 0, 0, 0}); err == nil {
		t.Error("expected error adding zero-norm vector")
	}
	if idx.Size() != 0 {
		t.Errorf("zero-norm vector was stored; Size() = %d, want 0", idx.Size())
	}
}

func TestBruteForceRejectsZeroNormQuery(t *testing.T) {
	idx := NewBruteForce(4)
	if err := idx.Add(1, []float32{1, 0, 0, 0}); err != nil {
		t.Fatal(err)
	}
	if _, err := idx.Search([]float32{0, 0, 0, 0}, 1); err == nil {
		t.Error("expected error searching with zero-norm query")
	}
}

// naiveTopK is an independent sequential implementation used as ground truth.
func naiveTopK(ids []uint64, vecs [][]float32, q []float32, k int) []uint64 {
	type sc struct {
		id  uint64
		sim float32
	}
	var all []sc
	var qn float32
	for _, x := range q {
		qn += x * x
	}
	qn = float32(math.Sqrt(float64(qn)))
	for i := range ids {
		var vn, dot float32
		for j := range vecs[i] {
			vn += vecs[i][j] * vecs[i][j]
			dot += vecs[i][j] * q[j]
		}
		vn = float32(math.Sqrt(float64(vn)))
		all = append(all, sc{ids[i], dot / (qn * vn)})
	}
	sort.Slice(all, func(a, b int) bool {
		if all[a].sim != all[b].sim {
			return all[a].sim > all[b].sim
		}
		return all[a].id < all[b].id
	})
	if len(all) > k {
		all = all[:k]
	}
	out := make([]uint64, len(all))
	for i, s := range all {
		out[i] = s.id
	}
	return out
}

// TestBruteForceMatchesSequentialScan pins the parallel striped merge against
// an independent sequential implementation, including deterministic id-order
// tie-breaking (duplicate vectors produce exact score ties across stripes).
func TestBruteForceMatchesSequentialScan(t *testing.T) {
	const dim = 32
	const n = 5000
	rng := rand.New(rand.NewSource(3))

	idx := NewBruteForce(dim)
	var ids []uint64
	var vecs [][]float32
	dup := randVec(rng, dim) // planted exact ties spanning stripe boundaries
	for i := 0; i < n; i++ {
		var v []float32
		if i%17 == 0 {
			v = append([]float32(nil), dup...)
		} else {
			v = randVec(rng, dim)
		}
		id := uint64(i * 3) // non-contiguous ids
		if err := idx.Add(id, v); err != nil {
			t.Fatalf("add: %v", err)
		}
		ids = append(ids, id)
		vecs = append(vecs, v)
	}

	for trial := 0; trial < 5; trial++ {
		q := randVec(rng, dim)
		got, err := idx.Search(q, 25)
		if err != nil {
			t.Fatalf("search: %v", err)
		}
		want := naiveTopK(ids, vecs, q, 25)
		if len(got) != len(want) {
			t.Fatalf("trial %d: got %d results, want %d", trial, len(got), len(want))
		}
		for i := range got {
			if got[i] != want[i] {
				t.Errorf("trial %d rank %d: got id %d, want id %d", trial, i, got[i], want[i])
			}
		}
	}
}

// TestBruteForceConcurrentAddSearch exercises Add and Search under real
// contention so `go test -race` observes the locking, converting the
// by-inspection safety argument into evidence.
func TestBruteForceConcurrentAddSearch(t *testing.T) {
	const dim = 16
	rng := rand.New(rand.NewSource(11))
	idx := NewBruteForce(dim)
	for i := 0; i < 100; i++ {
		if err := idx.Add(uint64(i), randVec(rng, dim)); err != nil {
			t.Fatal(err)
		}
	}

	var wg sync.WaitGroup
	stop := make(chan struct{})
	for w := 0; w < 4; w++ {
		wg.Add(1)
		go func(seed int64) {
			defer wg.Done()
			r := rand.New(rand.NewSource(seed))
			for {
				select {
				case <-stop:
					return
				default:
				}
				if _, err := idx.Search(randVec(r, dim), 10); err != nil {
					t.Errorf("concurrent search: %v", err)
					return
				}
			}
		}(int64(100 + w))
	}
	for i := 100; i < 600; i++ {
		if err := idx.Add(uint64(i), randVec(rng, dim)); err != nil {
			t.Errorf("concurrent add: %v", err)
			break
		}
	}
	close(stop)
	wg.Wait()

	if idx.Size() != 600 {
		t.Errorf("Size() = %d, want 600", idx.Size())
	}
}
