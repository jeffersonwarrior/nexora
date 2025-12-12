package tools

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strings"
)

// EditRetryStrategy provides auto-healing mechanisms for failed edit operations.
// When an edit fails (e.g., "old_string not found"), it can automatically:
// 1. View the target file to extract exact context
// 2. Use grep to find similar patterns
// 3. Regenerate the old_string with proper whitespace/indentation
type EditRetryStrategy struct {
	ctx    context.Context
	viewFn func(ctx context.Context, filePath string, offset, limit int) (string, error)
	grepFn func(ctx context.Context, pattern string, workingDir string) ([]string, error)
}

// NewEditRetryStrategy creates a new retry strategy for failed edits.
func NewEditRetryStrategy(ctx context.Context) *EditRetryStrategy {
	return &EditRetryStrategy{
		ctx: ctx,
		viewFn: func(ctx context.Context, filePath string, offset, limit int) (string, error) {
			content, err := os.ReadFile(filePath)
			return string(content), err
		},
		grepFn: func(ctx context.Context, pattern string, workingDir string) ([]string, error) {
			return []string{}, nil // Placeholder for grep integration
		},
	}
}

// RetryWithContext attempts to fix a failed edit by extracting context from the file.
// This is called when "old_string not found" error occurs.
func (s *EditRetryStrategy) RetryWithContext(
	filePath string,
	oldString string,
	newString string,
	failureReason string,
) (EditParams, error) {
	// Read the file to find the target pattern
	content, err := os.ReadFile(filePath)
	if err != nil {
		return EditParams{}, fmt.Errorf("failed to read file for retry: %w", err)
	}

	fileContent := string(content)

	// Try to find a similar pattern by removing extra whitespace
	normalized := normalizeWhitespace(oldString)
	if idx := findSimilarPattern(fileContent, normalized); idx >= 0 {
		// Extract context around the match (7 lines of context as per guidelines)
		lines := strings.Split(fileContent, "\n")
		matchLine := countNewlines(fileContent[:idx])
		startLine := max(0, matchLine-3)
		endLine := min(len(lines), matchLine+4)

		// Build the new old_string with proper context
		contextLines := lines[startLine:endLine]
		improvedOldString := strings.Join(contextLines, "\n")

		slog.Debug("self-healing edit",
			"file", filePath,
			"original_length", len(oldString),
			"improved_length", len(improvedOldString),
			"context_lines", len(contextLines),
		)

		return EditParams{
			FilePath:   filePath,
			OldString:  improvedOldString,
			NewString:  newString,
			ReplaceAll: false,
		}, nil
	}

	// Try advanced context extraction
	if improvedParams := s.extractContextFromPattern(fileContent, oldString, newString); improvedParams.OldString != "" {
		return improvedParams, nil
	}

	return EditParams{}, fmt.Errorf("could not recover from edit failure: %s", failureReason)
}

// extractContextFromPattern uses advanced pattern matching to extract better context
func (s *EditRetryStrategy) extractContextFromPattern(fileContent, oldString, newString string) EditParams {
	lines := strings.Split(fileContent, "\n")

	// Look for each line of the old_string in the file
	oldLines := strings.Split(oldString, "\n")

	for lineIdx := 0; lineIdx < len(lines); lineIdx++ {
		// Try to find the start of a matching sequence
		if strings.Contains(lines[lineIdx], oldLines[0]) {
			// Check if this could be the start of our target sequence
			remainingLines := len(lines) - lineIdx
			if remainingLines >= len(oldLines) {
				// Extract a 7-line block around this potential match
				startIdx := max(0, lineIdx-3)
				endIdx := min(len(lines), lineIdx+4)

				contextBlock := lines[startIdx:endIdx]
				contextString := strings.Join(contextBlock, "\n")

				// Check if the old_string pattern appears in this context block
				if strings.Contains(contextString, oldString) {
					slog.Debug("extracted context block",
						"start_line", startIdx,
						"end_line", endIdx,
						"block_length", len(contextBlock),
					)

					return EditParams{
						FilePath:   "", // Will be filled by caller
						OldString:  contextString,
						NewString:  newString,
						ReplaceAll: false,
					}
				}
			}
		}
	}

	return EditParams{} // Return empty if no context found
}

// normalizeWhitespace removes leading/trailing whitespace and collapses multiple spaces.
func normalizeWhitespace(s string) string {
	// Replace tabs with spaces for comparison
	s = strings.ReplaceAll(s, "\t", "    ")
	// Collapse multiple spaces in each line
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	return strings.Join(lines, "\n")
}

// findSimilarPattern searches for a pattern that matches the normalized version.
// Returns the index in the file where the pattern was found, or -1 if not found.
func findSimilarPattern(fileContent string, normalizedPattern string) int {
	// Try exact match first
	if idx := strings.Index(fileContent, normalizedPattern); idx >= 0 {
		return idx
	}

	// Try matching with normalized content
	normalized := normalizeWhitespace(fileContent)
	if idx := strings.Index(normalized, normalizedPattern); idx >= 0 {
		return idx
	}

	// Try matching individual lines
	lines := strings.Split(normalizedPattern, "\n")
	if len(lines) == 1 {
		// Single line search - try various whitespace variations
		return findLineVariation(fileContent, lines[0])
	}

	return -1
}

// findLineVariation tries to find a line accounting for different whitespace.
func findLineVariation(content string, targetLine string) int {
	lines := strings.Split(content, "\n")
	normTarget := strings.TrimSpace(targetLine)

	for i, line := range lines {
		if strings.TrimSpace(line) == normTarget {
			// Calculate the byte offset for this line
			offset := 0
			for j := range i {
				offset += len(lines[j]) + 1 // +1 for newline
			}
			return offset
		}
	}
	return -1
}

// countNewlines returns the number of newlines before the given byte position.
func countNewlines(s string) int {
	return strings.Count(s, "\n")
}

// ExtractContextLines extracts N lines around a target line for better edit matching.
// This helps create unique old_string matches by including surrounding context.
func ExtractContextLines(
	filePath string,
	targetLine string,
	contextLinesCount int,
) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(content), "\n")

	// Find the target line
	var targetIdx int = -1
	for i, line := range lines {
		if strings.Contains(line, strings.TrimSpace(targetLine)) {
			targetIdx = i
			break
		}
	}

	if targetIdx < 0 {
		return "", fmt.Errorf("target line not found in file")
	}

	// Extract context
	startIdx := max(0, targetIdx-contextLinesCount)
	endIdx := min(len(lines), targetIdx+contextLinesCount+1)

	contextLines := lines[startIdx:endIdx]
	return strings.Join(contextLines, "\n"), nil
}

// max returns the larger of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// min returns the smaller of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ValidateEditPattern checks if an old_string is unique in the file.
// Returns true if the pattern appears exactly once, false otherwise.
func ValidateEditPattern(filePath string, oldString string) (bool, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false, err
	}

	fileContent := string(content)
	matches := strings.Count(fileContent, oldString)

	if matches == 0 {
		return false, fmt.Errorf("pattern not found in file")
	}
	if matches > 1 {
		return false, fmt.Errorf("pattern appears %d times (must be unique)", matches)
	}

	return true, nil
}

// FindBestMatch searches for the best matching context around a target line.
// This is useful when the exact old_string fails to match.
func FindBestMatch(
	filePath string,
	targetContent string,
	contextSize int,
) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(content), "\n")

	// Find best matching line using regex
	pattern := regexp.MustCompile(regexp.QuoteMeta(strings.TrimSpace(targetContent)))
	var bestIdx int = -1

	for i, line := range lines {
		if pattern.MatchString(line) {
			bestIdx = i
			break
		}
	}

	if bestIdx < 0 {
		return "", fmt.Errorf("no matching line found")
	}

	// Extract context around the match
	start := max(0, bestIdx-contextSize)
	end := min(len(lines), bestIdx+contextSize+1)

	return strings.Join(lines[start:end], "\n"), nil
}

// tryNormalizedMatch attempts to find oldString in content after normalizing whitespace
func tryNormalizedMatch(content, oldString string) (string, bool) {
	normalized := normalizeWhitespace(oldString)
	if strings.Contains(content, normalized) {
		return normalized, true
	}
	return "", false
}
