# Installation Guide

## Prerequisites

### 1. Install Ollama

**macOS:**
```bash
# Download from https://ollama.ai
# Or use Homebrew:
brew install ollama
```

**Linux:**
```bash
curl -fsSL https://ollama.com/install.sh | sh
```

**Windows:**
Download from https://ollama.ai

### 2. Start Ollama

```bash
# Start Ollama service
ollama serve

# In another terminal, pull the embedding model
ollama pull nomic-embed-text
```

### 3. Build obsidx

```bash
cd obsidx
./build.sh
```

### 4. Run

**Easy mode (handles everything):**
```bash
./run.sh ~/notes
```

**Manual mode:**
```bash
# Make sure Ollama is running first
ollama serve &

# Then start indexing
./bin/obsidx-indexer --vault ~/notes --watch
```

## Verifying Installation

Test Ollama is working:
```bash
curl http://localhost:11434/api/tags
```

Should return JSON with installed models.

## Troubleshooting

**"Connection refused" error:**
- Ollama is not running. Start it with: `ollama serve`

**"Model not found" error:**
- Install the model: `ollama pull nomic-embed-text`

**Build errors:**
- Make sure Go 1.22+ is installed: `go version`
- Run: `go mod tidy`

## Quick Reference

```bash
# Start everything in one command (after Ollama is installed)
./run.sh ~/notes

# Search your vault
./bin/obsidx-recall "your query"

# Rebuild index
./bin/obsidx-rebuild

# Use different model
MODEL=all-minilm ./run.sh ~/notes
```
