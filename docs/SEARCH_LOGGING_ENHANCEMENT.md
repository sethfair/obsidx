# Search Logging Enhancement

**Date:** January 20, 2026  
**Status:** Complete

## Summary

Added comprehensive logging to `obsidx-recall` to show search activity, making it clear what's happening during search operations and helping users understand performance characteristics.

## Changes Made

### 1. Enhanced Search Logging

Added detailed logging throughout the search process:

**🔍 Query Information:**
- Shows the search query being executed
- Displays timing from start to finish

**📊 Database Info:**
- Shows embedding dimension
- Shows model name used for indexing

**🔌 Connection Status:**
- Logs Ollama connection attempts
- Confirms successful connection

**🏗️ Index Building:**
- Shows HNSW index construction progress
- Reports number of vectors loaded
- Shows progress every 1000 vectors or 2 seconds

**🔎 Search Stages:**
- **Stage 1:** HNSW candidate retrieval with count and timing
- **Stage 2:** Database chunk fetching with count
- **Stage 3:** Reranking with timing

**✅ Final Summary:**
- Total search time
- Number of results returned

### 2. Added Verbose Flag

New command-line flag: `--verbose` (default: `true`)

**Purpose:** Allow users to disable logging for:
- JSON output piping
- Shell scripting
- CI/CD pipelines
- Tool integration

**Usage:**
```bash
# With logging (default)
obsidx-recall "query"

# Without logging
obsidx-recall --verbose=false "query"
```

### 3. Example Output

**With verbose logging (default):**
```bash
$ obsidx-recall "authentication"
2026/01/20 19:00:00 🔍 Searching for: "authentication"
2026/01/20 19:00:00 📊 Database: dim=768, model=nomic-embed-text
2026/01/20 19:00:00 🔌 Connecting to Ollama at http://localhost:11434...
2026/01/20 19:00:00 ✓ Connected to Ollama
2026/01/20 19:00:00 🏗️  Building HNSW index...
2026/01/20 19:00:01 ✓ Loaded 5432 vectors into HNSW index
2026/01/20 19:00:01 🧮 Generating embedding for query...
2026/01/20 19:00:01 🔎 Stage 1: Searching HNSW for 200 candidates...
2026/01/20 19:00:01 ✓ Found 200 candidates in 2ms
2026/01/20 19:00:01 📥 Stage 2: Fetching chunks from database...
2026/01/20 19:00:01 ✓ Fetched 200 chunks
2026/01/20 19:00:01 🎯 Stage 3: Reranking with exact cosine + category weights...
2026/01/20 19:00:01 ✓ Reranked to top 12 results in 1ms
2026/01/20 19:00:01 ✅ Search complete in 1.5s

Found 12 results:
...
```

**With verbose disabled:**
```bash
$ obsidx-recall --verbose=false "authentication"
Found 12 results:
...
```

## Benefits

### For Users
- **Transparency:** See what's happening during search
- **Performance insight:** Understand where time is spent
- **Debugging:** Easier to diagnose issues
- **Progress feedback:** Know system is working on large indexes

### For Developers
- **Performance profiling:** Identify bottlenecks
- **Debugging:** Trace execution flow
- **Testing:** Verify each stage works correctly

### For Scripting
- **Clean output:** Disable logging when piping
- **No interference:** JSON output not polluted with logs
- **Automation-friendly:** Works in CI/CD pipelines

## Technical Details

### Code Changes

**File:** `cmd/obsidx-recall/main.go`

**Additions:**
1. Import `time` package for duration tracking
2. Import `io` package for output discard
3. New `--verbose` flag (default: true)
4. Log statements at each major step
5. Timing measurements for HNSW and rerank stages
6. Progress logging in `loadIndex()` function

**Key Implementation:**
```go
// Disable logging if not verbose
if !*verbose {
    log.SetOutput(io.Discard)
}

// Timing measurements
searchStart := time.Now()
// ... do work ...
totalDuration := time.Since(searchStart)
log.Printf("✅ Search complete in %v\n", totalDuration)
```

### Logging Strategy

**What gets logged:**
- ✅ User-facing operations (search, connect, load)
- ✅ Performance metrics (timing, counts)
- ✅ Stage completion markers
- ❌ Internal implementation details
- ❌ Debug-level information

**Why emoji indicators:**
- Quick visual scanning
- Differentiate log types
- Make logs friendlier
- Follow modern CLI conventions

### Performance Impact

**Negligible:**
- Logging adds < 1ms to total search time
- Most time is in Ollama embedding (~500ms) and HNSW (~2ms)
- Can be completely disabled with `--verbose=false`

## Documentation Updates

### 1. `docs/RETRIEVAL.md`
- Added "Search Activity Logging" section
- Showed example output
- Documented `--verbose=false` flag
- Updated quick reference examples

### 2. `README.md`
- Added note about search logging
- Listed what gets logged
- Documented verbose flag usage
- Updated search examples

## Usage Examples

### Interactive Use
```bash
# Default: see what's happening
obsidx-recall "authentication"
```

### Scripting
```bash
# Clean output for parsing
results=$(obsidx-recall --verbose=false --json "query" | jq -r '.[0].path')
```

### Performance Analysis
```bash
# See timing breakdown
obsidx-recall "complex query" 2>&1 | grep "complete in"
```

### Debugging
```bash
# Full visibility into search process
obsidx-recall "query" 2>&1 | tee search.log
```

## Backwards Compatibility

**✅ Fully compatible:**
- Default behavior: logging enabled (new feature, doesn't break anything)
- Existing scripts: work as before (just with additional stderr output)
- JSON output: unchanged (logs go to stderr, results to stdout)
- All flags: existing flags unchanged

**Migration path:** None needed - it just works better now

## Future Enhancements

Possible additions:
- [ ] `--debug` flag for even more detailed logging
- [ ] `--quiet` alias for `--verbose=false`
- [ ] Structured logging (JSON) for machine parsing
- [ ] Log levels (info, warn, error)
- [ ] Progress bar for large index loading

## Testing

Tested scenarios:
- [x] Basic search with default logging
- [x] Search with `--verbose=false`
- [x] JSON output (logging to stderr, results to stdout)
- [x] Large index (progress logging every 1000 vectors)
- [x] Failed searches (no results logged correctly)
- [x] Ollama connection failures (clear error messages)

## Summary

**Problem:** Users couldn't see search activity, making it unclear if the tool was working or stuck.

**Solution:** Added comprehensive logging with emoji indicators, timing information, and a `--verbose` flag.

**Result:** Users now have full visibility into search operations while maintaining clean output for scripting.

---

**Status:** ✅ Complete - Build successful, documentation updated, ready to use
