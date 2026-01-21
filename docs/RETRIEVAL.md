# Retrieval Guide

**How to search and retrieve knowledge from your obsidx-indexed vault.**

## Quick Reference

```bash
# Basic search (excludes archive by default)
obsidx-recall "your search query"

# Canon-only (authoritative decisions only)
obsidx-recall --canon-only "deployment process"

# Include archive
obsidx-recall --exclude-archive=false "historical context"

# Filter by categories
obsidx-recall --category "canon,project" "authentication"

# JSON output (for scripting)
obsidx-recall --json "query" | jq
```

---

## Search Commands

### Basic Search

```bash
obsidx-recall "async error handling"
```

**Default behavior:**
- Excludes `category=archive`
- Excludes `status IN ('superseded', 'deprecated')`
- Returns top 10 results
- Applies category-based ranking weights

### Canon-Only Search

```bash
obsidx-recall --canon-only "rate limiting"
```

**Filters to:**
- Only `category=canon`
- Only `status=active`
- Highest signal for "what should I follow?"

### Include Archive

```bash
obsidx-recall --exclude-archive=false "old decisions"
```

**Includes:**
- All categories including archive
- Useful for historical research
- Archive notes still weighted lower (0.60x)

### Category Filter

```bash
obsidx-recall --category "canon,project" "api design"
```

**Filters to:**
- Only specified categories
- Comma-separated list
- Case-insensitive

### Limit Results

```bash
obsidx-recall --limit 5 "authentication"
```

**Returns:**
- Top N results (default: 10)
- Useful for focused searches

### JSON Output

```bash
obsidx-recall --json "query" | jq '.[] | {path, score, category}'
```

**Returns:**
- Machine-readable JSON array
- Each result includes: `path`, `heading`, `content`, `category`, `status`, `scope`, `score`
- Pipe to `jq` for processing

---

## Understanding Results

### Output Format

```
Found 3 results:

─────────────────────────────────────────────────────────────
[1] Score: 0.8745 [📚 CANON]
Path: /canon/decisions/ADR-003-Rate-Limiting.md
Section: Decision > Implementation
Scope: mycompany
Status: active
Lines: 15-42

We use token bucket rate limiting with Redis backing.
Maximum 100 requests per minute per API key...
─────────────────────────────────────────────────────────────
[2] Score: 0.7621 [🔨 PROJECT]
Path: /projects/api-v2/rate-limit-impl.md
Section: Current Implementation
Lines: 8-25

The rate limiter is implemented in middleware/ratelimit.go...
─────────────────────────────────────────────────────────────
```

### Category Badges

| Badge | Category | Meaning |
|-------|----------|---------|
| 📚 **CANON** | `canon` | Authoritative, stable truth |
| 🔨 **PROJECT** | `project` | Active work, current implementation |
| 🧪 **WORKBENCH** | `workbench` | Drafts, experiments |
| 📦 **ARCHIVE** | `archive` | Historical reference |

### Score Interpretation

**Scores range from 0.0 to 1.0+**

- **0.85 - 1.0+** - Highly relevant, exact match
- **0.70 - 0.85** - Very relevant, strong semantic match
- **0.60 - 0.70** - Relevant, moderate match
- **< 0.60** - Weakly relevant

Scores can exceed 1.0 for canon notes (1.20x boost) with perfect matches.

---

## Retrieval Weights

### How Weights Work

Search results are ranked by: **similarity × category_weight × status_weight**

### Category Weights

| Category | Weight | Effect |
|----------|--------|--------|
| `canon` | 1.20x | 20% boost - prioritizes authoritative content |
| `project` | 1.05x | 5% boost - slightly favors active work |
| `workbench` | 0.90x | 10% penalty - de-prioritizes drafts |
| `archive` | 0.60x | 40% penalty - reduces historical noise |

### Status Weights

| Status | Weight | Effect |
|--------|--------|--------|
| `active` | 1.0x | No modification |
| `draft` | 0.90x | 10% penalty - indicates incomplete |
| `superseded` | 0.50x | 50% penalty (usually excluded) |
| `deprecated` | 0.50x | 50% penalty (usually excluded) |

### Combined Example

A note with `category: canon` and `status: active`:
- Base similarity: 0.85
- Category weight: 1.20
- Status weight: 1.0
- **Final score: 0.85 × 1.20 × 1.0 = 1.02**

A note with `category: workbench` and `status: draft`:
- Base similarity: 0.85
- Category weight: 0.90
- Status weight: 0.90
- **Final score: 0.85 × 0.90 × 0.90 = 0.69**

---

## Filtering Behavior

### Default Filters

By default, `obsidx-recall` excludes:

1. **Archive category** - Historical content not relevant to current decisions
2. **Superseded status** - Content replaced by newer versions
3. **Deprecated status** - Content no longer valid

### Override Filters

**Include archive:**
```bash
obsidx-recall --exclude-archive=false "old architecture"
```

**Show only specific categories:**
```bash
obsidx-recall --category "canon" "decisions"
```

**Include all statuses:**
Currently no flag; use `--exclude-archive=false` to get most content.

---

## Common Workflows

### 1. Finding Authoritative Decisions

**Goal:** What is our established approach?

```bash
obsidx-recall --canon-only "authentication strategy"
```

**Why:** Canon notes are authoritative and stable.

### 2. Understanding Current Implementation

**Goal:** What's the current state of this feature?

```bash
obsidx-recall --category "project" "rate limiting implementation"
```

**Why:** Project notes reflect active work.

### 3. Researching Historical Context

**Goal:** Why did we make this decision? What did we try before?

```bash
obsidx-recall --exclude-archive=false "authentication history"
```

**Why:** Archive contains historical decisions and deprecated approaches.

### 4. Finding Related Work

**Goal:** What else is being worked on in this area?

```bash
obsidx-recall --category "project,workbench" "api versioning"
```

**Why:** Includes both active work and explorations.

### 5. Scripting and Automation

**Goal:** Extract data for processing

```bash
# Get all canon notes about architecture
obsidx-recall --canon-only --json "architecture" | \
  jq -r '.[] | "\(.path): \(.heading)"'

# Count results by category
obsidx-recall --json "query" | \
  jq 'group_by(.category) | map({category: .[0].category, count: length})'
```

---

## Search Tips

### Good Queries

✅ **Specific concepts:**
- "OAuth implementation"
- "database migration strategy"
- "rate limiting algorithm"

✅ **Problem statements:**
- "how to handle async errors"
- "when to use caching"
- "user authentication flow"

✅ **Domain-specific terms:**
- "mycompany onboarding"
- "myproject event processing"

### Bad Queries

❌ **Too vague:**
- "how to"
- "best practices"
- "help"

❌ **Single words:**
- "auth"
- "api"
- "database"

Better: Add context - "auth implementation", "api versioning strategy", "database choice rationale"

### Query Techniques

**Use natural language:**
```bash
obsidx-recall "how do we prevent rate limit abuse"
```

**Use technical terms:**
```bash
obsidx-recall "token bucket rate limiter Redis"
```

**Ask questions:**
```bash
obsidx-recall "what is our deployment process"
```

**Specify scope in query:**
```bash
obsidx-recall "mycompany user onboarding flow"
```

---

## Advanced Usage

### Combining with Other Tools

**Grep through results:**
```bash
obsidx-recall "authentication" | grep -i "oauth"
```

**Count matches:**
```bash
obsidx-recall --json "api" | jq 'length'
```

**Extract specific fields:**
```bash
obsidx-recall --json "decision" | jq -r '.[] | "\(.category) - \(.path)"'
```

**Find notes by scope:**
```bash
obsidx-recall --json "query" | jq '.[] | select(.scope == "mycompany")'
```

### Shell Aliases

Add to `~/.zshrc` or `~/.bashrc`:

```bash
# Quick access
alias kb='obsidx-recall'
alias kb-canon='obsidx-recall --canon-only'
alias kb-all='obsidx-recall --exclude-archive=false'
alias kb-json='obsidx-recall --json'

# Domain-specific
alias kb-wf='obsidx-recall --json | jq ".[] | select(.scope == \"mycompany\")"'
```

Then use:
```bash
kb "authentication"
kb-canon "deployment"
kb-all "historical decisions"
```

---

## Troubleshooting

### No Results Returned

**Possible causes:**

1. **Index not built:** Run `./run.sh ~/notes` first
2. **Query too specific:** Try broader terms
3. **Wrong category filter:** Remove filters to search all categories
4. **Content not indexed:** Check if files are `.md` and in vault

**Debug:**
```bash
# Check if database exists
ls -lh .obsidian-index/obsidx.db

# Count indexed chunks
sqlite3 .obsidian-index/obsidx.db "SELECT COUNT(*) FROM chunks WHERE active = 1;"

# Try broad query
obsidx-recall --exclude-archive=false "the"
```

### Low Scores

**Possible causes:**

1. **Semantic mismatch:** Content exists but uses different terminology
2. **Wrong category weights:** Adjust if needed (see KNOWLEDGE_GOVERNANCE.md)
3. **Query too broad:** Be more specific

**Solutions:**
- Try synonyms or related terms
- Check actual note content to see terminology used
- Use more context in query

### Canon Notes Not Ranking First

**Check:**

1. **Metadata correct:** `category: canon` and `status: active`
2. **Reindex if metadata changed:** `./run.sh ~/notes`
3. **Verify in database:**
   ```bash
   sqlite3 .obsidian-index/obsidx.db \
     "SELECT path, category, category_weight FROM chunks WHERE category='canon' LIMIT 5;"
   ```

### Results Don't Match Query

**This is expected semantic search behavior:**
- Results are based on meaning, not keyword matching
- Related concepts will surface even without exact word matches
- Use category filters to narrow scope

---

## Integration Examples

### GitHub Copilot CLI

GitHub Copilot CLI can use obsidx through instruction-based tool invocation. There are two approaches:

#### Approach 1: Shell Alias (Recommended)

Create a shell alias that Copilot can reference:

```bash
# Add to ~/.zshrc or ~/.bashrc
alias kb='~/code/obsidx/bin/obsidx-recall'
alias kb-canon='~/code/obsidx/bin/obsidx-recall --canon-only'
alias kb-json='~/code/obsidx/bin/obsidx-recall --json'

# Reload shell
source ~/.zshrc
```

Then use in terminal:
```bash
# Search before asking Copilot
kb-canon "authentication strategy"

# Then ask Copilot based on results
gh copilot suggest "implement authentication based on our docs"
```

#### Approach 2: Direct Command

Run obsidx before asking Copilot questions:

```bash
# Get context first
~/code/obsidx/bin/obsidx-recall --canon-only "rate limiting"

# Use results to inform your Copilot query
gh copilot suggest "implement rate limiting using our documented approach"
```

**Note:** GitHub Copilot CLI does not automatically call external tools. You need to:
1. Run obsidx-recall manually to get context
2. Use that context to inform your Copilot questions
3. Reference the found documentation in your prompts

### GitHub Copilot (Editor Instructions)

For in-editor GitHub Copilot, use instruction files:

```markdown
Before answering, search knowledge base:

\`\`\`bash
obsidx-recall --canon-only "authentication strategy"
\`\`\`

Use results to inform code generation.
```

See [COPILOT_QUICKSTART.md](COPILOT_QUICKSTART.md) for full setup.

### CI/CD Scripts

```bash
#!/bin/bash
# Check if ADR exists before deploying

if obsidx-recall --canon-only "deployment to production" | grep -q "ADR-"; then
  echo "✅ Deployment ADR found"
  exit 0
else
  echo "❌ No deployment ADR found. Create one first."
  exit 1
fi
```

### Documentation Validation

```bash
#!/bin/bash
# Find undocumented features

features=("authentication" "rate limiting" "caching" "logging")

for feature in "${features[@]}"; do
  count=$(obsidx-recall --canon-only --json "$feature" | jq 'length')
  if [ "$count" -eq 0 ]; then
    echo "⚠️  No canon documentation for: $feature"
  fi
done
```

---

## Summary

**Key Commands:**

```bash
obsidx-recall "query"                    # Standard search
obsidx-recall --canon-only "query"        # Decisions only
obsidx-recall --category "canon,project"  # Filter categories
obsidx-recall --exclude-archive=false     # Include history
obsidx-recall --json "query" | jq         # Machine-readable
```

**Remember:**
- Canon notes rank highest (1.20x boost)
- Archive excluded by default
- Scores reflect both similarity and category weights
- Use specific queries for better results

---

**Next:**
- [KNOWLEDGE_GOVERNANCE.md](KNOWLEDGE_GOVERNANCE.md) - How to structure and categorize notes
- [COPILOT_QUICKSTART.md](COPILOT_QUICKSTART.md) - Integrate with AI assistants
- [../README.md](../README.md) - Full project documentation
