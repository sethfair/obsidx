# GitHub Copilot Integration Guide

Complete guide to integrating GitHub Copilot with obsidx for knowledge-aware AI assistance.

## Overview

Configure GitHub Copilot to search your Obsidian vault **before** answering questions or generating code. This ensures Copilot uses YOUR documented decisions, patterns, and knowledge instead of generic advice.

**Benefits:**
- ✅ Copilot respects your architectural decisions
- ✅ Generates code following your established patterns
- ✅ Flags conflicts with canon documentation
- ✅ Suggests ADRs for new decisions
- ✅ Cites specific notes in responses

---

## Quick Start (2 Minutes)

### 1. Start obsidx

```bash
cd ~/code/obsidx
./start-daemon.sh ~/notes
```

Wait for indexing to complete.

### 2. Add Copilot Instructions

**Global (all projects):**
```bash
cp .github/copilot-instructions.md ~/.github/copilot-instructions.md
```

**Per-project:**
```bash
cp .github/copilot-instructions.md /path/to/project/.github/
```

### 3. Add to PATH

```bash
echo 'export PATH="$HOME/code/obsidx/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

### 4. Verify

Ask Copilot: **"What's our authentication strategy?"**

**Expected:**
- Copilot runs `obsidx-recall --canon-only "authentication"`
- Returns answer based on YOUR notes
- Cites specific documents

✅ If it works, you're done!

---

## How It Works

### The Flow

1. **Copilot reads** `.github/copilot-instructions.md`
2. **Instructions tell Copilot** to search your vault first
3. **Copilot executes** `obsidx-recall` commands
4. **Results inform** Copilot's responses
5. **You get answers** based on YOUR knowledge

### Key Points

- Copilot executes shell commands as instructed
- obsidx must be in PATH (or use absolute paths)
- The instructions file guides Copilot's behavior
- Search results are injected as context
- No separate server or API needed

---

## Search Commands

### Basic Search

```bash
# Standard search
obsidx-recall "your query"

# Canon only (authoritative)
obsidx-recall --canon-only "query"

# Specific categories
obsidx-recall --category "canon,project" "query"

# Include archive
obsidx-recall --exclude-archive=false "query"

# JSON output (for parsing)
obsidx-recall --json "query" | head -c 2000
```

### Query Guidelines

**Good queries:**
- "authentication implementation strategy"
- "rate limiting approach"
- "database schema design decisions"
- "customer persona for content creators"

**Bad queries:**
- "how to" (too vague)
- "best practices" (too generic)
- Single words (not specific enough)

---

## Response Protocol

### When Copilot Should Search

Always search before answering questions about:
- ✅ Architectural decisions
- ✅ Design patterns
- ✅ Product requirements
- ✅ Technical implementation
- ✅ Business logic
- ✅ API contracts
- ✅ Customer personas
- ✅ Marketing strategy

### How Copilot Should Respond

**Step 1: Search**
```bash
obsidx-recall --json "relevant query" | head -c 2000
```

**Step 2: Parse results**
- Read top 3-5 results
- Note category badges (📚 CANON, 🔨 PROJECT, 🧪 WORKBENCH)
- Prioritize canon notes

**Step 3: Respond**
- Start with: "Based on the knowledge base..."
- Cite specific notes: "According to ADR-005..."
- Flag conflicts with canon
- Suggest ADRs for new decisions

---

## Example Workflows

### Example 1: Found in Canon

**User asks:** "How should we implement authentication?"

**Copilot searches:**
```bash
obsidx-recall --canon-only "authentication implementation"
```

**Copilot responds:**
```
Based on the knowledge base:

📚 CANON: /canon/decisions/ADR-005-Authentication-Strategy.md

We use OAuth 2.0 with JWT tokens because [reasons from ADR].
Here's the implementation following our established pattern...

[generates code matching documented approach]
```

### Example 2: Not Found - Suggest Documentation

**User asks:** "How should we handle caching?"

**Copilot searches:**
```bash
obsidx-recall --canon-only "caching strategy"
```

**No results. Copilot responds:**
```
⚠️ No caching strategy found in canon notes.

Before implementing, I recommend:
1. Creating ADR-XXX to document the caching decision
2. Defining cache invalidation strategy
3. Selecting cache technology (Redis, Memcached, etc.)

Would you like me to help draft an ADR?
```

### Example 3: Conflict with Canon

**User asks:** "Let's use MongoDB for this feature"

**Copilot searches and finds:**
```
📚 CANON: ADR-002 - Database Technology Selection
"Use PostgreSQL for all relational data"
```

**Copilot responds:**
```
⚠️ This conflicts with our established architecture.

📚 CANON: ADR-002 specifies PostgreSQL for relational data.

If MongoDB is needed for this use case, we should:
1. Document why PostgreSQL is insufficient
2. Create ADR-XXX to approve the exception
3. Update architecture guidelines

Would you like me to help draft this ADR?
```

---

## Canon Authority

When canon notes exist, treat them as **authoritative**:

- ❌ Don't contradict canon silently
- ✅ Flag conflicts and suggest ADRs
- ❌ Don't propose alternatives without discussion
- ✅ Implement according to canon guidance

### Confidence by Category

| Category | Copilot's Approach |
|----------|-------------------|
| 📚 **Canon** | "Our established approach is..." |
| 🔨 **Project** | "Current implementation shows..." |
| 🧪 **Workbench** | "Early exploration suggests... (not finalized)" |
| 📦 **Archive** | "Historical context (superseded): ..." |

---

## Integration with Code Generation

### Before Generating Code

1. **Search for patterns:**
   ```bash
   obsidx-recall "code patterns for <feature>"
   ```

2. **Check architecture:**
   ```bash
   obsidx-recall --canon-only "architecture <component>"
   ```

3. **Review requirements:**
   ```bash
   obsidx-recall "requirements <feature>"
   ```

4. **Generate code** following retrieved patterns

### Example: Creating an API Endpoint

```bash
# Step 1: Check API design principles
obsidx-recall --canon-only "API design principles"

# Step 2: Check existing patterns
obsidx-recall "API endpoint implementation"

# Step 3: Check authentication
obsidx-recall "API authentication"

# Step 4: Generate code following patterns
```

---

## Troubleshooting

### "Command not found: obsidx-recall"

**Fix:** Add to PATH or use absolute path

```bash
# Add to PATH
echo 'export PATH="$HOME/code/obsidx/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc

# Or use absolute path in instructions
/Users/yourname/code/obsidx/bin/obsidx-recall "query"
```

### "No results returned"

**Fix:** Ensure obsidx is running

```bash
./start-daemon.sh ~/notes
```

### Copilot Not Searching

**Fix:** Make instructions more explicit

Edit `.github/copilot-instructions.md`:
```markdown
**IMPORTANT: Before answering ANY question about our codebase,
architecture, or decisions, you MUST search the knowledge base first.**
```

### Wrong/Outdated Results

**Fix:** Indexer automatically re-indexes changes

Check logs:
```bash
tail -f .obsidian-index/indexer.log
```

---

## Advanced Usage

### Shell Aliases

Add to `~/.zshrc`:
```bash
alias kb='obsidx-recall'
alias kb-canon='obsidx-recall --canon-only'
alias kb-json='obsidx-recall --json'
```

Use in instructions:
```bash
kb-canon "authentication"
kb "current sprint features"
```

### VS Code Settings

Add to `.vscode/settings.json`:
```json
{
  "github.copilot.advanced": {
    "customInstructions": "Search knowledge base first: obsidx-recall"
  }
}
```

### Knowledge Gap Detection

If searches return no results for important topics, Copilot should suggest:

```
⚠️ Knowledge Gap Detected

I searched for "<topic>" but found no documentation.

I recommend:
1. Creating ADR-XXX-<topic>.md in canon/decisions/
2. Documenting: Context, Decision, Alternatives, Consequences
3. Adding implementation examples

Would you like me to create a template ADR?
```

---

## Quality Checklist

Before Copilot responds to design questions:

- [ ] Searched canon for established decisions
- [ ] Checked project notes for current implementation
- [ ] Reviewed workbench for ongoing exploration
- [ ] Flagged any conflicts with canon
- [ ] Cited specific notes in response
- [ ] Suggested ADR if new decision needed

---

## What Success Looks Like

### Before obsidx Integration

```
You: "Let's use MongoDB"
Copilot: "Sure! MongoDB is a great choice for..."
```

**Problem:** Generic advice, no context

### After obsidx Integration

```
You: "Let's use MongoDB"
Copilot: [searches canon]
Copilot: "⚠️ This conflicts with ADR-002 which specifies 
PostgreSQL. To use MongoDB, we need to create ADR-XXX. 
Would you like me to draft it?"
```

**Success:** Context-aware, respects your decisions

---

## Anti-Patterns to Avoid

### ❌ Don't Do This

- Answer from general knowledge without searching
- Ignore canon notes because you "know better"
- Mix generic advice with established patterns
- Create new patterns without checking existing ones
- Silently contradict documented decisions

### ✅ Do This Instead

- Always search first, answer second
- Respect canon as authoritative
- Distinguish "our approach" vs "general best practices"
- Suggest documenting new patterns as ADRs
- Flag conflicts clearly and suggest resolution

---

## Emergency Override

If obsidx is unavailable:

```
⚠️ Knowledge base unavailable, using general knowledge

[answer with disclaimer]

Recommend verifying against documentation when available.
```

---

## Testing

### Verification Checklist

- [ ] Copilot can execute obsidx-recall
- [ ] Searches return results with category badges
- [ ] Copilot cites specific notes
- [ ] Copilot flags conflicts with canon
- [ ] Copilot suggests ADRs for new decisions

### Test Queries

Try these with Copilot:

1. "What's our authentication strategy?" → Should search canon
2. "How do we handle rate limiting?" → Should find relevant docs
3. "Let's use a different database" → Should check canon first
4. "Implement feature X" → Should check requirements/patterns

---

## Related Documentation

- [RETRIEVAL.md](RETRIEVAL.md) - How to search effectively
- [KNOWLEDGE_GOVERNANCE.md](KNOWLEDGE_GOVERNANCE.md) - Metadata and categories
- [MODES.md](MODES.md) - Running obsidx

---

## Summary

**Goal:** Make Copilot use YOUR knowledge, not generic advice

**Method:** 
1. Add instructions file
2. Copilot searches your vault
3. Responses based on YOUR docs

**Result:**
- ✅ Context-aware code generation
- ✅ Respects your decisions
- ✅ Flags conflicts
- ✅ Suggests documentation gaps

**Remember:** "Search first, answer second."
