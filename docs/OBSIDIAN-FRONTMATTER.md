# Obsidian Frontmatter Guide for obsidx

How to organize your Obsidian vault using frontmatter metadata for smart semantic search.

## Overview

obsidx uses **frontmatter metadata** to organize and prioritize your Obsidian notes. This creates a knowledge hierarchy where some notes are authoritative canon, others are active projects, and some are experimental drafts.

**Why this matters:**
- ✅ Search finds authoritative content first
- ✅ AI assistants respect your established decisions
- ✅ Drafts don't pollute search results
- ✅ Knowledge has a clear lifecycle
- ✅ You control what's "truth" vs "experiment"

## Quick Start

Add this to the top of your Obsidian notes:

```yaml
---
category: canon        # canon | project | workbench | archive
scope: mycompany      # your organization/project
type: decision         # decision | principle | vision | spec | note | log
status: active         # active | draft | superseded | deprecated
last_reviewed: 2026-01-20
---
```

**That's it!** obsidx will automatically:
- Weight canon notes higher in search results
- Filter out archive by default
- Help AI understand what's authoritative

## Frontmatter Fields

### Required: None (All Optional)

If you don't add frontmatter, obsidx infers the category from your folder structure:
- `/canon/` → `category: canon`
- `/project/` → `category: project`
- `/workbench/` → `category: workbench`
- `/archive/` → `category: archive`

### Recommended Fields

#### `category` - Knowledge Lifecycle Stage

Where this note is in its evolution:

```yaml
category: canon        # Authoritative, stable (ADRs, principles)
category: project      # Active work, current implementation
category: workbench    # Drafts, experiments, brainstorming
category: archive      # Historical, deprecated, superseded
```

**Search impact:**
- Canon notes are weighted **1.20x** higher in results
- Archive is excluded by default (use `--exclude-archive=false` to include)

#### `type` - Document Purpose

What role this document serves:

```yaml
type: decision         # Architecture Decision Record (ADR)
type: principle        # Core belief, guideline, team value
type: vision           # Product vision, long-term direction
type: spec             # Technical specification, requirements
type: note             # General knowledge, reference material
type: log              # Journal, changelog, retrospective
```

#### `status` - Current Validity

Is this still relevant?

```yaml
status: active         # Current, follow this
status: draft          # Work in progress, not finalized
status: superseded     # Replaced (link to replacement)
status: deprecated     # No longer relevant (kept for history)
```

#### `scope` - Project/Domain

Organize by product, team, or domain:

```yaml
scope: mycompany       # Your organization
scope: myproject       # Specific project
scope: personal        # Personal notes
```

Search by scope:
```bash
./bin/obsidx-recall --scope "myproject" "authentication"
```

#### `last_reviewed` - Governance Date

When was this last validated? (Especially important for canon):

```yaml
last_reviewed: 2026-01-20
```

## Category System Explained

### Canon (Authoritative Truth)

**What it is:**
- Architecture Decision Records (ADRs)
- Core principles and values
- Established patterns and practices
- Team agreements and standards

**Examples:**
- `/canon/decisions/ADR-001-Use-PostgreSQL.md`
- `/canon/principles/API-Design-Principles.md`
- `/canon/architecture/System-Architecture.md`

**Frontmatter:**
```yaml
---
category: canon
type: decision
status: active
last_reviewed: 2026-01-20
---
```

**Search behavior:**
- Weighted **1.20x** higher in results
- AI treats as authoritative
- Use `--canon-only` to search only these

### Project (Active Work)

**What it is:**
- Current implementation docs
- Sprint/cycle documentation
- Technical notes for ongoing work
- Feature specifications

**Examples:**
- `/project/features/Authentication-System.md`
- `/project/specs/API-Endpoints.md`
- `/project/implementation/Database-Schema.md`

**Frontmatter:**
```yaml
---
category: project
type: spec
status: active
scope: myproject
---
```

**Search behavior:**
- Weighted **1.05x** (slight boost)
- Represents current state
- Included in default searches

### Workbench (Experiments & Drafts)

**What it is:**
- Brainstorming and exploration
- Research notes
- Draft proposals
- "Thinking out loud" notes

**Examples:**
- `/workbench/research/Caching-Options.md`
- `/workbench/drafts/Proposed-Architecture.md`
- `/workbench/experiments/New-Framework-Test.md`

**Frontmatter:**
```yaml
---
category: workbench
type: note
status: draft
---
```

**Search behavior:**
- Weighted **0.90x** (slightly lower)
- Clearly marked as experimental
- Useful for finding research

### Archive (Historical Reference)

**What it is:**
- Superseded decisions
- Old project documentation
- Deprecated approaches
- Historical context

**Examples:**
- `/archive/decisions/ADR-001-Old-Database-Choice.md`
- `/archive/projects/Legacy-System-Docs.md`

**Frontmatter:**
```yaml
---
category: archive
type: decision
status: superseded
superseded_by: "[[ADR-005-New-Database-Choice]]"
---
```

**Search behavior:**
- Weighted **0.60x** (much lower)
- **Excluded by default**
- Use `--exclude-archive=false` to include

## Knowledge Lifecycle

### The Lifecycle Flow

```
workbench (draft)
    ↓
  Refined & validated
    ↓
project (active implementation)
    ↓
  Becomes established pattern
    ↓
canon (authoritative)
    ↓
  Eventually replaced
    ↓
archive (historical)
```

### Promoting to Canon

When a workbench note becomes established:

1. **Validate the approach** - Has it been tested/proven?
2. **Add frontmatter:**
   ```yaml
   ---
   category: canon
   type: decision
   status: active
   last_reviewed: 2026-01-20
   ---
   ```
3. **Move the file:** `/workbench/idea.md` → `/canon/decisions/ADR-XXX-Idea.md`
4. **Reindex:** obsidx picks up the changes automatically

### Archiving Old Content

When canon becomes superseded:

1. **Create replacement** - New ADR or decision
2. **Update frontmatter:**
   ```yaml
   ---
   category: archive
   status: superseded
   superseded_by: "[[ADR-XXX-New-Approach]]"
   ---
   ```
3. **Move the file:** `/canon/...` → `/archive/...`
4. **Link from new doc** - Reference the historical decision

## Architecture Decision Records (ADRs)

ADRs are the primary canon document type.

### ADR Template

```markdown
---
category: canon
type: decision
status: active
scope: myproject
last_reviewed: 2026-01-20
---

# ADR-XXX: [Title]

**Status:** Active  
**Date:** 2026-01-20  
**Deciders:** [Names]

## Context

What problem are we solving? What constraints exist?

## Decision

What did we decide? Be specific.

## Rationale

Why this approach? What alternatives did we consider?

## Consequences

What are the trade-offs?

**Benefits:**
- Benefit 1
- Benefit 2

**Drawbacks:**
- Drawback 1
- Drawback 2

## Implementation

How do we apply this decision?

## References

- Related ADRs: [[ADR-YYY]]
- External resources: [Link]
```

### Naming Convention

Use consistent naming:
- `ADR-001-Database-Choice.md`
- `ADR-002-Authentication-Strategy.md`
- `ADR-003-API-Design.md`

## Search Examples

### Search Canon Only

```bash
# Find authoritative decisions
./bin/obsidx-recall --canon-only "authentication"
```

### Search by Category

```bash
# Active work only
./bin/obsidx-recall --category "project" "current features"

# Canon and project (exclude experiments)
./bin/obsidx-recall --category "canon,project" "API design"
```

### Include Archive

```bash
# See historical decisions
./bin/obsidx-recall --exclude-archive=false "old database"
```

### Search by Scope

```bash
# Project-specific
./bin/obsidx-recall --scope "myproject" "deployment"
```

## AI Integration

### How AI Uses Frontmatter

When GitHub Copilot searches your vault:

1. **Prioritizes canon** - Treats as authoritative
2. **Respects status** - Flags superseded decisions
3. **Understands hierarchy** - Canon > project > workbench > archive
4. **Cites sources** - References specific notes

### Example AI Behavior

**User asks:** "How should we handle authentication?"

**Copilot with obsidx:**
```
📚 CANON: ADR-005-Authentication-Strategy.md

Our established approach is OAuth 2.0 with JWT tokens.
[Implements according to canon]
```

**Without metadata:**
```
[Generic OAuth advice without context]
```

## Best Practices

### 1. Start Simple

Just add `category` to your most important notes:

```yaml
---
category: canon
---
```

### 2. Be Consistent

Pick a convention and stick to it:
- Folder structure: `/canon/`, `/project/`, etc.
- Naming: `ADR-XXX-Title.md`
- Frontmatter fields: Same fields across similar notes

### 3. Review Regularly

For canon notes, update `last_reviewed`:

```yaml
---
category: canon
last_reviewed: 2026-01-20  # Update quarterly
---
```

### 4. Link Between Notes

Connect related decisions:

```yaml
---
category: canon
related: "[[ADR-002]], [[ADR-005]]"
superseded_by: "[[ADR-010]]"  # If superseded
---
```

### 5. Use Folder Structure

Let folders provide defaults:
- `/canon/decisions/` → All get `category: canon`, `type: decision`
- `/project/specs/` → All get `category: project`, `type: spec`
- `/workbench/research/` → All get `category: workbench`

## Troubleshooting

### Search Not Finding Canon Notes

**Check:**
1. Is frontmatter valid YAML?
2. Is the file a `.md` file?
3. Has the indexer run? Check logs: `tail -f .obsidian-index/indexer.log`

### Results Seem Random

**Add metadata:**
- Canon notes should have `category: canon`
- Use `--canon-only` to see only authoritative content

### Too Many Old/Irrelevant Results

**Archive old content:**
- Move to `/archive/` or add `category: archive`
- Archive is excluded by default

## Summary

**Minimal setup:**
```yaml
---
category: canon  # Or project, workbench, archive
---
```

**Complete setup:**
```yaml
---
category: canon
type: decision
status: active
scope: myproject
last_reviewed: 2026-01-20
---
```

**The result:**
- ✅ Smart search that finds authoritative content first
- ✅ AI that respects your decisions
- ✅ Clear knowledge hierarchy
- ✅ Searchable, organized vault

**Philosophy:** Your Obsidian vault + frontmatter = AI-readable knowledge base

---

## Related Documentation

- [RETRIEVAL.md](RETRIEVAL.md) - How to search effectively
- [COPILOT.md](COPILOT.md) - AI integration guide
- [MODES.md](MODES.md) - Running obsidx
