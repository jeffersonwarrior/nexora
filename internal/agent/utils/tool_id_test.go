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

func TestHasTextToolCalls(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected bool
	}{
		{
			name:     "MiniMax tool call format",
			text:     "Let me view that file minimax:tool_call /home/user/file.go 30 310 </minimax:tool_call>",
			expected: true,
		},
		{
			name:     "XML tool_call format",
			text:     "I'll call a tool <tool_call>{\"name\": \"view\", \"arguments\": {\"file_path\": \"/test\"}}</tool_call>",
			expected: true,
		},
		{
			name:     "function_call format",
			text:     "Calling <function_call name=\"bash\">echo hello</function_call>",
			expected: true,
		},
		{
			name:     "Plain text without tool calls",
			text:     "This is just regular text without any tool calls",
			expected: false,
		},
		{
			name:     "Empty string",
			text:     "",
			expected: false,
		},
		{
			name:     "Partial match - only opening tag",
			text:     "Some text with minimax:tool_call but no closing tag",
			expected: true, // HasTextToolCalls only checks for presence of patterns
		},
		{
			name:     "xai:function_call format",
			text:     "<xai:function_call name=\"bash\">ls</xai:function_call>",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasTextToolCalls(tt.text)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestParseTextToolCalls(t *testing.T) {
	t.Run("MiniMax tool call format - view with path only", func(t *testing.T) {
		text := "Let me view that file minimax:tool_call /home/user/file.go </minimax:tool_call>"
		calls := ParseTextToolCalls(text)
		require.Len(t, calls, 1)
		require.Equal(t, "view", calls[0].Name)
		require.Contains(t, calls[0].Arguments, "/home/user/file.go")
	})

	t.Run("MiniMax tool call format - view with path, limit, offset", func(t *testing.T) {
		text := "Now viewing minimax:tool_call /home/nexora/internal/agent/agent.go 30 310 </minimax:tool_call>"
		calls := ParseTextToolCalls(text)
		require.Len(t, calls, 1)
		require.Equal(t, "view", calls[0].Name)
		require.Contains(t, calls[0].Arguments, "/home/nexora/internal/agent/agent.go")
		require.Contains(t, calls[0].Arguments, "30")
		require.Contains(t, calls[0].Arguments, "310")
	})

	t.Run("XML tool_call format", func(t *testing.T) {
		text := `I'll use a tool <tool_call>{"name": "bash", "arguments": {"command": "ls -la"}}</tool_call>`
		calls := ParseTextToolCalls(text)
		require.Len(t, calls, 1)
		require.Equal(t, "bash", calls[0].Name)
	})

	t.Run("function_call format", func(t *testing.T) {
		text := `Executing <function_call name="grep">{"pattern": "test"}</function_call>`
		calls := ParseTextToolCalls(text)
		require.Len(t, calls, 1)
		require.Equal(t, "grep", calls[0].Name)
	})

	t.Run("Multiple tool calls", func(t *testing.T) {
		text := `First call <tool_call>{"name": "view", "arguments": {"file_path": "/a"}}</tool_call> and another <tool_call>{"name": "bash", "arguments": {"command": "pwd"}}</tool_call>`
		calls := ParseTextToolCalls(text)
		require.Len(t, calls, 2)
		require.Equal(t, "view", calls[0].Name)
		require.Equal(t, "bash", calls[1].Name)
	})

	t.Run("No tool calls in text", func(t *testing.T) {
		text := "This is plain text without any tool call patterns"
		calls := ParseTextToolCalls(text)
		require.Empty(t, calls)
	})

	t.Run("Malformed tool call - no closing tag", func(t *testing.T) {
		text := "minimax:tool_call /some/path without closing tag"
		calls := ParseTextToolCalls(text)
		require.Empty(t, calls) // Should not parse without closing tag
	})

	t.Run("Generated IDs are unique", func(t *testing.T) {
		text := `<tool_call>{"name": "a", "arguments": {}}</tool_call><tool_call>{"name": "b", "arguments": {}}</tool_call>`
		calls := ParseTextToolCalls(text)
		require.Len(t, calls, 2)
		require.NotEqual(t, calls[0].ID, calls[1].ID, "Each tool call should have a unique ID")
	})

	t.Run("Real-world MiniMax sample from bug report", func(t *testing.T) {
		// This is the actual format seen in the bug report
		text := `that when the default max tokens are 0 or undefined, it gets passed as 0 to the backend.                             
                                                                                                                      
 Let me add a validation fix to ensure that  MaxOutputTokens  is always at least 1:                                   
                                                                                                                      
 Now I understand the issue! The problem is that  MaxOutputTokens  is being set to 0 somewhere, and this gets passed  
 to the backend as  max_completion_tokens  with a value of 0. The backend validates that  max_completion_tokens       
 must be at least 1, so it fails with the error you're seeing.                                                        
                                                                                                                      
 Let me fix this by adding validation to ensure  MaxOutputTokens  is always at least 1: minimax:tool_call             
 /home/nexora/internal/agent/agent.go 30 310  </minimax:tool_call>`

		calls := ParseTextToolCalls(text)
		require.Len(t, calls, 1)
		require.Equal(t, "view", calls[0].Name)
		require.Contains(t, calls[0].Arguments, "/home/nexora/internal/agent/agent.go")
	})

	t.Run("xai:function_call format", func(t *testing.T) {
		text := `I'll use the bash tool <xai:function_call name="bash">
<parameter name="command">make test-qa</parameter>
</xai:function_call>`
		calls := ParseTextToolCalls(text)
		require.Len(t, calls, 1)
		require.Equal(t, "bash", calls[0].Name)
		require.Contains(t, calls[0].Arguments, "make test-qa")
	})

	t.Run("Multiple xai:function_call", func(t *testing.T) {
		text := `<xai:function_call name="view"><parameter name="file_path">/test</parameter></xai:function_call> then <xai:function_call name="bash"><parameter name="command">ls</parameter></xai:function_call>`
		calls := ParseTextToolCalls(text)
		require.Len(t, calls, 2)
		require.Equal(t, "view", calls[0].Name)
		require.Equal(t, "bash", calls[1].Name)
	})
}
