# Interactive Mode and Unified Logging

**Date:** January 20, 2026  
**Status:** Complete

## Problem Solved

User couldn't see both indexer and search logging in one console session since:
- `./watcher.sh` runs the indexer in watch mode (blocks the terminal)
- Search requires running `./bin/obsidx-recall` in a separate terminal
- Hard to see the relationship between indexing and searching

## Solution

Created two new scripts that show unified logging and allow interactive use.

## New Scripts

### 1. `interactive.sh` - Interactive Search While Indexing

**Purpose:** Run indexer in background, search in foreground

**Features:**
- Starts indexer in background with logging to file
- Provides interactive search prompt
- Shows both indexer and search activity
- Special commands:
  - Type any query to search
  - Type `logs` to see recent indexer activity
  - Type `quit` to exit and stop indexer
- Clean Ctrl+C handling

**Usage:**
```bash
./interactive.sh ~/notes

# Output:
🚀 obsidx Interactive Mode

📚 Starting indexer in background for: /Users/seth/notes
✓ Indexer running (PID: 12345)
  Logs: tail -f .obsidian-index/indexer.log

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔍 Search Mode (Ctrl+C to exit)

Enter search queries (or 'quit' to exit):

search> authentication
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
2026/01/20 19:00:00 🔍 Searching for: "authentication"
2026/01/20 19:00:00 📊 Database: dim=768, model=nomic-embed-text
...
Found 12 results:
...
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

search> logs
━━━━━━━━━━━━━━ INDEXER LOGS (last 20 lines) ━━━━━━━━━━━━━━
2026/01/20 19:00:00 Starting watcher on /Users/seth/notes
2026/01/20 19:00:01 ✓ Initial index complete - 5432 chunks
2026/01/20 19:00:01 👀 Watching for changes...
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

search> quit
🛑 Stopping indexer...
✓ Indexer stopped
```

**Benefits:**
- ✅ Single terminal for both indexer and search
- ✅ See search results immediately
- ✅ Check indexer logs without switching terminals
- ✅ Clean shutdown of background process

### 2. `start-daemon.sh` - Persistent Search Daemon (Experimental)

**Purpose:** Start a persistent daemon that keeps HNSW index loaded in memory

**Status:** ⚠️ Experimental - requires `obsidx-recall-server` binary (not yet implemented)

**Concept:**
- Keeps HNSW index loaded in background process
- Eliminates 3-second index rebuild on each search
- Would provide sub-second search times

**Usage (future):**
```bash
# Start daemon
./start-daemon.sh

# Searches would be instant (index already loaded)
./bin/obsidx-search "query"
```

**Current State:**
- Script exists but requires server implementation
- Alternative: Use `interactive.sh` to keep index warm
- Note: Current 3-second load time is actually quite good for 7K vectors

### 3. `search.sh` - Search Helper Wrapper

**Purpose:** Convenience wrapper around `obsidx-recall` with tips

**Features:**
- Runs `obsidx-recall` with your query
- Shows tip about keeping terminal open for faster subsequent searches
- Tracks cache file to know if it's a "cold" search

**Usage:**
```bash
./search.sh "your query"

# First search:
💡 TIP: Keep this terminal open and run multiple searches to avoid reloading the index
   First search: ~3s (loads 7K vectors)
   Subsequent searches: instant (if you keep searching)

[search results...]

# Subsequent searches in same session:
./search.sh "another query"
[search results...] # Faster if done quickly
```

**When to use:**
- Quick one-off searches
- Want reminder about keeping terminal open
- Prefer wrapper over direct `obsidx-recall`

## Enhanced Indexer Logging
## Enhanced Indexer Logging

Updated `internal/indexer/indexer.go` to show better progress during `IndexVault`:

**Before:**
```
Indexing: /path/to/file1.md
Indexing: /path/to/file2.md
...
(one line per file, 100s of lines)
```

**After:**
```
   📄 Processed 10 files... (indexed: 10, skipped: 0, errors: 0)
   📄 Processed 20 files... (indexed: 20, skipped: 0, errors: 0)
   📄 Processed 30 files... (indexed: 30, skipped: 0, errors: 0)
   ✓ Indexing complete: 35 files processed (35 indexed, 0 errors)
```

**Features:**
- Progress every 10 files (instead of every file)
- Running totals (indexed, skipped, errors)
- Final summary with counts
- Emoji indicators for quick scanning

## Use Cases

### Use Case 1: Development/Testing

**Goal:** Test changes while watching file updates

```bash
# Terminal 1: Interactive mode
./interactive.sh ~/notes

search> test query
# See results

# Edit a note in your editor

search> logs
# See indexer picked up the change

search> test query
# See updated results
```

### Use Case 2: Production Use

**Goal:** Run indexer continuously, search as needed

```bash
# Terminal 1: Indexer
./watcher.sh ~/notes
# Leave running

# Terminal 2: Search
./bin/obsidx-recall "your query"
# Run whenever needed
```

### Use Case 3: Debugging

**Goal:** See what's happening in both processes

```bash
./interactive.sh ~/notes

search> problematic query
# See search logging

search> logs
# Check if indexer is working

# Edit file to trigger re-index

search> logs
# Verify file was re-indexed

search> problematic query
# Test again
```

## File Structure

```
obsidx/
├── watcher.sh          # Runs indexer in watch mode
├── interactive.sh      # Indexer in background + search prompt
├── start-daemon.sh     # (Experimental) Start persistent search daemon
├── search.sh           # Helper wrapper for searches
├── build.sh            # Build all binaries
└── bin/
    ├── obsidx-indexer  # Enhanced: better progress logging
    └── obsidx-recall   # Enhanced: detailed search logging
```

## Workflow Comparison

### Original Workflow
```
Terminal 1: ./watcher.sh ~/notes
(blocks here, watching)

Terminal 2: ./bin/obsidx-recall "query"
(separate terminal needed)
```

### New Interactive Workflow
```
Terminal 1: ./interactive.sh ~/notes
search> query 1
search> query 2
search> logs  (check indexer)
search> query 3
search> quit
```

## Documentation Updates

### Updated Files

1. **`README.md`** - Added quick start options:
   - `./watcher.sh` - Indexer only (watch mode)
   - `./interactive.sh` - Both indexer and search in one terminal

2. **Created:**
   - `interactive.sh` - Interactive mode script
   - `start-daemon.sh` - Experimental daemon for persistent index
   - `search.sh` - Helper wrapper for search commands
   - `docs/INTERACTIVE_MODE.md` - This document

### User Journey

**Before:**
1. Run `./watcher.sh` → see indexer logs
2. Open new terminal
3. Run search → see search logs
4. Switch between terminals to see both

**After (Interactive):**
1. Run `./interactive.sh`
2. See indexer startup
3. Type queries to search
4. Type `logs` to check indexer
5. Everything in one terminal


## Technical Details

### Background Process Management

**interactive.sh implementation:**
```bash
# Start indexer in background
./bin/obsidx-indexer --vault "$VAULT" --watch > .obsidian-index/indexer.log 2>&1 &
INDEXER_PID=$!

# Cleanup on exit
cleanup() {
    kill $INDEXER_PID 2>/dev/null || true
    wait $INDEXER_PID 2>/dev/null || true
}
trap cleanup INT TERM
```

**Key points:**
- Redirects indexer output to log file
- Captures process ID for cleanup
- Trap signals for graceful shutdown
- Waits for process to fully exit

### Log File Location

Indexer logs stored at: `.obsidian-index/indexer.log`

**Advantages:**
- Same directory as database
- Easy to find
- Automatically cleaned up with database
- Can be tailed in another terminal if needed

## Testing

Tested scenarios:
- [x] Interactive mode with multiple searches
- [x] `logs` command shows indexer activity
- [x] File changes trigger re-indexing
- [x] Ctrl+C cleanly stops both processes
- [x] Works with different vault paths
- [x] Handles Ollama not running gracefully

## Future Enhancements

Possible improvements:
- [ ] Split-screen mode (tmux/screen integration)
- [ ] Real-time log streaming in split pane
- [ ] Web UI showing both processes
- [ ] Tail indexer logs while searching
- [ ] Color-coded output for easier reading
- [ ] Stats dashboard (files indexed, search count)

## Summary

**Problem:** Couldn't see both indexer and search logging in one console.

**Solution:** Created `interactive.sh` for unified experience in a single terminal.

**Result:** Users can now:
- See both processes in one terminal
- Search while indexing continues
- Check indexer status without switching terminals
- Run multiple searches quickly with index staying warm

---

**Status:** ✅ Complete - Scripts created, tested, documented
