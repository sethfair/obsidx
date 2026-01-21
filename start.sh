#!/bin/bash
# Unified start script - runs indexer and search server as foreground processes
set -e

VAULT=${1:-~/notes}
VAULT=$(eval echo "$VAULT")

if [ ! -d "$VAULT" ]; then
    echo "❌ Vault directory not found: $VAULT"
    echo ""
    echo "Usage: $0 [vault-path]"
    echo "Example: $0 ~/notes"
    exit 1
fi

echo "🚀 Starting obsidx..."
echo "📚 Vault: $VAULT"
echo ""

# Check Ollama
if ! curl -s http://localhost:11434/api/tags &> /dev/null; then
    echo "⚠️  Ollama not running. Starting Ollama..."
    ollama serve > /dev/null 2>&1 &
    sleep 2
    if ! curl -s http://localhost:11434/api/tags &> /dev/null; then
        echo "❌ Failed to start Ollama"
        exit 1
    fi
    echo "✓ Ollama started"
fi

# Check/download model
if ! ollama list | grep -q "nomic-embed-text"; then
    echo "📥 Downloading embedding model (one-time setup)..."
    ollama pull nomic-embed-text
fi

DB_PATH=".obsidian-index/obsidx.db"
INDEXER_LOG=".obsidian-index/indexer.log"
SERVER_LOG=".obsidian-index/recall-server.log"
COMBINED_LOG=".obsidian-index/combined.log"

# Start indexer in background
echo "📥 Starting indexer..."
./bin/obsidx-indexer --vault "$VAULT" --watch > "$INDEXER_LOG" 2>&1 &
INDEXER_PID=$!
echo "✓ Indexer started (PID: $INDEXER_PID)"

# Wait for initial indexing
sleep 3

# Start search server in background
echo "🔍 Starting search server..."
./bin/obsidx-recall-server --db "$DB_PATH" > "$SERVER_LOG" 2>&1 &
SERVER_PID=$!
echo "✓ Search server started (PID: $SERVER_PID)"

# Wait for server to be ready
echo "⏳ Waiting for server to load index..."
for i in {1..30}; do
    if curl -s http://localhost:8765/health > /dev/null 2>&1; then
        echo "✅ Server ready!"
        break
    fi
    sleep 1
done

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🎉 obsidx is running!"
echo ""
echo "Usage:"
echo "  ./bin/obsidx-recall \"your query\"  # Fast searches (<100ms)"
echo ""
echo "Logs:"
echo "  tail -f $INDEXER_LOG  # Indexer activity"
echo "  tail -f $SERVER_LOG   # Search server activity"
echo ""
echo "Stop:"
echo "  Ctrl+C (will stop all processes)"
echo "  Or: kill $INDEXER_PID $SERVER_PID"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📊 Watching logs (Ctrl+C to stop)..."
echo ""

# Cleanup on exit
cleanup() {
    echo ""
    echo ""
    echo "🛑 Stopping obsidx..."
    kill $INDEXER_PID $SERVER_PID 2>/dev/null || true
    wait $INDEXER_PID $SERVER_PID 2>/dev/null || true
    echo "✓ Stopped"
    exit 0
}

trap cleanup INT TERM

# Tail both logs with labels
tail -f "$INDEXER_LOG" "$SERVER_LOG" 2>/dev/null | awk '
/obsidian-index\/indexer.log/ {prefix="[INDEXER] "; next}
/obsidian-index\/recall-server.log/ {prefix="[SERVER]  "; next}
{print prefix $0}
' || true

# If tail fails, just wait
wait
