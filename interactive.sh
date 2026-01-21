#!/bin/bash
# Run indexer in background and allow interactive searches
set -e

echo "🚀 obsidx Interactive Mode"
echo ""

# Get vault path from argument or use default
VAULT=${1:-~/notes}
VAULT=$(eval echo "$VAULT")  # Expand ~ and variables

if [ ! -d "$VAULT" ]; then
    echo "❌ Vault directory not found: $VAULT"
    echo ""
    echo "Usage: $0 [vault-path]"
    echo "Example: $0 ~/notes"
    exit 1
fi

# Check if Ollama is running
if ! curl -s http://localhost:11434/api/tags &> /dev/null; then
    echo "❌ Ollama is not running"
    echo "Start Ollama first: ollama serve"
    exit 1
fi

# Start indexer in background with logging
echo "📚 Starting indexer in background for: $VAULT"
./bin/obsidx-indexer --vault "$VAULT" --watch > .obsidian-index/indexer.log 2>&1 &
INDEXER_PID=$!

# Wait a moment for indexer to start
sleep 2

echo "✓ Indexer running (PID: $INDEXER_PID)"
echo "  Logs: tail -f .obsidian-index/indexer.log"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Cleanup function
cleanup() {
    echo ""
    echo ""
    echo "🛑 Stopping indexer..."
    kill $INDEXER_PID 2>/dev/null || true
    wait $INDEXER_PID 2>/dev/null || true
    echo "✓ Indexer stopped"
    exit 0
}

trap cleanup INT TERM

# Interactive search loop
echo "🔍 Search Mode (Ctrl+C to exit)"
echo ""
echo "Enter search queries (or 'quit' to exit):"
echo ""

while true; do
    echo -n "search> "
    read query

    if [ -z "$query" ]; then
        continue
    fi

    if [ "$query" = "quit" ] || [ "$query" = "exit" ]; then
        break
    fi

    if [ "$query" = "logs" ]; then
        echo ""
        echo "━━━━━━━━━━━━━━ INDEXER LOGS (last 20 lines) ━━━━━━━━━━━━━━"
        tail -20 .obsidian-index/indexer.log
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo ""
        continue
    fi

    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    ./bin/obsidx-recall "$query"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
done

cleanup
