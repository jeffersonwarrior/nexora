package tools

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
)

// EditDiagnosticsInfo captures detailed diagnostic information about edit failures
type EditDiagnosticsInfo struct {
	FilePath      string
	OldString     string
	NewString     string
	FileContent   string
	FileSizeBytes int
	LineCount     int
	FailureReason string
	Context       map[string]interface{}
}

// ViewDiagnosticsInfo captures diagnostic information about view operations
type ViewDiagnosticsInfo struct {
	FilePath  string
	Offset    int
	Limit     int
	FileSize  int64
	LineCount int
	Context   map[string]interface{}
}

// WhitespaceAnalysis provides detailed whitespace analysis for debugging
type WhitespaceAnalysis struct {
	ContainsTab        bool
	ContainsSpace      bool
	LeadingSpaces      int
	TrailingSpaces     int
	BlankLines         int
	HasMixedIndent     bool
	LineEndings        string // "LF" or "CRLF" or "Mixed"
	ByteRepresentation string
}

// AnalyzeWhitespace performs detailed whitespace analysis on a string
func AnalyzeWhitespace(s string) WhitespaceAnalysis {
	analysis := WhitespaceAnalysis{
		ByteRepresentation: formatBytesForDebug(s),
	}

	lines := strings.Split(s, "\n")
	indentationStyles := make(map[string]int)

	for _, line := range lines {
		if line == "" {
			analysis.BlankLines++
			continue
		}

		if strings.Contains(line, "\t") {
			analysis.ContainsTab = true
		}
		if strings.Contains(line, " ") {
			analysis.ContainsSpace = true
		}

		for i, ch := range line {
			if ch == '\t' {
				indentationStyles["tab"]++
			} else if ch == ' ' {
				indentationStyles["space"]++
			} else {
				analysis.LeadingSpaces = i
				break
			}
		}

		trimmed := strings.TrimRight(line, " \t")
		if len(trimmed) < len(line) {
			analysis.TrailingSpaces += len(line) - len(trimmed)
		}
	}

	if len(indentationStyles) > 1 {
		analysis.HasMixedIndent = true
	}

	analysis.LineEndings = detectLineEndings(s)
	return analysis
}

func detectLineEndings(s string) string {
	crlfCount := strings.Count(s, "\r\n")
	lfCount := strings.Count(s, "\n") - crlfCount

	if crlfCount > 0 && lfCount == 0 {
		return "CRLF"
	} else if lfCount > 0 && crlfCount == 0 {
		return "LF"
	} else if crlfCount > 0 && lfCount > 0 {
		return "Mixed"
	}
	return "Unknown"
}

func formatBytesForDebug(s string) string {
	var result strings.Builder
	for i, ch := range s {
		if i > 100 {
			result.WriteString(fmt.Sprintf("... (%d more chars)", len(s)-i))
			break
		}
		switch ch {
		case '\n':
			result.WriteString("\\n")
		case '\r':
			result.WriteString("\\r")
		case '\t':
			result.WriteString("\\t")
		case ' ':
			result.WriteString("·")
		default:
			if ch >= 32 && ch < 127 {
				result.WriteRune(ch)
			} else {
				result.WriteString(fmt.Sprintf("\\x%02x", ch))
			}
		}
	}
	return result.String()
}

// LogEditFailure captures comprehensive diagnostic data when an edit fails
func LogEditFailure(diag EditDiagnosticsInfo) {
	oldAnalysis := AnalyzeWhitespace(diag.OldString)
	newAnalysis := AnalyzeWhitespace(diag.NewString)
	fileAnalysis := AnalyzeWhitespace(diag.FileContent)

	slog.Error("Edit operation failed",
		"file", diag.FilePath,
		"failure_reason", diag.FailureReason,
		"old_string_length", len(diag.OldString),
		"new_string_length", len(diag.NewString),
		"file_content_length", len(diag.FileContent),
		"old_string_lines", strings.Count(diag.OldString, "\n")+1,
		"new_string_lines", strings.Count(diag.NewString, "\n")+1,
		"old_has_tabs", oldAnalysis.ContainsTab,
		"old_has_mixed_indent", oldAnalysis.HasMixedIndent,
		"old_line_endings", oldAnalysis.LineEndings,
		"new_has_tabs", newAnalysis.ContainsTab,
		"file_has_tabs", fileAnalysis.ContainsTab,
		"file_has_mixed_indent", fileAnalysis.HasMixedIndent,
		"file_line_endings", fileAnalysis.LineEndings,
		"old_string_preview", truncatePreview(diag.OldString, 80),
		"file_size", len(diag.FileContent),
	)

	slog.Debug("Edit failure detailed diagnostics",
		"old_string_bytes", oldAnalysis.ByteRepresentation,
		"new_string_bytes", newAnalysis.ByteRepresentation,
		"additional_context", diag.Context,
	)
}

// LogEditSuccess captures successful edit metrics
func LogEditSuccess(filePath string, oldLen, newLen int, replacementCount int, attemptCount int) {
	slog.Info("Edit operation succeeded",
		"file", filePath,
		"old_string_length", oldLen,
		"new_string_length", newLen,
		"replacement_count", replacementCount,
		"attempt_count", attemptCount,
	)
}

// LogViewOperation logs view tool operations
func LogViewOperation(diag ViewDiagnosticsInfo, duration float64) {
	slog.Info("View operation completed",
		"file", diag.FilePath,
		"offset", diag.Offset,
		"limit", diag.Limit,
		"file_size", diag.FileSize,
		"total_lines", diag.LineCount,
		"duration_ms", duration,
	)
}

// LogViewError logs view tool errors
func LogViewError(diag ViewDiagnosticsInfo, errReason string) {
	slog.Error("View operation failed",
		"file", diag.FilePath,
		"offset", diag.Offset,
		"limit", diag.Limit,
		"file_size", diag.FileSize,
		"error_reason", errReason,
	)
}

func truncatePreview(s string, maxLen int) string {
	if len(s) <= maxLen {
		return formatBytesForDebug(s)
	}
	return formatBytesForDebug(s[:maxLen]) + "..."
}

// LogSelfHealingAttempt logs self-healing retry attempts
func LogSelfHealingAttempt(filePath, reason string, success bool, originalLen, improvedLen int) {
	level := slog.LevelInfo
	if !success {
		level = slog.LevelWarn
	}
	slog.Log(context.TODO(), level,
		"Self-healing edit retry",
		"file", filePath,
		"reason", reason,
		"success", success,
		"original_length", originalLen,
		"improved_length", improvedLen,
	)
}

// PatternMatchAnalysis provides insight into why pattern matching failed
func PatternMatchAnalysis(fileContent, pattern string) map[string]interface{} {
	return map[string]interface{}{
		"file_length":        len(fileContent),
		"pattern_length":     len(pattern),
		"pattern_line_count": strings.Count(pattern, "\n") + 1,
		"file_line_count":    strings.Count(fileContent, "\n") + 1,
		"pattern_in_file":    strings.Contains(fileContent, pattern),
	}
}

// AnalyzeWhitespaceDifference compares whitespace between file content and pattern
func AnalyzeWhitespaceDifference(fileContent, pattern string) map[string]interface{} {
	fileTabs := strings.Count(fileContent, "\t")
	patternTabs := strings.Count(pattern, "\t")
	displayTabs := strings.Count(pattern, "→\t")

	fileSpaces := countLeadingSpaces(fileContent)
	patternSpaces := countLeadingSpaces(pattern)

	return map[string]interface{}{
		"has_tab_mismatch":            displayTabs > 0,
		"expected_tabs":               fileTabs,
		"found_tabs":                  patternTabs,
		"display_tabs":                displayTabs,
		"has_space_mismatch":          patternSpaces != fileSpaces,
		"expected_spaces":             fileSpaces,
		"found_spaces":                patternSpaces,
		"pattern_in_file":             strings.Contains(fileContent, pattern),
		"pattern_after_normalization": strings.Contains(fileContent, normalizeTabIndicators(pattern)),
	}
}

// Helper function for space counting
func countLeadingSpaces(content string) int {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return 0
	}

	// Count leading spaces on first non-empty line
	for _, line := range lines {
		trimmed := strings.TrimLeft(line, " ")
		if len(trimmed) < len(line) {
			return len(line) - len(trimmed)
		}
	}
	return 0
}

// createAIErrorMessage generates AI-friendly error messages with actionable guidance
func createAIErrorMessage(err error, fileContent, oldString string) string {
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

	// Pattern not found at all
	if !analysis["pattern_in_file"].(bool) {
		// Check if it would match after normalization
		if analysis["pattern_after_normalization"].(bool) {
			return fmt.Sprintf("PATTERN_FORMAT_MISMATCH: Your pattern would match after tab normalization. " +
				"Use AI mode (ai_mode=true) or normalize tabs manually.")
		}

		return fmt.Sprintf("PATTERN_NOT_FOUND: The text was not found in the file. " +
			"Common fixes: 1) Use AI mode (ai_mode=true), " +
			"2) Include more surrounding context (3-5 lines), " +
			"3) Check for tab/space differences, " +
			"4) Verify the file hasn't changed since you viewed it.")
	}

	// Fallback to original error
	return err.Error()
}
