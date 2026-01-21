# obsidx Documentation

Complete documentation for obsidx - semantic search for your Obsidian vault with AI-assisted development.

## Quick Start

**New user?** Start here:
1. Read the main [README.md](../README.md) in the root
2. Follow the quick start to get running
3. Come back here for detailed guides

## User Guides

### Core Functionality

**[RETRIEVAL.md](RETRIEVAL.md)** - Search Your Vault
- How to search effectively
- Query syntax and filters
- Category-based search (canon/project/workbench/archive)
- Performance tips

**[OBSIDIAN-FRONTMATTER.md](OBSIDIAN-FRONTMATTER.md)** - Organize with Metadata
- Frontmatter fields explained
- Category system (canon/project/workbench/archive)
- Knowledge lifecycle (draft → active → canon → archive)
- ADR (Architecture Decision Record) template
- Search examples by category

**[MODES.md](MODES.md)** - Running obsidx
- Daemon mode (background processes)
- Foreground mode (with live logs)
- Interactive mode (search prompt)
- Process management

### AI Integration

**[COPILOT.md](COPILOT.md)** - GitHub Copilot Integration
- Quick 2-minute setup
- How Copilot uses your vault
- Search commands and protocols
- Canon authority and conflict resolution
- Code generation workflows
- Troubleshooting and testing

## Documentation Map

```
User Journey:
1. Install & Setup → ../README.md (root)
2. Choose Mode → MODES.md
3. Search → RETRIEVAL.md
4. Organize → OBSIDIAN-FRONTMATTER.md
5. AI Integration → COPILOT.md
```

## Quick Reference

### Common Tasks

**Start obsidx:**
```bash
./start-daemon.sh ~/notes
```

**Search:**
```bash
./bin/obsidx-recall "your query"
```

**Search canon only:**
```bash
./bin/obsidx-recall --canon-only "query"
```

**Stop obsidx:**
```bash
./stop-daemon.sh
```

**Watch logs:**
```bash
tail -f .obsidian-index/indexer.log
tail -f .obsidian-index/recall-server.log
```

### File Organization

Add to your notes:
```yaml
---
category: canon        # canon | project | workbench | archive
scope: mycompany      # your organization/project
type: decision        # decision | principle | vision | spec | note | log
status: active        # active | draft | superseded | deprecated
---
```

### Search Filters

```bash
# Canon only (authoritative)
--canon-only

# Specific categories
--category "canon,project"

# Include archive
--exclude-archive=false

# JSON output
--json

# Quiet mode
--verbose=false
```

## Documentation Status

| Document | Lines | Purpose | Audience |
|----------|-------|---------|----------|
| README.md (root) | ~700 | Project overview, setup | All users |
| MODES.md | ~200 | Running modes | All users |
| RETRIEVAL.md | ~600 | Search guide | All users |
| OBSIDIAN-FRONTMATTER.md | ~500 | Frontmatter & organization | All users |
| COPILOT.md | ~400 | AI integration | Developers |

**Last Updated:** January 20, 2026

## Getting Help

1. **Search issues:** Check logs for errors
2. **Check logs:** `tail -f .obsidian-index/*.log`
3. **Read search guide:** See RETRIEVAL.md for advanced usage
4. **Check frontmatter:** See OBSIDIAN-FRONTMATTER.md for organization tips

## Contributing

When adding documentation:
- User-facing guides go in docs/
- Keep docs concise and actionable
- Update this README when adding new docs
- Use examples liberally

---

**Philosophy:** obsidx keeps your knowledge organized and makes AI assistants use YOUR decisions, not generic advice.
