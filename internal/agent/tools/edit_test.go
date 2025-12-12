package tools

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTabNormalization(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "real tab unchanged",
			input:    "func\tfoo()",
			expected: "func\tfoo()",
		},
		{
			name:     "display tab normalized",
			input:    "func→\tfoo()",
			expected: "func\tfoo()",
		},
		{
			name:     "partial display tab normalized",
			input:    "func→foo()",
			expected: "func\tfoo()",
		},
		{
			name:     "no tabs unchanged",
			input:    "func foo()",
			expected: "func foo()",
		},
		{
			name:     "multiple display tabs",
			input:    "func→\t→\tfoo()",
			expected: "func\t\tfoo()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeTabIndicators(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestAutoExpandContext(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := tmpDir + "/test.txt"

	// Create a longer file to test proper context expansion
	content := "line 1\nline 2\nline 3\ntarget line\nline 5\nline 6\nline 7"
	err := os.WriteFile(testFile, []byte(content), 0o644)
	require.NoError(t, err)

	// Test context expansion with a unique pattern
	expanded, err := autoExpandContext(testFile, "target line")
	require.NoError(t, err)

	// Should include lines before and after (2 before, 2 after)
	require.Contains(t, expanded, "line 2")
	require.Contains(t, expanded, "line 3")
	require.Contains(t, expanded, "target line")
	require.Contains(t, expanded, "line 5")
	require.Contains(t, expanded, "line 6")

	// Should not include lines that are too far
	require.NotContains(t, expanded, "line 1")
	require.NotContains(t, expanded, "line 7")
}

func TestAIErrorMessages(t *testing.T) {
	fileContent := "func\tmain() {\n\tfmt.Println(\"hello\")\n}"

	tests := []struct {
		name           string
		pattern        string
		expectedPrefix string
	}{
		{
			name:           "tab mismatch",
			pattern:        "func→\tmain()",
			expectedPrefix: "TAB_MISMATCH",
		},
		{
			name:           "pattern not found",
			pattern:        "nonexistent",
			expectedPrefix: "PATTERN_NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := createAIErrorMessage(
				fmt.Errorf("test error"),
				fileContent,
				tt.pattern,
			)

			require.Contains(t, msg, tt.expectedPrefix)
		})
	}
}

func TestCountLeadingSpaces(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "no leading spaces",
			input:    "func()",
			expected: 0,
		},
		{
			name:     "2 leading spaces",
			input:    "  func()",
			expected: 2,
		},
		{
			name:     "4 leading spaces",
			input:    "    func()",
			expected: 4,
		},
		{
			name:     "mixed lines - uses first",
			input:    "  func1()\n    func2()",
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countLeadingSpaces(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestWhitespaceAnalysis(t *testing.T) {
	fileContent := "func\tmain() {\n\tfmt.Println(\"hello\")\n}"
	patternWithDisplayTabs := "func→\tmain()"
	patternWithRealTabs := "func\tmain()"

	// Test tab mismatch detection
	analysis1 := AnalyzeWhitespaceDifference(fileContent, patternWithDisplayTabs)
	require.True(t, analysis1["has_tab_mismatch"].(bool))
	require.Equal(t, 1, analysis1["display_tabs"].(int))

	// Test no mismatch with real tabs
	analysis2 := AnalyzeWhitespaceDifference(fileContent, patternWithRealTabs)
	require.False(t, analysis2["has_tab_mismatch"].(bool))
	require.Equal(t, 0, analysis2["display_tabs"].(int))

	// Test that pattern with real tabs is found in file
	require.True(t, analysis2["pattern_in_file"].(bool))
}
