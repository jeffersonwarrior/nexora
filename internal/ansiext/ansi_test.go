package ansiext

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestEscape(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "regular text",
			input:    "Hello, World!",
			expected: "Hello, World!",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "null character",
			input:    "Hello\x00World",
			expected: "Hello\u2400World",
		},
		{
			name:     "tab character",
			input:    "Hello\tWorld",
			expected: "Hello\u2409World",
		},
		{
			name:     "newline character",
			input:    "Hello\nWorld",
			expected: "Hello\u240AWorld",
		},
		{
			name:     "carriage return",
			input:    "Hello\rWorld",
			expected: "Hello\u240DWorld",
		},
		{
			name:     "escape character",
			input:    "Hello\x1bWorld",
			expected: "Hello\u241BWorld",
		},
		{
			name:     "delete character",
			input:    string([]byte{'H', 'e', 'l', 'l', 'o', ansi.DEL, 'W', 'o', 'r', 'l', 'd'}),
			expected: "Hello\u2421World",
		},
		{
			name:     "multiple control characters",
			input:    "\x00\x01\x02\x03\x04\x05",
			expected: "\u2400\u2401\u2402\u2403\u2404\u2405",
		},
		{
			name:     "all control characters 0x00-0x1F",
			input:    string([]byte{0x00, 0x01, 0x02, 0x1E, 0x1F}),
			expected: "\u2400\u2401\u2402\u241E\u241F",
		},
		{
			name:     "bell character",
			input:    "Alert\x07!",
			expected: "Alert\u2407!",
		},
		{
			name:     "backspace character",
			input:    "Test\x08ing",
			expected: "Test\u2408ing",
		},
		{
			name:     "vertical tab",
			input:    "Line1\x0BLine2",
			expected: "Line1\u240BLine2",
		},
		{
			name:     "form feed",
			input:    "Page1\x0CPage2",
			expected: "Page1\u240CPage2",
		},
		{
			name:     "mixed text and control chars",
			input:    "Normal text\x00\x01\x02 more text\x1F",
			expected: "Normal text\u2400\u2401\u2402 more text\u241F",
		},
		{
			name:     "unicode characters preserved",
			input:    "Hello 世界 🌍",
			expected: "Hello 世界 🌍",
		},
		{
			name:     "unicode with control chars",
			input:    "世界\x00🌍\x01",
			expected: "世界\u2400🌍\u2401",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Escape(tt.input)
			if result != tt.expected {
				t.Errorf("Escape(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestEscapeAllControlCharacters tests all control characters from 0x00 to 0x1F
func TestEscapeAllControlCharacters(t *testing.T) {
	for i := 0; i <= 0x1F; i++ {
		input := string(rune(i))
		expected := string(rune(0x2400 + i))
		result := Escape(input)
		if result != expected {
			t.Errorf("Escape(0x%02X) = %q, want %q", i, result, expected)
		}
	}
}

// TestEscapeDelCharacter tests the DEL character (0x7F)
func TestEscapeDelCharacter(t *testing.T) {
	input := string(rune(ansi.DEL))
	expected := "\u2421"
	result := Escape(input)
	if result != expected {
		t.Errorf("Escape(DEL) = %q, want %q", result, expected)
	}
}

// TestEscapeLength ensures proper string builder growth
func TestEscapeLength(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"short", "test"},
		{"empty", ""},
		{"long", strings.Repeat("a", 1000)},
		{"long with controls", strings.Repeat("a\x00b\x01c", 100)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Escape(tt.input)
			// Result should have same number of runes as input
			// (control chars are replaced with single unicode chars)
			inputRunes := []rune(tt.input)
			resultRunes := []rune(result)
			if len(inputRunes) != len(resultRunes) {
				t.Errorf("Escape changed rune count: input=%d, result=%d",
					len(inputRunes), len(resultRunes))
			}
		})
	}
}

// BenchmarkEscape tests performance of the Escape function
func BenchmarkEscape(b *testing.B) {
	testCases := []struct {
		name  string
		input string
	}{
		{"plain text", "Hello, World! This is a test string with no control characters."},
		{"with controls", "Hello\x00World\x01Test\x02String\x03"},
		{"many controls", strings.Repeat("\x00\x01\x02\x03\x04\x05", 10)},
		{"large plain", strings.Repeat("abcdefghij", 100)},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				Escape(tc.input)
			}
		})
	}
}
