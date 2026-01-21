# obsidx Documentation

**Quick navigation for all documentation.**

## Getting Started

- **[../README.md](../README.md)** - Main project README with quick start and overview
- **[Installation & First Run](../INSTALL.md)** - How to install and run obsidx

## Core Concepts

- **[KNOWLEDGE_GOVERNANCE.md](KNOWLEDGE_GOVERNANCE.md)** - How to structure and govern your knowledge base
  - Category system (canon/project/workbench/archive)
  - Metadata schema and lifecycle management
  - ADR pattern for architectural decisions
  - Promotion workflow from draft to canon

## Database & Storage

**Database Location:** `.obsidian-index/obsidx.db` (SQLite database in current directory)

obsidx uses **SQLite** for local storage - there is no server to run. The database file is created automatically when you run the indexer:

```bash
# Database is created in current directory as:
.obsidian-index/
├── obsidx.db           # SQLite database (chunks + metadata)
└── hnsw/               # HNSW index files (optional, rebuilt from SQLite)
```

**Key Points:**
- **No server process** - SQLite is embedded in the binaries
- **Portable** - Copy `.obsidian-index/` to move your index
- **Local storage** - All data stays on your machine
- **Single writer** - Only run one indexer per vault at a time
- **Multiple readers** - Can run many search queries concurrently

**Default Location:** Current working directory  
**Custom Location:** Use `--db /path/to/obsidx.db` flag with any command

## Usage Guides

- **[RETRIEVAL.md](RETRIEVAL.md)** - How to search and retrieve knowledge
  - Search commands and filters
  - Category-aware retrieval
  - Understanding weights and ranking
  - Output format and badges

## AI Integration

- **[COPILOT_QUICKSTART.md](COPILOT_QUICKSTART.md)** - 2-minute setup for GitHub Copilot
- **[COPILOT_GUIDE.md](COPILOT_GUIDE.md)** - Complete AI integration reference

## Documentation Status

| Document | Lines | Purpose | Audience |
|----------|-------|---------|----------|
| README.md (root) | ~660 | Project overview, quick start | All users |
| KNOWLEDGE_GOVERNANCE.md | ~500 | Metadata system, lifecycle | Knowledge managers |
| RETRIEVAL.md | ~580 | Search and filtering | End users |
| COPILOT_QUICKSTART.md | ~180 | Fast Copilot setup | Developers |
| COPILOT_GUIDE.md | ~600 | Complete AI integration | Power users |

---

**Documentation Philosophy:**
- **Quick start first** - Get running in minutes
- **Depth when needed** - Reference guides for power users
- **No redundancy** - Each doc has a single responsibility
