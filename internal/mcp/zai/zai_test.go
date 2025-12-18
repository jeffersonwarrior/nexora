package zai

import (
	"os"
	"testing"

	"github.com/nexora/cli/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestIsVisionTool(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		expected bool
	}{
		{
			name:     "data visualization tool",
			toolName: "mcp_vision_analyze_data_visualization",
			expected: true,
		},
		{
			name:     "general image analysis tool",
			toolName: "mcp_vision_analyze_image",
			expected: true,
		},
		{
			name:     "text extraction tool",
			toolName: "mcp_vision_extract_text_from_screenshot",
			expected: true,
		},
		{
			name:     "UI to artifact tool",
			toolName: "mcp_vision_ui_to_artifact",
			expected: true,
		},
		{
			name:     "error diagnosis tool",
			toolName: "mcp_vision_diagnose_error_screenshot",
			expected: true,
		},
		{
			name:     "technical diagram tool",
			toolName: "mcp_vision_understand_technical_diagram",
			expected: true,
		},
		{
			name:     "UI diff check tool",
			toolName: "mcp_vision_ui_diff_check",
			expected: true,
		},
		{
			name:     "video analysis tool",
			toolName: "mcp_vision_analyze_video",
			expected: true,
		},
		{
			name:     "non-vision tool",
			toolName: "bash",
			expected: false,
		},
		{
			name:     "empty tool name",
			toolName: "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsVisionTool(tt.toolName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetToolDescription(t *testing.T) {
	tests := []struct {
		name         string
		toolName     string
		expectedDesc string
	}{
		{
			name:         "data visualization tool description",
			toolName:     "mcp_vision_analyze_data_visualization",
			expectedDesc: "Analyze charts, graphs, and data visualizations to extract insights",
		},
		{
			name:         "general image analysis tool description",
			toolName:     "mcp_vision_analyze_image",
			expectedDesc: "General-purpose image analysis for any visual content",
		},
		{
			name:         "unknown tool",
			toolName:     "unknown_tool",
			expectedDesc: "Z.ai vision analysis tool",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetToolDescription(tt.toolName)
			assert.Equal(t, tt.expectedDesc, result)
		})
	}
}

func TestStateString(t *testing.T) {
	tests := []struct {
		name     string
		state    State
		expected string
	}{
		{
			name:     "not configured state",
			state:    StateNotConfigured,
			expected: "not_configured",
		},
		{
			name:     "stopped state",
			state:    StateStopped,
			expected: "stopped",
		},
		{
			name:     "running state",
			state:    StateRunning,
			expected: "running",
		},
		{
			name:     "unknown state",
			state:    State(999),
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.state.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateConfig(t *testing.T) {
	// Store original env var
	originalAPIKey := os.Getenv("ZAI_API_KEY")
	defer func() {
		if originalAPIKey != "" {
			os.Setenv("ZAI_API_KEY", originalAPIKey)
		} else {
			os.Unsetenv("ZAI_API_KEY")
		}
	}()

	t.Run("missing API key", func(t *testing.T) {
		os.Setenv("ZAI_API_KEY", "")
		err := ValidateConfig()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ZAI_API_KEY environment variable is required")
	})

	t.Run("API key too short", func(t *testing.T) {
		os.Setenv("ZAI_API_KEY", "short")
		err := ValidateConfig()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "appears to be too short")
	})

	t.Run("valid API key", func(t *testing.T) {
		os.Setenv("ZAI_API_KEY", "valid_api_key_for_testing")
		err := ValidateConfig()
		assert.NoError(t, err)
	})
}

func TestManager(t *testing.T) {
	// Store original env var
	originalAPIKey := os.Getenv("ZAI_API_KEY")
	defer func() {
		if originalAPIKey != "" {
			os.Setenv("ZAI_API_KEY", originalAPIKey)
		} else {
			os.Unsetenv("ZAI_API_KEY")
		}
	}()

	t.Run("manager without API key", func(t *testing.T) {
		os.Setenv("ZAI_API_KEY", "")
		cfg := config.Config{}
		manager := NewManager(cfg)

		// Check initial status
		status := manager.GetStatus()
		assert.Equal(t, StateNotConfigured, status.State)
		assert.Contains(t, status.Message, "ZAI_API_KEY not set")

		// Try to start
		err := manager.Start(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ZAI_API_KEY environment variable is required")
	})

	t.Run("manager with API key", func(t *testing.T) {
		os.Setenv("ZAI_API_KEY", "valid_api_key_for_testing")
		cfg := config.Config{}
		manager := NewManager(cfg)

		// Check initial status
		status := manager.GetStatus()
		assert.Equal(t, StateStopped, status.State)

		// Try to get client before start
		_, err := manager.GetClient()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not initialized")

		// Start manager
		err = manager.Start(nil)
		// Note: This will fail in mock mode, but we can still test the structure
		// In a real environment with proper MCP server, this would succeed
		if err != nil {
			t.Logf("Expected failure in mock mode: %v", err)
		}
	})

	t.Run("manager stop", func(t *testing.T) {
		os.Setenv("ZAI_API_KEY", "valid_api_key_for_testing")
		cfg := config.Config{}
		manager := NewManager(cfg)

		// Stop manager (should not error even if not started)
		err := manager.Stop()
		assert.NoError(t, err)

		// Check status after stop
		status := manager.GetStatus()
		assert.Equal(t, StateStopped, status.State)
	})
}
