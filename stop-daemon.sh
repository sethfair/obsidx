#!/bin/bash
# Stop both indexer and search server daemons

INDEXER_PID_FILE=".obsidian-index/indexer.pid"
SERVER_PID_FILE=".obsidian-index/recall-server.pid"

STOPPED=0

# Stop indexer
if [ -f "$INDEXER_PID_FILE" ]; then
    PID=$(cat "$INDEXER_PID_FILE")

    if kill -0 "$PID" 2>/dev/null; then
        echo "🛑 Stopping indexer (PID: $PID)..."
        kill "$PID"

        # Wait for graceful shutdown
        for i in {1..5}; do
            if ! kill -0 "$PID" 2>/dev/null; then
                echo "✓ Indexer stopped"
                STOPPED=$((STOPPED + 1))
                break
            fi
            sleep 1
        done

        # Force kill if still running
        if kill -0 "$PID" 2>/dev/null; then
            echo "⚠️  Force killing indexer..."
            kill -9 "$PID"
            sleep 1
        fi
    else
        echo "✓ Indexer not running (stale PID file)"
    fi

    rm -f "$INDEXER_PID_FILE"
else
    echo "✓ Indexer not running"
fi

# Stop search server
if [ -f "$SERVER_PID_FILE" ]; then
    PID=$(cat "$SERVER_PID_FILE")

    if kill -0 "$PID" 2>/dev/null; then
        echo "🛑 Stopping search server (PID: $PID)..."
        kill "$PID"

        # Wait for graceful shutdown
        for i in {1..5}; do
            if ! kill -0 "$PID" 2>/dev/null; then
                echo "✓ Server stopped"
                STOPPED=$((STOPPED + 1))
                break
            fi
            sleep 1
        done

        # Force kill if still running
        if kill -0 "$PID" 2>/dev/null; then
            echo "⚠️  Force killing server..."
            kill -9 "$PID"
            sleep 1
        fi
    else
        echo "✓ Server not running (stale PID file)"
    fi

    rm -f "$SERVER_PID_FILE"
else
    echo "✓ Server not running"
fi

if [ $STOPPED -gt 0 ]; then
    echo ""
    echo "✅ All daemons stopped"
fi
