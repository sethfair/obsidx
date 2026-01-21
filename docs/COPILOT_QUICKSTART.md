# Copilot Integration - Quick Start

## What This Does

Configures GitHub Copilot to search your Obsidian vault (via obsidx) **before** answering questions, ensuring it uses YOUR decisions and documentation instead of generic knowledge.

## Setup (2 minutes)

### Step 1: Index Your Vault

```bash
cd ~/code/obsidx
./watcher.sh ~/notes  # Wait for initial indexing to complete
```

### Step 2: Add Copilot Instructions

Choose global (all projects) or per-project setup:

**Option A: Global Setup (All Projects)**

Copy the template to your home directory:

```bash
cp .github/copilot-instructions.md ~/.github/copilot-instructions.md
```

**Option B: Per-Project Setup**

Copy to each project:

```bash
cp .github/copilot-instructions.md /path/to/your/project/.github/
```

### Step 3: Add obsidx to PATH

Make obsidx commands available to Copilot:

```bash
echo 'export PATH="$HOME/code/obsidx/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc

# Verify
which obsidx-recall
```

## Verify It Works

Ask Copilot: **"What's our authentication strategy?"**

**Expected behavior:**
1. Copilot reads `.github/copilot-instructions.md`
2. Executes: `obsidx-recall --canon-only "authentication"`
3. Uses results from your vault to inform its answer
4. Responds based on YOUR notes, not generic advice

If it works: ✅ You're done!

---

## How It Works

GitHub Copilot reads your `.github/copilot-instructions.md` file and follows the instructions to:

1. **Search your knowledge base** using obsidx-recall commands
2. **Retrieve authoritative canon notes** first
3. **Use that context** to inform code generation and answers
4. **Cite specific notes** in responses

**Key Points:**
- Copilot executes shell commands as instructed
- obsidx must be in PATH or use absolute paths
- The instructions file guides Copilot's behavior
- No separate MCP server needed

---

## Common Issues

### "Command not found: obsidx-recall"

**Fix:** Add obsidx to PATH

```bash
# Add to PATH
echo 'export PATH="$HOME/code/obsidx/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc

# Verify
which obsidx-recall
# Should output: /Users/yourname/code/obsidx/bin/obsidx-recall
```

**Alternative:** Use absolute path in instructions

Edit `.github/copilot-instructions.md` and replace:
```bash
obsidx-recall --json "<query>"
```

With:
```bash
/Users/yourname/code/obsidx/bin/obsidx-recall --json "<query>"
```

### "No results returned"

**Fix:** Make sure indexer has run

```bash
cd ~/code/obsidx
./watcher.sh ~/notes
# Wait for "✓ Initial index complete"
```

### Copilot not searching

**Fix:** Make instructions more explicit

Edit `.github/copilot-instructions.md` and change:
```markdown
you MUST retrieve → you MUST ALWAYS retrieve
```

Add emphasis:
```markdown
**IMPORTANT: Before answering ANY question about our codebase,
architecture, or decisions, you MUST search the knowledge base first.**
```

### Wrong results or outdated content

**Fix:** Keep indexer running to auto-index changes

```bash
# Terminal 1: Keep indexer running
cd ~/code/obsidx
./watcher.sh ~/notes

# Terminal 2: Work normally
# Changes are automatically indexed
```

---

## Key Commands for Copilot

These commands should be in your `.github/copilot-instructions.md`:

```bash
# Standard search
obsidx-recall --json "query" | head -c 2000

# Canon only (authoritative decisions)
obsidx-recall --canon-only "query"

# Specific categories
obsidx-recall --category "canon,project" "query"

# Include archive (historical context)
obsidx-recall --exclude-archive=false "query"
```

---

## Example Interactions

### ✅ Good: Copilot searches first

**You:** "How do we handle rate limiting?"

**Copilot:**
```
Let me check our documentation...
[runs: obsidx-recall --canon-only "rate limiting"]

Based on ADR-003 in your knowledge base, you use token bucket 
rate limiting with Redis backing. Here's the implementation...
```

### ❌ Bad: Copilot answers without searching

**You:** "How do we handle rate limiting?"

**Copilot:**
```
There are several approaches to rate limiting...
[generic answer without checking your docs]
```

**Fix:** Strengthen instructions to make searching mandatory.

---

## Advanced Usage

### Create Shell Aliases

Add to `~/.zshrc`:

```bash
alias kb='obsidx-recall'
alias kb-canon='obsidx-recall --canon-only'
alias kb-json='obsidx-recall --json'
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
    "customInstructions": "Search knowledge base first: obsidx-recall"
  }
}
```

---

## Testing Checklist

- [ ] Copilot can run obsidx-recall commands
- [ ] Searches return results with category badges
- [ ] Copilot cites specific notes in answers
- [ ] Copilot flags conflicts with canon
- [ ] Copilot suggests ADRs for new decisions

---

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

---

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
