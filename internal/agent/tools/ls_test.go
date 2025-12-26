package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"charm.land/fantasy"
	"github.com/nexora/nexora/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLsTool(t *testing.T) {
	tmpDir := t.TempDir()
	mockPerms := &mockPermissionService{}
	lsConfig := config.ToolLs{}

	tool := NewLsTool(mockPerms, tmpDir, lsConfig)
	assert.NotNil(t, tool)
	assert.Equal(t, LSToolName, tool.Info().Name)
}

func TestLsTool_ListCurrentDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test structure
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("test"), 0644)
	os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "subdir", "nested.txt"), []byte("test"), 0644)

	mockPerms := &mockPermissionService{}
	lsConfig := config.ToolLs{}
	tool := NewLsTool(mockPerms, tmpDir, lsConfig)

	ctx := context.Background()
	params := LSParams{}
	paramsJSON, _ := json.Marshal(params)
	toolCall := fantasy.ToolCall{
		ID:    "test",
		Name:  LSToolName,
		Input: string(paramsJSON),
	}

	resp, err := tool.Run(ctx, toolCall)

	require.NoError(t, err)
	assert.False(t, resp.IsError)

	output := resp.Content
	assert.Contains(t, output, "file1.txt")
	assert.Contains(t, output, "file2.txt")
	assert.Contains(t, output, "subdir/")
}

func TestLsTool_NonExistentDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	mockPerms := &mockPermissionService{}
	lsConfig := config.ToolLs{}
	tool := NewLsTool(mockPerms, tmpDir, lsConfig)

	ctx := context.Background()
	params := LSParams{
		Path: filepath.Join(tmpDir, "nonexistent"),
	}
	paramsJSON, _ := json.Marshal(params)
	toolCall := fantasy.ToolCall{
		ID:    "test",
		Name:  LSToolName,
		Input: string(paramsJSON),
	}

	_, err := tool.Run(ctx, toolCall)

	// The tool returns an error for non-existent directories
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestListDirectoryTree(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test structure
	os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("test"), 0644)
	os.Mkdir(filepath.Join(tmpDir, "dir"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "dir", "nested.txt"), []byte("test"), 0644)

	output, metadata, err := ListDirectoryTree(tmpDir, LSParams{}, config.ToolLs{})

	require.NoError(t, err)
	assert.Greater(t, metadata.NumberOfFiles, 0)
	assert.Contains(t, output, "file.txt")
	assert.Contains(t, output, "dir/")
}
