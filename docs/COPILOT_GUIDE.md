# GitHub Copilot Instructions for obsidx Integration

## Overview

Configure GitHub Copilot to use **obsidx** for retrieving context from your Obsidian vault before answering questions or generating code. This ensures Copilot uses your established knowledge, decisions, and canon documentation instead of making assumptions.

## Setup Instructions

### 1. Create Copilot Instructions File

Create or update your `.github/copilot-instructions.md` file in your workspace:

```bash
mkdir -p .github
touch .github/copilot-instructions.md
```

### 2. Add obsidx Integration Instructions

Copy this template into `.github/copilot-instructions.md`:

```markdown
# GitHub Copilot - Knowledge Base Integration

## Context Retrieval Protocol

Before answering questions or implementing features, you MUST retrieve relevant context from the obsidx knowledge base.

### When to Use obsidx

Use obsidx for:
- ✅ Architectural decisions
- ✅ Design patterns and principles
- ✅ Product requirements and vision
- ✅ Technical implementation details
- ✅ Business logic and rules
- ✅ API design and contracts
- ✅ Customer personas and use cases
- ✅ Marketing strategy and positioning

### How to Retrieve Context

Execute this command to search the knowledge base:

\`\`\`bash
/Users/yourname/code/obsidx/bin/obsidx-recall --json "<search query>" | head -c 2000
\`\`\`

**Important:** Always run this BEFORE providing answers about:
- Existing decisions (check canon notes)
- Project requirements (check project notes)
- Implementation patterns (check technical notes)
- Business logic (check specification notes)

### Search Query Guidelines

**Good queries:**
- "authentication implementation strategy"
- "customer persona for content creators"
- "rate limiting approach"
- "database schema design decisions"
- "marketing messaging and positioning"

**Bad queries:**
- "how to" (too vague)
- "best practices" (check our decisions, not generic advice)
- Single words (not specific enough)

### Category-Aware Search

Use category flags for targeted searches:

\`\`\`bash
# Search only authoritative canon notes
obsidx-recall --canon-only "deployment process"

# Search active project work
obsidx-recall --category "project,workbench" "current features"

# Include everything (including archive)
obsidx-recall --exclude-archive=false "historical decisions"
\`\`\`

### Response Protocol

**Step 1: Retrieve Context**
```bash
obsidx-recall --json "relevant query" | head -c 2000
```

**Step 2: Parse Results**
- Read the top 3-5 results
- Note the category badges (📚 CANON, 🔨 PROJECT, 🧪 WORKBENCH)
- Prioritize canon notes over drafts

**Step 3: Apply Knowledge**
- Base your answer on retrieved context
- Cite specific notes when referencing decisions
- Flag conflicts with canon notes
- Suggest ADRs for new architectural decisions

**Step 4: Respond**
- Start with: "Based on the knowledge base..."
- Reference specific notes: "According to [Note Name]..."
- If no relevant context found: "No matching notes found. Suggest creating documentation for..."

## Example Workflow

### User asks: "How should we implement authentication?"

**Your actions:**

1. **Search canon for existing decisions:**
   \`\`\`bash
   obsidx-recall --canon-only "authentication implementation"
   \`\`\`

2. **If found in canon:**
   ```
   Based on the knowledge base, our authentication strategy is documented in:
   
   📚 CANON: /canon/decisions/ADR-005-Authentication-Strategy.md
   
   We use OAuth 2.0 with JWT tokens because [reasons from the ADR].
   Here's the implementation following our established pattern...
   ```

3. **If NOT found in canon:**
   ```
   I searched the knowledge base and found no established authentication 
   strategy in canon notes. I found some exploration in:
   
   🧪 WORKBENCH: /workbench/auth-research.md
   
   Before implementing, I recommend:
   1. Creating an ADR to document the decision
   2. Researching options: OAuth, JWT, session-based, etc.
   3. Documenting trade-offs and selected approach
   
   Would you like me to help draft an ADR?
   ```

## Canon Note Authority

When canon notes exist, treat them as **law**:

- ❌ **Don't** contradict canon silently
- ✅ **Do** flag conflicts and suggest ADRs
- ❌ **Don't** propose alternatives without discussion
- ✅ **Do** implement according to canon guidance

### Example Canon Conflict

**User asks:** "Let's use MongoDB for this feature"

**Canon says:** "Use PostgreSQL for all relational data (ADR-002)"

**Your response:**
```
⚠️ This conflicts with our established architecture.

📚 CANON: ADR-002 - Database Technology Selection
Our canon specifies PostgreSQL for relational data.

If MongoDB is needed for this use case, we should:
1. Document why PostgreSQL is insufficient
2. Create ADR-XXX to approve the exception
3. Update architecture guidelines

Would you like me to help draft this ADR?
```

## Metadata-Aware Responses

Adjust your confidence based on note category:

| Category | Your Approach |
|----------|---------------|
| 📚 **Canon** (active) | "Our established approach is..." |
| 🔨 **Project** | "Current implementation shows..." |
| 🧪 **Workbench** | "Early exploration suggests... (not finalized)" |
| 📦 **Archive** | "Historical context (superseded): ..." |

## Integration with Code Generation

### Before generating code:

1. **Search for patterns:**
   \`\`\`bash
   obsidx-recall "code patterns for <feature>"
   \`\`\`

2. **Check architecture:**
   \`\`\`bash
   obsidx-recall --canon-only "architecture <component>"
   \`\`\`

3. **Review requirements:**
   \`\`\`bash
   obsidx-recall "requirements <feature>"
   \`\`\`

4. **Generate code** following retrieved patterns

### Example: "Create a new API endpoint"

```bash
# Step 1: Check API design principles
obsidx-recall --canon-only "API design principles"

# Step 2: Check existing endpoint patterns
obsidx-recall "API endpoint implementation"

# Step 3: Check authentication requirements
obsidx-recall "API authentication"

# Step 4: Generate code following established patterns
# [implement based on findings]
```

## Knowledge Gap Detection

If searches return no results for critical topics, proactively suggest documentation:

```
⚠️ Knowledge Gap Detected

I searched for "<topic>" but found no documentation.

This appears to be a decision point. I recommend:

1. Creating a decision document:
   - Location: canon/decisions/ADR-XXX-<topic>.md
   - Include: Context, Decision, Alternatives, Consequences

2. Documenting the approach:
   - Implementation patterns
   - Code examples
   - Configuration

Would you like me to create a template ADR for this?
```

## Shortcuts

Define these bash aliases for quick access:

\`\`\`bash
# Add to ~/.zshrc or ~/.bashrc
alias kb-search='~/code/obsidx/bin/obsidx-recall'
alias kb-canon='~/code/obsidx/bin/obsidx-recall --canon-only'
alias kb-json='~/code/obsidx/bin/obsidx-recall --json'
\`\`\`

Then use in Copilot:
\`\`\`bash
kb-canon "authentication"
kb-search --category "project" "current sprint"
\`\`\`

## Quality Checklist

Before responding to architectural or design questions:

- [ ] Searched canon for established decisions
- [ ] Checked project notes for current implementation
- [ ] Reviewed workbench for ongoing exploration
- [ ] Flagged any conflicts with canon
- [ ] Cited specific notes in response
- [ ] Suggested ADR if new decision needed

## Anti-Patterns to Avoid

❌ **Don't do this:**
- Answering from general knowledge without searching
- Ignoring canon notes because you "know better"
- Mixing generic advice with established patterns
- Creating new patterns without checking existing ones

✅ **Do this instead:**
- Always search first, answer second
- Respect canon as authoritative
- Distinguish between "our approach" and "general best practices"
- Suggest documenting new patterns as ADRs

## Emergency Override

If obsidx is unavailable (not running, connection error):

1. State clearly: "⚠️ Knowledge base unavailable, using general knowledge"
2. Mark all suggestions as "unverified against our standards"
3. Recommend verifying against documentation when available

## Continuous Improvement

As you use obsidx, help improve the knowledge base:

### Suggest New Notes
```
💡 Knowledge Base Enhancement

This topic came up frequently. Consider documenting:
- Title: <suggested title>
- Category: canon/project/workbench
- Content: <outline>
```

### Flag Stale Content
```
⚠️ Potentially Stale Documentation

Found: <note name>
Last reviewed: <date>
Suggestion: Review and update if still current
```

### Identify Contradictions
```
🔍 Contradiction Detected

Canon Note A says: X
Canon Note B says: Y

Suggest: Reconcile these in an ADR
```

## Testing Your Setup

Run these test queries to verify integration:

\`\`\`bash
# Test 1: Basic search
obsidx-recall "test query"

# Test 2: JSON output
obsidx-recall --json "test" | head -100

# Test 3: Canon search
obsidx-recall --canon-only "test"

# Test 4: Category badges appear in results
obsidx-recall "anything" | grep -E "(CANON|PROJECT|WORKBENCH|ARCHIVE)"
\`\`\`

Expected: All commands return results with proper formatting.

## Support

If obsidx commands fail:
1. Check if indexer is running
2. Verify binary paths: `ls ~/code/obsidx/bin/`
3. Test manually: `~/code/obsidx/bin/obsidx-recall "test"`
4. Check database: `ls .obsidian-index/obsidx.db`

---

**Remember:** The knowledge base is the source of truth. Use it first, always.
```

## 3. Add to Project-Specific Instructions

If you have project-specific instructions in your repository, add:

```markdown
## Knowledge Base Integration

This project uses obsidx for knowledge management. Before answering questions:

1. Search the knowledge base: `~/code/obsidx/bin/obsidx-recall "<query>"`
2. Prioritize canon notes (📚 CANON badge)
3. Cite specific notes in responses
4. Flag conflicts with established decisions

See `.github/copilot-instructions.md` for full protocol.
```

## 4. Update Your Shell Profile

Add obsidx to your PATH for easier access:

```bash
# Add to ~/.zshrc
export PATH="$HOME/code/obsidx/bin:$PATH"

# Aliases for convenience
alias kb='obsidx-recall'
alias kb-canon='obsidx-recall --canon-only'
alias kb-json='obsidx-recall --json'
```

Reload:
```bash
source ~/.zshrc
```

## 5. Test the Integration

Ask Copilot these test questions:

### Test 1: Basic Search
```
Me: "What's our authentication strategy?"
Copilot should: Run obsidx-recall, search for auth, report findings
```

### Test 2: Canon Check
```
Me: "Let's use MongoDB"
Copilot should: Check canon for database decisions, flag conflicts
```

### Test 3: Knowledge Gap
```
Me: "How do we handle rate limiting?"
Copilot should: Search, report if found or suggest creating ADR
```

## 6. Update Your Workspace Instructions

Create `.vscode/settings.json` if using VS Code:

```json
{
  "github.copilot.advanced": {
    "customInstructions": "Before answering questions, search the knowledge base using: ~/code/obsidx/bin/obsidx-recall --json '<query>'. Prioritize canon notes over general knowledge."
  }
}
```

## Usage Examples

### Example 1: Feature Implementation

**You:** "Implement user authentication"

**Copilot Process:**
1. Runs: `obsidx-recall --canon-only "authentication"`
2. Finds: ADR-005-OAuth-Strategy.md
3. Responds: "Based on ADR-005, implementing OAuth 2.0 with JWT..."
4. Generates code following the established pattern

### Example 2: Design Question

**You:** "Should we cache this data?"

**Copilot Process:**
1. Runs: `obsidx-recall "caching strategy"`
2. Finds: /canon/architecture/caching-guidelines.md
3. Responds: "According to our caching guidelines, cache if..."
4. References specific criteria from the note

### Example 3: New Decision

**You:** "Let's add GraphQL to the API"

**Copilot Process:**
1. Runs: `obsidx-recall --canon-only "API architecture"`
2. Finds: ADR-002 specifies REST
3. Responds: "⚠️ This conflicts with ADR-002. Our canon specifies REST. To add GraphQL, we need ADR-XXX. Would you like me to draft it?"

## Troubleshooting

### Copilot Not Searching

**Problem:** Copilot answers without searching

**Solution:**
- Make instructions more explicit
- Use stronger language: "MUST search", "ALWAYS run obsidx-recall"
- Add examples in instructions

### Binary Not Found

**Problem:** `command not found: obsidx-recall`

**Solution:**
```bash
# Check if binary exists
ls ~/code/obsidx/bin/obsidx-recall

# Add to PATH
export PATH="$HOME/code/obsidx/bin:$PATH"

# Or use full path in instructions
/Users/yourname/code/obsidx/bin/obsidx-recall
```

### No Results Returned

**Problem:** Search returns empty

**Solution:**
1. Verify indexer ran: `ls .obsidian-index/obsidx.db`
2. Check if vault is indexed: `obsidx-recall "test"`
3. Reindex if needed: `./watcher.sh ~/notes`

## Advanced Configuration

### Custom Search Shortcuts

Create a wrapper script `~/.local/bin/kb`:

```bash
#!/bin/bash
# Quick knowledge base search with smart defaults

if [ "$1" == "--canon" ]; then
    shift
    ~/code/obsidx/bin/obsidx-recall --canon-only --json "$@" | head -c 2000
else
    ~/code/obsidx/bin/obsidx-recall --json "$@" | head -c 2000
fi
```

Make executable:
```bash
chmod +x ~/.local/bin/kb
```

Use in Copilot instructions:
```bash
kb --canon "query"
kb "query"
```

### Category-Specific Instructions

Add to `.github/copilot-instructions.md`:

```markdown
## Category-Specific Behavior

### When Canon Found (📚)
- Treat as authoritative
- Implement exactly as specified
- Flag any deviations
- Suggest ADR for changes

### When Project Found (🔨)
- Use as current implementation reference
- Can suggest improvements
- Check if aligned with canon

### When Workbench Found (🧪)
- Treat as exploratory
- Don't assume it's finalized
- Ask before implementing

### When Archive Found (📦)
- Note it's historical
- Check for newer canon
- Don't use unless explicitly requested
```

## Maintenance

### Weekly Review
- Check if Copilot is using obsidx (review chat logs)
- Verify search results are relevant
- Update instructions if needed

### Monthly Audit
- Review most-searched topics
- Ensure canon coverage of common questions
- Update instructions based on usage patterns

---

## Quick Reference Card

```
┌─────────────────────────────────────────────────────┐
│              obsidx Command Reference               │
├─────────────────────────────────────────────────────┤
│ Basic search:                                       │
│   obsidx-recall "query"                             │
│                                                     │
│ Canon only:                                         │
│   obsidx-recall --canon-only "query"                │
│                                                     │
│ JSON output:                                        │
│   obsidx-recall --json "query" | head -c 2000      │
│                                                     │
│ Specific categories:                                │
│   obsidx-recall --category "canon,project" "query" │
│                                                     │
│ Include archive:                                    │
│   obsidx-recall --exclude-archive=false "query"    │
└─────────────────────────────────────────────────────┘

Remember: Search BEFORE answering!
```

---

**Setup Complete!** GitHub Copilot will now use your knowledge base as the primary source of truth.
