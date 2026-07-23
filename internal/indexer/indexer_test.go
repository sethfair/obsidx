package indexer

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sethfair/obsidx/internal/ann"
	"github.com/sethfair/obsidx/internal/chunker"
	"github.com/sethfair/obsidx/internal/store"

	_ "github.com/mattn/go-sqlite3"
)

// fakeEmbedder records every text it is asked to embed and returns a
// deterministic distinct vector per call.
type fakeEmbedder struct {
	calls []string
}

func (f *fakeEmbedder) Embed(_ context.Context, text string) ([]float32, error) {
	f.calls = append(f.calls, text)
	vec := make([]float32, 8)
	vec[0] = float32(len(f.calls)) // distinct per call
	vec[1] = 1
	return vec, nil
}
func (f *fakeEmbedder) Dimension() int               { return 8 }
func (f *fakeEmbedder) ModelName() string            { return "fake" }
func (f *fakeEmbedder) Ping(_ context.Context) error { return nil }

func newTestIndexer(t *testing.T) (*Indexer, *fakeEmbedder, string, string) {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	st, err := store.Open(dbPath, 8)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	emb := &fakeEmbedder{}
	return New(st, emb, ann.NewBruteForce(8), dir), emb, dir, dbPath
}

func writeNote(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	return path
}

// activeChunkContents reads the content of every active chunk for path
// straight from the database (the store has no read-back-by-path method).
func activeChunkContents(t *testing.T, dbPath, notePath string) []string {
	t.Helper()
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	rows, err := db.Query("SELECT content FROM chunks WHERE path = ? AND active = 1", notePath)
	if err != nil {
		t.Fatalf("query chunks: %v", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			t.Fatalf("scan: %v", err)
		}
		out = append(out, c)
	}
	return out
}

func TestIndexFileSkipsHeadingOnlyChunks(t *testing.T) {
	idx, emb, dir, dbPath := newTestIndexer(t)
	ctx := context.Background()

	// "## Empty Section" is followed directly by another heading, so the
	// chunker emits it as a heading-only chunk; the second chunk has body.
	path := writeNote(t, dir, "note.md", "## Empty Section\n\n## Real Section\n\nThis body paragraph is long enough to index.\n")

	if err := idx.IndexFile(ctx, path); err != nil {
		t.Fatalf("IndexFile: %v", err)
	}

	if len(emb.calls) == 0 {
		t.Fatal("embedder was never called; expected the body chunk to be embedded")
	}
	for _, call := range emb.calls {
		if chunker.IsHeadingOnly(strings.TrimSpace(call)) {
			t.Errorf("embedder was called with heading-only content %q", call)
		}
	}
	for _, c := range activeChunkContents(t, dbPath, path) {
		if chunker.IsHeadingOnly(strings.TrimSpace(c)) {
			t.Errorf("heading-only chunk stored as active: %q", c)
		}
	}
}

// Regression test: a file whose new revision produces zero indexable chunks
// must still deactivate its old chunks and record the new file hash.
// Previously IndexFile returned before MarkChunksInactive in that path, so
// search kept serving deleted content forever.
func TestIndexFileDeactivatesStaleChunksWhenAllChunksSkipped(t *testing.T) {
	idx, _, dir, dbPath := newTestIndexer(t)
	ctx := context.Background()

	path := writeNote(t, dir, "note.md", "## Section\n\nOriginal body content that is definitely long enough.\n")
	if err := idx.IndexFile(ctx, path); err != nil {
		t.Fatalf("initial IndexFile: %v", err)
	}
	if got := activeChunkContents(t, dbPath, path); len(got) == 0 {
		t.Fatal("setup: expected active chunks after first index")
	}

	// Rewrite the file to headings only — every chunk should now be skipped.
	writeNote(t, dir, "note.md", "## Section\n\n## Another Section\n")
	if err := idx.IndexFile(ctx, path); err != nil {
		t.Fatalf("re-IndexFile: %v", err)
	}

	if got := activeChunkContents(t, dbPath, path); len(got) != 0 {
		t.Errorf("stale chunks still active after headings-only rewrite: %v", got)
	}

	// FileInfo must be updated so the file is not re-read on every pass.
	fi, err := idx.store.GetFileInfo(ctx, path)
	if err != nil {
		t.Fatalf("get file info: %v", err)
	}
	if fi == nil {
		t.Fatal("file info missing after headings-only re-index")
	}
	if err := idx.IndexFile(ctx, path); err != nil {
		t.Fatalf("third IndexFile (hash-skip pass): %v", err)
	}
}

func TestIndexFileEmptiedFileDeactivatesChunks(t *testing.T) {
	idx, _, dir, dbPath := newTestIndexer(t)
	ctx := context.Background()

	path := writeNote(t, dir, "note.md", "## Section\n\nBody content long enough to be embedded and stored.\n")
	if err := idx.IndexFile(ctx, path); err != nil {
		t.Fatalf("initial IndexFile: %v", err)
	}

	writeNote(t, dir, "note.md", "")
	if err := idx.IndexFile(ctx, path); err != nil {
		t.Fatalf("re-IndexFile empty: %v", err)
	}

	if got := activeChunkContents(t, dbPath, path); len(got) != 0 {
		t.Errorf("stale chunks still active after file emptied: %v", got)
	}
}
