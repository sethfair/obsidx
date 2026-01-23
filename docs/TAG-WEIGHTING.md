# Tag-Based Weighting System

## Overview

ObsIDX now uses a flexible, tag-based weighting system instead of hardcoded "canon" categories. This allows you to customize retrieval weights based on your own knowledge management system (Zettelkasten, PARA, etc.).

## Quick Start

### 1. Initialize Weight Configuration

```bash
./bin/obsidx-weights --init
```

This creates `.obsidian-index/weights.json` with sensible defaults for:
- **Zettelkasten tags**: `#permanent-note`, `#literature-note`, `#fleeting-notes`
- **PARA tags**: `#writerflow`, `#mvp-1`, `#archive`, `#marketing`, `#product`
- **Writerflow tags**: `#customer-research`, `#vision`, `#icp`, `#positioning`

### 2. Add Tags to Your Notes

In your Obsidian note frontmatter, use either format:

**Inline format** (recommended for Obsidian users):
```yaml
---
tags: #permanent-note #customer-research #writerflow
status: active
scope: mycompany
type: decision
---
```

**Array format** (also supported):
```yaml
---
tags: [permanent-note, customer-research, writerflow]
status: active
scope: mycompany
type: decision
---
```

Both formats work identically - the system normalizes tags internally.

### 3. Index Your Vault

The indexer automatically loads the weight config:

```bash
./bin/obsidx-indexer --vault ~/obsidian-vault --watch
```

## Weight Configuration

### View Current Weights

```bash
# Show active configuration
./bin/obsidx-weights

# Show defaults
./bin/obsidx-weights --defaults
```

### Customize Weights

Edit `.obsidian-index/weights.json`:

```json
{
  "tag_weights": [
    {
      "tag": "permanent-note",
      "weight": 1.3
    },
    {
      "tag": "fleeting-notes",
      "weight": 0.8
    },
    {
      "tag": "customer-research",
      "weight": 1.25
    },
    {
      "tag": "archive",
      "weight": 0.6
    }
  ],
  "status_weights": [
    {
      "status": "active",
      "weight": 1.0
    },
    {
      "status": "draft",
      "weight": 0.9
    },
    {
      "status": "superseded",
      "weight": 0.5
    }
  ],
  "default_weight": 1.0,
  "multiply_tag_weights": false
}
```

### Weight Calculation Modes

**Max Mode** (`multiply_tag_weights: false`, default):
- Uses the highest weight among matching tags
- Good for most use cases
- Example: Note with `[permanent-note, customer-research]` gets weight 1.3

**Multiply Mode** (`multiply_tag_weights: true`):
- Multiplies all matching tag weights together
- Can compound importance for notes with multiple high-value tags
- Example: Note with `[permanent-note, customer-research]` gets weight 1.3 × 1.25 = 1.625

## Recommended Tag Weights

### By System

**Zettelkasten**:
- `permanent-note`: 1.3 (refined insights)
- `literature-note`: 1.1 (curated sources)
- `fleeting-notes`: 0.8 (quick captures)
- `reference`: 1.0 (lookup material)

**PARA (Second Brain)**:
- Active projects: 1.15-1.25
- Areas of responsibility: 1.1
- Resources: 1.0
- Archive: 0.6

**Custom Domain (e.g., Writerflow)**:
- Customer insights: 1.2-1.3
- Product vision: 1.3
- Validation data: 1.2
- Competitive research: 1.0-1.15

### By Importance Level

- **Critical/Canonical**: 1.3-1.5
- **High Priority**: 1.15-1.25
- **Normal**: 1.0-1.1
- **Low Priority**: 0.8-0.95
- **Archived/Deprecated**: 0.5-0.7

## Migration from Old Category System

The old `category` field is no longer used. Replace with tags in either format:

**Old:**
```yaml
---
category: canon
---
```

**New (inline format):**
```yaml
---
tags: #permanent-note
---
```

**Or (array format):**
```yaml
---
tags: [permanent-note]
---
```

### Tag Equivalents

- `category: canon` → `tags: [permanent-note]` or custom high-priority tag
- `category: project` → `tags: [your-project-name]`
- `category: workbench` → `tags: [draft]` or `status: draft`
- `category: archive` → `tags: [archive]` or `status: deprecated`

## How Weights Affect Retrieval

Weights are applied during re-ranking:

```
final_score = cosine_similarity × tag_weight × status_weight
```

Higher weights boost search relevance, making those notes appear higher in results.

### Example

Query: "customer validation approach"

Without weights (all 1.0):
1. Draft note: similarity=0.85, weight=1.0 → **score=0.85**
2. Permanent note: similarity=0.82, weight=1.0 → **score=0.82**

With tag weights:
1. Draft note: similarity=0.85, weight=0.8 → score=0.68
2. Permanent note (`#permanent-note, #customer-research`): similarity=0.82, weight=1.3 → **score=1.066**

The refined permanent note now ranks higher despite slightly lower similarity.

## Best Practices

1. **Start with defaults** - Run `obsidx-weights --init` and adjust gradually
2. **Use semantic tags** - Match your existing knowledge system
3. **Weight by trust** - Higher weights for validated, refined content
4. **Don't over-weight** - Keep most weights between 0.8-1.5
5. **Archive old content** - Use low weights (0.5-0.7) for deprecated notes
6. **Reindex after changes** - Weight changes only affect newly indexed content

## Troubleshooting

### Weights Not Applied

Make sure:
- Config file exists at `.obsidian-index/weights.json`
- Tags in your notes match config (case-sensitive, no `#` in config)
- You've reindexed after changing weights

### Check What's Being Used

```bash
# See active weight config
./bin/obsidx-weights

# Search with verbose mode to see scores
./bin/obsidx-recall --verbose "your query"
```

### Reset to Defaults

```bash
rm .obsidian-index/weights.json
./bin/obsidx-weights --init
```

## Examples

### Zettelkasten Setup

```json
{
  "tag_weights": [
    {"tag": "permanent-note", "weight": 1.3},
    {"tag": "literature-note", "weight": 1.1},
    {"tag": "fleeting-notes", "weight": 0.8},
    {"tag": "index", "weight": 1.4},
    {"tag": "moc", "weight": 1.35}
  ],
  "default_weight": 1.0,
  "multiply_tag_weights": false
}
```

### Software Development

```json
{
  "tag_weights": [
    {"tag": "architecture-decision", "weight": 1.4},
    {"tag": "production-incident", "weight": 1.3},
    {"tag": "api-design", "weight": 1.2},
    {"tag": "meeting-notes", "weight": 0.9},
    {"tag": "spike", "weight": 0.95},
    {"tag": "archive", "weight": 0.6}
  ],
  "default_weight": 1.0
}
```

### Research/Academic

```json
{
  "tag_weights": [
    {"tag": "primary-source", "weight": 1.4},
    {"tag": "peer-reviewed", "weight": 1.3},
    {"tag": "methodology", "weight": 1.25},
    {"tag": "literature-review", "weight": 1.15},
    {"tag": "hypothesis", "weight": 1.2},
    {"tag": "raw-notes", "weight": 0.85}
  ],
  "default_weight": 1.0
}
```

## Advanced: Per-Project Weights

You can create project-specific configs:

```bash
# Default config
.obsidian-index/weights.json

# Load custom config
./bin/obsidx-indexer --vault ~/vault --weights ~/vault/.weights-custom.json
```

This allows different weighting schemes for different vaults or contexts.
