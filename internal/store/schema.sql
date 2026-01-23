-- SQLite schema for obsidx
-- This is the source of truth for all chunk data

-- Files table: tracks processed files to avoid reprocessing unchanged content
CREATE TABLE IF NOT EXISTS files (
  path TEXT PRIMARY KEY,
  sha256 TEXT NOT NULL,
  mtime_unix INTEGER NOT NULL,
  indexed_at_unix INTEGER NOT NULL
);

-- Chunks table: markdown chunks with metadata
CREATE TABLE IF NOT EXISTS chunks (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  path TEXT NOT NULL,
  heading_path TEXT,
  chunk_index INTEGER NOT NULL,
  content TEXT NOT NULL,
  content_sha256 TEXT NOT NULL,
  start_line INTEGER,
  end_line INTEGER,
  active INTEGER NOT NULL DEFAULT 1,
  created_at_unix INTEGER NOT NULL,
  -- Metadata fields
  status TEXT,
  scope TEXT,
  note_type TEXT,
  category_weight REAL DEFAULT 1.0,
  tags TEXT  -- JSON array of tags
);

-- Embeddings table: raw vectors for chunks
CREATE TABLE IF NOT EXISTS embeddings (
  chunk_id INTEGER PRIMARY KEY,
  dim INTEGER NOT NULL,
  vec BLOB NOT NULL,
  FOREIGN KEY(chunk_id) REFERENCES chunks(id) ON DELETE CASCADE
);

-- Index metadata: tracks HNSW index state
CREATE TABLE IF NOT EXISTS index_meta (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_chunks_path ON chunks(path);
CREATE INDEX IF NOT EXISTS idx_chunks_active ON chunks(active);
CREATE INDEX IF NOT EXISTS idx_chunks_status ON chunks(status);
CREATE INDEX IF NOT EXISTS idx_files_mtime ON files(mtime_unix);
