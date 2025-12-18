package tools

import (
	"testing"
)

func TestFuzzyMatching(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		target     string
		shouldFind bool
		confidence float64
	}{
		{
			name:       "exact match",
			content:    "func main() {\n\tfmt.Println(\"hello\")\n}",
			target:     "func main() {\n\tfmt.Println(\"hello\")\n}",
			shouldFind: true,
			confidence: 1.0,
		},
		{
			name:       "tab normalization",
			content:    "func main() {\n\tfmt.Println(\"hello\")\n}",
			target:     "func main() {\n→\tfmt.Println(\"hello\")\n}", // View tool display
			shouldFind: true,
			confidence: 0.95,
		},
		{
			name:       "line content match (different indentation)",
			content:    "func main() {\n\tfmt.Println(\"hello\")\n}",
			target:     "func main() {\n  fmt.Println(\"hello\")\n}", // 2 spaces instead of tab
			shouldFind: true,
			confidence: 0.90,
		},
		{
			name:       "no match",
			content:    "func main() {\n\tfmt.Println(\"hello\")\n}",
			target:     "func foo() {\n\treturn nil\n}",
			shouldFind: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findBestMatch(tt.content, tt.target)

			if tt.shouldFind {
				if result == nil {
					t.Errorf("Expected to find match but got nil")
					return
				}
				if result.confidence != tt.confidence {
					t.Errorf("Expected confidence %v, got %v", tt.confidence, result.confidence)
				}
			} else {
				if result != nil && result.confidence >= 0.90 {
					t.Errorf("Expected no match but found one with confidence %v", result.confidence)
				}
			}
		})
	}
}

func TestLineContentMatching(t *testing.T) {
	content := `package main

import "fmt"

func main() {
	fmt.Println("hello")
	fmt.Println("world")
}`

	tests := []struct {
		name       string
		target     string
		shouldFind bool
	}{
		{
			name:       "exact line match",
			target:     "\tfmt.Println(\"hello\")\n\tfmt.Println(\"world\")",
			shouldFind: true,
		},
		{
			name:       "different indent same content",
			target:     "  fmt.Println(\"hello\")\n  fmt.Println(\"world\")", // spaces instead of tabs
			shouldFind: true,
		},
		{
			name:       "wrong content",
			target:     "\tfmt.Println(\"goodbye\")\n\tfmt.Println(\"world\")",
			shouldFind: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			index := matchByLineContent(content, tt.target)
			found := index != -1

			if found != tt.shouldFind {
				t.Errorf("Expected shouldFind=%v, got index=%d", tt.shouldFind, index)
			}
		})
	}
}
