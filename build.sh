#!/bin/bash
set -e

echo "Building obsidx..."

# Create bin directory
mkdir -p bin

# Build all commands
echo "→ Building obsidx-indexer..."
go build -o bin/obsidx-indexer ./cmd/obsidx-indexer

echo "→ Building obsidx-recall..."
go build -o bin/obsidx-recall ./cmd/obsidx-recall

echo "→ Building obsidx-rebuild..."
go build -o bin/obsidx-rebuild ./cmd/obsidx-rebuild

echo "→ Building obsidx-recall-server..."
go build -o bin/obsidx-recall-server ./cmd/obsidx-recall-server

echo ""
echo "✓ Build complete!"
echo ""
echo "Binaries created:"
echo "  bin/obsidx-indexer        # Index your vault (watch mode)"
echo "  bin/obsidx-recall         # Search (uses daemon for speed)"
echo "  bin/obsidx-recall-server  # Search daemon (persistent index)"
echo "  bin/obsidx-rebuild        # Rebuild HNSW index"
echo ""
echo "Quick start:"
echo "  ./start-daemon.sh ~/notes     # Start both indexer + search server"
echo "  ./bin/obsidx-recall \"query\"    # Search (fast, <100ms)"
echo "  ./stop-daemon.sh              # Stop everything"


