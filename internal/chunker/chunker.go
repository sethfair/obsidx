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
