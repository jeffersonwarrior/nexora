package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// BenchmarkEditSmallFile measures edit performance on small files
func BenchmarkEditSmallFile(b *testing.B) {
	tempDir := b.TempDir()
	filePath := filepath.Join(tempDir, "small.txt")

	// Create small file (100 lines)
	content := strings.Repeat("line of text here\n", 100)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		b.Fatalf("Failed to create file: %v", err)
	}

	oldString := "line of text here"
	newString := "modified line here"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset file content
		os.WriteFile(filePath, []byte(content), 0644)

		// Perform replace operation (simulating edit core logic)
		data, _ := os.ReadFile(filePath)
		modified := strings.Replace(string(data), oldString, newString, 1)
		os.WriteFile(filePath, []byte(modified), 0644)
	}
}

// BenchmarkEditLargeFile measures edit performance on large files
func BenchmarkEditLargeFile(b *testing.B) {
	tempDir := b.TempDir()
	filePath := filepath.Join(tempDir, "large.txt")

	// Create large file (10000 lines)
	content := strings.Repeat("this is a line of text in a large file\n", 10000)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		b.Fatalf("Failed to create file: %v", err)
	}

	oldString := "this is a line of text in a large file"
	newString := "this is a modified line in a large file"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset file content
		os.WriteFile(filePath, []byte(content), 0644)

		// Perform replace operation
		data, _ := os.ReadFile(filePath)
		modified := strings.Replace(string(data), oldString, newString, 1)
		os.WriteFile(filePath, []byte(modified), 0644)
	}
}

// BenchmarkFuzzyMatch measures fuzzy matching performance
func BenchmarkFuzzyMatch(b *testing.B) {
	content := strings.Repeat("func example() {\n    return nil\n}\n\n", 1000)
	target := "func example() {\n    return nil\n}"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simple fuzzy match simulation
		_ = strings.Contains(content, target)
		_ = strings.Index(content, target)
	}
}

// BenchmarkNormalizeTabIndicators measures tab normalization performance
func BenchmarkNormalizeTabIndicators(b *testing.B) {
	// Content with many tab indicators
	content := strings.Repeat("→\tfunc test() {\n→\t→\treturn nil\n→\t}\n", 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = normalizeTabIndicators(content)
	}
}

// BenchmarkEditReplaceAll measures replace_all performance
func BenchmarkEditReplaceAll(b *testing.B) {
	content := strings.Repeat("old_value = something\n", 5000)
	oldString := "old_value"
	newString := "new_value"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strings.ReplaceAll(content, oldString, newString)
	}
}

// BenchmarkEditCountOccurrences measures occurrence counting
func BenchmarkEditCountOccurrences(b *testing.B) {
	content := strings.Repeat("pattern appears here pattern and here pattern\n", 2000)
	pattern := "pattern"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strings.Count(content, pattern)
	}
}
