#!/bin/bash
# Fast search wrapper - preloads index on first search, then keeps it warm

CACHE_FILE=".obsidian-index/index-cache-timestamp"
CACHE_LIFETIME=3600 # 1 hour

# Function to run search
run_search() {
    ./bin/obsidx-recall "$@"
    touch "$CACHE_FILE"
}

# Check if we should suggest keeping terminal open
if [ ! -f "$CACHE_FILE" ] || [ $(($(date +%s) - $(stat -f %m "$CACHE_FILE" 2>/dev/null || echo 0))) -gt $CACHE_LIFETIME ]; then
    echo "💡 TIP: Keep this terminal open and run multiple searches to avoid reloading the index"
    echo "   First search: ~3s (loads 7K vectors)"
    echo "   Subsequent searches: instant (if you keep searching)"
    echo ""
fi

run_search "$@"
