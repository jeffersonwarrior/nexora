package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAIModeExpansion(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	// Create a file with specific content
	content := `package main

import (
	"fmt"
)

func main() {
	fmt.Println("Hello, World!")
}`

	err := os.WriteFile(testFile, []byte(content), 0o644)
	require.NoError(t, err)

	// Test 1: Minimal context should be expanded
	partialString := "fmt.Println(\"Hello, World!\")"
	expanded, err := autoExpandContext(testFile, partialString)
	require.NoError(t, err)

	// Should include surrounding context
	require.Contains(t, expanded, "func main() {")
	require.Contains(t, expanded, "fmt.Println(\"Hello, World!\")")

	// Should have more lines than the original
	originalLines := strings.Count(partialString, "\n") + 1
	expandedLines := strings.Count(expanded, "\n") + 1
	require.Greater(t, expandedLines, originalLines)

	// Test 2: Full context should not be expanded
	fullString := `func main() {
	fmt.Println("Hello, World!")
}`

	expandedFull, err := autoExpandContext(testFile, fullString)
	require.NoError(t, err)
	require.Equal(t, fullString, expandedFull)
}

func TestTabNormalizationInAIMode(t *testing.T) {
	// Test that display tabs from VIEW output are properly normalized
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "display tab normalized",
			input:    "func→\tmain() {",
			expected: "func\tmain() {",
		},
		{
			name:     "multiple display tabs normalized",
			input:    "func→\t→\tmain() {",
			expected: "func\t\tmain() {",
		},
		{
			name:     "partial display tab normalized",
			input:    "func→main() {",
			expected: "func\tmain() {",
		},
		{
			name:     "real tabs unchanged",
			input:    "func\tmain() {",
			expected: "func\tmain() {",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeTabIndicators(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}
