#!/bin/bash
# Show logs from obsidx daemons

INDEXER_LOG=".obsidian-index/indexer.log"
SERVER_LOG=".obsidian-index/recall-server.log"

# Parse command line options
FOLLOW=false
LINES=50
WHICH="both"

usage() {
    echo "Usage: $0 [options]"
    echo ""
    echo "Options:"
    echo "  -f, --follow        Follow log output (like tail -f)"
    echo "  -n, --lines NUM     Number of lines to show (default: 50)"
    echo "  -i, --indexer       Show only indexer logs"
    echo "  -s, --server        Show only server logs"
    echo "  -a, --all           Show all logs (no limit)"
    echo "  -h, --help          Show this help"
    echo ""
    echo "Examples:"
    echo "  $0                  # Show last 50 lines of both logs"
    echo "  $0 -f               # Follow both logs in real-time"
    echo "  $0 -i -n 100        # Show last 100 lines of indexer log"
    echo "  $0 -s -f            # Follow server log"
    echo "  $0 -a               # Show all logs"
    exit 0
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -f|--follow)
            FOLLOW=true
            shift
            ;;
        -n|--lines)
            LINES="$2"
            shift 2
            ;;
        -i|--indexer)
            WHICH="indexer"
            shift
            ;;
        -s|--server)
            WHICH="server"
            shift
            ;;
        -a|--all)
            LINES="all"
            shift
            ;;
        -h|--help)
            usage
            ;;
        *)
            echo "Unknown option: $1"
            usage
            ;;
    esac
done

# Check if logs exist
if [ ! -f "$INDEXER_LOG" ] && [ ! -f "$SERVER_LOG" ]; then
    echo "❌ No log files found"
    echo ""
    echo "Logs will be created when you start the daemons:"
    echo "  ./start-daemon.sh"
    exit 1
fi

# Helper function to show logs
show_log() {
    local log_file=$1
    local label=$2

    if [ ! -f "$log_file" ]; then
        echo "⚠️  $label log not found: $log_file"
        return
    fi

    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "📋 $label ($log_file)"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    if [ "$FOLLOW" = true ]; then
        tail -f "$log_file"
    elif [ "$LINES" = "all" ]; then
        cat "$log_file"
    else
        tail -n "$LINES" "$log_file"
    fi

    echo ""
}

# Show requested logs
case $WHICH in
    indexer)
        show_log "$INDEXER_LOG" "Indexer"
        ;;
    server)
        show_log "$SERVER_LOG" "Search Server"
        ;;
    both)
        if [ "$FOLLOW" = true ]; then
            # For follow mode, we need to tail both files
            echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
            echo "📋 Following both logs (Ctrl+C to stop)"
            echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
            echo ""

            if [ -f "$INDEXER_LOG" ] && [ -f "$SERVER_LOG" ]; then
                tail -f "$INDEXER_LOG" "$SERVER_LOG"
            elif [ -f "$INDEXER_LOG" ]; then
                tail -f "$INDEXER_LOG"
            elif [ -f "$SERVER_LOG" ]; then
                tail -f "$SERVER_LOG"
            fi
        else
            show_log "$INDEXER_LOG" "Indexer"
            show_log "$SERVER_LOG" "Search Server"
        fi
        ;;
esac

# Show helpful footer (unless following)
if [ "$FOLLOW" = false ]; then
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "💡 Tips:"
    echo "  ./logs.sh -f          # Follow logs in real-time"
    echo "  ./logs.sh -i          # Show only indexer logs"
    echo "  ./logs.sh -s          # Show only server logs"
    echo "  ./logs.sh -a          # Show all logs"
    echo "  ./logs.sh -n 100      # Show last 100 lines"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
fi
