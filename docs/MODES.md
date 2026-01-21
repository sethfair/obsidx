# Interactive & Daemon Modes

Quick guide to different ways of running obsidx.

## Primary Method: Background Daemon

Start both indexer and search server as background processes:

```bash
./start-daemon.sh ~/notes
```

**What it does:**
- Starts indexer in watch mode (background)
- Starts search server with persistent HNSW index (background)
- Both processes run independently
- Returns immediately

**Usage:**
```bash
# Start everything
./start-daemon.sh ~/notes

# Search anytime (fast: <100ms)
./bin/obsidx-recall "your query"

# Stop everything
./stop-daemon.sh
```

**Logs:**
```bash
tail -f .obsidian-index/indexer.log         # Indexer activity
tail -f .obsidian-index/recall-server.log   # Search server activity
```

## Alternative: Foreground with Live Logs

Start everything and tail logs in the foreground:

```bash
./start.sh ~/notes
```

**What it does:**
- Starts indexer (background)
- Starts search server (background)
- Tails logs from both processes
- Ctrl+C stops everything

**When to use:**
- When you want to watch activity in real-time
- During development/debugging
- When you don't need the terminal for other tasks

## Alternative: Interactive Search Prompt

Start processes and get an interactive search prompt:

```bash
./interactive.sh ~/notes
```

**What it does:**
- Starts indexer (background)
- Starts search server (background)
- Provides interactive search prompt

**Commands:**
- Type any query → runs search
- Type `logs` → shows recent indexer activity
- Type `quit` → stops everything and exits

**When to use:**
- Quick testing and experimentation
- Running multiple searches in sequence
- Development work

## Comparison

| Mode | Command | Processes | Terminal | Best For |
|------|---------|-----------|----------|----------|
| **Daemon** | `./start-daemon.sh` | Background | Free | Production use |
| **Foreground** | `./start.sh` | Background | Blocked (logs) | Watching activity |
| **Interactive** | `./interactive.sh` | Background | Prompt | Testing/dev |

## Process Management

### PIDs
- `.obsidian-index/indexer.pid` - Indexer process
- `.obsidian-index/recall-server.pid` - Search server process

### Check Status
```bash
# Check if running
ps aux | grep obsidx-indexer
ps aux | grep obsidx-recall-server

# Or check health endpoint
curl http://localhost:8765/health
```

### Manual Control
```bash
# Stop everything
./stop-daemon.sh

# Stop just indexer
kill $(cat .obsidian-index/indexer.pid)

# Stop just server
kill $(cat .obsidian-index/recall-server.pid)
```

## Performance

All modes provide the same fast search performance:
- **First search after server start:** ~5 seconds (loads HNSW index into memory)
- **Subsequent searches:** <100ms (index stays loaded)
- **No rebuild:** Index persists in memory between searches

## Troubleshooting

### Server won't start
```bash
# Check logs
tail -f .obsidian-index/recall-server.log

# Common issues:
# - Database doesn't exist (run indexer first)
# - Port 8765 already in use
# - Ollama not running
```

### Indexer not picking up changes
```bash
# Check if running
ps aux | grep obsidx-indexer

# Check logs
tail -f .obsidian-index/indexer.log

# Restart
./stop-daemon.sh
./start-daemon.sh ~/notes
```

### Search gets "server not running"
```bash
# The client auto-starts server if needed
# Just run the search again:
./bin/obsidx-recall "your query"
```

## Summary

**Recommended workflow:**
```bash
# One-time setup
./build.sh

# Start daemons
./start-daemon.sh ~/notes

# Search whenever
./bin/obsidx-recall "query"

# When done
./stop-daemon.sh
```

Simple, fast, and always ready to search!
