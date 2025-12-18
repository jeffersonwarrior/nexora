package config

import (
	"os"
	"testing"
)

func TestApplyEnvironmentVariables(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		input    ProviderConfig
		expected ProviderConfig
	}{
		{
			name: "apply authorization header",
			envVars: map[string]string{
				"NEXORA_PROVIDER_HEADER_AUTHORIZATION": "Bearer test-token",
			},
			input: ProviderConfig{
				BaseURL: "https://api.example.com",
			},
			expected: ProviderConfig{
				BaseURL: "https://api.example.com",
				ExtraHeaders: map[string]string{
					"authorization": "Bearer test-token",
				},
			},
		},
		{
			name: "apply custom header and base URL",
			envVars: map[string]string{
				"NEXORA_PROVIDER_HEADER_X_CUSTOM": "custom-value",
				"NEXORA_PROVIDER_BASE_URL":        "https://custom.api.com",
			},
			input: ProviderConfig{
				BaseURL: "https://original.api.com",
			},
			expected: ProviderConfig{
				BaseURL: "https://custom.api.com",
				ExtraHeaders: map[string]string{
					"x-custom": "custom-value",
				},
			},
		},
		{
			name: "merge with existing headers",
			envVars: map[string]string{
				"NEXORA_PROVIDER_HEADER_AUTHORIZATION": "Bearer new-token",
			},
			input: ProviderConfig{
				BaseURL: "https://api.example.com",
				ExtraHeaders: map[string]string{
					"x-existing": "existing-value",
				},
			},
			expected: ProviderConfig{
				BaseURL: "https://api.example.com",
				ExtraHeaders: map[string]string{
					"x-existing":    "existing-value",
					"authorization": "Bearer new-token",
				},
			},
		},
		{
			name: "no environment variables",
			input: ProviderConfig{
				BaseURL: "https://api.example.com",
			},
			expected: ProviderConfig{
				BaseURL: "https://api.example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Store original environment variables for this test
			originalVars := make(map[string]string)
			for k := range tt.envVars {
				originalVars[k] = os.Getenv(k)
			}

			// Clean up environment variables for this test
			t.Cleanup(func() {
				for k, v := range originalVars {
					if v != "" {
						os.Setenv(k, v)
					} else {
						os.Unsetenv(k)
					}
				}
			})

			// Clear all test environment variables first
			for k := range tt.envVars {
				os.Unsetenv(k)
			}

			// Set test environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			// Apply environment variables deep copy of input
			config := tt.input
			if config.ExtraHeaders != nil {
				// Make a copy of headers to avoid test pollution
				config.ExtraHeaders = make(map[string]string)
				for k, v := range tt.input.ExtraHeaders {
					config.ExtraHeaders[k] = v
				}
			}

			applyEnvironmentVariables(&config)

			// Compare results
			if config.BaseURL != tt.expected.BaseURL {
				t.Errorf("BaseURL = %v, want %v", config.BaseURL, tt.expected.BaseURL)
			}

			if len(config.ExtraHeaders) != len(tt.expected.ExtraHeaders) {
				t.Errorf("ExtraHeaders length = %d, want %d", len(config.ExtraHeaders), len(tt.expected.ExtraHeaders))
			}

			for k, v := range tt.expected.ExtraHeaders {
				if config.ExtraHeaders[k] != v {
					t.Errorf("ExtraHeaders[%s] = %v, want %v", k, config.ExtraHeaders[k], v)
				}
			}
		})
	}
}

func TestApplyEnvironmentVariablesWithMCP(t *testing.T) {
	// Test MCP-specific environment variables
	origHeader := os.Getenv("NEXORA_MCP_HEADER_AUTHORIZATION")
	t.Cleanup(func() {
		if origHeader != "" {
			os.Setenv("NEXORA_MCP_HEADER_AUTHORIZATION", origHeader)
		} else {
			os.Unsetenv("NEXORA_MCP_HEADER_AUTHORIZATION")
		}
	})

	os.Setenv("NEXORA_MCP_HEADER_AUTHORIZATION", "Bearer mcp-token")

	config := &ProviderConfig{
		BaseURL: "https://mcp.example.com",
	}

	applyEnvironmentVariables(config)

	if config.ExtraHeaders["authorization"] != "Bearer mcp-token" {
		t.Errorf("Expected MCP authorization header, got: %v", config.ExtraHeaders)
	}
}
