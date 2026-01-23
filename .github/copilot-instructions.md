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
# Search all content
~/code/obsidx/bin/obsidx-recall "authentication strategy"

# Find implementation patterns
~/code/obsidx/bin/obsidx-recall "API design patterns"

# Search project docs
~/code/obsidx/bin/obsidx-recall "current features"
```

## Response Protocol

**Step 1:** Search the knowledge base first
**Step 2:** Read top 3-5 results and note their tags
**Step 3:** Base answer on retrieved context
**Step 4:** Cite specific notes in your response

### Tag Interpretation

Results show tags that indicate note importance and type:
- **#permanent-note** - Refined insights, highly authoritative
- **#literature-note** - Curated sources and references
- **#customer-research** - Validated customer insights
- **#vision** - Strategic direction and goals
- **#archive** - Historical, likely superseded

**Higher-weighted tags appear more prominently in search results.**

## Tag-Based Authority

When reviewing search results:
- ✅ Trust highly-weighted tags (#permanent-note, #customer-research, #vision)
- ✅ Prefer notes with relevant project tags (#writerflow, #mvp-1, etc.)
- ⚠️ Be cautious with #fleeting-notes or #draft - may be unrefined
- ❌ #archive or deprecated status - historical reference only

## Example Response

User asks: "How should we implement caching?"

Your process:
```bash
~/code/obsidx/bin/obsidx-recall "caching strategy"
```

If found with #permanent-note:
```
Based on the knowledge base:

Note: /architecture/caching-guidelines.md
Tags: #permanent-note, #architecture-decision

Our established caching strategy is [summary from note].

Implementing according to documented guidelines:
[code following the guidelines]
```

If only #fleeting-notes found:
```
I found a draft note on caching, but no finalized decision.

The draft suggests [summary], but this should be validated.

Recommend:
1. Creating a permanent note with the decided approach
2. Defining cache invalidation strategy
3. Documenting the decision

Would you like me to help formalize this?
```

## Knowledge Gap Detection

If no results found for critical topics, suggest creating documentation:

```
⚠️ Knowledge Gap Detected

No documentation found for "<topic>".

Recommend creating a note with appropriate tags:
- Use #permanent-note for finalized decisions
- Use #draft or #fleeting-notes for work in progress
- Add relevant project tags (#writerflow, etc.)

Would you like me to create a template?
```

## Tag Weight Customization

The system uses configurable tag weights (see docs/TAG-WEIGHTING.md).

Current defaults:
- #permanent-note: 1.3
- #customer-research: 1.25  
- #vision: 1.3
- #literature-note: 1.1
- #fleeting-notes: 0.8
- #archive: 0.6

Users can customize weights via `.obsidian-index/weights.json`.

## Quality Checklist

Before responding:
- [ ] Searched knowledge base
- [ ] Checked tags on top results
- [ ] Prioritized high-weight tags
- [ ] Cited specific notes
- [ ] Distinguished refined notes from drafts

---

**Remember: Search first, answer second. Tag weights indicate source authority.**

