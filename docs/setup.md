# Obsidian Knowledge System — Metadata & Lifecycle Setup

This document defines how to structure, govern, and retrieve knowledge in your Obsidian vault using **metadata-driven categories** and **semantic recall**. The goal is to keep your knowledge base coherent at scale, reduce drift, and give AI agents a reliable mental model of “what is true vs. what is exploratory.”

---

## Objectives

* Prevent conceptual drift as the vault grows
* Separate *truth* from *drafts*
* Make retrieval predictable and low-noise
* Enable agents to reason with your knowledge, not merely over it
* Reduce prompt context size by retrieving only the most relevant, authoritative material

---

## Core Concepts

Every note participates in a **knowledge lifecycle** defined by metadata.
Folders are optional; *metadata is authoritative*.

Each note can declare:

```yaml
---
category: canon        # canon | project | workbench | archive
scope: writerflow      # writerflow | ventio | personal | …
type: decision         # vision | principle | decision | spec | note | log | glossary
status: active         # active | draft | superseded | deprecated
last_reviewed: 2026-01-20
tags: [ai, onboarding]
---
```

### Field Semantics

| Field           | Purpose                               |
| --------------- | ------------------------------------- |
| `category`      | Where this note sits in the lifecycle |
| `scope`         | Domain or product this applies to     |
| `type`          | Semantic role of the note             |
| `status`        | Current validity                      |
| `last_reviewed` | Governance hygiene                    |
| `tags`          | Optional topical hints                |

---

## Categories (Lifecycle Tiers)

| Category    | Meaning             | Agent Behavior                                                        |
| ----------- | ------------------- | --------------------------------------------------------------------- |
| `canon`     | Authoritative truth | Treat as law. Do not contradict silently. Propose changes explicitly. |
| `project`   | Active work         | May be edited, extended, refined.                                     |
| `workbench` | Drafts, experiments | Safe to rewrite, discard, refactor.                                   |
| `archive`   | Historical          | Do not drive new decisions unless explicitly requested.               |

Defaults:

* If no front matter exists: `category=project`
* If a note is moved to an “Archive” folder without metadata, infer `category=archive` (optional fallback)

---

## Retrieval Policy

Your recall engine enforces these rules:

1. **Exclude by default**

    * `category=archive`
    * `status IN ('superseded','deprecated')`

2. **Canon-first retrieval**

    * First pass: `category=canon AND status=active`
    * Second pass: `category IN ('project','workbench') AND status IN ('active','draft')`

3. **Rerank with weights**

| Condition                          | Multiplier             |
| ---------------------------------- | ---------------------- |
| `category=canon AND status=active` | `* 1.20`               |
| `category=project`                 | `* 1.05`               |
| `category=workbench`               | `* 0.90`               |
| `category=archive`                 | `* 0.60` (or excluded) |
| `status=draft`                     | `* 0.90`               |
| `status=superseded/deprecated`     | exclude                |

4. **Scope awareness**

* If a query has a scope (e.g., `writerflow`), boost matching notes:

    * `* 1.10` for `scope=writerflow`

---

## Authoring Workflow

### New Thinking

* Create in `@workbench/`
* Use `category: workbench`
* Be messy, speculative, incomplete

### Active Work

* Move to `@projects/<name>/`
* Set `category: project`
* Refine, structure, evolve

### Promotion to Canon

Promote when the content is:

* A decision you intend to follow
* A principle you want enforced
* A definition that must remain stable
* An architectural invariant

Steps:

1. Add or move to a note with:

   ```yaml
   category: canon
   status: active
   ```
2. Set `type` appropriately (`decision`, `principle`, etc.)
3. Update `last_reviewed`
4. Replace the source note with a link to the canon note

### Archival

* When work is complete or obsolete:

  ```yaml
  category: archive
  status: deprecated
  ```
* Keep for reference only

---

## ADR Pattern (Decisions)

Create:

```
@canon/decisions/
  ADR-000 - Index.md
  ADR-001 - Use HNSW for Recall.md
```

Template:

```markdown
---
category: canon
scope: writerflow
type: decision
status: active
last_reviewed: 2026-01-20
---

# ADR-001 — Use HNSW for Recall

## Context
…

## Decision
…

## Alternatives Considered
…

## Consequences
…

## Date
2026-01-20
```

All future architectural or product decisions become ADRs.

---

## Tooling Requirements

Your Go indexer must:

1. Parse YAML front matter
2. Store extracted fields per chunk:

    * `category`
    * `scope`
    * `type`
    * `status`
    * `meta_json` (full blob)
3. Index these fields in SQLite
4. Apply:

    * category-based filtering
    * canon-first retrieval
    * metadata-based reranking

Your recall CLI should support:

```bash
recall --scope writerflow "What are our onboarding principles?"
```

---

## Linting & Hygiene

Add a `obsidx-lint` command:

Rules:

* If `category=canon`:

    * `status` must exist
    * `last_reviewed` must exist
* Validate enums for:

    * `category`
    * `status`
    * `type`
* Warn if `last_reviewed` > 90 days

This keeps canon credible and prevents silent decay.

---

## Agent Policy

Embed these rules in Copilot instructions:

* Always retrieve context before writing.
* Prioritize `category=canon AND status=active`.
* Treat canon as authoritative.
* If a proposal conflicts with canon:

    * Call it out
    * Suggest an ADR
* Use `workbench` for drafts and exploration.
* Do not revive `archive` content without flagging it.

---

This system turns your vault into a governed, evolving knowledge base:

* Canon = truth
* Project = work
* Workbench = thought
* Archive = memory

Everything else—semantic indexing, HNSW, context reduction—becomes far more effective once these lifecycle semantics exist.
