# obsidx

A semantic search tool for Obsidian vaults using SQLite as source of truth and HNSW for fast vector similarity search.

## Architecture

- **SQLite** is the authoritative store for chunks, metadata, and vectors
- **HNSW** (Weaviate's implementation) is a rebuildable acceleration structure
- **Two-stage recall**: HNSW candidates → exact cosine rerank in Go

## Components

- `obsidx-indexer`: Daemon that watches vault and indexes changes
- `obsidx-recall`: CLI for semantic search queries
- `obsidx-rebuild`: CLI to rebuild HNSW index from SQLite

## Project Structure

```
obsidx/
  cmd/
    obsidx-indexer/     # daemon: watch vault & index
    obsidx-recall/      # CLI: recall "query"
    obsidx-rebuild/     # CLI: rebuild HNSW from SQLite
  internal/
    chunker/            # markdown chunking
    embed/              # embedder interface + implementations
    store/              # sqlite store
    ann/                # Weaviate HNSW wrapper
    rank/               # cosine rerank, top-k heap
    watcher/            # fs notify + debounce
```

## Quick Start

```bash
# Build all commands
go build ./cmd/obsidx-indexer
go build ./cmd/obsidx-recall
go build ./cmd/obsidx-rebuild

# Index a vault
./obsidx-indexer --vault /path/to/vault --db .obsidian-index/obsidx.db

# Query
./obsidx-recall --db .obsidian-index/obsidx.db "your semantic query"

# Rebuild HNSW index
./obsidx-rebuild --db .obsidian-index/obsidx.db
```

## Configuration

By default, the tool expects:
- Embedding dimension: 768 (configurable)
- Distance metric: cosine
- HNSW parameters: M=16, efConstruction=128, efSearch=64

## Embedding Provider

You need to provide an embedding service. The default implementation expects
a local HTTP endpoint at `http://localhost:8080/embed` that accepts JSON:

```json
{"text": "your text here"}
```

And returns:

```json
{"vector": [0.1, 0.2, ...]}
```

See `internal/embed/` for implementing custom providers.
