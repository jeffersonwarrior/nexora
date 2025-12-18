package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestNoCrushReferences ensures CRUSH.md references are replaced with NEXORA.md
func TestNoCrushReferences(t *testing.T) {
	t.Parallel()

	// Check defaultContextPaths doesn't contain CRUSH
	for _, path := range defaultContextPaths {
		require.NotContains(t, strings.ToUpper(path), "CRUSH",
			"defaultContextPaths should not contain CRUSH references, found: %s", path)
	}

	// Check project root for CRUSH.md files
	root := filepath.Join("..", "..")
	entries, err := os.ReadDir(root)
	require.NoError(t, err)

	for _, entry := range entries {
		if !entry.IsDir() {
			name := entry.Name()
			if strings.Contains(strings.ToUpper(name), "CRUSH") && strings.HasSuffix(name, ".md") {
				t.Errorf("Found CRUSH markdown file in project root: %s (should use NEXORA.md or AGENTS.md)", name)
			}
		}
	}
}

// TestConfigSchemaUsesNexora ensures JSON schema examples use NEXORA.md not CRUSH.md
func TestConfigSchemaUsesNexora(t *testing.T) {
	t.Parallel()

	// Read the config.go file
	configPath := "config.go"
	content, err := os.ReadFile(configPath)
	require.NoError(t, err)

	lines := strings.Split(string(content), "\n")
	
	for i, line := range lines {
		// Check jsonschema tags and comments
		if (strings.Contains(line, "jsonschema") || strings.Contains(line, "example=")) &&
			strings.Contains(strings.ToUpper(line), "CRUSH") {
			t.Errorf("Line %d contains CRUSH in schema/example: %s\nShould use NEXORA.md instead", i+1, strings.TrimSpace(line))
		}
	}
}

// TestAgentsMarkdownUsesNexora checks AGENTS.md doesn't reference CRUSH.md
func TestAgentsMarkdownUsesNexora(t *testing.T) {
	t.Parallel()

	agentsPath := filepath.Join("..", "..", "AGENTS.md")
	
	// Skip if AGENTS.md doesn't exist
	if _, err := os.Stat(agentsPath); os.IsNotExist(err) {
		t.Skip("AGENTS.md not found, skipping")
		return
	}

	content, err := os.ReadFile(agentsPath)
	require.NoError(t, err)

	lines := strings.Split(string(content), "\n")
	
	for i, line := range lines {
		if strings.Contains(strings.ToUpper(line), "CRUSH.MD") {
			t.Errorf("AGENTS.md line %d references CRUSH.md: %s\nShould use NEXORA.md instead", 
				i+1, strings.TrimSpace(line))
		}
	}
}

// TestTaskfileUsesNexora checks Taskfile.yaml doesn't use CRUSH_PROFILE
func TestTaskfileUsesNexora(t *testing.T) {
	t.Parallel()

	taskfilePath := filepath.Join("..", "..", "Taskfile.yaml")
	
	// Skip if Taskfile.yaml doesn't exist
	if _, err := os.Stat(taskfilePath); os.IsNotExist(err) {
		t.Skip("Taskfile.yaml not found, skipping")
		return
	}

	content, err := os.ReadFile(taskfilePath)
	require.NoError(t, err)

	if strings.Contains(string(content), "CRUSH_PROFILE") {
		t.Error("Taskfile.yaml contains CRUSH_PROFILE, should use NEXORA_PROFILE instead")
	}
}
