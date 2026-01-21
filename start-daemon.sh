#!/bin/bash
# Start HNSW index daemon that stays loaded in memory
# This eliminates the 3-second rebuild on every search

VAULT=${1:-~/notes}
VAULT=$(eval echo "$VAULT")

DB_PATH=".obsidian-index/obsidx.db"
DAEMON_PID_FILE=".obsidian-index/recall-daemon.pid"

# Check if daemon is already running
if [ -f "$DAEMON_PID_FILE" ]; then
    PID=$(cat "$DAEMON_PID_FILE")
    if kill -0 "$PID" 2>/dev/null; then
        echo "✓ Recall daemon already running (PID: $PID)"
        exit 0
    fi
fi

echo "🚀 Starting obsidx recall daemon..."
echo "   This keeps the HNSW index loaded in memory for fast searches"
echo ""

# Start the recall server (we'll create this next)
./bin/obsidx-recall-server --db "$DB_PATH" &
DAEMON_PID=$!

echo "$DAEMON_PID" > "$DAEMON_PID_FILE"
echo "✓ Daemon started (PID: $DAEMON_PID)"
echo "  Use ./bin/obsidx-search to run fast searches"
echo ""
