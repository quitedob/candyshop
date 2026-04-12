package rag

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func loadCorpus(corpusDir string, opts Options) ([]indexedChunk, error) {
	dirInfo, err := os.Stat(corpusDir)
	if err != nil {
		return nil, fmt.Errorf("corpus directory error: %w", err)
	}
	if !dirInfo.IsDir() {
		return nil, fmt.Errorf("corpus path is not a directory: %s", corpusDir)
	}

	files, err := filepath.Glob(filepath.Join(corpusDir, "*.md"))
	if err != nil {
		return nil, fmt.Errorf("failed to list corpus markdown files: %w", err)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no markdown files found in corpus: %s", corpusDir)
	}

	chunks := make([]indexedChunk, 0, len(files)*4)
	for _, file := range files {
		content, readErr := os.ReadFile(file)
		if readErr != nil {
			if opts.StrictCorpus {
				return nil, fmt.Errorf("failed to read %s: %w", file, readErr)
			}
			continue
		}

		fileChunks, parseErr := parseMarkdownToChunks(filepath.Base(file), string(content), opts)
		if parseErr != nil {
			if opts.StrictCorpus {
				return nil, fmt.Errorf("failed to parse %s: %w", file, parseErr)
			}
			continue
		}
		chunks = append(chunks, fileChunks...)
	}

	if len(chunks) == 0 {
		return nil, fmt.Errorf("corpus loaded with zero chunks: %s", corpusDir)
	}
	return chunks, nil
}

func parseMarkdownToChunks(source, markdown string, opts Options) ([]indexedChunk, error) {
	markdown = strings.TrimSpace(markdown)
	if markdown == "" {
		return nil, fmt.Errorf("empty markdown in %s", source)
	}
	meta := resolveSourceMetadata(source)
	domain := sourceDomain(meta.URL)

	var (
		chunks         []indexedChunk
		section        = "General"
		paragraphLines []string
	)

	flushParagraph := func() {
		if len(paragraphLines) == 0 {
			return
		}
		block := strings.TrimSpace(strings.Join(paragraphLines, " "))
		paragraphLines = paragraphLines[:0]
		if block == "" {
			return
		}

		for _, chunkText := range splitLongText(block, opts.ChunkSize, opts.ChunkOverlapWord) {
			chunkText = strings.TrimSpace(chunkText)
			if chunkText == "" {
				continue
			}

			chunks = append(chunks, indexedChunk{
				Source:            source,
				Section:           section,
				Text:              chunkText,
				NormalizedText:    normalizeForSearch(chunkText),
				NormalizedSection: normalizeForSearch(section),
				TokenSet:          toSet(tokenize(chunkText)),
				SourceURL:         meta.URL,
				SourceDomain:      domain,
				IsOfficial:        meta.Official,
			})
		}
	}

	lines := strings.Split(markdown, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			flushParagraph()
			continue
		}
		if isMarkdownHeading(trimmed) {
			flushParagraph()
			section = normalizeHeading(trimmed)
			continue
		}
		paragraphLines = append(paragraphLines, trimmed)
	}
	flushParagraph()

	if len(chunks) == 0 {
		return nil, fmt.Errorf("no parsable content in %s", source)
	}
	return chunks, nil
}

func splitLongText(text string, maxChars, overlapWord int) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	if maxChars <= 0 || len([]rune(text)) <= maxChars {
		return []string{text}
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{text}
	}

	var chunks []string
	start := 0

	for start < len(words) {
		end := start
		charCount := 0

		for end < len(words) {
			wordLen := len([]rune(words[end]))
			if charCount == 0 {
				charCount = wordLen
				end++
				continue
			}
			if charCount+1+wordLen > maxChars {
				break
			}
			charCount += 1 + wordLen
			end++
		}

		if end <= start {
			end = start + 1
		}

		chunks = append(chunks, strings.Join(words[start:end], " "))
		if end >= len(words) {
			break
		}

		nextStart := end - overlapWord
		if nextStart <= start {
			nextStart = end
		}
		if nextStart < 0 {
			nextStart = 0
		}
		start = nextStart
	}

	return chunks
}

func isMarkdownHeading(line string) bool {
	return strings.HasPrefix(line, "#")
}

func normalizeHeading(line string) string {
	trimmed := strings.TrimLeft(line, "#")
	trimmed = strings.TrimSpace(trimmed)
	if trimmed == "" {
		return "General"
	}
	return trimmed
}
