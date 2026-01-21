package metadata

import (
	"strings"
	"time"
)

// NoteMetadata represents extracted front matter
type NoteMetadata struct {
	Category     string // canon, project, workbench, archive
	Scope        string // writerflow, ventio, personal, etc.
	Type         string // decision, principle, vision, spec, note, log, glossary
	Status       string // active, draft, superseded, deprecated
	LastReviewed time.Time
	Tags         []string

	// Derived fields
	InferredCategory string // from folder structure if no explicit category
}

// CategoryWeight returns the retrieval weight for a category
func (m *NoteMetadata) CategoryWeight() float32 {
	switch m.EffectiveCategory() {
	case "canon":
		return 1.20
	case "project":
		return 1.05
	case "workbench":
		return 0.90
	case "archive":
		return 0.60
	default:
		return 1.0
	}
}

// StatusWeight returns the retrieval weight for a status
func (m *NoteMetadata) StatusWeight() float32 {
	switch m.Status {
	case "active":
		return 1.0
	case "draft":
		return 0.90
	case "superseded", "deprecated":
		return 0.50
	default:
		return 1.0
	}
}

// CombinedWeight returns the combined retrieval weight
func (m *NoteMetadata) CombinedWeight() float32 {
	return m.CategoryWeight() * m.StatusWeight()
}

// EffectiveCategory returns the category to use (explicit or inferred)
func (m *NoteMetadata) EffectiveCategory() string {
	if m.Category != "" {
		return m.Category
	}
	if m.InferredCategory != "" {
		return m.InferredCategory
	}
	return "project" // default
}

// IsActive returns whether this note should be included in standard retrieval
func (m *NoteMetadata) IsActive() bool {
	// Exclude archived notes by default
	if m.EffectiveCategory() == "archive" {
		return false
	}
	// Exclude superseded/deprecated
	if m.Status == "superseded" || m.Status == "deprecated" {
		return false
	}
	return true
}

// ParseFrontMatter extracts YAML front matter from markdown
func ParseFrontMatter(markdown string) *NoteMetadata {
	meta := &NoteMetadata{
		Status: "active", // default
	}

	lines := strings.Split(markdown, "\n")
	if len(lines) < 3 {
		return meta
	}

	// Check for YAML front matter (---...---)
	if !strings.HasPrefix(lines[0], "---") {
		return meta
	}

	endIndex := -1
	for i := 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "---") || strings.HasPrefix(lines[i], "...") {
			endIndex = i
			break
		}
	}

	if endIndex == -1 {
		return meta
	}

	// Parse YAML-like key: value pairs (simple parser)
	for i := 1; i < endIndex; i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		value = strings.Trim(value, `"'`)

		switch key {
		case "category", "tier":
			meta.Category = normalizeCategory(value)
		case "scope":
			meta.Scope = value
		case "type":
			meta.Type = value
		case "status":
			meta.Status = normalizeStatus(value)
		case "last_reviewed", "lastReviewed":
			if t, err := time.Parse("2006-01-02", value); err == nil {
				meta.LastReviewed = t
			}
		case "tags":
			// Simple tag parsing (comma or space separated)
			value = strings.Trim(value, "[]")
			for _, tag := range strings.FieldsFunc(value, func(r rune) bool {
				return r == ',' || r == ' '
			}) {
				tag = strings.TrimSpace(tag)
				if tag != "" {
					meta.Tags = append(meta.Tags, tag)
				}
			}
		}
	}

	return meta
}

// InferCategoryFromPath attempts to infer category from folder structure
func InferCategoryFromPath(path string) string {
	lower := strings.ToLower(path)

	// Check for explicit folders
	if strings.Contains(lower, "/canon/") || strings.Contains(lower, "/@canon/") {
		return "canon"
	}
	if strings.Contains(lower, "/archive/") {
		return "archive"
	}
	if strings.Contains(lower, "/workbench/") || strings.Contains(lower, "/drafts/") {
		return "workbench"
	}
	if strings.Contains(lower, "/projects/") {
		return "project"
	}

	return "" // no inference
}

func normalizeCategory(cat string) string {
	cat = strings.ToLower(strings.TrimSpace(cat))
	switch cat {
	case "canon", "canonical":
		return "canon"
	case "project", "projects":
		return "project"
	case "workbench", "draft", "drafts", "wip":
		return "workbench"
	case "archive", "archived":
		return "archive"
	default:
		return cat
	}
}

func normalizeStatus(status string) string {
	status = strings.ToLower(strings.TrimSpace(status))
	switch status {
	case "active", "live", "current":
		return "active"
	case "draft", "wip", "in-progress":
		return "draft"
	case "superseded", "replaced":
		return "superseded"
	case "deprecated", "obsolete":
		return "deprecated"
	default:
		return status
	}
}
