# Category System & Tiered Retrieval

The obsidx system now supports a sophisticated **category-based tiered retrieval** system that treats notes based on their authority level and lifecycle stage.

## Overview

Notes are classified into categories that determine:
- **Retrieval priority** (canon notes rank higher)
- **Agent behavior** (different edit permissions per category)
- **Filtering** (exclude archive by default)

## Metadata Schema

Add front matter to your markdown notes:

```yaml
---
category: canon     # canon | project | workbench | archive
scope: writerflow   # writerflow | ventio | personal | etc.
type: decision      # decision | principle | vision | spec | note | log
status: active      # active | draft | superseded | deprecated
last_reviewed: 2026-01-20
---
```

### Category Definitions

| Category | Meaning | Retrieval Weight | Use Case |
|----------|---------|------------------|----------|
| `canon` | Curated, stable, authoritative | **1.20x** | Core principles, established decisions |
| `project` | Active work, evolving truth | **1.05x** | Current projects, active specs |
| `workbench` | Drafts, experiments, thinking out loud | **0.90x** | WIP notes, brainstorming |
| `archive` | Historical reference | **0.60x** | Old decisions, deprecated docs |

### Status Definitions

| Status | Meaning | Weight | Filtering |
|--------|---------|--------|-----------|
| `active` | Current, in-use | 1.0x | Always included |
| `draft` | Work in progress | 0.90x | Included by default |
| `superseded` | Replaced by newer version | 0.50x | Excluded by default |
| `deprecated` | No longer relevant | 0.50x | Excluded by default |

## How It Works

### 1. Indexing

When a note is indexed:
1. Front matter is parsed to extract metadata
2. If no explicit `category`, it's inferred from folder structure:
   - `/canon/` or `/@canon/` → `canon`
   - `/archive/` → `archive`
   - `/workbench/` or `/drafts/` → `workbench`
   - `/projects/` → `project`
   - Default: `project`
3. Combined weight is calculated: `category_weight * status_weight`
4. Each chunk inherits the note's metadata

### 2. Retrieval

During search:
1. **HNSW** finds vector-similar candidates
2. **Reranking** applies category weights:
   - Canon notes get 20% boost
   - Workbench notes get 10% penalty
   - Archive notes get 40% penalty
3. Results are sorted by weighted similarity

### 3. Filtering

By default:
- Archive notes excluded (use `--exclude-archive=false` to include)
- Superseded/deprecated status excluded
- Canon-only mode available with `--canon-only`

## Usage Examples

### Index a vault

```bash
./obsidx-indexer --vault ~/notes --watch
```

The indexer automatically:
- Extracts front matter metadata
- Infers category from folder structure as fallback
- Applies appropriate weights

### Search with defaults

```bash
./obsidx-recall "how to handle async errors"
```

This automatically:
- Excludes archive
- Boosts canon results
- Shows category badges in output

### Canon-only search

```bash
./obsidx-recall --canon-only "deployment process"
```

Only searches notes with `category: canon`.

### Include archive

```bash
./obsidx-recall --exclude-archive=false "old architecture"
```

Searches all categories including archived notes.

### Filter by category

```bash
./obsidx-recall --category "canon,project" "auth flow"
```

Only searches canon and project categories.

## Output Format

Results show category badges:
```
[1] Score: 0.8432 [📚 CANON]
Path: /canon/decisions/auth-strategy.md
Section: Authentication > OAuth Flow
Scope: writerflow
...
```

Badges:
- 📚 CANON - Authoritative reference
- 🔨 PROJECT - Active work
- 🧪 WORKBENCH - Experimental
- 📦 ARCHIVE - Historical

## Best Practices

### 1. Start with explicit categories

Add `category:` to your important notes immediately:

```yaml
---
category: canon
status: active
---
```

### 2. Use folder inference as convenience

Structure your vault:
```
notes/
  canon/         # Auto-inferred as canon
  projects/      # Auto-inferred as project
  workbench/     # Auto-inferred as workbench
  archive/       # Auto-inferred as archive
```

### 3. Review and promote

Workflow:
1. Start in `workbench` (drafts)
2. Promote to `project` when active
3. Promote to `canon` when stable
4. Move to `archive` when superseded

### 4. Keep enums tight

Don't create category drift:
- Stick to the 4 core categories
- Use `scope` for domain separation
- Use `type` for document classification

## Agent Integration

The category system enables smart agent behavior:

| Category | Agent Permission |
|----------|------------------|
| `workbench` | Can edit freely, restructure |
| `project` | Can update, should preserve intent |
| `canon` | Propose changes as diffs, require approval |
| `archive` | Read-only, flag if attempting to revive |

## Migration

Existing vaults work immediately:
- Notes without metadata get `category: project` by default
- Folder structure provides fallback inference
- No database migration needed (schema auto-updates)

## SQL Queries

### Find all canon notes
```sql
SELECT path, heading_path, category, status 
FROM chunks 
WHERE category = 'canon' AND active = 1;
```

### Category distribution
```sql
SELECT category, COUNT(*) as count
FROM chunks
WHERE active = 1
GROUP BY category;
```

### Stale canon (not reviewed recently)
```sql
SELECT DISTINCT path, last_reviewed
FROM chunks
WHERE category = 'canon' 
  AND last_reviewed < date('now', '-90 days');
```

## Advanced: Custom Weights

Edit `internal/metadata/metadata.go` to adjust weights:

```go
func (m *NoteMetadata) CategoryWeight() float32 {
    switch m.EffectiveCategory() {
    case "canon":
        return 1.30  // Increase canon boost
    case "workbench":
        return 0.80  // Stronger penalty
    // ...
    }
}
```

## Troubleshooting

### "Canon notes not ranking higher"

Check:
1. Front matter syntax is correct
2. `category_weight` is set in database: `SELECT path, category, category_weight FROM chunks LIMIT 5;`
3. Rerank is using weights: look for `weightedScore` in logs

### "Categories not detected"

1. Verify front matter starts with `---` on first line
2. Check folder paths match inference patterns
3. Manually set category if inference fails

### "Archive notes still appearing"

Use `--exclude-archive=false` intentionally or check that status is not `superseded`.
