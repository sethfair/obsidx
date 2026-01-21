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

echo ""
echo "✓ Build complete!"
echo ""
echo "Binaries created:"
echo "  bin/obsidx-indexer"
echo "  bin/obsidx-recall"
echo "  bin/obsidx-rebuild"
echo ""
echo "Quick start:"
echo "  ./watcher.sh ~/notes     # Start with auto-setup"
echo ""
echo "Or manually:"
echo "  ollama serve               # (in another terminal)"
echo "  ./bin/obsidx-indexer --vault ~/notes --watch"
echo "  ./bin/obsidx-recall \"your search query\""


