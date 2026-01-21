#!/bin/bash
# Start both indexer and search server as background daemons

VAULT=${1:-~/notes}
VAULT=$(eval echo "$VAULT")

if [ ! -d "$VAULT" ]; then
    echo "❌ Vault directory not found: $VAULT"
    echo ""
    echo "Usage: $0 [vault-path]"
    echo "Example: $0 ~/notes"
    exit 1
fi

DB_PATH=".obsidian-index/obsidx.db"
INDEXER_PID_FILE=".obsidian-index/indexer.pid"
SERVER_PID_FILE=".obsidian-index/recall-server.pid"
INDEXER_LOG=".obsidian-index/indexer.log"
SERVER_LOG=".obsidian-index/recall-server.log"
PORT=8765

echo "🚀 Starting obsidx daemons..."
echo "📚 Vault: $VAULT"
echo ""

# Check if indexer is already running
if [ -f "$INDEXER_PID_FILE" ]; then
    PID=$(cat "$INDEXER_PID_FILE")
    if kill -0 "$PID" 2>/dev/null; then
        echo "✓ Indexer already running (PID: $PID)"
        INDEXER_PID=$PID
    else
        rm -f "$INDEXER_PID_FILE"
    fi
fi

# Check if server is already running
if [ -f "$SERVER_PID_FILE" ]; then
    PID=$(cat "$SERVER_PID_FILE")
    if kill -0 "$PID" 2>/dev/null; then
        echo "✓ Search server already running (PID: $PID)"
        SERVER_PID=$PID
    else
        rm -f "$SERVER_PID_FILE"
    fi
fi

# Start indexer if not running
if [ -z "$INDEXER_PID" ]; then
    echo "📥 Starting indexer..."
    ./bin/obsidx-indexer --vault "$VAULT" --watch > "$INDEXER_LOG" 2>&1 &
    INDEXER_PID=$!
    echo "$INDEXER_PID" > "$INDEXER_PID_FILE"
    echo "✓ Indexer started (PID: $INDEXER_PID)"

    # Wait for initial indexing
    sleep 3
fi

# Start search server if not running
if [ -z "$SERVER_PID" ]; then
    # Check if database exists
    if [ ! -f "$DB_PATH" ]; then
        echo "⏳ Waiting for indexer to create database..."
        for i in {1..10}; do
            if [ -f "$DB_PATH" ]; then
                break
            fi
            sleep 1
        done

        if [ ! -f "$DB_PATH" ]; then
            echo "❌ Database not created. Check indexer logs:"
            echo "   tail -f $INDEXER_LOG"
            exit 1
        fi
    fi

    echo "🔍 Starting search server..."
    ./bin/obsidx-recall-server --db "$DB_PATH" --port "$PORT" > "$SERVER_LOG" 2>&1 &
    SERVER_PID=$!
    echo "$SERVER_PID" > "$SERVER_PID_FILE"

    # Wait for server to start
    sleep 2

    # Verify server is running
    if ! kill -0 "$SERVER_PID" 2>/dev/null; then
        echo "❌ Server failed to start. Check logs:"
        echo "   tail $SERVER_LOG"
        rm -f "$SERVER_PID_FILE"
        exit 1
    fi

    # Wait for server to be ready
    echo "⏳ Waiting for server to load index..."
    for i in {1..30}; do
        if curl -s http://localhost:$PORT/health > /dev/null 2>&1; then
            break
        fi
        sleep 1
    done

    if curl -s http://localhost:$PORT/health > /dev/null 2>&1; then
        echo "✓ Search server ready (PID: $SERVER_PID)"
    else
        echo "⚠️  Server started but not responding. Check logs:"
        echo "   tail -f $SERVER_LOG"
    fi
fi

echo ""
echo "✅ All daemons running!"
echo ""
echo "PIDs:"
echo "  Indexer: $INDEXER_PID"
echo "  Server:  $SERVER_PID"
echo ""
echo "Usage:"
echo "  ./bin/obsidx-recall \"your query\"  # Instant search (<100ms)"
echo ""
echo "Logs:"
echo "  tail -f $INDEXER_LOG   # Indexer activity"
echo "  tail -f $SERVER_LOG    # Search server activity"
echo ""
echo "Stop:"
echo "  ./stop-daemon.sh       # Stop both processes"

