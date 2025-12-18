package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestGlobFiles(t *testing.T) {
	ctx := context.Background()
	
	tmpDir := t.TempDir()
	
	// Create test directory structure
	testFiles := []string{
		"file1.go",
		"file2.go",
		"test.txt",
		"subdir/nested.go",
		"subdir/data.json",
		"another/deep/path/file.go",
	}
	
	for _, f := range testFiles {
		fullPath := filepath.Join(tmpDir, f)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		if err := os.WriteFile(fullPath, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}

	t.Run("basic pattern matching", func(t *testing.T) {
		files, truncated, err := globFiles(ctx, "*.go", tmpDir, 100)
		
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		
		if len(files) != 2 {
			t.Errorf("Expected 2 .go files in root, got %d", len(files))
		}
		
		if truncated {
			t.Error("Expected truncated to be false")
		}
		
		// Verify filenames
		found := make(map[string]bool)
		for _, f := range files {
			base := filepath.Base(f)
			found[base] = true
		}
		
		if !found["file1.go"] || !found["file2.go"] {
			t.Error("Expected to find file1.go and file2.go")
		}
	})

	t.Run("recursive pattern", func(t *testing.T) {
		files, truncated, err := globFiles(ctx, "**/*.go", tmpDir, 100)
		
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		
		// Should find all .go files including nested ones
		if len(files) < 4 {
			t.Errorf("Expected at least 4 .go files recursively, got %d", len(files))
		}
		
		if truncated {
			t.Error("Expected truncated to be false")
		}
	})

	t.Run("respects limit", func(t *testing.T) {
		// Create many test files
		for i := 0; i < 20; i++ {
			filename := filepath.Join(tmpDir, "many_"+string(rune('a'+i))+".txt")
			if err := os.WriteFile(filename, []byte("test"), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}
		}
		
		files, truncated, err := globFiles(ctx, "*.txt", tmpDir, 10)
		
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		
		if len(files) > 10 {
			t.Errorf("Expected at most 10 files, got %d", len(files))
		}
		
		if !truncated {
			t.Error("Expected truncated to be true when limit is exceeded")
		}
	})

	t.Run("no matches", func(t *testing.T) {
		files, truncated, err := globFiles(ctx, "*.nonexistent", tmpDir, 100)
		
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		
		if len(files) != 0 {
			t.Errorf("Expected no files, got %d", len(files))
		}
		
		if truncated {
			t.Error("Expected truncated to be false with no matches")
		}
	})

	t.Run("multiple extensions", func(t *testing.T) {
		files, truncated, err := globFiles(ctx, "**/*.{go,json}", tmpDir, 100)
		
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		
		// Should find both .go and .json files
		foundGo := false
		foundJson := false
		for _, f := range files {
			if filepath.Ext(f) == ".go" {
				foundGo = true
			}
			if filepath.Ext(f) == ".json" {
				foundJson = true
			}
		}
		
		if !foundGo {
			t.Error("Expected to find .go files")
		}
		if !foundJson {
			t.Error("Expected to find .json files")
		}
		
		if truncated {
			t.Error("Expected truncated to be false")
		}
	})
}

func TestNormalizeFilePaths(t *testing.T) {
	// Test that normalizeFilePaths uses filepath.ToSlash
	tests := []struct {
		name  string
		input []string
	}{
		{
			name:  "unix paths",
			input: []string{"/home/user/file.go", "path/to/file.txt"},
		},
		{
			name:  "empty input",
			input: []string{},
		},
		{
			name:  "single path",
			input: []string{"test/file.go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paths := make([]string, len(tt.input))
			copy(paths, tt.input)
			
			normalizeFilePaths(paths)
			
			// Verify all paths use forward slashes (filepath.ToSlash behavior)
			for i, p := range paths {
				normalized := filepath.ToSlash(tt.input[i])
				if p != normalized {
					t.Errorf("Path %d not normalized: expected %s, got %s", i, normalized, p)
				}
			}
		})
	}
}
