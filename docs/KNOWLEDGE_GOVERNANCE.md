# Knowledge Governance Guide

**How to structure, categorize, and govern knowledge in your Obsidian vault using obsidx.**

## Table of Contents

1. [Overview](#overview)
2. [Metadata Schema](#metadata-schema)
3. [Category System](#category-system)
4. [Retrieval Weights](#retrieval-weights)
5. [Knowledge Lifecycle](#knowledge-lifecycle)
6. [ADR Pattern](#adr-pattern)
7. [Best Practices](#best-practices)
8. [Agent Integration](#agent-integration)

---

## Overview

obsidx uses **metadata-driven categories** to govern your knowledge base. This prevents conceptual drift, separates truth from drafts, and gives AI agents a reliable mental model of "what is authoritative vs. what is exploratory."

### Goals

- Prevent conceptual drift as the vault grows
- Separate *truth* from *drafts* from *experiments*
- Make retrieval predictable and low-noise
- Enable agents to reason *with* your knowledge, not merely *over* it
- Reduce AI context size by retrieving only relevant, authoritative material

### How It Works

Every note can declare metadata in YAML front matter:

```yaml
---
category: canon        # canon | project | workbench | archive
scope: mycompany      # your domain/project
type: decision         # vision | principle | decision | spec | note | log
status: active         # active | draft | superseded | deprecated
last_reviewed: 2026-01-20
---
```

Notes without explicit metadata inherit their category from folder structure (e.g., `/canon/` → `canon`).

---

## Metadata Schema

### Field Reference

| Field | Purpose | Values | Required |
|-------|---------|--------|----------|
| `category` | Lifecycle stage | `canon`, `project`, `workbench`, `archive` | No (inferred) |
| `scope` | Domain/product | Your project names | No |
| `type` | Document role | `decision`, `principle`, `vision`, `spec`, `note`, `log` | No |
| `status` | Current validity | `active`, `draft`, `superseded`, `deprecated` | No |
| `last_reviewed` | Governance date | ISO date (YYYY-MM-DD) | Recommended for canon |
| `tags` | Topics | Array of strings | No |

### Field Semantics

**category** - Where this note sits in the knowledge lifecycle:
- `canon` - Authoritative, stable truth (ADRs, principles, architectural invariants)
- `project` - Active work, evolving documentation
- `workbench` - Drafts, experiments, "thinking out loud"
- `archive` - Historical context, deprecated content

**scope** - Domain or product boundary:
- Use to separate concerns: `mycompany`, `myproject`, `personal`
- Enables domain-specific search: `--scope mycompany`
- Multiple scopes not recommended; use tags instead

**type** - Semantic role of the document:
- `decision` - ADR, architectural decision record
- `principle` - Core belief, guideline, invariant
- `vision` - Product vision, long-term direction
- `spec` - Technical specification, requirements
- `note` - General knowledge, reference
- `log` - Journal, changelog, retrospective

**status** - Current validity:
- `active` - Current, in-use, follow this
- `draft` - Work in progress, not yet finalized
- `superseded` - Replaced by a newer version (link to replacement)
- `deprecated` - No longer relevant, kept for historical reference

---

## Category System

### Category Definitions

| Category | Meaning | Retrieval Weight | Agent Permission |
|----------|---------|------------------|------------------|
| **canon** | Authoritative truth | **1.20x** | Propose changes as diffs, require approval |
| **project** | Active work | **1.05x** | Can update, preserve intent |
| **workbench** | Drafts, experiments | **0.90x** | Can edit freely, restructure |
| **archive** | Historical reference | **0.60x** | Read-only, flag if reviving |

### Status Modifiers

Status values modify the base category weight:

| Status | Multiplier | Default Filter |
|--------|------------|----------------|
| `active` | 1.0x | ✅ Included |
| `draft` | 0.90x | ✅ Included |
| `superseded` | 0.50x | ❌ Excluded by default |
| `deprecated` | 0.50x | ❌ Excluded by default |

**Combined weight** = `category_weight × status_weight`

Example: `canon` + `active` = 1.20 × 1.0 = **1.20**  
Example: `workbench` + `draft` = 0.90 × 0.90 = **0.81**

---

## Retrieval Weights

### How Weights Work

When you search with `obsidx-recall`:

1. **HNSW** finds vector-similar candidates
2. **Reranking** applies category weights to similarity scores:
   ```
   weighted_score = cosine_similarity × category_weight × status_weight
   ```
3. **Filtering** excludes notes based on status and flags:
   - By default: exclude `archive` and `superseded`/`deprecated`
   - Use `--exclude-archive=false` to include everything
4. **Results** sorted by weighted score

### Retrieval Policies

**Default search** (`obsidx-recall "query"`):
- Excludes `category=archive`
- Excludes `status IN ('superseded', 'deprecated')`
- Boosts canon, slightly penalizes workbench

**Canon-only search** (`obsidx-recall --canon-only "query"`):
- Only `category=canon AND status=active`
- Highest signal for "what should I follow?"

**Include everything** (`obsidx-recall --exclude-archive=false "query"`):
- Includes all categories and statuses
- Useful for historical research

**Category filter** (`obsidx-recall --category "canon,project" "query"`):
- Only searches specified categories
- Useful for scoped queries

### Scope Boosting

If a query or note has a scope (e.g., `mycompany`), matching notes get **1.10x boost**.

---

## Knowledge Lifecycle

### Promotion Workflow

Notes move through lifecycle stages:

```
workbench → project → canon → archive
```

#### 1. **Start in Workbench**

Create drafts, brainstorm, explore:

```yaml
---
category: workbench
status: draft
---

# Exploring Authentication Options

Maybe OAuth? Or JWT? Still researching...
```

**Folder:** `@workbench/` or `drafts/`  
**Agent behavior:** Can edit freely, restructure, discard

#### 2. **Promote to Project**

When work becomes structured and active:

```yaml
---
category: project
status: active
scope: mycompany
---

# Authentication Implementation

Using OAuth 2.0 with Okta. In progress...
```

**Folder:** `@projects/<name>/`  
**Agent behavior:** Can update, should preserve intent

#### 3. **Promote to Canon**

When content becomes stable and authoritative:

```yaml
---
category: canon
scope: mycompany
type: decision
status: active
last_reviewed: 2026-01-20
---

# ADR-005 — OAuth 2.0 Authentication

## Decision
We use OAuth 2.0 with Okta for all authentication...
```

**Folder:** `@canon/decisions/` or `@canon/principles/`  
**Agent behavior:** Treat as law, propose changes as ADRs  
**Requirements:** Must have `status` and `last_reviewed`

#### 4. **Archive When Superseded**

When decisions are replaced or deprecated:

```yaml
---
category: archive
status: superseded
superseded_by: "[[ADR-012 - Zero Trust Authentication]]"
last_reviewed: 2026-01-20
---

# ADR-005 — OAuth 2.0 Authentication (SUPERSEDED)

Replaced by ADR-012 which adopts zero-trust model.
...
```

**Folder:** `archive/` or `@archive/`  
**Agent behavior:** Read-only, historical reference only

---

## ADR Pattern

**Architectural Decision Records (ADRs)** are canon documents that capture key decisions.

### Structure

```
@canon/decisions/
  ADR-000 - Index.md           # Table of contents
  ADR-001 - Use HNSW.md        # First decision
  ADR-002 - Database Choice.md # Second decision
  ...
```

### Template

```markdown
---
category: canon
scope: <your-project>
type: decision
status: active
last_reviewed: 2026-01-20
tags: [architecture, <topic>]
---

# ADR-XXX — <Title>

## Context

What is the issue we're facing? What are the constraints?

## Decision

What did we decide? Be specific.

## Alternatives Considered

1. **Option A** - Why not?
2. **Option B** - Why not?
3. **Selected Option** - Why yes?

## Consequences

### Positive
- What benefits?

### Negative
- What trade-offs?

### Neutral
- What else changes?

## Status

Active as of 2026-01-20

## Notes

Any follow-up thoughts or related decisions.
```

### When to Create an ADR

Create an ADR when:
- Making architectural decisions (tech stack, patterns, infrastructure)
- Choosing between competing approaches
- Establishing team conmyprojectns or standards
- Documenting "why" for future reference
- Agents propose changes to existing canon

### ADR Index (ADR-000)

Maintain an index of all ADRs:

```markdown
---
category: canon
type: decision
status: active
---

# ADR Index

| ID | Title | Status | Date |
|----|-------|--------|------|
| [001](ADR-001%20-%20Use%20HNSW.md) | Use HNSW for Recall | Active | 2026-01-15 |
| [002](ADR-002%20-%20PostgreSQL.md) | Use PostgreSQL | Active | 2026-01-18 |
| [003](ADR-003%20-%20Rate%20Limiting.md) | Token Bucket Rate Limiting | Active | 2026-01-20 |
```

---

## Best Practices

### 1. Start with Explicit Categories

Add front matter to important notes immediately:

```yaml
---
category: canon
status: active
last_reviewed: 2026-01-20
---
```

Don't rely solely on folder inference.

### 2. Use Folder Structure for Convenience

Organize your vault:

```
notes/
  @canon/
    decisions/      # ADRs
    principles/     # Core beliefs
    architecture/   # System design
  @projects/
    mycompany/
    myproject/
  @workbench/
    drafts/
    research/
  archive/
```

Folder names with `@` prevent confusion with regular notes.

### 3. Review and Promote Regularly

**Weekly:** Review workbench notes
- Promote useful drafts to project
- Delete dead experiments

**Monthly:** Review project notes
- Promote stable patterns to canon
- Archive completed work

**Quarterly:** Review canon notes
- Update `last_reviewed` dates
- Check for staleness (>90 days)
- Archive superseded decisions

### 4. Keep Enums Tight

**Don't create category drift:**
- Stick to 4 core categories (canon/project/workbench/archive)
- Use `scope` for domain separation
- Use `type` for document classification
- Use `tags` for topics

**Bad:** Creating custom categories like "important" or "reference"  
**Good:** Use `canon` + `type: principle` or `project` + `tags: [reference]`

### 5. Link Superseded Content

When archiving, always link to replacement:

```yaml
superseded_by: "[[ADR-012 - New Approach]]"
```

This creates a trail for future readers.

### 6. Scope for Separation

Use scopes to separate domains:

```yaml
scope: mycompany  # Content creation product
scope: myproject      # Event platform product
scope: personal    # Personal knowledge
```

Then search scoped:
```bash
obsidx-recall --scope mycompany "authentication"
```

---

## Agent Integration

### Agent Behavior by Category

AI agents should adjust behavior based on category:

| Category | Agent Permission | Example |
|----------|------------------|---------|
| **canon** | Propose changes as diffs, require approval | "⚠️ This conflicts with ADR-002. Suggest creating ADR-XXX?" |
| **project** | Can update, should preserve intent | "Updating implementation based on current spec..." |
| **workbench** | Can edit freely, restructure | "Refactoring draft based on feedback..." |
| **archive** | Read-only, flag if attempting to revive | "⚠️ This is archived content. Use current approach instead?" |

### Canon Authority

When canon notes exist, agents **must**:
- ❌ **Never** contradict canon silently
- ✅ **Always** flag conflicts and suggest ADRs
- ❌ **Never** propose alternatives without discussion
- ✅ **Always** implement according to canon guidance

### Agent Instructions

Add to your agent's system prompt:

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

---

## Advanced: Custom Weights

Want to adjust retrieval weights? Edit `internal/metadata/metadata.go`:

```go
func (m *NoteMetadata) CategoryWeight() float32 {
    switch m.EffectiveCategory() {
    case "canon":
        return 1.30  // Increase canon boost to 30%
    case "project":
        return 1.05
    case "workbench":
        return 0.85  // Stronger draft penalty
    case "archive":
        return 0.50  // Even lower archive weight
    default:
        return 1.0
    }
}
```

Rebuild after changes:
```bash
./build.sh
./bin/obsidx-rebuild --db .obsidian-index/obsidx.db --dim 768
```

---

## Migration Guide

### From Unstructured Vault

If you have an existing vault without metadata:

1. **Start fresh:** New notes get metadata immediately
2. **Gradual migration:** Add metadata to notes as you touch them
3. **Bulk tagging:** Use scripts to add default metadata

Example migration script:

```bash
#!/bin/bash
# Add default metadata to all markdown files in a directory

for file in canon/**/*.md; do
  if ! grep -q "^---$" "$file"; then
    # No front matter, add it
    echo "Adding metadata to $file"
    cat > "$file.tmp" <<EOF
---
category: canon
status: active
last_reviewed: $(date +%Y-%m-%d)
---

$(cat "$file")
EOF
    mv "$file.tmp" "$file"
  fi
done
```

### From Existing ADR System

If you already have ADRs:

1. Move ADRs to `@canon/decisions/`
2. Add front matter with `category: canon` and `type: decision`
3. Set `status: active` or `status: superseded` appropriately
4. Add `last_reviewed` dates
5. Reindex: `./watcher.sh ~/notes`

---

## Troubleshooting

### Canon Notes Not Ranking Higher

Check:
1. Front matter syntax is correct (starts with `---` on line 1)
2. Category is set: `category: canon`
3. Status is active: `status: active`
4. Database reflects metadata: 
   ```bash
   sqlite3 .obsidian-index/obsidx.db "SELECT path, category, category_weight FROM chunks LIMIT 5;"
   ```
5. Reindex if needed: `./watcher.sh ~/notes`

### Categories Not Detected

1. Verify front matter format (YAML, starts/ends with `---`)
2. Check folder paths match inference patterns
3. Manually set category if inference fails
4. Check logs during indexing: `📝 Detected change: ...`

### Stale Canon Detection

Find canon notes not reviewed recently:

```bash
sqlite3 .obsidian-index/obsidx.db "
SELECT DISTINCT path, 
       json_extract(meta_json, '$.last_reviewed') as last_reviewed
FROM chunks
WHERE category = 'canon' 
  AND date(json_extract(meta_json, '$.last_reviewed')) < date('now', '-90 days');
"
```

---

## Summary

**Knowledge Governance = Metadata + Lifecycle + Retrieval**

1. **Add metadata** to notes (or use folder inference)
2. **Follow lifecycle** (workbench → project → canon → archive)
3. **Let retrieval weights** surface the right content
4. **Treat canon as law** for AI agents
5. **Review regularly** to prevent drift

This system turns your Obsidian vault into a governed, evolving knowledge base where **canon = truth**, **project = work**, **workbench = thought**, and **archive = memory**.

---

**Next Steps:**
- Read [RETRIEVAL.md](RETRIEVAL.md) for search commands
- Read [COPILOT_QUICKSTART.md](COPILOT_QUICKSTART.md) for AI integration
- See [../README.md](../README.md) for full project documentation
