package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"charm.land/fantasy"
	"github.com/nexora/nexora/internal/csync"
	"github.com/nexora/nexora/internal/lsp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteTool(t *testing.T) {
	tmpDir := t.TempDir()
	lspClients := csync.NewMap[string, *lsp.Client]()
	mockPerms := &mockPermissionService{}
	mockFiles := &mockHistoryService{}

	tool := NewWriteTool(lspClients, mockPerms, mockFiles, tmpDir)
	assert.NotNil(t, tool)
	assert.Equal(t, WriteToolName, tool.Info().Name)
}

func TestWriteTool_CreateNewFile(t *testing.T) {
	tmpDir := t.TempDir()
	lspClients := csync.NewMap[string, *lsp.Client]()
	mockPerms := &mockPermissionService{}
	mockFiles := &mockHistoryService{}
	tool := NewWriteTool(lspClients, mockPerms, mockFiles, tmpDir)

	testFile := filepath.Join(tmpDir, "newfile.txt")
	content := "Hello, World!"

	ctx := context.WithValue(context.Background(), SessionIDContextKey, "test-session")
	params := WriteParams{
		FilePath: testFile,
		Content:  content,
	}
	paramsJSON, _ := json.Marshal(params)
	toolCall := fantasy.ToolCall{
		ID:    "test",
		Name:  WriteToolName,
		Input: string(paramsJSON),
	}

	resp, err := tool.Run(ctx, toolCall)

	require.NoError(t, err)
	assert.False(t, resp.IsError)

	// Verify file was created
	data, err := os.ReadFile(testFile)
	require.NoError(t, err)
	assert.Equal(t, content, string(data))
}

func TestWriteTool_OverwriteFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "existing.txt")

	// Create initial file
	err := os.WriteFile(testFile, []byte("old content"), 0644)
	require.NoError(t, err)

	// Record read to pass modification time check
	recordFileRead(testFile)
	time.Sleep(10 * time.Millisecond)

	lspClients := csync.NewMap[string, *lsp.Client]()
	mockPerms := &mockPermissionService{}
	mockFiles := &mockHistoryService{}
	tool := NewWriteTool(lspClients, mockPerms, mockFiles, tmpDir)

	newContent := "new content"
	ctx := context.WithValue(context.Background(), SessionIDContextKey, "test-session")
	params := WriteParams{
		FilePath: testFile,
		Content:  newContent,
	}
	paramsJSON, _ := json.Marshal(params)
	toolCall := fantasy.ToolCall{
		ID:    "test",
		Name:  WriteToolName,
		Input: string(paramsJSON),
	}

	resp, err := tool.Run(ctx, toolCall)

	require.NoError(t, err)
	assert.False(t, resp.IsError)

	// Verify file was updated
	data, err := os.ReadFile(testFile)
	require.NoError(t, err)
	assert.Equal(t, newContent, string(data))
}

func TestWriteTool_DirectoryCreation(t *testing.T) {
	tmpDir := t.TempDir()
	lspClients := csync.NewMap[string, *lsp.Client]()
	mockPerms := &mockPermissionService{}
	mockFiles := &mockHistoryService{}
	tool := NewWriteTool(lspClients, mockPerms, mockFiles, tmpDir)

	testFile := filepath.Join(tmpDir, "subdir", "nested", "file.txt")
	content := "nested file"

	ctx := context.WithValue(context.Background(), SessionIDContextKey, "test-session")
	params := WriteParams{
		FilePath: testFile,
		Content:  content,
	}
	paramsJSON, _ := json.Marshal(params)
	toolCall := fantasy.ToolCall{
		ID:    "test",
		Name:  WriteToolName,
		Input: string(paramsJSON),
	}

	resp, err := tool.Run(ctx, toolCall)

	require.NoError(t, err)
	assert.False(t, resp.IsError)

	// Verify file and directories were created
	data, err := os.ReadFile(testFile)
	require.NoError(t, err)
	assert.Equal(t, content, string(data))
}

func TestWriteTool_FileIsDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	dirPath := filepath.Join(tmpDir, "testdir")
	err := os.Mkdir(dirPath, 0755)
	require.NoError(t, err)

	lspClients := csync.NewMap[string, *lsp.Client]()
	mockPerms := &mockPermissionService{}
	mockFiles := &mockHistoryService{}
	tool := NewWriteTool(lspClients, mockPerms, mockFiles, tmpDir)

	ctx := context.WithValue(context.Background(), SessionIDContextKey, "test-session")
	params := WriteParams{
		FilePath: dirPath,
		Content:  "test",
	}
	paramsJSON, _ := json.Marshal(params)
	toolCall := fantasy.ToolCall{
		ID:    "test",
		Name:  WriteToolName,
		Input: string(paramsJSON),
	}

	resp, err := tool.Run(ctx, toolCall)

	require.NoError(t, err)
	assert.True(t, resp.IsError)
	assert.Contains(t, resp.Content, "directory")
}
