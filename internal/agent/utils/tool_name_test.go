package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSanitizeToolName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "clean tool name",
			input:    "grep",
			expected: "grep",
		},
		{
			name:     "tool name with underscores",
			input:    "grep_path_pattern",
			expected: "grep_path_pattern",
		},
		{
			name:     "tool name with hyphens",
			input:    "web-fetch",
			expected: "web-fetch",
		},
		{
			name:     "XML corruption - arg_key",
			input:    "grep_path_pattern</arg_key><arg_value>internal/permission</arg_value>",
			expected: "grep_path_pattern",
		},
		{
			name:     "XML corruption - closing tag",
			input:    "view_path</arg_key><arg_value>/home/nexora/cmd/nexora/main.go</arg_value>",
			expected: "view_path",
		},
		{
			name:     "JSON corruption",
			input:    "bash{\"command\":\"ls\"}",
			expected: "bash",
		},
		{
			name:     "special characters",
			input:    "tool!@#$%^&*()name",
			expected: "toolname",
		},
		{
			name:     "spaces",
			input:    "tool name with spaces",
			expected: "toolnamewithspaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeToolName(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}
