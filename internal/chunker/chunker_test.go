package chunker

import "testing"

func TestIsHeadingOnly(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{"single h2 heading", "## Key Workflows", true},
		{"single h1 heading", "# Credit Rollover at 2x Plan Size", true},
		{"heading with trailing newline", "## Meeting Transcript\n", true},
		{"multiple headings no body", "## Summary\n\n### Details", true},
		{"heading with body", "## Core Insight\n\nA 2x rollover cap on monthly AI credits maximizes retention.", false},
		{"body only", "Rollover policy is one of the highest-leverage retention levers.", false},
		{"heading then horizontal rule only", "## Summary\n\n---", true},
		{"heading then real text after rule", "## Summary\n\n---\n\nActual content here.", false},
		{"empty", "", true},
		{"whitespace only", "   \n\t\n", true},
		// Any #-prefixed line counts, not just well-formed headings — a
		// tag-only chunk is just as unembeddable as a heading-only one.
		{"obsidian tag lines only", "#permanent-note #writerflow", true},
		{"heading with trailing hashes", "## Foo ##", true},
		// Markdown horizontal rules allow 3+ marker chars and interior spaces.
		{"long dash rule", "## Summary\n\n----", true},
		{"long asterisk rule", "## Summary\n\n*****", true},
		{"spaced dash rule", "## Summary\n\n- - -", true},
		{"two dashes is not a rule", "## Summary\n\n--", false},
		{"mixed rule chars are not a rule", "## Summary\n\n-*-", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsHeadingOnly(tt.content); got != tt.want {
				t.Errorf("IsHeadingOnly(%q) = %v, want %v", tt.content, got, tt.want)
			}
		})
	}
}
