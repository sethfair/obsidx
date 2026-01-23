# Tag Format Reference

## Supported Formats

ObsIDX supports **both** of these tag formats in YAML frontmatter. Use whichever you prefer:

### Format 1: Inline with Hashtags (Obsidian-style)

```yaml
---
tags: #permanent-note #customer-research #writerflow
---
```

**Pros:**
- Matches Obsidian's inline tag syntax
- More compact
- Easier to type
- Familiar to Obsidian users

### Format 2: Array (YAML standard)

```yaml
---
tags: [permanent-note, customer-research, writerflow]
---
```

**Pros:**
- Standard YAML array syntax
- Works with YAML parsers
- Clean, structured format

## Both Work Identically

The parser normalizes tags internally:
- Removes `#` prefix if present
- Splits on spaces, commas, or tabs
- Strips `[]` brackets if present
- Trims whitespace

## Examples

All of these are equivalent:

```yaml
# Inline with hashtags
tags: #permanent-note #customer-research

# Array format
tags: [permanent-note, customer-research]

# Array with hashtags (also works!)
tags: [#permanent-note, #customer-research]

# Comma-separated inline
tags: #permanent-note, #customer-research

# Comma-separated array
tags: [permanent-note, customer-research]
```

Result: `["permanent-note", "customer-research"]`

## Weight Configuration

In `.obsidian-index/weights.json`, **always use tags without `#` prefix**:

```json
{
  "tag_weights": [
    {"tag": "permanent-note", "weight": 1.3},
    {"tag": "customer-research", "weight": 1.25}
  ]
}
```

The system will match these weights regardless of which format you use in your notes.

## Search Results

Search results show tags **without `#` prefix**:

```
[1] Score: 0.8532 [permanent-note, customer-research]
Path: insights/customer-validation.md
```

## Recommendations

**For Obsidian users:**
- Use inline hashtag format: `tags: #permanent-note #customer-research`
- Matches your existing workflow
- Works with Obsidian's tag pane
- Faster to type

**For YAML purists:**
- Use array format: `tags: [permanent-note, customer-research]`
- Standard YAML syntax
- Works with any YAML parser
- More portable

**For teams:**
- Pick one format and be consistent
- Document your choice in your vault's README
- Consider Obsidian users' familiarity with hashtags

## Testing Your Format

Create a test note and index it:

```bash
# Create test note
cat > /tmp/test-tags.md <<EOF
---
tags: #permanent-note #test
---
# Test Note
Content here
EOF

# Index and search
./bin/obsidx-indexer --vault /tmp
./bin/obsidx-recall --verbose "test"
```

Check that:
1. Tags appear in search results: `[permanent-note, test]`
2. Weight is applied (check score with --verbose)
3. Both tags are recognized

## Troubleshooting

### Tags Not Recognized

**Problem:** Tags not showing in search results

**Check:**
1. **Format:** Ensure one of the supported formats above
2. **Spacing:** Check for correct spacing after `tags:`
3. **Frontmatter:** Ensure `---` delimiters are present
4. **Rebuild:** Delete `.obsidian-index/obsidx.db` and reindex

**Common mistakes:**
```yaml
# ❌ WRONG: No space after tags:
tags:#permanent-note

# ✅ CORRECT:
tags: #permanent-note

# ❌ WRONG: Newlines in array (not supported)
tags:
  - permanent-note
  - customer-research

# ✅ CORRECT (inline array):
tags: [permanent-note, customer-research]
```

### Weights Not Applied

**Problem:** All notes have same priority

**Check:**
1. Tags in config match tags in notes (case-sensitive)
2. No `#` in config file: `"tag": "permanent-note"` not `"tag": "#permanent-note"`
3. Config file exists at `.obsidian-index/weights.json`
4. Reindexed after creating config

### Mixed Formats

**Question:** Can I use different formats in different notes?

**Answer:** Yes! The system handles both formats seamlessly. Each note can use whichever format you prefer.

## Format Migration

If you want to switch formats:

### Inline to Array

```bash
# Convert inline hashtags to array format
find ~/vault -name "*.md" | while read file; do
    sed -i.bak 's/^tags: \(.*\)$/tags: [\1]/' "$file"
    # Remove # symbols
    sed -i.bak 's/\[#/[/g; s/ #/ /g' "$file"
done
```

### Array to Inline

```bash
# Convert array to inline hashtags
find ~/vault -name "*.md" | while read file; do
    sed -i.bak 's/^tags: \[\(.*\)\]$/tags: #\1/' "$file"
    # Add # to remaining tags
    sed -i.bak 's/ \([a-z]\)/ #\1/g' "$file"
    # Remove commas
    sed -i.bak 's/,//g' "$file"
done
```

**Note:** Test on a few files first before batch converting!

## Related Documentation

- `docs/TAG-WEIGHTING.md` - Complete weighting system guide
- `docs/MIGRATION-TAGS.md` - Migration from category system
- `docs/OBSIDIAN-FRONTMATTER.md` - Full frontmatter reference
