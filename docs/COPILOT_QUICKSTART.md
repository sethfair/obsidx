# Copilot Integration - Quick Start

## What This Does

Configures GitHub Copilot to search your Obsidian vault (via obsidx) **before** answering questions, ensuring it uses YOUR decisions and documentation instead of generic knowledge.

## Setup (2 minutes)

### Option 1: Global Setup (All Projects)

Copy the template to your home directory:

```bash
cp .github/copilot-instructions.md ~/.github/copilot-instructions.md
```

### Option 2: Per-Project Setup

Copy to each project:

```bash
cp .github/copilot-instructions.md /path/to/your/project/.github/
```

## Verify It Works

Ask Copilot: **"What's our authentication strategy?"**

**Expected behavior:**
1. Copilot runs: `obsidx-recall --canon-only "authentication"`
2. Shows results from your vault
3. Answers based on YOUR notes, not generic advice

If it works: ✅ You're done!

## Common Issues

### "Command not found: obsidx-recall"

**Fix:** Add to PATH

```bash
echo 'export PATH="$HOME/code/obsidx/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

### "No results returned"

**Fix:** Make sure indexer has run

```bash
cd ~/code/obsidx
./run.sh ~/notes
```

Wait for indexing to complete, then try again.

### Copilot not searching

**Fix:** Make instructions more explicit

Edit `.github/copilot-instructions.md` and change:
```markdown
you MUST retrieve → you MUST ALWAYS retrieve
```

## Key Commands for Copilot

These should be in your instructions:

```bash
# Standard search
~/code/obsidx/bin/obsidx-recall --json "query" | head -c 2000

# Canon only (decisions)
~/code/obsidx/bin/obsidx-recall --canon-only "query"

# Specific categories
~/code/obsidx/bin/obsidx-recall --category "canon,project" "query"
```

## Example Interactions

### ✅ Good: Copilot searches first

**You:** "How do we handle rate limiting?"

**Copilot:**
```
Let me check our documentation...
[runs: obsidx-recall --canon-only "rate limiting"]

Based on ADR-003, we use token bucket rate limiting with...
```

### ❌ Bad: Copilot answers without searching

**You:** "How do we handle rate limiting?"

**Copilot:**
```
There are several approaches to rate limiting...
[generic answer without checking your docs]
```

**Fix:** Strengthen instructions to make searching mandatory.

## Advanced Usage

### Create Aliases

Add to `~/.zshrc`:

```bash
alias kb='~/code/obsidx/bin/obsidx-recall'
alias kb-canon='~/code/obsidx/bin/obsidx-recall --canon-only'
alias kb-json='~/code/obsidx/bin/obsidx-recall --json'
```

Then in Copilot instructions, use shorter commands:
```bash
kb-canon "query"
kb-json "query" | head -c 2000
```

### VS Code Settings

Add to `.vscode/settings.json`:

```json
{
  "github.copilot.advanced": {
    "customInstructions": "Search knowledge base first: ~/code/obsidx/bin/obsidx-recall"
  }
}
```

## Testing Checklist

- [ ] Copilot can run obsidx-recall commands
- [ ] Searches return results with category badges
- [ ] Copilot cites specific notes in answers
- [ ] Copilot flags conflicts with canon
- [ ] Copilot suggests ADRs for new decisions

## What Success Looks Like

**Before obsidx integration:**
```
You: "Let's use MongoDB"
Copilot: "Sure! MongoDB is a great choice for..."
```

**After obsidx integration:**
```
You: "Let's use MongoDB"
Copilot: [searches canon]
Copilot: "⚠️ This conflicts with ADR-002 which specifies 
PostgreSQL. To use MongoDB, we need to create ADR-XXX. 
Would you like me to draft it?"
```

## Next Steps

1. ✅ Copy `.github/copilot-instructions.md` to your projects
2. ✅ Test with a few queries
3. ✅ Refine instructions based on Copilot's behavior
4. 📚 Read full guide: [COPILOT_GUIDE.md](COPILOT_GUIDE.md)

## Support

- Full setup guide: [COPILOT_GUIDE.md](COPILOT_GUIDE.md)
- Knowledge governance: [KNOWLEDGE_GOVERNANCE.md](KNOWLEDGE_GOVERNANCE.md)
- Retrieval commands: [RETRIEVAL.md](RETRIEVAL.md)

---

**Remember:** The goal is "Search first, answer second."
