package diff

import (
	"strings"
	"testing"
)

func TestGenerateDiff(t *testing.T) {
	tests := []struct {
		name            string
		beforeContent   string
		afterContent    string
		fileName        string
		expectAdditions int
		expectRemovals  int
		checkDiff       func(t *testing.T, diff string)
	}{
		{
			name:            "no changes",
			beforeContent:   "line1\nline2\nline3\n",
			afterContent:    "line1\nline2\nline3\n",
			fileName:        "test.txt",
			expectAdditions: 0,
			expectRemovals:  0,
			checkDiff: func(t *testing.T, diff string) {
				if diff != "" {
					t.Errorf("Expected empty diff for identical content, got: %s", diff)
				}
			},
		},
		{
			name:            "add one line",
			beforeContent:   "line1\nline2\n",
			afterContent:    "line1\nline2\nline3\n",
			fileName:        "test.txt",
			expectAdditions: 1,
			expectRemovals:  0,
			checkDiff: func(t *testing.T, diff string) {
				if !strings.Contains(diff, "+line3") {
					t.Errorf("Expected diff to contain '+line3', got: %s", diff)
				}
				if !strings.Contains(diff, "a/test.txt") {
					t.Error("Expected diff to contain 'a/test.txt'")
				}
				if !strings.Contains(diff, "b/test.txt") {
					t.Error("Expected diff to contain 'b/test.txt'")
				}
			},
		},
		{
			name:            "remove one line",
			beforeContent:   "line1\nline2\nline3\n",
			afterContent:    "line1\nline3\n",
			fileName:        "test.txt",
			expectAdditions: 1, // line1 is shown as context
			expectRemovals:  2, // line1 and line2
			checkDiff: func(t *testing.T, diff string) {
				if !strings.Contains(diff, "-line2") {
					t.Errorf("Expected diff to contain '-line2', got: %s", diff)
				}
			},
		},
		{
			name:            "modify line",
			beforeContent:   "line1\nold line\nline3\n",
			afterContent:    "line1\nnew line\nline3\n",
			fileName:        "test.txt",
			expectAdditions: 1,
			expectRemovals:  1,
			checkDiff: func(t *testing.T, diff string) {
				if !strings.Contains(diff, "-old line") {
					t.Error("Expected diff to contain '-old line'")
				}
				if !strings.Contains(diff, "+new line") {
					t.Error("Expected diff to contain '+new line'")
				}
			},
		},
		{
			name:            "multiple additions",
			beforeContent:   "line1\n",
			afterContent:    "line1\nline2\nline3\nline4\n",
			fileName:        "test.txt",
			expectAdditions: 3,
			expectRemovals:  0,
			checkDiff: func(t *testing.T, diff string) {
				if !strings.Contains(diff, "+line2") {
					t.Error("Expected diff to contain '+line2'")
				}
				if !strings.Contains(diff, "+line3") {
					t.Error("Expected diff to contain '+line3'")
				}
				if !strings.Contains(diff, "+line4") {
					t.Error("Expected diff to contain '+line4'")
				}
			},
		},
		{
			name:            "multiple removals",
			beforeContent:   "line1\nline2\nline3\nline4\n",
			afterContent:    "line1\n",
			fileName:        "test.txt",
			expectAdditions: 0,
			expectRemovals:  3,
			checkDiff: func(t *testing.T, diff string) {
				if !strings.Contains(diff, "-line2") {
					t.Error("Expected diff to contain '-line2'")
				}
				if !strings.Contains(diff, "-line3") {
					t.Error("Expected diff to contain '-line3'")
				}
				if !strings.Contains(diff, "-line4") {
					t.Error("Expected diff to contain '-line4'")
				}
			},
		},
		{
			name:            "empty to content",
			beforeContent:   "",
			afterContent:    "line1\nline2\n",
			fileName:        "test.txt",
			expectAdditions: 2,
			expectRemovals:  0,
			checkDiff: func(t *testing.T, diff string) {
				if !strings.Contains(diff, "+line1") {
					t.Error("Expected diff to contain '+line1'")
				}
				if !strings.Contains(diff, "+line2") {
					t.Error("Expected diff to contain '+line2'")
				}
			},
		},
		{
			name:            "content to empty",
			beforeContent:   "line1\nline2\n",
			afterContent:    "",
			fileName:        "test.txt",
			expectAdditions: 0,
			expectRemovals:  2,
			checkDiff: func(t *testing.T, diff string) {
				if !strings.Contains(diff, "-line1") {
					t.Error("Expected diff to contain '-line1'")
				}
				if !strings.Contains(diff, "-line2") {
					t.Error("Expected diff to contain '-line2'")
				}
			},
		},
		{
			name:            "filename with leading slash",
			beforeContent:   "old\n",
			afterContent:    "new\n",
			fileName:        "/path/to/file.txt",
			expectAdditions: 1,
			expectRemovals:  1,
			checkDiff: func(t *testing.T, diff string) {
				// Should strip leading slash
				if !strings.Contains(diff, "a/path/to/file.txt") {
					t.Errorf("Expected 'a/path/to/file.txt', got: %s", diff)
				}
				if !strings.Contains(diff, "b/path/to/file.txt") {
					t.Errorf("Expected 'b/path/to/file.txt', got: %s", diff)
				}
			},
		},
		{
			name:            "filename without leading slash",
			beforeContent:   "old\n",
			afterContent:    "new\n",
			fileName:        "path/to/file.txt",
			expectAdditions: 1,
			expectRemovals:  1,
			checkDiff: func(t *testing.T, diff string) {
				if !strings.Contains(diff, "a/path/to/file.txt") {
					t.Error("Expected 'a/path/to/file.txt'")
				}
				if !strings.Contains(diff, "b/path/to/file.txt") {
					t.Error("Expected 'b/path/to/file.txt'")
				}
			},
		},
		{
			name:            "complex code diff",
			beforeContent:   "func main() {\n\tfmt.Println(\"Hello\")\n}\n",
			afterContent:    "func main() {\n\tfmt.Println(\"Hello\")\n\tfmt.Println(\"World\")\n}\n",
			fileName:        "main.go",
			expectAdditions: 2, // Hello line shown as context + World line
			expectRemovals:  1, // Original Hello line
			checkDiff: func(t *testing.T, diff string) {
				if !strings.Contains(diff, "+\tfmt.Println(\"World\")") {
					t.Error("Expected diff to contain new println statement")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff, additions, removals := GenerateDiff(tt.beforeContent, tt.afterContent, tt.fileName)

			if additions != tt.expectAdditions {
				t.Errorf("Expected %d additions, got %d", tt.expectAdditions, additions)
			}

			if removals != tt.expectRemovals {
				t.Errorf("Expected %d removals, got %d", tt.expectRemovals, removals)
			}

			if tt.checkDiff != nil {
				tt.checkDiff(t, diff)
			}
		})
	}
}

// TestGenerateDiffEdgeCases tests edge cases and special scenarios
func TestGenerateDiffEdgeCases(t *testing.T) {
	t.Run("very long lines", func(t *testing.T) {
		longLine := strings.Repeat("a", 10000)
		before := longLine + "\n"
		after := strings.Repeat("b", 10000) + "\n"
		
		diff, additions, removals := GenerateDiff(before, after, "long.txt")
		
		if additions != 1 {
			t.Errorf("Expected 1 addition, got %d", additions)
		}
		if removals != 1 {
			t.Errorf("Expected 1 removal, got %d", removals)
		}
		if diff == "" {
			t.Error("Expected non-empty diff")
		}
	})

	t.Run("special characters in filename", func(t *testing.T) {
		diff, _, _ := GenerateDiff("old\n", "new\n", "file with spaces.txt")
		if !strings.Contains(diff, "file with spaces.txt") {
			t.Error("Expected filename with spaces to be preserved")
		}
	})

	t.Run("unicode content", func(t *testing.T) {
		before := "Hello 世界\n"
		after := "Hello World\n"
		
		diff, additions, removals := GenerateDiff(before, after, "unicode.txt")
		
		if additions != 1 {
			t.Errorf("Expected 1 addition, got %d", additions)
		}
		if removals != 1 {
			t.Errorf("Expected 1 removal, got %d", removals)
		}
		if !strings.Contains(diff, "Hello World") {
			t.Error("Expected diff to contain new content")
		}
	})

	t.Run("lines with plus and minus but not diff markers", func(t *testing.T) {
		before := "normal line\n"
		after := "normal line\n+ this is content, not diff\n- also content\n"
		
		diff, additions, removals := GenerateDiff(before, after, "test.txt")
		
		// Should count actual additions (the two new lines)
		if additions != 2 {
			t.Errorf("Expected 2 additions, got %d", additions)
		}
		if removals != 0 {
			t.Errorf("Expected 0 removals, got %d", removals)
		}
		if diff == "" {
			t.Error("Expected non-empty diff")
		}
	})
}

// TestGenerateDiffCountAccuracy verifies that +++ and --- headers are not counted
func TestGenerateDiffCountAccuracy(t *testing.T) {
	before := "line1\nline2\nline3\n"
	after := "line1\nmodified\nline3\n"
	
	diff, additions, removals := GenerateDiff(before, after, "test.txt")
	
	// Should have exactly 1 addition and 1 removal
	if additions != 1 {
		t.Errorf("Expected 1 addition, got %d", additions)
	}
	if removals != 1 {
		t.Errorf("Expected 1 removal, got %d", removals)
	}
	
	// Verify the diff contains the header markers but they're not counted
	if !strings.Contains(diff, "+++") {
		t.Error("Expected diff to contain +++ header")
	}
	if !strings.Contains(diff, "---") {
		t.Error("Expected diff to contain --- header")
	}
}
