# obsidx

**A metadata-aware semantic search engine for Obsidian vaults** with tiered retrieval, canon enforcement, and agent-ready knowledge governance.

## Why obsidx?

Traditional search finds keywords. obsidx **understands your knowledge lifecycle**:

- 📚 **Canon notes** are authoritative truth (boosted 20%)
- 🔨 **Project notes** are active work (boosted 5%)
- 🧪 **Workbench notes** are drafts/experiments (reduced 10%)
- 📦 **Archive notes** are historical context (reduced 40%, excluded by default)

This prevents AI agents from latching onto old drafts instead of established decisions.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│  Obsidian Vault                                             │
│  ├── @canon/          (authoritative, stable)               │
│  ├── @projects/       (active work)                         │
│  ├── @workbench/      (drafts, experiments)                 │
│  └── archive/         (historical reference)                │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
              ┌────────────────────────┐
              │  obsidx-indexer        │
              │  • Parse front matter  │
              │  • Extract metadata    │
              │  • Chunk markdown      │
              │  • Generate embeddings │
              └────────────────────────┘
                           │
                           ▼
              ┌────────────────────────┐
              │  SQLite (source truth) │
              │  • Chunks + vectors    │
              │  • Category metadata   │
              │  • Active/inactive     │
              └────────────────────────┘
                           │
                           ▼
              ┌────────────────────────┐
              │  Exact Search Index    │
              │  • In-memory, flat     │
              │  • Cosine similarity   │
              │  • Parallel full scan  │
              │  • Rebuilt at startup  │
              └────────────────────────┘
                           │
                           ▼
              ┌────────────────────────┐
              │  obsidx-recall         │
              │  • Two-stage retrieval │
              │  • Category weighting  │
              │  • Exact rerank        │
              └────────────────────────┘
```

**Key Design Decisions:**

- **SQLite is authoritative**: the in-memory search index is derived and rebuilt at startup
- **Soft deletes**: File changes mark old chunks inactive, not deleted
- **Metadata inheritance**: All chunks inherit note-level category/scope/status
- **Two-stage recall**: exact scan → candidates, then rerank with category weights

### Search Index Details

**Implementation:** exact brute-force cosine search (`internal/ann/brute.go`) — flat storage of L2-normalized vectors, parallel striped scan with per-worker top-k heaps, deterministic tie-breaking (similarity desc, chunk id asc).

**Key Features:**
- **Exact by construction:** recall is always 100% — no graph pathologies possible
- **Cosine Similarity:** vectors normalized once at insert; search is a pure dot product
- **Thread-Safe:** read-write locks; the indexer can Add while the server Searches
- **Zero-norm rejection:** unembeddable vectors are rejected at Add and query time

**Performance (measured 2026-07-23, ~80k chunks × 768 dims):** search stage 11–14 ms, total query ~45–60 ms including query embedding — within the <100 ms budget. Complexity is O(N × dim) per query; at ~10× current vault size revisit ANN (with the duplicate-vector caution from ADR-002 below).

> **History:** obsidx used [coder/hnsw](https://github.com/coder/hnsw) approximate search until 2026-07-23, when it was replaced after showing near-zero recall on real vault embeddings. See ADR-002 below.

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
- Starts the search server with the in-memory search index (background)

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

# Search server activity
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

This creates:
- `bin/obsidx-indexer` - watches vault and indexes changes
- `bin/obsidx-recall` - semantic search with category awareness
- `bin/obsidx-rebuild` - validates that all stored embeddings decode and load into a search index

### 2. Index Your Vault

```bash
# Auto-detect embedder (tries Ollama, falls back to local)
./bin/obsidx-indexer --vault ~/notes --watch

# Or force a specific embedder
./bin/obsidx-indexer --vault ~/notes --embed-mode ollama --embed-model nomic-embed-text
```

The indexer:
- Watches for file changes (debounced)
- Parses YAML front matter
- Infers category from folder structure if no metadata
- Generates embeddings (Ollama, local TF-IDF, or HTTP)
- Stores in SQLite with full metadata

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

### 3. Search

The search server keeps the search index loaded in memory for instant searches.

```bash
# Standard search (fast: <100ms)
./bin/obsidx-recall "how do we handle authentication"

# Canon-only (authoritative decisions only)
./bin/obsidx-recall --canon-only "deployment process"

# Include archive
./bin/obsidx-recall --exclude-archive=false "old architecture decisions"

# Filter by categories
./bin/obsidx-recall --category "canon,project" "error handling strategy"

# JSON output (for tooling)
./bin/obsidx-recall --json "api design principles" | jq

# Quiet mode (disable timing info for scripting)
./bin/obsidx-recall --verbose=false "query"
```

**Performance:**
- First time: Server auto-starts if not running (takes ~5 seconds to load index)
- Subsequent searches: <100ms (index stays in memory)
- No index rebuild on every search!

**Search Activity:**

By default, obsidx-recall shows:
- Query being searched
- Timing breakdown (embed, search, fetch, rerank)
- Total time

Use `--verbose=false` to disable timing output for scripting.

## Metadata System

Add front matter to your notes:

```yaml
---
category: canon        # canon | project | workbench | archive
scope: mycompany      # your domain/project
type: decision         # decision | principle | vision | spec | note | log
status: active         # active | draft | superseded | deprecated
last_reviewed: 2026-01-20
---
```

### Category Hierarchy

| Category | Meaning | Weight | Use Case |
|----------|---------|--------|----------|
| **canon** | Authoritative, stable truth | **1.20x** | ADRs, core principles, architectural invariants |
| **project** | Active work, evolving | **1.05x** | Current specs, project docs, evolving designs |
| **workbench** | Drafts, experiments | **0.90x** | Brainstorming, sketches, "thinking out loud" |
| **archive** | Historical context | **0.60x** | Deprecated decisions, old projects |

### Status Values

| Status | Meaning | Default Filter |
|--------|---------|----------------|
| `active` | Current, in-use | ✅ Included |
| `draft` | Work in progress | ✅ Included (0.90x weight) |
| `superseded` | Replaced by newer | ❌ Excluded |
| `deprecated` | No longer relevant | ❌ Excluded |

### Folder Inference

If a note lacks front matter, category is inferred:
- `/canon/` or `/@canon/` → `canon`
- `/projects/` → `project`
- `/workbench/` or `/drafts/` → `workbench`
- `/archive/` → `archive`
- Otherwise: `project` (default)

**Precedence:** Front matter > folder inference > default

## Embeddings

### Auto-detection (Recommended)

```bash
./bin/obsidx-indexer --vault ~/notes --embed-mode auto
```

Tries in order:
1. **Ollama** at `localhost:11434` (if available)
2. **Local TF-IDF** (fallback, no external dependencies)

### Ollama (Best Quality)

```bash
# Install Ollama: https://ollama.ai
ollama pull nomic-embed-text

# Then index
./bin/obsidx-indexer --vault ~/notes --embed-mode ollama --embed-model nomic-embed-text
```

Recommended models:
- `nomic-embed-text` (768 dim, best balance)
- `all-minilm` (384 dim, faster)

### Local (No Dependencies)

```bash
./bin/obsidx-indexer --vault ~/notes --embed-mode local --dim 384
```

Uses simple TF-IDF hashing. Good for:
- Testing
- Privacy-sensitive vaults
- No network access

**Trade-off:** Lower quality vs. neural embeddings, but still useful.

### Custom HTTP Endpoint

```bash
./bin/obsidx-indexer --vault ~/notes --embed-mode http --http-url http://localhost:8080/embed --dim 768
```

Expects JSON API:
```json
POST /embed
{"text": "your content"}

Response:
{"vector": [0.1, 0.2, ...]}
```

## Usage Examples

### Daily Workflow

```bash
# Morning: What did we decide about X?
./bin/obsidx-recall --canon-only "rate limiting strategy"

# During work: Find related project context
./bin/obsidx-recall --category "project,workbench" "user session management"

# Research: Include everything
./bin/obsidx-recall --exclude-archive=false --category "canon,project,archive" "authentication history"
```

### Agent Integration

```bash
# Before writing code, retrieve canon context
context=$(./bin/obsidx-recall --canon-only --json "deployment" | jq -r '.[].content')

# Then pass to AI agent
echo "$context" | your-agent-tool
```

### Search Results

```
Found 3 results:

─────────────────────────────────────────────────────────────
[1] Score: 0.8745 [📚 CANON]
Path: /canon/decisions/ADR-003-Rate-Limiting.md
Section: Decision > Implementation
Scope: mycompany
Lines: 15-42

We use token bucket rate limiting with Redis backing.
Maximum 100 requests per minute per API key...
─────────────────────────────────────────────────────────────
[2] Score: 0.7621 [🔨 PROJECT]
Path: /projects/api-v2/rate-limit-impl.md
Section: Current Implementation
Lines: 8-25

The rate limiter is implemented in middleware/ratelimit.go...
```

## Tag-Based Weighting & Knowledge Governance

### Customizable Priority System

ObsIDX uses **tag weights** to prioritize content. Configure for your knowledge system:

```bash
# Initialize with defaults (Zettelkasten, PARA, Writerflow)
./bin/obsidx-weights --init

# View current weights
./bin/obsidx-weights

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

**Inline format:**
```yaml
---
tags: #permanent-note #architecture-decision
scope: mycompany
type: decision
status: active
last_reviewed: 2026-01-20
---
```

**Or array format:**
```yaml
---
tags: [permanent-note, architecture-decision]
scope: mycompany
type: decision
status: active
last_reviewed: 2026-01-20
---
```
# ADR-001 — Use HNSW for Semantic Search  [SUPERSEDED by ADR-002, 2026-07-23]

## Context
We need fast semantic search over 100k+ note chunks...

## Decision
Use in-memory HNSW with SQLite as source of truth...

## Alternatives Considered
1. Pinecone (external dependency)
2. PostgreSQL pgvector (slower)
3. Pure vector similarity (no structure)

## Consequences
Superseded: see ADR-002.

# ADR-002 — Exact Brute-Force Search Replaces HNSW (2026-07-23)

## Context
Two compounding failures surfaced on the production vault index:
1. nomic-embed-text (via Ollama, both embed endpoints) collapses
   heading-only lines to identical vectors — one constant vector per
   heading level (`# A B` ≡ `# C D`; `## A B` ≡ `## C D`; …). ~5,765 such
   chunks formed identical-vector clusters (largest: 1,063 nodes).
2. coder/hnsw v0.6.1 (latest upstream) returned near-zero recall on the
   vault's embeddings: identical-vector cliques trapped greedy search, and
   even after the clusters were purged, brute-force ground truth put the
   correct match (cosine 0.837) at rank 2 while the graph never returned
   it at any candidate depth.

## Decision
- Skip heading-only chunks at embed time (`chunker.IsHeadingOnly`).
- Replace HNSW with exact brute-force search (`ann.BruteForce`): at ~80k
  chunks × 768 dims a parallel exact scan takes 11–14 ms — correct by
  construction, well inside the <100 ms budget.

## Remediation note (pre-fix databases)
Databases indexed before this change still carry active heading-only
chunks. The live DB was purged in place on 2026-07-23 (5,765 chunks
marked inactive, matching the `IsHeadingOnly` predicate). A DB restored
from an earlier backup must either be re-purged the same way or fully
reindexed from scratch — otherwise recall silently regresses.

## Consequences
- 100% recall, deterministic results; no ANN tuning surface.
- O(N × dim) per query; revisit ANN only if the vault grows ~10×, and
  only with the duplicate-vector hazard above in mind.
```

All ADRs should use `tags: [permanent-note, architecture-decision]` for high search priority.

## Advanced Configuration

### Search Index Management

The search index is **in-memory only** — it is rebuilt from SQLite every
time a binary starts (a few seconds at current vault size). There is no
on-disk index file and nothing to persist or invalidate.

```bash
# Validate that every stored embedding decodes and loads:
./bin/obsidx-rebuild --db .obsidian-index/obsidx.db --dim 768
```

`obsidx-rebuild` streams all active embeddings into a fresh index and
reports progress; its index dies with the process. Use it as a data
integrity check after bulk imports or suspected corruption.

### Tune Retrieval Weights

Edit `internal/metadata/metadata.go`:

```go
func (m *NoteMetadata) CategoryWeight() float32 {
    switch m.EffectiveCategory() {
    case "canon":
        return 1.30  // More aggressive canon boost
    case "workbench":
        return 0.75  // Stronger draft penalty
    // ...
    }
}
```

### Search Tuning

There are no ANN parameters to tune — search is exact. The only knobs are
`top_n` (results returned, default 12) and `candidate_k` (candidates
fetched for reranking, default 200) on the `/search` API, and the category
weights below.

### Custom Categories

Add to `internal/metadata/metadata.go`:

```go
func normalizeCategory(cat string) string {
    switch cat {
    case "reference":
        return "reference"  // New category
    // ...
    }
}
```

Then update schema and weights.

## Database Location & Storage

**Where is the database?**

obsidx uses **SQLite** - a local, embedded database with no server process. The database is stored in:

```
.obsidian-index/obsidx.db
```

This file is created in your **current working directory** when you run the indexer.

**Example:**
```bash
# If you run from your home directory:
cd ~
./code/obsidx/start-daemon.sh ~/notes

# Database is created at:
~/.obsidian-index/obsidx.db
```

**Structure:**
```
.obsidian-index/
└── obsidx.db           # SQLite database (all chunks + embeddings + metadata)
```

**Key Points:**
- ✅ **No server** - SQLite is embedded in the binaries (no daemon, no service)
- ✅ **Local storage** - All data stays on your machine
- ✅ **Portable** - Copy `.obsidian-index/` folder to move your index
- ✅ **Single writer** - Only run one indexer per vault at a time
- ✅ **Multiple readers** - Run as many searches as you want concurrently

**Custom Location:**

Use the `--db` flag to specify a different location:

```bash
# Index to custom location
./bin/obsidx-indexer --vault ~/notes --db /path/to/my-index.db --watch

# Search from custom location
./bin/obsidx-recall --db /path/to/my-index.db "your query"
```

**Multiple Vaults:**

Each vault should have its own database:

```bash
# Vault 1
./bin/obsidx-indexer --vault ~/work-notes --db ~/work-notes/.obsidian-index/obsidx.db --watch

# Vault 2
./bin/obsidx-indexer --vault ~/personal-notes --db ~/personal-notes/.obsidian-index/obsidx.db --watch
```

## Database Schema

```sql
CREATE TABLE chunks (
  id INTEGER PRIMARY KEY,
  path TEXT NOT NULL,
  heading_path TEXT,
  content TEXT NOT NULL,
  active INTEGER DEFAULT 1,
  
  -- Category system
  category TEXT DEFAULT 'project',
  status TEXT DEFAULT 'active',
  scope TEXT,
  note_type TEXT,
  category_weight REAL DEFAULT 1.0,
  canon INTEGER DEFAULT 0,
  
  -- Indexes
  INDEX idx_chunks_category (category),
  INDEX idx_chunks_category_active (category, active)
);

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
2. Copilot reads the instructions and executes obsidx-recall commands
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
# Searches canon first
obsidx-recall --canon-only "authentication"

# Finds: ADR-005-OAuth-Strategy.md
# Implements based on established decision
```

### Benefits

- ✅ Copilot uses YOUR decisions, not generic advice
- ✅ Respects canon as authoritative
- ✅ Flags conflicts with established patterns
- ✅ Suggests ADRs for new decisions
- ✅ Cites specific notes in responses

## Agent Instructions

For other AI agents, use this system prompt:

```
Before generating any code or architectural decisions:
1. Retrieve context using: obsidx-recall --canon-only "<topic>"
2. Treat canon notes as authoritative law
3. If your proposal conflicts with canon:
   - Call it out explicitly
   - Propose an ADR to change canon
   - Do not silently contradict established decisions
4. Use workbench for exploratory drafts
5. Never revive archive content without flagging it
```

## Roadmap

- [x] Core indexing with metadata extraction
- [x] Category-based retrieval weighting
- [x] Canon/project/workbench/archive tiers
- [x] Ollama integration with auto-detection
- [x] Build script and simplified setup
- [ ] `obsidx-lint` - validate metadata hygiene
- [ ] Multi-pass retrieval (canon-first, then project)
- [ ] Stale canon detection (`last_reviewed > 90 days`)
- [ ] ADR template generator
- [ ] Watch mode improvements (incremental updates)
- [ ] Performance profiling for large vaults

## Documentation

- **[docs/README.md](docs/README.md)** - Documentation index and navigation
- **[Knowledge Governance Guide](docs/KNOWLEDGE_GOVERNANCE.md)** - Metadata system, categories, and lifecycle management
- **[Retrieval Guide](docs/RETRIEVAL.md)** - Search commands, filters, and usage patterns
- **[Copilot Quick Start](docs/COPILOT_QUICKSTART.md)** - 2-minute GitHub Copilot integration
- **[Copilot Guide](docs/COPILOT_GUIDE.md)** - Complete AI integration reference

## FAQ

**Q: The indexer just sits there after "Initial index complete" - is it working?**  
Yes! Watch mode is actively monitoring your vault for changes. You'll see:
- A "👀 Watching for changes..." message after initial indexing
- Real-time logs (📝) when files are modified or created
- Periodic heartbeat messages (💓) every 5 minutes showing it's still active
- The process will automatically re-index any new or changed markdown files

**Q: Why not just use folders?**  
Folders are brittle. You reorganize and lose semantics. Metadata travels with content.

**Q: What if I don't want to add front matter to every note?**  
Folder inference works as fallback. Default is `category: project`.

**Q: Can I use this without Ollama?**  
Yes! Use `--embed-mode local` for TF-IDF (no external deps) or `--embed-mode http` for custom embedders.

**Q: How do I mark something as authoritative?**
```yaml
category: canon
status: active
```
That's it. It gets 20% boost in retrieval.

**Q: What happens when I edit a canon note?**  
File change triggers reindex. Old chunks marked inactive, new chunks inserted. The in-memory index picks up additions incrementally; a restart rebuilds it from SQLite.

**Q: How big can my vault be?**  
The exact scan handles the current ~80k chunks in 11-14 ms per query (measured 2026-07-23) and scales linearly, so ~10x the vault size stays comfortably interactive. Beyond that, revisit approximate search — with the duplicate-vector hazard documented in ADR-002 in mind.

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
- Go 1.21+
- SQLite (via mattn/go-sqlite3)
- Ollama (for embeddings)

**Philosophy:** Your knowledge base should reflect reality: some things are canon, some are drafts, some are experiments. The retrieval system should honor that hierarchy.
