# GitHub Copilot - Knowledge Base Integration

## Context Retrieval Protocol

Before answering questions or implementing features, you MUST retrieve relevant context from the obsidx knowledge base.

### Command to Search Knowledge Base

```bash
~/code/obsidx/bin/obsidx-recall --json "<search query>" | head -c 2000
```

### When to Search

Always search before:
- Making architectural decisions
- Implementing features
- Answering design questions
- Proposing solutions
- Writing code for existing systems

### Search Examples

```bash
# Check for existing decisions
~/code/obsidx/bin/obsidx-recall --canon-only "authentication"

# Find implementation patterns
~/code/obsidx/bin/obsidx-recall "API design patterns"

# Search project docs
~/code/obsidx/bin/obsidx-recall --category "project" "current features"
```

## Response Protocol

**Step 1:** Search the knowledge base first
**Step 2:** Read top 3-5 results and note categories
**Step 3:** Base answer on retrieved context
**Step 4:** Cite specific notes in your response

### Category Meanings

- 📚 **CANON** - Authoritative, treat as law
- 🔨 **PROJECT** - Current implementation
- 🧪 **WORKBENCH** - Exploratory, not finalized
- 📦 **ARCHIVE** - Historical, likely superseded

## Canon Authority

When canon notes exist:
- ❌ Don't contradict them silently
- ✅ Flag conflicts and suggest ADRs
- ❌ Don't propose alternatives without discussion
- ✅ Implement according to canon guidance

## Example Response

User asks: "How should we implement caching?"

Your process:
```bash
# Search canon first
~/code/obsidx/bin/obsidx-recall --canon-only "caching strategy"
```

If found:
```
Based on the knowledge base:

📚 CANON: /canon/architecture/caching-guidelines.md

Our established caching strategy is [summary from note].

Implementing according to canon guidelines:
[code following the guidelines]
```

If NOT found:
```
I searched the knowledge base and found no established caching strategy.

Before implementing, I recommend:
1. Creating ADR-XXX to document the caching decision
2. Defining cache invalidation strategy
3. Selecting cache technology

Would you like me to help draft an ADR?
```

## Knowledge Gap Detection

If no results found for critical topics, suggest creating documentation:

```
⚠️ Knowledge Gap Detected

No documentation found for "<topic>".

Recommend creating:
- ADR if it's a decision
- Technical doc if it's implementation
- Architecture doc if it's system design

Would you like me to create a template?
```

## Quality Checklist

Before responding:
- [ ] Searched knowledge base
- [ ] Checked canon for conflicts
- [ ] Cited specific notes
- [ ] Distinguished our patterns from generic advice

---

**Remember: Search first, answer second. The knowledge base is the source of truth.**
