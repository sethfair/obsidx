package chunker

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"strings"
)

// Chunk represents a markdown chunk
type Chunk struct {
	HeadingPath string // e.g., "Introduction > Setup > Installation"
	Content     string
	ChunkIndex  int
	StartLine   int
	EndLine     int
	// Metadata inherited from note
	Status         string
	Scope          string
	NoteType       string
	CategoryWeight float32  // Calculated from tags using weight config
	Tags           []string // Tags from frontmatter (e.g., #permanent-note, #writerflow)
}

// IsHeadingOnly reports whether content has no body text: every non-blank
// line is either #-prefixed (ATX headings, but also Obsidian tag lines like
// "#permanent-note #writerflow" — both are equally unembeddable) or a
// horizontal rule. Such chunks carry no embeddable content — nomic-embed-text
// collapses "## Word Word" lines to identical vectors, and thousands of
// identical vectors form cliques in the HNSW graph that trap greedy search
// (observed 2026-07-23: searches could only reach ~200 clique nodes out
// of an 85k-vector index). Callers should skip these chunks at embed time.
//
// Note: lines are trimmed before the # check, so an indented "# shell
// comment" inside an unfenced code block also matches. Accepted trade-off —
// aligning with ChunkMarkdown's untrimmed heading detection is not worth
// keeping such fragments in the index.
func IsHeadingOnly(content string) bool {
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		if isHorizontalRule(trimmed) {
			continue
		}
		return false
	}
	return true
}

// isHorizontalRule matches markdown thematic breaks: 3+ of the same marker
// char (-, *, _), optionally space-separated ("---", "-----", "- - -").
func isHorizontalRule(trimmed string) bool {
	compact := strings.ReplaceAll(trimmed, " ", "")
	if len(compact) < 3 {
		return false
	}
	marker := compact[0]
	if marker != '-' && marker != '*' && marker != '_' {
		return false
	}
	for i := 1; i < len(compact); i++ {
		if compact[i] != marker {
			return false
		}
	}
	return true
}

// ChunkMarkdown splits markdown into semantic chunks
// Strategy: chunk by headings, keeping reasonable size limits
func ChunkMarkdown(markdown string) []Chunk {
	var chunks []Chunk
	lines := strings.Split(markdown, "\n")

	var (
		currentChunk strings.Builder
		currentPath  []string
		chunkStart   int
		lineNum      int
		chunkIndex   int
	)

	const maxChunkSize = 1000 // characters

	flushChunk := func(endLine int) {
		if currentChunk.Len() == 0 {
			return
		}

		chunks = append(chunks, Chunk{
			HeadingPath: strings.Join(currentPath, " > "),
			Content:     strings.TrimSpace(currentChunk.String()),
			ChunkIndex:  chunkIndex,
			StartLine:   chunkStart,
			EndLine:     endLine,
		})

		currentChunk.Reset()
		chunkIndex++
		chunkStart = endLine + 1
	}

	for lineNum = 0; lineNum < len(lines); lineNum++ {
		line := lines[lineNum]

		// Check for markdown heading
		if strings.HasPrefix(line, "#") {
			// Flush previous chunk before starting new section
			if currentChunk.Len() > 0 {
				flushChunk(lineNum - 1)
			}

			// Parse heading level and text
			level := 0
			for i := 0; i < len(line) && line[i] == '#'; i++ {
				level++
			}
			headingText := strings.TrimSpace(line[level:])

			// Update heading path
			if level <= len(currentPath) {
				currentPath = currentPath[:level-1]
			}
			if level == 0 {
				currentPath = nil
			}
			currentPath = append(currentPath, headingText)

			// Add heading to chunk
			currentChunk.WriteString(line)
			currentChunk.WriteByte('\n')
			continue
		}

		// Regular content line
		currentChunk.WriteString(line)
		currentChunk.WriteByte('\n')

		// Split if chunk gets too large
		if currentChunk.Len() > maxChunkSize {
			flushChunk(lineNum)
		}
	}

	// Flush final chunk
	if currentChunk.Len() > 0 {
		flushChunk(lineNum - 1)
	}

	return chunks
}

// ComputeContentHash returns SHA256 hash of content
func ComputeContentHash(content string) string {
	h := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", h)
}

// ReadMarkdownFile reads a file line by line (for large files)
func ReadMarkdownFile(scanner *bufio.Scanner) []string {
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}
