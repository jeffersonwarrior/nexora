package tools

import (
	"fmt"
	"strings"
)

// findMostSimilarContent finds the closest matching content in the file
func findMostSimilarContent(fileContent, target string) string {
	targetLines := strings.Split(strings.TrimSpace(target), "\n")
	if len(targetLines) == 0 {
		return ""
	}

	fileLines := strings.Split(fileContent, "\n")
	bestMatch := ""
	bestScore := 0

	// Look for sequences with at least 50% line overlap
	for i := 0; i <= len(fileLines)-len(targetLines); i++ {
		matchCount := 0
		for j, targetLine := range targetLines {
			if i+j < len(fileLines) {
				fileLine := fileLines[i+j]
				if strings.TrimSpace(fileLine) == strings.TrimSpace(targetLine) {
					matchCount++
				}
			}
		}

		if matchCount > bestScore {
			bestScore = matchCount
			endIdx := i + len(targetLines)
			if endIdx > len(fileLines) {
				endIdx = len(fileLines)
			}
			bestMatch = strings.Join(fileLines[i:endIdx], "\n")
		}
	}

	if bestScore >= len(targetLines)/2 {
		return bestMatch
	}
	return ""
}

// visualizeDifference creates a visual diff showing expected vs actual
func visualizeDifference(expected, actual string) string {
	var result strings.Builder
	result.WriteString("EXPECTED (your old_string):\n")

	expectedLines := strings.Split(expected, "\n")
	for i, line := range expectedLines {
		// Show whitespace visually
		visualLine := strings.ReplaceAll(line, "\t", "→")
		visualLine = strings.ReplaceAll(visualLine, " ", "·")
		result.WriteString(fmt.Sprintf("  %d: %s\n", i+1, visualLine))
	}

	result.WriteString("\nACTUAL (closest match in file):\n")
	actualLines := strings.Split(actual, "\n")
	for i, line := range actualLines {
		visualLine := strings.ReplaceAll(line, "\t", "→")
		visualLine = strings.ReplaceAll(visualLine, " ", "·")

		// Mark differing lines
		marker := " "
		if i < len(expectedLines) && strings.TrimSpace(line) != strings.TrimSpace(expectedLines[i]) {
			marker = "✗"
		}
		result.WriteString(fmt.Sprintf("%s %d: %s\n", marker, i+1, visualLine))
	}

	return result.String()
}

// createEnhancedErrorMessage generates AI-friendly error messages with visual diffs
func createEnhancedErrorMessage(err error, fileContent, oldString string) string {
	analysis := AnalyzeWhitespaceDifference(fileContent, oldString)

	// Tab mismatch - most common issue
	if analysis["has_tab_mismatch"].(bool) {
		return fmt.Sprintf("TAB_MISMATCH: The VIEW tool shows tabs as '→\t' but EDIT needs raw tabs. "+
			"Found %d display tabs in your pattern. Try: "+
			"1) Use AI mode (ai_mode=true) for automatic normalization, or "+
			"2) Replace '→\t' with actual tab characters in your pattern.",
			analysis["display_tabs"].(int))
	}

	// Space mismatch
	if analysis["has_space_mismatch"].(bool) {
		return fmt.Sprintf("SPACE_MISMATCH: Expected %d leading spaces but found %d. "+
			"Count spaces carefully. AI mode (ai_mode=true) can help with this.",
			analysis["expected_spaces"].(int), analysis["found_spaces"].(int))
	}

	// Pattern not found at all - show visual diff
	if !analysis["pattern_in_file"].(bool) {
		// Check if it would match after normalization
		if analysis["pattern_after_normalization"].(bool) {
			return fmt.Sprintf("PATTERN_FORMAT_MISMATCH: Your pattern would match after tab normalization. " +
				"Use AI mode (ai_mode=true) or normalize tabs manually.")
		}

		// Try to find similar content and show diff
		similarContent := findMostSimilarContent(fileContent, oldString)
		if similarContent != "" {
			diff := visualizeDifference(oldString, similarContent)
			return fmt.Sprintf("PATTERN_NOT_FOUND: The text was not found in the file.\n\n"+
				"Most similar content found:\n%s\n\n"+
				"Common fixes:\n"+
				"1) Use `sed -n 'START,ENDp' file` to extract exact text (RECOMMENDED)\n"+
				"2) Use AI mode (ai_mode=true)\n"+
				"3) If edit fails once, use `write` tool instead\n"+
				"4) Use `smart_edit` tool with line numbers for 100%% reliability",
				diff)
		}

		return fmt.Sprintf("PATTERN_NOT_FOUND: The text was not found in the file. " +
			"Common fixes:\n" +
			"1) Use `grep -n 'pattern' file` to find line numbers, then use `smart_edit`\n" +
			"2) Use `sed -n 'START,ENDp' file` to extract exact text\n" +
			"3) Include more surrounding context (3-5 lines)\n" +
			"4) If edit fails once, use `write` tool instead")
	}

	// Fallback to original error
	return err.Error()
}
