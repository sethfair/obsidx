#!/bin/bash
set -e

echo "🚀 Starting obsidx..."
echo ""

# Check if Ollama is installed
if ! command -v ollama &> /dev/null; then
    echo "❌ Ollama is not installed"
    echo ""
    echo "Installation options:"
    echo ""
    echo "  macOS:"
    echo "    brew install ollama"
    echo "    # or download from https://ollama.ai"
    echo ""
    echo "  Linux:"
    echo "    curl -fsSL https://ollama.com/install.sh | sh"
    echo ""
    echo "After installing, run this script again:"
    echo "  ./run.sh $1"
    echo ""
    echo "See INSTALL.md for detailed instructions"
    exit 1
fi

# Check if Ollama is running
if ! curl -s http://localhost:11434/api/tags &> /dev/null; then
    echo "⚠️  Ollama is not running"
    echo "Starting Ollama in the background..."

    # Start Ollama in the background
    ollama serve > /tmp/ollama.log 2>&1 &
    OLLAMA_PID=$!

    echo "Waiting for Ollama to start..."
    for i in {1..10}; do
        if curl -s http://localhost:11434/api/tags &> /dev/null; then
            echo "✓ Ollama started (PID: $OLLAMA_PID)"
            break
        fi
        sleep 1
        if [ $i -eq 10 ]; then
            echo "❌ Ollama failed to start. Check /tmp/ollama.log"
            exit 1
        fi
    done
else
    echo "✓ Ollama is already running"
fi

# Check if model is installed
MODEL=${MODEL:-nomic-embed-text}
echo ""
echo "Checking for model: $MODEL"

if ! ollama list | grep -q "$MODEL"; then
    echo "⚠️  Model '$MODEL' not found"
    echo "Pulling model (this may take a few minutes)..."
    ollama pull "$MODEL"
    echo "✓ Model downloaded"
else
    echo "✓ Model '$MODEL' is installed"
fi

echo ""
echo "─────────────────────────────────────────"
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

echo "📚 Indexing vault: $VAULT"
echo ""

# Check if database needs migration
DB_PATH=".obsidian-index/obsidx.db"
if [ -f "$DB_PATH" ]; then
    # Check if category column exists
    if ! sqlite3 "$DB_PATH" "PRAGMA table_info(chunks);" | grep -q "category"; then
        echo "⚠️  Database needs migration"
        echo "Running migration script..."
        ./migrate.sh "$DB_PATH"
        echo ""
    fi
fi

# Run the indexer
./bin/obsidx-indexer --vault "$VAULT" --watch
