package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateToolCallID(t *testing.T) {
	t.Run("Mistral ID generation", func(t *testing.T) {
		id := GenerateToolCallID("mistral")
		require.Len(t, id, 9)
		require.True(t, ValidateMistralID(id), "Generated ID should be valid for Mistral")

		// Test uniqueness
		id2 := GenerateToolCallID("mistral")
		require.NotEqual(t, id, id2, "Generated IDs should be unique")

		// Test alphanumeric only
		for _, char := range id {
			require.True(t,
				(char >= 'a' && char <= 'z') ||
					(char >= 'A' && char <= 'Z') ||
					(char >= '0' && char <= '9'),
				"ID should contain only alphanumeric characters: %s", string(char))
		}
	})

	t.Run("Mistral-native ID generation", func(t *testing.T) {
		id := GenerateToolCallID("mistral-native")
		require.Len(t, id, 9)
		require.True(t, ValidateMistralID(id))
	})

	t.Run("OpenAI ID generation", func(t *testing.T) {
		id := GenerateToolCallID("openai")
		require.True(t, len(id) > 12, "OpenAI ID should include call_ prefix + random chars")
		require.Contains(t, id, "call_")
	})

	t.Run("Default provider", func(t *testing.T) {
		id := GenerateToolCallID("unknown")
		require.Contains(t, id, "call_")
	})
}

func TestValidateMistralID(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		expected bool
	}{
		{"Valid 9 chars", "abc123XYZ", true},
		{"Valid all letters", "ABCDEFGHI", true},
		{"Valid all numbers", "123456789", true},
		{"Valid mixed case", "AbCdEfGhI", true},
		{"Invalid too short", "abc123", false},
		{"Invalid too long", "abcdefghijk", false},
		{"Invalid contains underscore", "abc_12345", false},
		{"Invalid contains dash", "abc-12345", false},
		{"Invalid empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateMistralID(tt.id)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeToolCallID(t *testing.T) {
	t.Run("Mistral sanitization", func(t *testing.T) {
		tests := []struct {
			input    string
			expected int // expected length
		}{
			{"call_61626542", 9}, // Error case from bug report
			{"call_abcdefghijklmnopqrstu", 9},
			{"call_1234567890", 9},
			{"abc_123", 9}, // Non-standard format
			{"test123456789", 9},
			{"", 9}, // Empty input
		}

		for _, tt := range tests {
			t.Run(tt.input, func(t *testing.T) {
				result := SanitizeToolCallID(tt.input, "mistral")
				require.Len(t, result, 9)
				require.True(t, ValidateMistralID(result),
					"Sanitized ID should be valid: %s from %s", result, tt.input)
			})
		}
	})

	t.Run("OpenAI passthrough", func(t *testing.T) {
		input := "call_61626542"
		result := SanitizeToolCallID(input, "openai")
		require.Equal(t, input, result, "OpenAI IDs should pass through unchanged")
	})

	t.Run("Mistral-nativE sanitization", func(t *testing.T) {
		input := "call_61626542"
		result := SanitizeToolCallID(input, "mistral-native")
		require.Len(t, result, 9)
		require.True(t, ValidateMistralID(result))
	})
}

func TestEdgeCases(t *testing.T) {
	t.Run("Multiple sanitizations consistency", func(t *testing.T) {
		input := "call_61626542"
		result1 := SanitizeToolCallID(input, "mistral")
		result2 := SanitizeToolCallID(input, "mistral")

		// Results might be different due to random padding, but both should be valid
		require.True(t, ValidateMistralID(result1))
		require.True(t, ValidateMistralID(result2))
	})

	t.Run("All invalid chars input", func(t *testing.T) {
		input := "!@#$%^&*()_+-=[]{}|;':\",./<>?"
		result := SanitizeToolCallID(input, "mistral")
		require.Len(t, result, 9)
		require.True(t, ValidateMistralID(result))
	})
}
