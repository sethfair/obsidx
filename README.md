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
              │  HNSW Index (fast ANN) │
              │  • In-memory           │
              │  • Rebuildable         │
              │  • <1ms search         │
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

- **SQLite is authoritative**: HNSW index is derived and rebuildable
- **Soft deletes**: File changes mark old chunks inactive, not deleted
- **Metadata inheritance**: All chunks inherit note-level category/scope/status
- **Two-stage recall**: HNSW → candidates, then exact cosine with category weights

## Quick Start

**One command to run everything:**

```bash
git clone https://github.com/sethfair/obsidx
cd obsidx
./build.sh
./run.sh ~/MyObsidianVault
```

This automatically:
- Starts Ollama if not running
- Downloads the embedding model if needed
- Begins indexing your vault in watch mode

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
- `bin/obsidx-rebuild` - rebuilds HNSW from SQLite

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

### 3. Search

```bash
# Standard search (excludes archive, weights by category)
./bin/obsidx-recall "how do we handle authentication"

# Canon-only (authoritative decisions only)
./bin/obsidx-recall --canon-only "deployment process"

# Include archive
./bin/obsidx-recall --exclude-archive=false "old architecture decisions"

# Filter by categories
./bin/obsidx-recall --category "canon,project" "error handling strategy"

# JSON output (for tooling)
./bin/obsidx-recall --json "api design principles" | jq
```

## Metadata System

Add front matter to your notes:

```yaml
---
category: canon        # canon | project | workbench | archive
scope: writerflow      # your domain/project
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
Scope: writerflow
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

## Knowledge Governance

### Promotion Workflow

```
workbench → project → canon → archive
```

1. **Draft in workbench** (`category: workbench`)
   - Speculative, incomplete, messy

2. **Refine in project** (`category: project`)
   - Structured, evolving, under active development

3. **Promote to canon** (`category: canon, status: active`)
   - Stable, authoritative, enforced
   - Requires `last_reviewed` date
   - Becomes law for AI agents

4. **Archive when superseded** (`category: archive, status: deprecated`)
   - Historical reference only
   - Excluded from default retrieval

### ADR Pattern

Create architectural decisions as canon:

```yaml
---
category: canon
scope: writerflow
type: decision
status: active
last_reviewed: 2026-01-20
---

# ADR-001 — Use HNSW for Semantic Search

## Context
We need fast semantic search over 100k+ note chunks...

## Decision
Use in-memory HNSW with SQLite as source of truth...

## Alternatives Considered
1. Pinecone (external dependency)
2. PostgreSQL pgvector (slower)
3. Pure vector similarity (no structure)

## Consequences
...
```

All ADRs live in `@canon/decisions/`.

## Advanced Configuration

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

### HNSW Parameters

Edit `internal/ann/hnsw.go`:

```go
func DefaultHNSWConfig(dim int) HNSWConfig {
    return HNSWConfig{
        M:              32,   // More connections (slower build, better recall)
        EfConstruction: 256,  // Higher quality (slower indexing)
        EfSearch:       128,  // More candidates (slower search, better recall)
    }
}
```

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

Configure Copilot to use obsidx as your knowledge source:

**Quick Setup:**
1. Copy `.github/copilot-instructions.md` to your project root
2. Copilot will now search your knowledge base before answering

**Key Command:**
```bash
~/code/obsidx/bin/obsidx-recall --json "your query" | head -c 2000
```

**Full Setup Guide:** See [docs/COPILOT_SETUP.md](docs/COPILOT_SETUP.md)

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

- **[Category System Guide](docs/CATEGORY_SYSTEM.md)** - Full metadata reference and retrieval behavior
- **[Setup Guide](docs/setup.md)** - Knowledge governance workflow and best practices
- Architecture decisions (see ADRs in your vault)

## FAQ

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
File change triggers reindex. Old chunks marked inactive, new chunks inserted. HNSW index is append-only until rebuild.

**Q: How big can my vault be?**  
Tested with 100k chunks. Brute-force search is ~1ms. For larger vaults (>500k), consider proper HNSW integration.

## Contributing

1. Fork the repo
2. Create feature branch
3. Add tests
4. Update docs
5. Submit PR

## License

MIT

---

**Built with:** Go, SQLite, HNSW, Ollama (optional)

**Philosophy:** Your knowledge base should reflect reality: some things are canon, some are drafts, some are experiments. The retrieval system should honor that hierarchy.
