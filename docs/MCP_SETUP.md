# MCP Tool Setup for GitHub Copilot CLI

**Configure obsidx as an automatic tool in GitHub Copilot CLI using the Model Context Protocol (MCP).**

## What is MCP?

The Model Context Protocol (MCP) allows you to configure external tools that GitHub Copilot CLI can automatically invoke. When you ask Copilot a question, it can call obsidx to search your knowledge base and use that context in its response.

## Benefits

- ✅ **Automatic context retrieval** - No manual commands needed
- ✅ **Seamless integration** - Copilot decides when to search your vault
- ✅ **Always up-to-date** - Uses live indexed data from your vault
- ✅ **Canon-first approach** - Copilot respects your authoritative decisions
- ✅ **Natural workflow** - Just ask questions, Copilot handles the rest

## Prerequisites

1. **GitHub Copilot CLI installed**
   ```bash
   gh extension install github/gh-copilot
   ```

2. **obsidx installed and indexed**
   ```bash
   cd ~/code/obsidx
   ./run.sh ~/notes  # Index your vault
   ```

3. **Indexer running in background** (optional but recommended)
   ```bash
   # Keep indexer running to auto-index changes
   ./run.sh ~/notes  # Leave this terminal open
   ```

## Setup Steps

### Step 1: Locate Config File

GitHub Copilot CLI stores MCP server configuration in:

**macOS:**
```
~/Library/Application Support/github-copilot-cli/config.json
```

**Linux:**
```
~/.config/github-copilot-cli/config.json
```

**Windows:**
```
%APPDATA%\github-copilot-cli\config.json
```

### Step 2: Create or Edit Config

If the file doesn't exist, create it:

```bash
# macOS
mkdir -p ~/Library/Application\ Support/github-copilot-cli
touch ~/Library/Application\ Support/github-copilot-cli/config.json
```

### Step 3: Add obsidx Configuration

**Option A: Use example config (easiest)**

Copy the example configuration:

```bash
# macOS
cp ~/code/obsidx/.github/copilot-cli-mcp-config.json \
   ~/Library/Application\ Support/github-copilot-cli/config.json

# Verify path to obsidx-recall
which obsidx-recall
# Update "command" in config.json if path differs
```

**Option B: Add to existing config (if you have other tools)**

Edit the config file and add the obsidx MCP server:

```json
{
  "mcpServers": {
    "obsidx": {
      "command": "/Users/seth/code/obsidx/bin/obsidx-recall",
      "args": ["--json"],
      "env": {},
      "description": "Search Obsidian vault for authoritative decisions and documentation"
    }
  }
}
```

**Important:** Replace `/Users/seth/code/obsidx/bin/obsidx-recall` with your actual path:

```bash
# Find your path
which obsidx-recall
# Or if not in PATH:
echo "$HOME/code/obsidx/bin/obsidx-recall"
```

### Step 4: Validate Configuration

Check that your JSON is valid:

```bash
# macOS/Linux
cat ~/Library/Application\ Support/github-copilot-cli/config.json | jq .

# Should output formatted JSON without errors
```

### Step 5: Restart Copilot CLI

Kill any running Copilot CLI processes:

```bash
pkill -f "github-copilot-cli"
```

### Step 6: Test Integration

Try asking Copilot a question about your documented knowledge:

```bash
gh copilot suggest "what is our authentication strategy"
```

**Expected behavior:**
- Copilot automatically searches your vault using obsidx
- Returns context from your canon/project notes
- Generates response based on YOUR documentation

## Advanced Configuration

### Custom Tool Parameters

You can configure how obsidx is called by adding tool definitions:

```json
{
  "mcpServers": {
    "obsidx": {
      "command": "/Users/seth/code/obsidx/bin/obsidx-recall",
      "args": ["--json"],
      "env": {}
    }
  },
  "tools": {
    "search-canon": {
      "server": "obsidx",
      "command": "obsidx-recall",
      "args": ["--canon-only", "--json", "${query}"],
      "description": "Search only authoritative canon notes"
    },
    "search-all": {
      "server": "obsidx",
      "command": "obsidx-recall",
      "args": ["--exclude-archive=false", "--json", "${query}"],
      "description": "Search all notes including archive"
    }
  }
}
```

### Environment Variables

If obsidx needs specific environment variables:

```json
{
  "mcpServers": {
    "obsidx": {
      "command": "/Users/seth/code/obsidx/bin/obsidx-recall",
      "args": ["--json"],
      "env": {
        "OBSIDX_DB": "/path/to/.obsidian-index/obsidx.db",
        "OBSIDX_VAULT": "/path/to/vault"
      }
    }
  }
}
```

### Multiple Vaults

If you have multiple vaults, configure separate tools:

```json
{
  "mcpServers": {
    "obsidx-work": {
      "command": "/Users/seth/code/obsidx/bin/obsidx-recall",
      "args": ["--db", "/path/to/work/.obsidian-index/obsidx.db", "--json"]
    },
    "obsidx-personal": {
      "command": "/Users/seth/code/obsidx/bin/obsidx-recall",
      "args": ["--db", "/path/to/personal/.obsidian-index/obsidx.db", "--json"]
    }
  }
}
```

## Usage Examples

Once configured, Copilot CLI automatically uses obsidx:

### Architecture Questions

```bash
gh copilot suggest "implement rate limiting for our API"
```

**What happens:**
1. Copilot calls: `obsidx-recall --canon-only "rate limiting"`
2. Finds ADR-003 in your vault
3. Generates implementation matching your documented approach

### Reviewing Existing Decisions

```bash
gh copilot explain "why do we use OAuth instead of JWT"
```

**What happens:**
1. Copilot searches: `obsidx-recall --canon-only "authentication OAuth JWT"`
2. Retrieves your ADR explaining the decision
3. Explains based on YOUR rationale, not generic advice

### Implementation Guidance

```bash
gh copilot suggest "add caching to user profile endpoint"
```

**What happens:**
1. Copilot searches: `obsidx-recall "caching strategy"`
2. Finds your caching guidelines
3. Implements according to your patterns

## Troubleshooting

### "Tool not available" error

**Cause:** Config file not found or invalid JSON

**Fix:**
```bash
# Check if file exists
ls -la ~/Library/Application\ Support/github-copilot-cli/config.json

# Validate JSON
cat ~/Library/Application\ Support/github-copilot-cli/config.json | jq .

# Check for syntax errors
```

### "Command not found: obsidx-recall"

**Cause:** Path to binary is incorrect

**Fix:**
```bash
# Find correct path
which obsidx-recall

# Or use absolute path
echo "$HOME/code/obsidx/bin/obsidx-recall"

# Update config.json with correct path
```

### "No results returned"

**Cause:** Vault not indexed

**Fix:**
```bash
cd ~/code/obsidx
./run.sh ~/notes

# Wait for "✓ Initial index complete"
```

### Copilot doesn't call obsidx

**Cause:** Tool not registered or query not relevant

**Fix:**
1. Verify config is loaded: restart Copilot CLI
   ```bash
   pkill -f "github-copilot-cli"
   ```

2. Make queries more specific:
   - ❌ "help with auth"
   - ✅ "what is our authentication strategy from canon notes"

3. Check Copilot CLI logs (if available)

### JSON parse errors

**Cause:** obsidx-recall not returning valid JSON

**Fix:**
```bash
# Test obsidx manually
obsidx-recall --json "test query"

# Should output valid JSON array
```

## Best Practices

### 1. Keep Indexer Running

Run indexer in watch mode to keep index up-to-date:

```bash
# Terminal 1: Keep indexer running
cd ~/code/obsidx
./run.sh ~/notes

# Terminal 2: Use Copilot CLI
gh copilot suggest "your question"
```

### 2. Use Specific Queries

Help Copilot call obsidx by being specific:

**Good:**
- "what is our database choice and why" ✅
- "show me our API design principles" ✅
- "explain our rate limiting implementation" ✅

**Less Good:**
- "database" ❌ (too vague)
- "help" ❌ (no context)
- "what should I do" ❌ (not specific)

### 3. Structure Your Vault

Use clear category system:
- `@canon/` for authoritative decisions
- `@projects/` for current work
- `@workbench/` for drafts

This helps Copilot find the right information.

### 4. Maintain Canon Notes

Keep your canon notes:
- Up-to-date with `last_reviewed` dates
- Well-structured with clear headings
- Tagged with relevant `scope` and `type`

### 5. Review Copilot's Sources

When Copilot answers based on your vault:
- Verify it used the correct notes
- Check if canon was prioritized
- Update your docs if answers are off

## Comparison: MCP vs Editor Instructions

| Feature | MCP Tool | Editor Instructions |
|---------|----------|---------------------|
| **Automation** | Fully automatic | Manual command execution |
| **Setup** | Moderate (JSON config) | Simple (copy file) |
| **Works with** | Copilot CLI only | Any editor |
| **Context retrieval** | Seamless | Requires Copilot to run commands |
| **Best for** | CLI workflows | Editor workflows |
| **Maintenance** | Config file updates | Instruction file updates |

**Recommendation:**
- Use **MCP** if you primarily use `gh copilot` commands
- Use **Editor Instructions** if you primarily use Copilot in your editor
- Use **Both** for comprehensive coverage

## Next Steps

1. ✅ Configure MCP tool
2. ✅ Test with simple queries
3. ✅ Verify canon notes are used
4. 📚 Read [COPILOT_GUIDE.md](COPILOT_GUIDE.md) for advanced usage
5. 📚 Read [KNOWLEDGE_GOVERNANCE.md](KNOWLEDGE_GOVERNANCE.md) for best practices

---

**Summary:** With MCP configured, GitHub Copilot CLI automatically searches your vault and uses YOUR established knowledge instead of generic advice.
