package metadata

import (
	"strings"
	"time"

	"github.com/sethfair/obsidx/internal/config"
)

// NoteMetadata represents extracted front matter
type NoteMetadata struct {
	Scope        string // mycompany, myproject, personal, etc.
	Type         string // decision, principle, vision, spec, note, log, glossary
	Status       string // active, draft, superseded, deprecated
	LastReviewed time.Time
	Tags         []string
}

// CalculateWeight returns the retrieval weight using the provided config
func (m *NoteMetadata) CalculateWeight(weightConfig *config.WeightConfig) float32 {
	// If no config provided, use default weights
	if weightConfig == nil {
		weightConfig = config.DefaultWeightConfig()
	}

	return weightConfig.CalculateWeight(m.Tags, m.Status)
}

// IsActive returns whether this note should be included in standard retrieval
func (m *NoteMetadata) IsActive() bool {
	// Exclude superseded/deprecated
	if m.Status == "superseded" || m.Status == "deprecated" {
		return false
	}

	// Check if archive tag is present
	for _, tag := range m.Tags {
		normalized := strings.ToLower(strings.TrimPrefix(tag, "#"))
		if normalized == "archive" || normalized == "archived" {
			return false
		}
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
			// Support multiple tag formats:
			// 1. Inline with hashtags: tags: #permanent-note #customer-research
			// 2. Array format: tags: [permanent-note, customer-research]
			// 3. Array with hashtags: tags: [#permanent-note, #customer-research]

			value = strings.Trim(value, "[]")

			// Split on common separators (space, comma)
			for _, tag := range strings.FieldsFunc(value, func(r rune) bool {
				return r == ',' || r == ' ' || r == '\t'
			}) {
				tag = strings.TrimSpace(tag)
				// Remove # prefix if present (normalize to no prefix)
				tag = strings.TrimPrefix(tag, "#")
				if tag != "" {
					meta.Tags = append(meta.Tags, tag)
				}
			}
		}
	}

	return meta
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
