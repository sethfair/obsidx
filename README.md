# obsidx

**A metadata-aware semantic search engine for Obsidian vaults** with tag-weighted retrieval, scope filtering, and agent-ready knowledge governance.

> 🔍 **Search from inside Obsidian** with the companion plugin — [**obsidx-obsidian**](https://github.com/sethfair/obsidx-obsidian): an Omnisearch-style modal backed by the obsidx recall server.

## Why obsidx?

Traditional search finds keywords. obsidx **understands your knowledge lifecycle** by weighting results from the tags and metadata already in your notes:

- 🧠 **Permanent notes** are refined, authoritative insights (boosted)
- 📖 **Literature notes** are structured, paraphrased sources (slightly boosted)
- 💭 **Fleeting notes** are quick, unrefined captures (reduced)
- 🗂️ **Superseded / deprecated** notes are down-weighted so fresher thinking ranks first

This prevents AI agents from latching onto old drafts instead of your established thinking. Weights are fully configurable (`weights.json`) for whatever tag system you use — Zettelkasten, PARA, or your own.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│  Obsidian Vault (markdown + YAML front matter + tags)        │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
              ┌────────────────────────┐
              │  obsidx-indexer        │
              │  • Parse front matter  │
              │  • Extract tags/scope  │
              │  • Chunk markdown      │
              │  • Generate embeddings │
              └────────────────────────┘
                           │
                           ▼
              ┌────────────────────────┐
              │  SQLite (source truth) │
              │  • Chunks + vectors    │
              │  • Tag/scope/status    │
              │  • Active/inactive     │
              └────────────────────────┘
                           │
                           ▼
              ┌────────────────────────┐
              │  HNSW Index (fast ANN) │
              │  • In-memory graph     │
              │  • Cosine distance     │
              │  • Persistable         │
              │  • Rebuildable         │
              └────────────────────────┘
                           │
                           ▼
              ┌────────────────────────┐
              │  obsidx-recall-server  │
              │  • HTTP API (:8765)    │
              │  • Two-stage retrieval │
              │  • Tag weighting       │
              │  • Exact rerank        │
              └────────────────────────┘
                     │            │
                     ▼            ▼
          obsidx-recall (CLI)   obsidx-obsidian (plugin)
```

**Key Design Decisions:**

- **SQLite is authoritative**: HNSW index is derived and rebuildable
- **Soft deletes**: File changes mark old chunks inactive, not deleted
- **Metadata inheritance**: All chunks inherit note-level tags/scope/status
- **Tag-weighted recall**: HNSW → candidates, then exact cosine scaled by a tag-derived weight
- **Server-backed search**: `obsidx-recall-server` holds the HNSW index in memory; the CLI and the Obsidian plugin are both thin HTTP clients

### HNSW Technical Details

**Implementation:** Uses the [coder/hnsw](https://github.com/coder/hnsw) library for approximate nearest neighbor search.

**Key Features:**
- **Cosine Distance:** Measures semantic similarity via vector angles (1 - cosine_similarity)
- **Thread-Safe:** Read-write locks protect concurrent access
- **Incremental Updates:** Add vectors without full rebuild
- **Persistent:** Save/load index to disk for fast startup
- **Memory Efficient:** Hierarchical graph structure with configurable connectivity

**Search Algorithm:**
1. Query vector enters at top layer
2. Greedy search finds closest neighbors at each layer
3. Descends through layers refining candidates
4. Returns top-k results from base layer
5. Results are re-ranked with exact cosine + tag weights

**Performance Characteristics:**
- **Build Time:** O(N × log(N) × M × EfConstruction)
- **Search Time:** O(log(N) × EfSearch)
- **Memory:** O(N × M × layers)
- **Accuracy:** ~95%+ recall@10 with default params

## Quick Start

**One command to run everything:**

```bash
git clone https://github.com/sethfair/obsidx
cd obsidx
./build.sh
./start-daemon.sh ~/MyObsidianVault  # Runs in background
```

This automatically:
- Starts Ollama if not running
- Downloads the embedding model if needed
- Starts the indexer in watch mode (background)
- Starts the recall server with a persistent HNSW index (background, port 8765)

**Search (instant, <100ms):**

```bash
./bin/obsidx-recall "your query"
```

**Stop everything:**

```bash
./stop-daemon.sh
```

**Watch logs:**

```bash
# Indexer activity
tail -f .obsidian-index/indexer.log

# Recall server activity
tail -f .obsidian-index/recall-server.log
```

**Interactive mode (foreground with live logs):**

```bash
./interactive.sh ~/MyObsidianVault
```

This runs the indexer in the background and gives you an interactive search prompt:
- Type queries to search your vault
- Type `logs` to see recent indexer activity
- Type `quit` to exit
- Shows both indexer and search logging


### Manual Setup

If you prefer more control:

### 1. Install

```bash
git clone https://github.com/sethfair/obsidx
cd obsidx
go build -o bin/ ./cmd/...
```

This creates all five binaries:
- `bin/obsidx-indexer` - watches the vault and indexes changes into SQLite + HNSW
- `bin/obsidx-recall-server` - long-lived HTTP search server (used by the CLI and the Obsidian plugin)
- `bin/obsidx-recall` - command-line search client (talks to the recall server)
- `bin/obsidx-rebuild` - rebuilds the HNSW index from SQLite
- `bin/obsidx-weights` - manages the tag-weight configuration

### 2. Index Your Vault

```bash
# Index once
./bin/obsidx-indexer --vault ~/notes

# Index and keep watching for changes
./bin/obsidx-indexer --vault ~/notes --watch
```

The indexer:
- Watches for file changes (debounced, `--debounce` ms)
- Parses YAML front matter (scope, status, type, tags)
- Computes a per-chunk weight from the note's tags + status
- Generates embeddings via Ollama (`--ollama-url`, `--model`)
- Stores chunks, vectors, and metadata in SQLite

**Common flags:** `--vault` (required), `--db`, `--index`, `--weights`, `--ollama-url`, `--model` (default `nomic-embed-text`), `--watch`, `--debounce`.

**Watch Mode Behavior:**
- Performs initial full index of all markdown files
- Monitors vault directory recursively for changes
- Automatically re-indexes when files are created, modified, or moved
- Shows activity log with emoji indicators:
  - 📝 File change detected
  - ✓ Successfully re-indexed
  - ❌ Error occurred
  - 💓 Periodic heartbeat (every 5 minutes) showing it's still active
- Debounces rapid changes (500ms default) to avoid thrashing
- Press Ctrl+C to gracefully shutdown

### 3. Run the recall server

Search is served by `obsidx-recall-server`, which loads the HNSW index into memory and exposes an HTTP API:

```bash
./bin/obsidx-recall-server --db .obsidian-index/obsidx.db --port 8765
```

**Flags:** `--db` (default `.obsidian-index/obsidx.db`), `--port` (default `8765`), `--ollama-url`, `--model`.

> The `start-daemon.sh` script starts this for you. Keep it running while you search from the CLI or the Obsidian plugin.

### 4. Search

```bash
# Standard search (fast: <100ms once the server is warm)
./bin/obsidx-recall "how do we handle authentication"

# More results
./bin/obsidx-recall --top 25 "error handling strategy"

# Point at a non-default server
./bin/obsidx-recall --server http://localhost:8765 "deployment process"

# JSON output (for tooling)
./bin/obsidx-recall --json "api design principles" | jq

# Quiet mode (disable timing info for scripting)
./bin/obsidx-recall --verbose=false "query"
```

**`obsidx-recall` flags:** `--server` (default `http://localhost:8765`), `--top` (default 12), `--candidates` (default 200), `--json`, `--verbose` (default true). The CLI is a thin client — it does **not** open the database directly; it calls the recall server.

**Search Activity:**

By default, `obsidx-recall` shows the query and a timing breakdown (embed, search, fetch, rerank, total). Use `--verbose=false` to disable timing output for scripting.

**Example output:**

```
Found 3 results:

─────────────────────────────────────────────────────────────
[1] Score: 0.8745 [permanent-note, architecture-decision]
Path: /decisions/ADR-003-Rate-Limiting.md
Section: Decision > Implementation
Scope: mycompany
Status: active
Lines: 15-42

We use token bucket rate limiting with Redis backing.
Maximum 100 requests per minute per API key...
─────────────────────────────────────────────────────────────
[2] Score: 0.7621 [literature-note]
Path: /sources/rate-limit-patterns.md
Section: Token Bucket
Lines: 8-25

The rate limiter is implemented in middleware/ratelimit.go...
```

## Recall Server & HTTP API

`obsidx-recall-server` exposes a small JSON HTTP API on port `8765` (configurable). This is the integration surface for `obsidx-recall`, the Obsidian plugin, and any custom tooling.

| Method | Path | Purpose |
|--------|------|---------|
| `POST` | `/search` | Run a semantic search |
| `GET` | `/health` | Liveness check |
| `GET` | `/stats` | Index stats (chunk counts, dimension, model) |

**Search request:**

```bash
curl -s http://localhost:8765/search \
  -H 'Content-Type: application/json' \
  -d '{"query": "rate limiting", "top_n": 10, "candidate_k": 200}'
```

```json
POST /search
{ "query": "rate limiting", "top_n": 10, "candidate_k": 200 }
```

**Search response:**

```json
{
  "results": [
    {
      "score": 0.8745,
      "path": "/decisions/ADR-003-Rate-Limiting.md",
      "heading_path": "Decision > Implementation",
      "status": "active",
      "scope": "mycompany",
      "start_line": 15,
      "end_line": 42,
      "content": "We use token bucket rate limiting...",
      "category_weight": 1.3,
      "tags": ["permanent-note", "architecture-decision"]
    }
  ],
  "timing": { "embed_ms": 8, "search_ms": 2, "fetch_ms": 1, "rerank_ms": 1, "total_ms": 12 }
}
```

> `category_weight` is the tag-derived multiplier applied during reranking (see [Tag-Based Weighting](#tag-based-weighting--knowledge-governance)). Despite the name, it is computed from tags + status, not a folder category.

## Obsidian Plugin

[**obsidx-obsidian**](https://github.com/sethfair/obsidx-obsidian) brings an Omnisearch-style search modal to Obsidian, backed by the recall server. Type a natural-language query and get semantically ranked notes — with heading paths, scope/status/tags, and excerpts — without leaving the app.

**Setup:**
1. Index your vault and start the recall server (steps 2–3 above).
2. Install the plugin (Community plugins, or [BRAT](https://github.com/TfTHacker/obsidian42-brat) with `sethfair/obsidx-obsidian`).
3. Point it at your server URL in settings (default `http://localhost:8765`) and run **Obsidx Semantic Search**.

The plugin calls `POST /search` and degrades gracefully (with a clear message) when the server isn't running.

## Metadata System

Add front matter to your notes. obsidx reads `scope`, `status`, `type`, and `tags`:

```yaml
---
tags: [permanent-note, architecture-decision]
scope: mycompany       # your domain/project
type: decision         # decision | principle | vision | spec | note | log
status: active         # active | draft | superseded | deprecated
last_reviewed: 2026-01-20
---
```

### Status Values

| Status | Meaning | Effect on ranking |
|--------|---------|-------------------|
| `active` | Current, in-use | Full weight (1.0×) |
| `draft` | Work in progress | Slightly reduced (0.9×) |
| `superseded` | Replaced by newer | Down-weighted (0.5×) |
| `deprecated` | No longer relevant | Down-weighted (0.5×) |

> Status affects **weight**, not inclusion — superseded/deprecated notes still appear in results, just ranked lower. (Status weights are configurable in `weights.json`.)

### Scope

`scope` is a free-form domain label (e.g. `mycompany`, `personal`, a project name). It travels with every chunk and is returned in results, so you can tell at a glance which domain a hit belongs to.

Retrieval **priority** comes from tags, not folders — see the next section.

## Tag-Based Weighting & Knowledge Governance

### Customizable Priority System

obsidx uses **tag weights** to prioritize content. Configure for your knowledge system:

```bash
# Initialize the config file with defaults (Zettelkasten, PARA, Writerflow)
./bin/obsidx-weights --init

# Show the built-in default weights
./bin/obsidx-weights --defaults

# View / edit the active config
./bin/obsidx-weights            # prints current config
# Edit: .obsidian-index/weights.json
```

### Knowledge Maturity Workflow

```
fleeting → literature → permanent → archive
```

1. **Capture** (`tags: [fleeting-notes]`, weight: 0.8)
   - Quick captures, unrefined ideas

2. **Curate** (`tags: [literature-note]`, weight: 1.1)
   - Paraphrased sources, structured notes

3. **Synthesize** (`tags: [permanent-note]`, weight: 1.3)
   - Refined insights, authoritative
   - Highest priority in search results

4. **Archive** (`tags: [archive]` or `status: deprecated`, weight: 0.6)
   - Historical reference only

### Tag Examples

Both formats work - use inline hashtags OR array format:

**Zettelkasten**:
```yaml
tags: #permanent-note #moc  # Inline (Obsidian-style)
# OR
tags: [permanent-note, moc]  # Array format
```

**PARA**:
```yaml
tags: #writerflow #mvp-1 #product-development  # Inline
# OR
tags: [writerflow, mvp-1, product-development]  # Array
```

**Domain Knowledge**:
```yaml
tags: #customer-research #validation #icp  # Inline
# OR
tags: [customer-research, validation, icp]  # Array
```

See `docs/TAG-WEIGHTING.md` and `docs/TAG-FORMAT.md` for complete documentation.

### ADR Pattern

Create architectural decisions with authoritative tags:

```yaml
---
tags: [permanent-note, architecture-decision]
scope: mycompany
type: decision
status: active
last_reviewed: 2026-01-20
---
# ADR-001 — Use HNSW for Semantic Search

## Context
We need fast semantic search over 100k+ note chunks...

## Decision
Use in-memory HNSW with SQLite as source of truth...

## Consequences
...
```

Tagging ADRs `permanent-note` gives them the highest retrieval weight, so agents and searches surface your decisions first.

## Embeddings

obsidx generates embeddings with **[Ollama](https://ollama.com)**.

```bash
# Install Ollama, then pull an embedding model
ollama pull nomic-embed-text

# Index using that model
./bin/obsidx-indexer --vault ~/notes --model nomic-embed-text --ollama-url http://localhost:11434
```

Recommended models:
- `nomic-embed-text` (768 dim, best balance — default)
- `all-minilm` (384 dim, faster)

The embedding model and dimension are recorded in the index. If you change models, rebuild the HNSW index (`obsidx-rebuild`) so dimensions stay consistent.

## Usage Examples

### Daily Workflow

```bash
# What did we decide about X?
./bin/obsidx-recall "rate limiting strategy"

# More breadth
./bin/obsidx-recall --top 25 "user session management"
```

Notes tagged `permanent-note` / `literature-note` naturally rank above fleeting drafts; superseded and deprecated notes are excluded by default.

### Agent Integration

```bash
# Before writing code, retrieve high-signal context as JSON
context=$(./bin/obsidx-recall --json "deployment" | jq -r '.results[].content')

# Then pass to your AI agent
echo "$context" | your-agent-tool
```

## Advanced Configuration

### HNSW Index Management

**Persistence:**
The HNSW index is saved to and loaded from disk for faster startup:

```bash
# The index is automatically saved/loaded from:
.obsidian-index/hnsw_<model>_<dim>.bin

# Force rebuild from SQLite (e.g., after changing models or params)
./bin/obsidx-rebuild --db .obsidian-index/obsidx.db --dim 768
```

**When to rebuild:**
- After changing HNSW parameters (M, EfConstruction, EfSearch)
- After changing the embedding model or dimension
- After bulk imports or major vault changes
- To optimize after many incremental updates

**Index lifecycle:**
1. `obsidx-indexer` builds HNSW incrementally during watch mode
2. Index is persisted to `.bin` on shutdown
3. On startup, the recall server loads the existing index (if compatible) or rebuilds
4. Use `obsidx-rebuild` to force full reconstruction from SQLite

### Tune Retrieval Weights

Weights are data, not code — edit `.obsidian-index/weights.json` (or run `obsidx-weights --init` to scaffold it):

```json
{
  "tag_weights": [
    { "tag": "permanent-note", "weight": 1.3 },
    { "tag": "literature-note", "weight": 1.1 },
    { "tag": "fleeting-notes", "weight": 0.8 }
  ],
  "default_weight": 1.0,
  "multiply_tag_weights": false
}
```

By default a chunk takes the **max** weight among its tags; set `multiply_tag_weights: true` to compound them. The resolved value becomes each chunk's `category_weight`, applied during exact reranking. The weighting logic lives in `internal/config/config.go`.

### HNSW Parameters

The HNSW (Hierarchical Navigable Small World) index uses cosine distance for similarity and can be tuned for performance vs. accuracy trade-offs.

**Default Configuration** (`internal/ann/hnsw.go`):

```go
func DefaultHNSWConfig(dim int) HNSWConfig {
    return HNSWConfig{
        Dim:            dim,  // Vector dimension (e.g., 768 for nomic-embed-text)
        M:              16,   // Connections per layer (higher = better recall, more memory)
        EfConstruction: 200,  // Build quality (higher = better index, slower build)
        EfSearch:       100,  // Search quality (higher = better recall, slower search)
    }
}
```

**Tuning Guide:**

| Parameter | Lower Value | Higher Value | Default |
|-----------|-------------|--------------|---------|
| `M` | Faster, less memory | Better recall, more memory | 16 |
| `EfConstruction` | Faster indexing | Higher quality index | 200 |
| `EfSearch` | Faster search | Better recall | 100 |

**Common Configurations:**

```go
// Fast, lower quality (small vaults, speed critical)
M: 8, EfConstruction: 100, EfSearch: 50

// Balanced (default - recommended for most use cases)
M: 16, EfConstruction: 200, EfSearch: 100

// High quality (large vaults, accuracy critical)
M: 32, EfConstruction: 256, EfSearch: 128
```

**Distance Metric:**

obsidx uses **cosine distance** (1 - cosine_similarity) for text embeddings:
- Normalized vectors (angles matter, magnitude doesn't)
- Range: 0 (identical) to 2 (opposite)
- Optimal for semantic similarity

## Database Location & Storage

**Where is the database?**

obsidx stores everything in a local **SQLite** database (no external database server):

```
.obsidian-index/obsidx.db
```

This file is created relative to your **current working directory** (or wherever `--db` points).

**Structure:**
```
.obsidian-index/
├── obsidx.db               # SQLite database (chunks + embeddings + metadata)
├── hnsw_<model>_<dim>.bin  # persisted HNSW index (rebuilt from SQLite as needed)
├── weights.json            # tag-weight configuration
├── indexer.log             # indexer daemon log
└── recall-server.log       # recall server daemon log
```

**Key Points:**
- ✅ **Embedded database** - SQLite, no separate DB service to run
- ✅ **One long-lived process for search** - `obsidx-recall-server` keeps the HNSW index hot
- ✅ **Local storage** - All data stays on your machine
- ✅ **Portable** - Copy `.obsidian-index/` to move your index
- ✅ **Single writer** - Only run one indexer per vault at a time
- ✅ **Multiple readers** - Run as many searches as you want concurrently

**Custom Location:**

```bash
# Index to a custom DB location
./bin/obsidx-indexer --vault ~/notes --db /path/to/my-index.db --watch

# Serve from that DB
./bin/obsidx-recall-server --db /path/to/my-index.db
```

**Multiple Vaults:**

Each vault should have its own database (and its own recall server on a distinct port):

```bash
# Vault 1
./bin/obsidx-indexer --vault ~/work-notes --db ~/work-notes/.obsidian-index/obsidx.db --watch
./bin/obsidx-recall-server --db ~/work-notes/.obsidian-index/obsidx.db --port 8765

# Vault 2
./bin/obsidx-indexer --vault ~/personal-notes --db ~/personal-notes/.obsidian-index/obsidx.db --watch
./bin/obsidx-recall-server --db ~/personal-notes/.obsidian-index/obsidx.db --port 8766
```

## Database Schema

```sql
CREATE TABLE chunks (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  path TEXT NOT NULL,
  heading_path TEXT,
  chunk_index INTEGER NOT NULL,
  content TEXT NOT NULL,
  content_sha256 TEXT NOT NULL,
  start_line INTEGER,
  end_line INTEGER,
  active INTEGER NOT NULL DEFAULT 1,
  created_at_unix INTEGER NOT NULL,

  -- Metadata fields
  status TEXT,
  scope TEXT,
  note_type TEXT,
  category_weight REAL DEFAULT 1.0,  -- computed from tags + status
  tags TEXT                          -- JSON array of tags
);

CREATE INDEX idx_chunks_path ON chunks(path);
CREATE INDEX idx_chunks_active ON chunks(active);
CREATE INDEX idx_chunks_status ON chunks(status);

CREATE TABLE embeddings (
  chunk_id INTEGER PRIMARY KEY,
  dim INTEGER NOT NULL,
  vec BLOB NOT NULL,
  FOREIGN KEY(chunk_id) REFERENCES chunks(id)
);
```

## GitHub Copilot Integration

Configure Copilot to use obsidx as your knowledge source through instruction files.

**How it works:**
1. Add `.github/copilot-instructions.md` to your project
2. Copilot reads the instructions and executes `obsidx-recall` commands
3. Results from your vault inform Copilot's answers
4. You get responses based on YOUR documented decisions, not generic advice

**Quick Setup:**

1. Start daemons: `./start-daemon.sh ~/notes` (runs in background)
2. Copy instructions: `cp .github/copilot-instructions.md ~/.github/` (global)
   or `cp .github/copilot-instructions.md /path/to/project/.github/` (per-project)
3. Add to PATH: `echo 'export PATH="$HOME/code/obsidx/bin:$PATH"' >> ~/.zshrc && source ~/.zshrc`
4. Test: Ask Copilot "what is our authentication strategy"

**Full Setup Guide:** See [docs/COPILOT_QUICKSTART.md](docs/COPILOT_QUICKSTART.md)

### Example Workflow

**You:** "Implement authentication"

**Copilot:**
```bash
# Retrieves your high-signal notes first
obsidx-recall "authentication strategy"

# Finds: ADR-005-OAuth-Strategy.md (tagged permanent-note)
# Implements based on your established decision
```

### Benefits

- ✅ Copilot uses YOUR decisions, not generic advice
- ✅ Surfaces high-weight (permanent) notes first
- ✅ Flags conflicts with established patterns
- ✅ Suggests ADRs for new decisions
- ✅ Cites specific notes in responses

## Agent Instructions

For other AI agents, use this system prompt:

```
Before generating any code or architectural decisions:
1. Retrieve context using: obsidx-recall "<topic>"
2. Treat high-weight notes (permanent-note, architecture-decision) as authoritative
3. If your proposal conflicts with an established decision:
   - Call it out explicitly
   - Propose an ADR to change it
   - Do not silently contradict documented decisions
4. Treat fleeting/draft notes as exploratory, not settled
5. Never revive superseded/deprecated content without flagging it
```

## Roadmap

- [x] Core indexing with metadata extraction
- [x] Tag-based retrieval weighting (configurable via `weights.json`)
- [x] Ollama embeddings
- [x] HTTP recall server + CLI client
- [x] Build script and simplified setup
- [x] Obsidian plugin ([obsidx-obsidian](https://github.com/sethfair/obsidx-obsidian))
- [ ] `obsidx-lint` - validate metadata hygiene
- [ ] Stale-note detection (`last_reviewed > 90 days`)
- [ ] ADR template generator
- [ ] Performance profiling for large vaults
- [ ] Automated test suite

## Documentation

- **[docs/README.md](docs/README.md)** - Documentation index and navigation
- **[Knowledge Governance Guide](docs/KNOWLEDGE_GOVERNANCE.md)** - Metadata, tags, and lifecycle management
- **[Retrieval Guide](docs/RETRIEVAL.md)** - Search commands and usage patterns
- **[Copilot Quick Start](docs/COPILOT_QUICKSTART.md)** - 2-minute GitHub Copilot integration
- **[Copilot Guide](docs/COPILOT_GUIDE.md)** - Complete AI integration reference

## FAQ

**Q: The indexer just sits there after "Initial index complete" - is it working?**
Yes! Watch mode is actively monitoring your vault for changes. You'll see:
- A "👀 Watching for changes..." message after initial indexing
- Real-time logs (📝) when files are modified or created
- Periodic heartbeat messages (💓) every 5 minutes showing it's still active
- The process will automatically re-index any new or changed markdown files

**Q: Why weight by tags instead of folders?**
Folders are brittle — you reorganize and lose semantics. Tags and metadata travel with the content, so a note's priority follows it wherever it lives.

**Q: What if I don't tag my notes?**
Untagged notes use `default_weight` (1.0). Tag the notes you most want surfaced (e.g. `permanent-note`) to boost them.

**Q: Can I use this without Ollama?**
Not currently — embeddings are generated via Ollama. Run Ollama locally (`ollama pull nomic-embed-text`) before indexing.

**Q: How do I mark something as authoritative?**
```yaml
tags: [permanent-note]
status: active
```
`permanent-note` carries the highest default weight (1.3×), so it ranks first.

**Q: What happens when I edit a note?**
File change triggers reindex. Old chunks are marked inactive, new chunks inserted. The HNSW index handles additions incrementally.

**Q: How big can my vault be?**
The HNSW index efficiently handles 100k+ chunks with sub-millisecond search times. The cosine distance metric and hierarchical graph structure provide O(log N) search complexity. For vaults over 1M chunks, consider tuning EfSearch for the speed/accuracy trade-off you need.

## Contributing

1. Fork the repo
2. Create feature branch
3. Add tests
4. Update docs
5. Submit PR

## License

MIT

---

**Built with:**
- Go 1.22+
- SQLite (via mattn/go-sqlite3)
- [coder/hnsw](https://github.com/coder/hnsw) - Hierarchical Navigable Small World graphs
- Ollama (for embeddings)

**Philosophy:** Your knowledge base should reflect reality: some notes are refined and authoritative, some are drafts, some are experiments. The retrieval system should honor that hierarchy — driven by the tags and metadata you already keep.
