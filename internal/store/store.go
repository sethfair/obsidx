package store

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schemaSQL string

// SQLite is the authoritative store for chunks and embeddings
type SQLite struct {
	db  *sql.DB
	dim int
}

// FileInfo tracks processed files
type FileInfo struct {
	Path          string
	SHA256        string
	MtimeUnix     int64
	IndexedAtUnix int64
}

// Chunk represents a text chunk with metadata
type Chunk struct {
	ID            int64
	Path          string
	HeadingPath   string
	ChunkIndex    int
	Content       string
	ContentSHA256 string
	StartLine     int
	EndLine       int
	Active        bool
	CreatedAtUnix int64
	// Metadata fields
	Status         string
	Scope          string
	NoteType       string
	CategoryWeight float32
	Tags           []string
}

// Embedding represents a vector embedding
type Embedding struct {
	ChunkID int64
	Dim     int
	Vec     []float32
}

// ChunkWithEmbedding combines chunk metadata with its vector
type ChunkWithEmbedding struct {
	Chunk
	Vec []float32
}

// Open creates or opens a SQLite database
func Open(path string, dimension int) (*SQLite, error) {
	// Enhanced connection string for better concurrency handling
	connStr := path + "?_journal_mode=WAL&_synchronous=NORMAL&_cache_size=-64000&_busy_timeout=5000&_txlock=immediate"

	db, err := sql.Open("sqlite3", connStr)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	// Set connection pool settings for better concurrency
	db.SetMaxOpenConns(1) // SQLite works best with single writer
	db.SetMaxIdleConns(1)

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	// Initialize schema
	if _, err := db.Exec(schemaSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("init schema: %w", err)
	}

	return &SQLite{db: db, dim: dimension}, nil
}

// Close closes the database
func (s *SQLite) Close() error {
	return s.db.Close()
}

// Dim returns the embedding dimension
func (s *SQLite) Dim() int {
	return s.dim
}

// GetFileInfo retrieves file tracking info
func (s *SQLite) GetFileInfo(ctx context.Context, path string) (*FileInfo, error) {
	var fi FileInfo
	err := s.db.QueryRowContext(ctx,
		"SELECT path, sha256, mtime_unix, indexed_at_unix FROM files WHERE path = ?",
		path,
	).Scan(&fi.Path, &fi.SHA256, &fi.MtimeUnix, &fi.IndexedAtUnix)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &fi, nil
}

// UpsertFileInfo updates or inserts file tracking info
func (s *SQLite) UpsertFileInfo(ctx context.Context, fi *FileInfo) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO files (path, sha256, mtime_unix, indexed_at_unix)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT(path) DO UPDATE SET
		   sha256 = excluded.sha256,
		   mtime_unix = excluded.mtime_unix,
		   indexed_at_unix = excluded.indexed_at_unix`,
		fi.Path, fi.SHA256, fi.MtimeUnix, fi.IndexedAtUnix,
	)
	return err
}

// UpsertFileInfoTx updates or inserts file tracking info within a transaction
func (s *SQLite) UpsertFileInfoTx(ctx context.Context, tx *sql.Tx, fi *FileInfo) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO files (path, sha256, mtime_unix, indexed_at_unix)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT(path) DO UPDATE SET
		   sha256 = excluded.sha256,
		   mtime_unix = excluded.mtime_unix,
		   indexed_at_unix = excluded.indexed_at_unix`,
		fi.Path, fi.SHA256, fi.MtimeUnix, fi.IndexedAtUnix,
	)
	return err
}

// MarkChunksInactive marks all chunks for a file as inactive (soft delete)
func (s *SQLite) MarkChunksInactive(ctx context.Context, tx *sql.Tx, path string) error {
	_, err := tx.ExecContext(ctx,
		"UPDATE chunks SET active = 0 WHERE path = ?",
		path,
	)
	return err
}

// InsertChunk inserts a new chunk and returns its ID
func (s *SQLite) InsertChunk(ctx context.Context, tx *sql.Tx, c *Chunk) (int64, error) {
	c.CreatedAtUnix = time.Now().Unix()
	c.Active = true

	// Convert tags to JSON string
	tagsJSON := "[]"
	if len(c.Tags) > 0 {
		tagParts := make([]string, len(c.Tags))
		for i, tag := range c.Tags {
			tagParts[i] = fmt.Sprintf(`"%s"`, tag)
		}
		tagsJSON = "[" + strings.Join(tagParts, ",") + "]"
	}

	result, err := tx.ExecContext(ctx,
		`INSERT INTO chunks (path, heading_path, chunk_index, content, content_sha256, 
		                     start_line, end_line, active, created_at_unix,
		                     status, scope, note_type, category_weight, tags)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		c.Path, c.HeadingPath, c.ChunkIndex, c.Content, c.ContentSHA256,
		c.StartLine, c.EndLine, 1, c.CreatedAtUnix,
		c.Status, c.Scope, c.NoteType, c.CategoryWeight, tagsJSON,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// InsertEmbedding inserts a vector for a chunk
func (s *SQLite) InsertEmbedding(ctx context.Context, tx *sql.Tx, e *Embedding) error {
	vecBlob := Float32ToBytes(e.Vec)
	_, err := tx.ExecContext(ctx,
		"INSERT INTO embeddings (chunk_id, dim, vec) VALUES (?, ?, ?)",
		e.ChunkID, e.Dim, vecBlob,
	)
	return err
}

// BeginTx starts a transaction
func (s *SQLite) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return s.db.BeginTx(ctx, nil)
}

// GetActiveChunkCount returns the number of active chunks
func (s *SQLite) GetActiveChunkCount(ctx context.Context) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM chunks WHERE active = 1").Scan(&count)
	return count, err
}

// StreamActiveEmbeddings streams all active chunk embeddings
func (s *SQLite) StreamActiveEmbeddings(ctx context.Context) (*sql.Rows, error) {
	return s.db.QueryContext(ctx,
		`SELECT c.id, e.vec 
		 FROM chunks c
		 JOIN embeddings e ON c.id = e.chunk_id
		 WHERE c.active = 1
		 ORDER BY c.id`,
	)
}

// GetChunksByIDs fetches chunks with embeddings by their IDs
func (s *SQLite) GetChunksByIDs(ctx context.Context, ids []uint64) ([]ChunkWithEmbedding, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	// Build query with placeholders
	query := `SELECT c.id, c.path, c.heading_path, c.chunk_index, c.content, 
	                 c.content_sha256, c.start_line, c.end_line, c.active, 
	                 c.created_at_unix, c.status, c.scope, c.note_type,
	                 c.category_weight, c.tags, e.dim, e.vec
	          FROM chunks c
	          JOIN embeddings e ON c.id = e.chunk_id
	          WHERE c.active = 1 AND c.id IN (`

	args := make([]interface{}, len(ids))
	for i, id := range ids {
		if i > 0 {
			query += ","
		}
		query += "?"
		args[i] = id
	}
	query += ")"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []ChunkWithEmbedding
	for rows.Next() {
		var cwe ChunkWithEmbedding
		var vecBlob []byte
		var tagsJSON string
		var dim int
		var active int

		err := rows.Scan(
			&cwe.ID, &cwe.Path, &cwe.HeadingPath, &cwe.ChunkIndex, &cwe.Content,
			&cwe.ContentSHA256, &cwe.StartLine, &cwe.EndLine, &active,
			&cwe.CreatedAtUnix, &cwe.Status, &cwe.Scope, &cwe.NoteType,
			&cwe.CategoryWeight, &tagsJSON, &dim, &vecBlob,
		)
		if err != nil {
			return nil, err
		}

		cwe.Active = active == 1

		// Parse tags from JSON
		if tagsJSON != "" && tagsJSON != "[]" {
			// Simple JSON array parsing
			tagsJSON = strings.Trim(tagsJSON, "[]")
			if tagsJSON != "" {
				for _, tag := range strings.Split(tagsJSON, ",") {
					tag = strings.Trim(strings.TrimSpace(tag), `"`)
					if tag != "" {
						cwe.Tags = append(cwe.Tags, tag)
					}
				}
			}
		}

		vec, err := BytesToFloat32(vecBlob)
		if err != nil {
			return nil, fmt.Errorf("decode vec for chunk %d: %w", cwe.ID, err)
		}
		cwe.Vec = vec
		results = append(results, cwe)
	}

	return results, rows.Err()
}

// GetIndexMeta retrieves index metadata value
func (s *SQLite) GetIndexMeta(ctx context.Context, key string) (string, error) {
	var value string
	err := s.db.QueryRowContext(ctx,
		"SELECT value FROM index_meta WHERE key = ?",
		key,
	).Scan(&value)

	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

// SetIndexMeta sets multiple index metadata values
func (s *SQLite) SetIndexMeta(ctx context.Context, kvs map[string]string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for k, v := range kvs {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO index_meta (key, value) VALUES (?, ?)
			 ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
			k, v,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Float32ToBytes converts float32 slice to little-endian bytes
func Float32ToBytes(vec []float32) []byte {
	b := make([]byte, 4*len(vec))
	for i, f := range vec {
		u := math.Float32bits(f)
		binary.LittleEndian.PutUint32(b[i*4:], u)
	}
	return b
}

// BytesToFloat32 converts little-endian bytes to float32 slice
func BytesToFloat32(b []byte) ([]float32, error) {
	if len(b)%4 != 0 {
		return nil, fmt.Errorf("invalid vec blob: length %d not divisible by 4", len(b))
	}
	n := len(b) / 4
	out := make([]float32, n)
	for i := 0; i < n; i++ {
		u := binary.LittleEndian.Uint32(b[i*4:])
		out[i] = math.Float32frombits(u)
	}
	return out, nil
}

// GetIndexMetaInt retrieves integer metadata
func (s *SQLite) GetIndexMetaInt(ctx context.Context, key string) (int, error) {
	val, err := s.GetIndexMeta(ctx, key)
	if err != nil || val == "" {
		return 0, err
	}
	return strconv.Atoi(val)
}

// GetCanonChunkIDs fetches only canon chunk IDs for priority retrieval
func (s *SQLite) GetCanonChunkIDs(ctx context.Context, limit int) ([]uint64, error) {
	query := `SELECT id FROM chunks 
	          WHERE active = 1 AND category = 'canon' AND status = 'active'
	          ORDER BY id
	          LIMIT ?`

	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []uint64
	for rows.Next() {
		var id uint64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
