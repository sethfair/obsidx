# Copilot Integration - Quick Start

## What This Does

Configures GitHub Copilot to search your Obsidian vault (via obsidx) **before** answering questions, ensuring it uses YOUR decisions and documentation instead of generic knowledge.

## Two Integration Methods

### Method 1: Copilot CLI Tool (Recommended)

Configure obsidx as an MCP tool that Copilot CLI can call automatically.

**Pros:** Seamless, automatic context retrieval  
**Cons:** Requires GitHub Copilot CLI

### Method 2: Editor Instructions

Use instruction files to tell Copilot to run obsidx commands.

**Pros:** Works with any editor  
**Cons:** Requires manual command execution

---

## Method 1: Copilot CLI Tool Setup

### Step 1: Index Your Vault

```bash
cd ~/code/obsidx
./run.sh ~/notes  # Wait for initial indexing to complete
```

### Step 2: Configure MCP Tool

Edit your Copilot CLI config file:

**macOS:** `~/Library/Application Support/github-copilot-cli/config.json`  
**Linux:** `~/.config/github-copilot-cli/config.json`  
**Windows:** `%APPDATA%\github-copilot-cli\config.json`

Add this configuration:

```json
{
  "mcpServers": {
    "obsidx": {
      "command": "/Users/seth/code/obsidx/bin/obsidx-recall",
      "args": ["--json"],
      "env": {}
    }
  }
}
```

**Note:** Adjust the path to match your obsidx installation.

### Step 3: Restart Copilot CLI

```bash
# Kill any running Copilot CLI processes
pkill -f "github-copilot-cli"

# Test the integration
gh copilot suggest "what is our authentication strategy"
```

### Step 4: Verify It Works

Ask Copilot CLI a question about your documented knowledge:

```bash
gh copilot suggest "how do we handle rate limiting"
```

**Expected behavior:**
- Copilot automatically calls obsidx to search your vault
- Returns results based on your canon notes
- Generates code/answers using YOUR documentation

---

## Method 2: Editor Instructions Setup

### Option A: Global Setup (All Projects)

### Option A: Global Setup (All Projects)

Copy the template to your home directory:

```bash
cp .github/copilot-instructions.md ~/.github/copilot-instructions.md
```

### Option B: Per-Project Setup

Copy to each project:

```bash
cp .github/copilot-instructions.md /path/to/your/project/.github/
```

## Verify Method 2 Works

Ask Copilot: **"What's our authentication strategy?"**

**Expected behavior:**
1. Copilot runs: `obsidx-recall --canon-only "authentication"`
2. Shows results from your vault
3. Answers based on YOUR notes, not generic advice

If it works: ✅ You're done!

---

## Comparison: CLI Tool vs Editor Instructions

| Feature | CLI Tool (Method 1) | Editor Instructions (Method 2) |
|---------|---------------------|--------------------------------|
| **Setup complexity** | Moderate (JSON config) | Simple (copy file) |
| **Automation** | Fully automatic | Requires Copilot to execute commands |
| **Context retrieval** | Seamless | Manual command in instructions |
| **Works with** | Copilot CLI only | Any editor with Copilot |
| **Real-time updates** | Yes | Yes (if indexer running) |
| **Best for** | CLI-heavy workflows | Editor-heavy workflows |

**Recommendation:** Use Method 1 (CLI Tool) if you use `gh copilot` commands. Use Method 2 (Editor Instructions) if you primarily use Copilot in your editor.

## Common Issues

### Method 1 (CLI Tool) Issues

#### "Tool not available" or "obsidx not found"

**Fix 1: Check config file location**

Verify you edited the correct config file:
```bash
# macOS
ls -la ~/Library/Application\ Support/github-copilot-cli/config.json

# Linux
ls -la ~/.config/github-copilot-cli/config.json
```

**Fix 2: Verify JSON syntax**

Your config must be valid JSON. Use a JSON validator or:
```bash
cat ~/Library/Application\ Support/github-copilot-cli/config.json | jq .
```

**Fix 3: Check binary path**

Ensure the path to obsidx-recall is correct:
```bash
ls -la /Users/seth/code/obsidx/bin/obsidx-recall
```

Update config with your actual path if different.

**Fix 4: Restart Copilot CLI**

```bash
# Kill all Copilot CLI processes
pkill -f "github-copilot-cli"

# Try again
gh copilot suggest "test query"
```

#### "Command not found: obsidx-recall"

**Fix:** Add obsidx to PATH or use absolute path in config

```bash
# Option 1: Add to PATH (recommended)
echo 'export PATH="$HOME/code/obsidx/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc

# Option 2: Use absolute path in config.json
{
  "mcpServers": {
    "obsidx": {
      "command": "/Users/seth/code/obsidx/bin/obsidx-recall",
      ...
    }
  }
}
```

#### "No results returned"

**Fix:** Make sure indexer has run

```bash
cd ~/code/obsidx
./run.sh ~/notes
# Wait for "✓ Initial index complete"
```

### Method 2 (Editor Instructions) Issues

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
