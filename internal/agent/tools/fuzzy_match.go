package tools

import (
	"strings"
)

// normalizeAllWhitespace normalizes all whitespace for fuzzy matching
// Converts tabs to spaces, trims lines, but preserves structure
func normalizeAllWhitespace(s string) string {
	lines := strings.Split(s, "\n")
	normalized := make([]string, len(lines))
	for i, line := range lines {
		// Replace tabs with 4 spaces
		line = strings.ReplaceAll(line, "\t", "    ")
		// Preserve relative indentation but normalize
		normalized[i] = line
	}
	return strings.Join(normalized, "\n")
}

// matchByLineContent attempts to match content line-by-line, ignoring leading whitespace
// Returns the byte offset if found, -1 otherwise
func matchByLineContent(content, target string) int {
	targetLines := strings.Split(target, "\n")
	contentLines := strings.Split(content, "\n")

	if len(targetLines) == 0 || len(targetLines) > len(contentLines) {
		return -1
	}

	// Try to find a sequence of lines that match when trimmed
	for i := 0; i <= len(contentLines)-len(targetLines); i++ {
		match := true
		for j, targetLine := range targetLines {
			contentLine := contentLines[i+j]
			// Compare trimmed versions
			if strings.TrimSpace(contentLine) != strings.TrimSpace(targetLine) {
				match = false
				break
			}
		}
		if match {
			// Calculate byte offset for this match
			offset := 0
			for k := 0; k < i; k++ {
				offset += len(contentLines[k]) + 1 // +1 for newline
			}
			return offset
		}
	}
	return -1
}

// matchResult represents the result of a fuzzy match attempt
type matchResult struct {
	exactMatch    string
	confidence    float64
	byteOffset    int
	matchStrategy string
}

// findBestMatch attempts fuzzy matching with confidence scoring
// Returns nil if no match found above confidence threshold
func findBestMatch(content, target string) *matchResult {
	// 1. Try exact match first
	if idx := strings.Index(content, target); idx != -1 {
		return &matchResult{
			exactMatch:    target,
			confidence:    1.0,
			byteOffset:    idx,
			matchStrategy: "exact",
		}
	}

	// 2. Try with tab normalization
	normalizedTarget := normalizeTabIndicators(target)
	if idx := strings.Index(content, normalizedTarget); idx != -1 {
		return &matchResult{
			exactMatch:    normalizedTarget,
			confidence:    0.95,
			byteOffset:    idx,
			matchStrategy: "tab_normalized",
		}
	}

	// 3. Try line-by-line matching (ignoring leading whitespace)
	if idx := matchByLineContent(content, target); idx != -1 {
		// Extract the actual content at this location
		lines := strings.Split(content, "\n")
		targetLines := strings.Split(target, "\n")
		startLine := strings.Count(content[:idx], "\n")
		endLine := startLine + len(targetLines)
		if endLine <= len(lines) {
			actualMatch := strings.Join(lines[startLine:endLine], "\n")
			return &matchResult{
				exactMatch:    actualMatch,
				confidence:    0.90,
				byteOffset:    idx,
				matchStrategy: "line_content_match",
			}
		}
	}

	// 4. Try with full whitespace normalization (lower confidence)
	normalizedContent := normalizeAllWhitespace(content)
	normalizedTarget = normalizeAllWhitespace(target)
	if idx := strings.Index(normalizedContent, normalizedTarget); idx != -1 {
		return &matchResult{
			exactMatch:    "",
			confidence:    0.80,
			byteOffset:    idx,
			matchStrategy: "whitespace_normalized",
		}
	}

	return nil
}
