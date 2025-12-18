package stringext

import (
	"testing"
)

func TestCapitalize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "lowercase word",
			input:    "hello",
			expected: "Hello",
		},
		{
			name:     "uppercase word",
			input:    "HELLO",
			expected: "Hello",
		},
		{
			name:     "mixed case word",
			input:    "hElLo",
			expected: "Hello",
		},
		{
			name:     "multiple words",
			input:    "hello world",
			expected: "Hello World",
		},
		{
			name:     "sentence with punctuation",
			input:    "hello, world!",
			expected: "Hello, World!",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "single character lowercase",
			input:    "a",
			expected: "A",
		},
		{
			name:     "single character uppercase",
			input:    "A",
			expected: "A",
		},
		{
			name:     "numbers",
			input:    "123 test",
			expected: "123 Test",
		},
		{
			name:     "all caps sentence",
			input:    "THIS IS A TEST",
			expected: "This Is A Test",
		},
		{
			name:     "camelCase",
			input:    "camelCaseString",
			expected: "Camelcasestring",
		},
		{
			name:     "snake_case",
			input:    "snake_case_string",
			expected: "Snake_case_string",
		},
		{
			name:     "kebab-case",
			input:    "kebab-case-string",
			expected: "Kebab-Case-String",
		},
		{
			name:     "leading whitespace",
			input:    "  hello",
			expected: "  Hello",
		},
		{
			name:     "trailing whitespace",
			input:    "hello  ",
			expected: "Hello  ",
		},
		{
			name:     "unicode characters",
			input:    "café résumé",
			expected: "Café Résumé",
		},
		{
			name:     "mixed unicode and ascii",
			input:    "hello 世界",
			expected: "Hello 世界",
		},
		{
			name:     "apostrophes",
			input:    "it's a beautiful day",
			expected: "It's A Beautiful Day",
		},
		{
			name:     "acronyms",
			input:    "http api endpoint",
			expected: "Http Api Endpoint",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Capitalize(tt.input)
			if result != tt.expected {
				t.Errorf("Capitalize(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestContainsAny(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		args     []string
		expected bool
	}{
		{
			name:     "contains first arg",
			str:      "hello world",
			args:     []string{"hello", "goodbye"},
			expected: true,
		},
		{
			name:     "contains second arg",
			str:      "hello world",
			args:     []string{"goodbye", "world"},
			expected: true,
		},
		{
			name:     "contains all args",
			str:      "hello world",
			args:     []string{"hello", "world"},
			expected: true,
		},
		{
			name:     "contains none",
			str:      "hello world",
			args:     []string{"foo", "bar"},
			expected: false,
		},
		{
			name:     "empty string",
			str:      "",
			args:     []string{"hello"},
			expected: false,
		},
		{
			name:     "empty args",
			str:      "hello world",
			args:     []string{},
			expected: false,
		},
		{
			name:     "single arg matches",
			str:      "hello",
			args:     []string{"hello"},
			expected: true,
		},
		{
			name:     "single arg doesn't match",
			str:      "hello",
			args:     []string{"world"},
			expected: false,
		},
		{
			name:     "substring match",
			str:      "hello world",
			args:     []string{"llo wo"},
			expected: true,
		},
		{
			name:     "case sensitive - no match",
			str:      "hello world",
			args:     []string{"HELLO", "WORLD"},
			expected: false,
		},
		{
			name:     "partial match in middle",
			str:      "hello world",
			args:     []string{"o w"},
			expected: true,
		},
		{
			name:     "empty string in args",
			str:      "hello",
			args:     []string{""},
			expected: true, // empty string is contained in any string
		},
		{
			name:     "multiple empty strings in args",
			str:      "hello",
			args:     []string{"", ""},
			expected: true,
		},
		{
			name:     "unicode characters",
			str:      "hello 世界",
			args:     []string{"世界"},
			expected: true,
		},
		{
			name:     "special characters",
			str:      "hello, world!",
			args:     []string{", "},
			expected: true,
		},
		{
			name:     "newline in string",
			str:      "hello\nworld",
			args:     []string{"\n"},
			expected: true,
		},
		{
			name:     "tab in string",
			str:      "hello\tworld",
			args:     []string{"\t"},
			expected: true,
		},
		{
			name:     "exact match",
			str:      "hello",
			args:     []string{"hello"},
			expected: true,
		},
		{
			name:     "multiple args, none match",
			str:      "test string",
			args:     []string{"foo", "bar", "baz"},
			expected: false,
		},
		{
			name:     "multiple args, last one matches",
			str:      "test string",
			args:     []string{"foo", "bar", "string"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ContainsAny(tt.str, tt.args...)
			if result != tt.expected {
				t.Errorf("ContainsAny(%q, %v) = %v, want %v", tt.str, tt.args, result, tt.expected)
			}
		})
	}
}

// TestContainsAnyPerformance ensures the function short-circuits on first match
func TestContainsAnyPerformance(t *testing.T) {
	// If the function doesn't short-circuit, this would take longer
	largeArgs := make([]string, 10000)
	for i := range largeArgs {
		largeArgs[i] = "nonexistent"
	}
	// Put a match at the beginning
	largeArgs[0] = "match"
	
	result := ContainsAny("this string contains a match", largeArgs...)
	if !result {
		t.Error("Expected to find match")
	}
	
	// Now test with match at the end
	largeArgs[0] = "nonexistent"
	largeArgs[9999] = "match"
	result = ContainsAny("this string contains a match", largeArgs...)
	if !result {
		t.Error("Expected to find match at end")
	}
}

// BenchmarkCapitalize tests performance of Capitalize
func BenchmarkCapitalize(b *testing.B) {
	testCases := []struct {
		name  string
		input string
	}{
		{"short", "hello"},
		{"medium", "hello world this is a test"},
		{"long", "the quick brown fox jumps over the lazy dog multiple times in this longer sentence"},
	}
	
	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				Capitalize(tc.input)
			}
		})
	}
}

// BenchmarkContainsAny tests performance of ContainsAny
func BenchmarkContainsAny(b *testing.B) {
	testCases := []struct {
		name string
		str  string
		args []string
	}{
		{"match_first", "hello world", []string{"hello", "foo", "bar"}},
		{"match_last", "hello world", []string{"foo", "bar", "world"}},
		{"no_match", "hello world", []string{"foo", "bar", "baz"}},
	}
	
	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ContainsAny(tc.str, tc.args...)
			}
		})
	}
}
