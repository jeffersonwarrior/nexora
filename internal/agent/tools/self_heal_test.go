package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeWhitespace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "  func foo() {",
			expected: "func foo() {",
		},
		{
			input:    "func\t\tfoo()\t{",
			expected: "func        foo()    {",
		},
		{
			input:    "line1\n  line2\nline3  ",
			expected: "line1\nline2\nline3",
		},
	}

	for _, tt := range tests {
		t.Run(strings.TrimSpace(tt.input), func(t *testing.T) {
			result := normalizeWhitespace(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestFindSimilarPattern(t *testing.T) {
	t.Parallel()

	content := `package main

import (
	"fmt"
)

func main() {
	fmt.Println("hello")
}
`

	tests := []struct {
		pattern string
		found   bool
	}{
		{"package main", true},
		{"fmt.Println", true},
		{"hello", true},
		{"nonexistent", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			idx := findSimilarPattern(content, tt.pattern)
			if tt.found {
				require.NotEqual(t, -1, idx, "pattern should be found")
			} else {
				require.Equal(t, -1, idx, "pattern should not be found")
			}
		})
	}
}

func TestExtractContextLines(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	testFile := filepath.Join(dir, "test.go")
	content := `package main

import "fmt"

func main() {
	fmt.Println("test")
}
`
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0o644))

	result, err := ExtractContextLines(testFile, "func main", 1)
	require.NoError(t, err)

	// Should contain the function declaration and surrounding lines
	require.Contains(t, result, "func main()")
	require.Contains(t, result, "fmt.Println")
}

func TestValidateEditPattern(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	testFile := filepath.Join(dir, "test.txt")
	content := `line 1
line 2
line 3
unique line
line 5
duplicate line
duplicate line
`
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0o644))

	tests := []struct {
		pattern   string
		wantValid bool
		wantErr   bool
	}{
		{"line 1", true, false},
		{"unique line", true, false},
		{"duplicate line", false, true},
		{"nonexistent", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			valid, err := ValidateEditPattern(testFile, tt.pattern)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.wantValid, valid)
		})
	}
}

func TestRetryWithContext(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	testFile := filepath.Join(dir, "test.go")
	content := `package main

func hello() {
	return "hello"
}

func world() {
	return "world"
}
`
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0o644))

	strategy := NewEditRetryStrategy(context.Background())

	// Attempt to retry with context when exact match fails
	improvedParams, err := strategy.RetryWithContext(
		testFile,
		"func hello() {",
		"func greeting() {",
		"old_string not found",
	)

	// The retry should either succeed or fail gracefully
	if err == nil {
		require.Equal(t, testFile, improvedParams.FilePath)
		require.Equal(t, "func greeting() {", improvedParams.NewString)
	}
}

func TestFindBestMatch(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	testFile := filepath.Join(dir, "test.go")
	content := `func foo() {
	bar()
}

func baz() {
	qux()
}
`
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0o644))

	result, err := FindBestMatch(testFile, "func foo()", 1)
	require.NoError(t, err)
	require.Contains(t, result, "func foo()")
	require.Contains(t, result, "bar()")
}

func TestCountNewlines(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected int
	}{
		{"hello", 0},
		{"hello\nworld", 1},
		{"a\nb\nc\n", 3},
		{"\n\n\n", 3},
	}

	for _, tt := range tests {
		t.Run(strings.TrimSpace(tt.input), func(t *testing.T) {
			result := countNewlines(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}
