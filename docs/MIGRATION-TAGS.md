# Migration Guide: Category to Tag-Based Weighting

## Overview

ObsIDX has migrated from hardcoded `category` fields to a flexible tag-based weighting system. This guide will help you migrate your existing vault.

## What Changed

### Before (Category-Based)
```yaml
---
category: canon
scope: mycompany
status: active
---
```

### After (Tag-Based)

ObsIDX supports **two tag formats** - use whichever matches your workflow:

**Option 1: Inline with hashtags** (Obsidian-style)
```yaml
---
tags: #permanent-note #customer-research
scope: mycompany
status: active
---
```

**Option 2: Array format**
```yaml
---
tags: [permanent-note, customer-research]
scope: mycompany
status: active
---
```

Both formats work identically. The system normalizes tags internally (removing `#` prefix).

## Why This Change?

1. **Flexibility**: Support any knowledge management system (Zettelkasten, PARA, Second Brain, custom)
2. **Customization**: Configure weights per-tag to match your workflow
3. **Multiple Tags**: Notes can have multiple priority indicators
4. **No Hardcoding**: No more "canon" terminology baked into the system
5. **Better Knowledge Systems**: Align with established methodologies (Zettelkasten, Building a Second Brain)

## Knowledge Management Integration

This migration enables integration with popular knowledge management systems:

### Zettelkasten Method
- `#fleeting-notes` → Quick captures (24-48 hour lifespan)
- `#literature-note` → Paraphrased sources (`@authorTitle[Year].md`)
- `#permanent-note` → Refined, atomic, evergreen insights
- `#reference` → Source material

### Building a Second Brain (PARA)
- `#project-name` → Projects (active work with deadlines)
- `#area-name` → Areas (ongoing responsibilities)
- `#topic-name` → Resources (reference material)
- `#archive` → Archives (completed work)

### Hybrid Approach (Recommended)
- Use **Zettelkasten tags** for knowledge development
- Use **project tags** for actionable work
- **Permanent notes** provide conceptual foundation
- **Project docs** reference permanent notes for context

## Migration Steps

### 1. Update ObsIDX

```bash
cd ~/code/obsidx
git pull
./build.sh  # or go build commands
```

### 2. Initialize Weight Configuration

```bash
./bin/obsidx-weights --init
```

This creates `.obsidian-index/weights.json` with defaults that match the old category system:
- `#permanent-note` = old "canon" (weight: 1.3)
- Tags for projects (weight: 1.05-1.2)
- `#fleeting-notes` = old "workbench" (weight: 0.8)
- `#archive` = old "archive" (weight: 0.6)

### 3. Update Your Notes

You have two options:

#### Option A: Automated (Recommended)

Create a script to batch-update your notes:

**For Inline Hashtag Format:**
```bash
#!/bin/bash
# migrate-categories-inline.sh

find ~/obsidian-vault -name "*.md" -type f | while read file; do
    # Replace category: canon with tags: #permanent-note
    sed -i.bak 's/^category: canon$/tags: #permanent-note/' "$file"
    
    # Replace category: project with relevant project tags
    sed -i.bak 's/^category: project$/tags: #your-project-name/' "$file"
    
    # Replace category: workbench with tags: #fleeting-notes
    sed -i.bak 's/^category: workbench$/tags: #fleeting-notes/' "$file"
    
    # Replace category: archive with tags: #archive
    sed -i.bak 's/^category: archive$/tags: #archive/' "$file"
done

# Clean up backup files
find ~/obsidian-vault -name "*.bak" -delete
```

**For Array Format:**
```bash
#!/bin/bash
# migrate-categories-array.sh

find ~/obsidian-vault -name "*.md" -type f | while read file; do
    # Replace category: canon with tags: [permanent-note]
    sed -i.bak 's/^category: canon$/tags: [permanent-note]/' "$file"
    
    # Replace category: project with relevant project tags
    sed -i.bak 's/^category: project$/tags: [your-project-name]/' "$file"
    
    # Replace category: workbench with tags: [fleeting-notes]
    sed -i.bak 's/^category: workbench$/tags: [fleeting-notes]/' "$file"
    
    # Replace category: archive with tags: [archive]
    sed -i.bak 's/^category: archive$/tags: [archive]/' "$file"
done

# Clean up backup files
find ~/obsidian-vault -name "*.bak" -delete
```

#### Option B: Manual (For Small Vaults)

Search for `category:` in your vault and update each note.

**Choose your preferred format:**

**Inline hashtag format** (recommended for Obsidian users):
1. Open note
2. Find `category: canon` → Change to `tags: #permanent-note`
3. Find `category: project` → Change to `tags: #your-project-name`
4. Find `category: workbench` → Change to `tags: #fleeting-notes` or `status: draft`
5. Find `category: archive` → Change to `tags: #archive`

**Array format** (recommended for YAML purists):
1. Open note
2. Find `category: canon` → Change to `tags: [permanent-note]`
3. Find `category: project` → Change to `tags: [your-project-name]`
4. Find `category: workbench` → Change to `tags: [fleeting-notes]` or `status: draft`
5. Find `category: archive` → Change to `tags: [archive]`

### 4. Customize Your Weights (Recommended)

Edit `.obsidian-index/weights.json` to match your knowledge management system:

#### For Zettelkasten Users

```json
{
  "tag_weights": [
    {"tag": "permanent-note", "weight": 1.3},
    {"tag": "literature-note", "weight": 1.15},
    {"tag": "reference", "weight": 1.0},
    {"tag": "fleeting-notes", "weight": 0.8},
    {"tag": "archive", "weight": 0.6}
  ],
  "status_weights": [
    {"status": "active", "weight": 1.0},
    {"status": "draft", "weight": 0.9},
    {"status": "deprecated", "weight": 0.5}
  ],
  "default_weight": 1.0,
  "multiply_tag_weights": false
}
```

#### For Second Brain (PARA) Users

```json
{
  "tag_weights": [
    {"tag": "your-main-project", "weight": 1.25},
    {"tag": "important-area", "weight": 1.15},
    {"tag": "secondary-project", "weight": 1.1},
    {"tag": "reference-topic", "weight": 1.0},
    {"tag": "archive", "weight": 0.6}
  ],
  "status_weights": [
    {"status": "active", "weight": 1.0},
    {"status": "draft", "weight": 0.9},
    {"status": "someday", "weight": 0.7},
    {"status": "deprecated", "weight": 0.5}
  ],
  "default_weight": 1.0,
  "multiply_tag_weights": false
}
```

#### For Hybrid Users (Zettelkasten + Second Brain)

```json
{
  "tag_weights": [
    {"tag": "permanent-note", "weight": 1.3},
    {"tag": "your-main-project", "weight": 1.25},
    {"tag": "literature-note", "weight": 1.15},
    {"tag": "important-area", "weight": 1.15},
    {"tag": "reference", "weight": 1.0},
    {"tag": "fleeting-notes", "weight": 0.8},
    {"tag": "archive", "weight": 0.6}
  ],
  "status_weights": [
    {"status": "active", "weight": 1.0},
    {"status": "draft", "weight": 0.9},
    {"status": "deprecated", "weight": 0.5}
  ],
  "default_weight": 1.0,
  "multiply_tag_weights": false
}
```

**Note:** With `multiply_tag_weights: false`, the **highest weight** wins. Set to `true` to multiply weights (e.g., `permanent-note` + `your-project` = 1.3 × 1.25 = 1.625).

### 5. Rebuild Index

After updating your notes:

```bash
# Stop existing services
./stop-daemon.sh

# Delete old index (schema changed)
rm -rf .obsidian-index/obsidx.db
rm -rf .obsidian-index/*.hnsw.bin

# Rebuild from scratch
./bin/obsidx-indexer --vault ~/obsidian-vault

# Start server
./start-daemon.sh
```

## Category to Tag Mappings

### Direct Equivalents (Basic Migration)

| Old Category | New Tag | Weight | Notes |
|--------------|---------|--------|-------|
| `canon` | `permanent-note` | 1.3 | Authoritative, refined insights |
| `project` | Project name (e.g., `writerflow`) | 1.15-1.25 | Active work with deadlines |
| `workbench` | `fleeting-notes` or `draft` status | 0.8-0.9 | Work in progress |
| `archive` | `archive` | 0.6 | Historical reference |

### Enhanced Mappings for Knowledge Management Systems

#### Zettelkasten Users

| Old | New (Zettelkasten) | Why Better |
|-----|---------------------|------------|
| `canon` | `permanent-note` | Atomic, evergreen insights |
| `project, type: note` | `literature-note` | Paraphrased sources |
| `workbench` | `fleeting-notes` | Temporary captures |
| N/A | `reference` | Raw source material |

**Recommended tag combinations:**
```yaml
# Permanent note on customer research (inline format)
tags: #permanent-note #customer-research #validation

# OR (array format - both work!)
tags: [permanent-note, customer-research, validation]

# Literature note from a book
tags: #literature-note #reference #mom-test

# Fleeting note to process
tags: #fleeting-notes #customer-research
```

#### Second Brain (PARA) Users

| Old | New (PARA) | Why Better |
|-----|------------|------------|
| `project` | Specific project name (e.g., `writerflow`, `mvp-1`) | Clear project context |
| `canon, scope: area` | Area name + `permanent-note` | Ongoing responsibility with authority |
| `workbench` | Project tag + `status: draft` | WIP within project context |
| `archive` | `archive` + original tags | Maintains searchability |

**Recommended tag combinations:**
```yaml
# Active project work (inline format)
tags: #writerflow #product-development #mvp-1

# OR (array format)
tags: [writerflow, product-development, mvp-1]

# Ongoing area of responsibility
tags: #marketing #permanent-note #strategy

# Archived project
tags: #old-project #archive
```

#### Hybrid Users (Zettelkasten + Second Brain)

| Old | New (Hybrid) | Why Better |
|-----|--------------|------------|
| `canon` | `permanent-note` + domain tags | Conceptual authority |
| `project` | Project name + document type | Context and purpose clear |
| `workbench, type: learning` | `literature-note` or `fleeting-notes` | Knowledge development stage |
| `workbench, type: building` | Project tag + `status: draft` | Project development stage |

**Recommended tag combinations:**
```yaml
# Permanent insight applied to project
tags: [permanent-note, customer-research, validation]
# Project doc references this

# Project document with conceptual grounding
tags: [writerflow, customer-research, validation-strategy]
# References permanent notes

# Literature note for learning
tags: [literature-note, reference, mom-test, customer-research]
# Feeds into permanent notes
```

## Verification

### Check Your Migration

```bash
# Search for old category syntax (should return nothing)
grep -r "^category:" ~/obsidian-vault

# Verify tags are working
./bin/obsidx-recall --verbose "test query"
```

Results should show tags like `[permanent-note, ...]` instead of categories.

### Test Weight Impact

```bash
# Create two test notes
echo "---
tags: [permanent-note]
---
# High Priority Note
Important content here" > /tmp/test-high.md

echo "---
tags: [fleeting-notes]
---
# Low Priority Note
Same content here" > /tmp/test-low.md

# Index and search - high priority should rank higher
./bin/obsidx-recall "important content"
```

## Backwards Compatibility

The old `category` field is **no longer used**. If you leave it in your notes:
- It will be ignored during indexing
- No errors will occur
- Search will use tags and status only

## Rollback (If Needed)

If you need to rollback:

1. **Restore note backups** (`.bak` files from sed)
```bash
find ~/obsidian-vault -name "*.bak" | while read file; do
    mv "$file" "${file%.bak}"
done
```

2. **Restore old ObsIDX version**
```bash
git checkout <previous-commit>
./build.sh
```

3. **Rebuild index with old version**

## Common Issues

### Tags Not Showing in Search Results

**Problem**: Results don't show tags
**Solution**: Make sure you rebuilt the index after updating notes

### Weights Not Applied

**Problem**: All notes seem to have same priority
**Solution**: 
1. Check `.obsidian-index/weights.json` exists
2. Verify tags match config (case-sensitive, no `#` in config file)
3. Rebuild index

### Multiple Tags Conflict

**Problem**: Not sure which weight is used
**Solution**: By default, **max weight** is used. Set `multiply_tag_weights: true` to multiply them instead.

## Get Help

- Read `docs/TAG-WEIGHTING.md` for complete documentation
- Check weight config: `./bin/obsidx-weights`
- View search results with verbose mode: `./bin/obsidx-recall --verbose "query"`

## Example Migration

### Example 1: Zettelkasten Note

Before:
```yaml
---
category: canon
scope: mycompany
type: principle
status: active
last_reviewed: 2026-01-20
---

# Never Pitch During Discovery

The best customer conversations happen when you never mention your idea...
```

After:
```yaml
---
tags: [permanent-note, customer-research, validation]
scope: mycompany
type: principle
status: active
last_reviewed: 2026-01-20
---

# Never Pitch During Discovery

The best customer conversations happen when you never mention your idea...
```

Benefits:
- `permanent-note` gives high base weight (1.3)
- `customer-research` enables topic filtering
- `validation` shows domain context
- All three tags make this highly discoverable

### Example 2: Second Brain (PARA) Note

Before:
```yaml
---
category: project
scope: mycompany
status: active
---

# Writerflow MVP 1 Features

## Core Features
...
```

After:
```yaml
---
tags: [writerflow, mvp-1, product-development]
scope: mycompany
status: active
---

# Writerflow MVP 1 Features

## Core Features
...
```

Benefits:
- `writerflow` identifies project (weight: 1.25)
- `mvp-1` shows specific phase
- `product-development` shows work type
- Easy to filter by project, phase, or work type

### Example 3: Hybrid Note (Permanent + Project)

Before:
```yaml
---
category: canon
scope: mycompany
type: decision
status: active
last_reviewed: 2026-01-20
---

# ADR-001: Use HNSW

## Decision
...
```

After:
```yaml
---
tags: [permanent-note, architecture-decision, search-system, writerflow]
scope: mycompany
type: decision
status: active
last_reviewed: 2026-01-20
---

# ADR-001: Use HNSW

## Decision
...
```

Benefits:
- `permanent-note` marks as authoritative (1.3)
- `architecture-decision` provides clear type
- `search-system` enables topic-based filtering
- `writerflow` links to project context
- All four tags maximize discoverability
