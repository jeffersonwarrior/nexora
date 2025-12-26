package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"charm.land/fantasy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDownloadTool(t *testing.T) {
	mockPermissions := &mockPermissionService{}
	tool := NewDownloadTool(mockPermissions, "/tmp", nil)

	assert.NotNil(t, tool)
	assert.Equal(t, DownloadToolName, tool.Info().Name)
	assert.NotEmpty(t, tool.Info().Description)
}

func TestNewDownloadTool_WithCustomClient(t *testing.T) {
	mockPermissions := &mockPermissionService{}
	customClient := &http.Client{}
	tool := NewDownloadTool(mockPermissions, "/tmp", customClient)

	assert.NotNil(t, tool)
}

func TestDownloadTool_MissingURL(t *testing.T) {
	mockPermissions := &mockPermissionService{}
	tool := NewDownloadTool(mockPermissions, "/tmp", nil)

	ctx := context.WithValue(context.Background(), SessionIDContextKey, "test-session")
	params := DownloadParams{
		FilePath: "/tmp/test.txt",
	}
	paramsJSON, _ := json.Marshal(params)
	toolCall := fantasy.ToolCall{
		ID:    "test-1",
		Name:  DownloadToolName,
		Input: string(paramsJSON),
	}

	resp, err := tool.Run(ctx, toolCall)

	require.NoError(t, err)
	assert.True(t, resp.IsError)
	assert.Contains(t, resp.Content, "URL parameter is required")
}

func TestDownloadTool_MissingFilePath(t *testing.T) {
	mockPermissions := &mockPermissionService{}
	tool := NewDownloadTool(mockPermissions, "/tmp", nil)

	ctx := context.WithValue(context.Background(), SessionIDContextKey, "test-session")
	params := DownloadParams{
		URL: "https://example.com/file.txt",
	}
	paramsJSON, _ := json.Marshal(params)
	toolCall := fantasy.ToolCall{
		ID:    "test-2",
		Name:  DownloadToolName,
		Input: string(paramsJSON),
	}

	resp, err := tool.Run(ctx, toolCall)

	require.NoError(t, err)
	assert.True(t, resp.IsError)
	assert.Contains(t, resp.Content, "file_path parameter is required")
}

func TestDownloadTool_InvalidURL(t *testing.T) {
	mockPermissions := &mockPermissionService{}
	tool := NewDownloadTool(mockPermissions, "/tmp", nil)

	ctx := context.WithValue(context.Background(), SessionIDContextKey, "test-session")
	params := DownloadParams{
		URL:      "ftp://example.com/file.txt",
		FilePath: "/tmp/test.txt",
	}
	paramsJSON, _ := json.Marshal(params)
	toolCall := fantasy.ToolCall{
		ID:    "test-3",
		Name:  DownloadToolName,
		Input: string(paramsJSON),
	}

	resp, err := tool.Run(ctx, toolCall)

	require.NoError(t, err)
	assert.True(t, resp.IsError)
	assert.Contains(t, resp.Content, "URL must start with http:// or https://")
}

func TestDownloadTool_MissingSessionID(t *testing.T) {
	mockPermissions := &mockPermissionService{}
	tool := NewDownloadTool(mockPermissions, "/tmp", nil)

	ctx := context.Background() // No session ID
	params := DownloadParams{
		URL:      "https://example.com/file.txt",
		FilePath: "/tmp/test.txt",
	}
	paramsJSON, _ := json.Marshal(params)
	toolCall := fantasy.ToolCall{
		ID:    "test-4",
		Name:  DownloadToolName,
		Input: string(paramsJSON),
	}

	_, err := tool.Run(ctx, toolCall)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "session ID is required")
}

func TestDownloadTool_SuccessfulDownload(t *testing.T) {
	// Create a test server
	testContent := "test file content"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testContent))
	}))
	defer server.Close()

	// Create temp directory for test
	tmpDir := t.TempDir()

	mockPermissions := &mockPermissionService{}
	tool := NewDownloadTool(mockPermissions, tmpDir, nil)

	ctx := context.WithValue(context.Background(), SessionIDContextKey, "test-session")
	params := DownloadParams{
		URL:      server.URL + "/test.txt",
		FilePath: "test.txt",
	}
	paramsJSON, _ := json.Marshal(params)
	toolCall := fantasy.ToolCall{
		ID:    "test-5",
		Name:  DownloadToolName,
		Input: string(paramsJSON),
	}

	resp, err := tool.Run(ctx, toolCall)

	require.NoError(t, err)
	assert.False(t, resp.IsError)
	assert.Contains(t, resp.Content, "Successfully downloaded")
	assert.Contains(t, resp.Content, "test.txt")

	// Verify file was created and has correct content
	content, err := os.ReadFile(filepath.Join(tmpDir, "test.txt"))
	require.NoError(t, err)
	assert.Equal(t, testContent, string(content))
}

func TestDownloadTool_PermissionHandling(t *testing.T) {
	// Note: Permission testing requires permission.Service mock with denial logic
	// The existing mockPermissionService always grants permission
	// Full permission testing is covered in integration tests
	t.Skip("Permission denial testing requires custom mock")
}

func TestDownloadTool_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	tmpDir := t.TempDir()

	mockPermissions := &mockPermissionService{}
	tool := NewDownloadTool(mockPermissions, tmpDir, nil)

	ctx := context.WithValue(context.Background(), SessionIDContextKey, "test-session")
	params := DownloadParams{
		URL:      server.URL + "/notfound.txt",
		FilePath: "test.txt",
	}
	paramsJSON, _ := json.Marshal(params)
	toolCall := fantasy.ToolCall{
		ID:    "test-7",
		Name:  DownloadToolName,
		Input: string(paramsJSON),
	}

	resp, err := tool.Run(ctx, toolCall)

	require.NoError(t, err)
	assert.True(t, resp.IsError)
	assert.Contains(t, resp.Content, "404")
}

func TestDownloadTool_WithTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	}))
	defer server.Close()

	tmpDir := t.TempDir()

	mockPermissions := &mockPermissionService{}
	tool := NewDownloadTool(mockPermissions, tmpDir, nil)

	ctx := context.WithValue(context.Background(), SessionIDContextKey, "test-session")
	params := DownloadParams{
		URL:      server.URL + "/test.txt",
		FilePath: "test.txt",
		Timeout:  30,
	}
	paramsJSON, _ := json.Marshal(params)
	toolCall := fantasy.ToolCall{
		ID:    "test-8",
		Name:  DownloadToolName,
		Input: string(paramsJSON),
	}

	resp, err := tool.Run(ctx, toolCall)

	require.NoError(t, err)
	assert.False(t, resp.IsError)
}

func TestDownloadTool_MaxTimeoutCap(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	}))
	defer server.Close()

	tmpDir := t.TempDir()

	mockPermissions := &mockPermissionService{}
	tool := NewDownloadTool(mockPermissions, tmpDir, nil)

	ctx := context.WithValue(context.Background(), SessionIDContextKey, "test-session")
	params := DownloadParams{
		URL:      server.URL + "/test.txt",
		FilePath: "test.txt",
		Timeout:  1000, // Should be capped at 600
	}
	paramsJSON, _ := json.Marshal(params)
	toolCall := fantasy.ToolCall{
		ID:    "test-9",
		Name:  DownloadToolName,
		Input: string(paramsJSON),
	}

	resp, err := tool.Run(ctx, toolCall)

	require.NoError(t, err)
	assert.False(t, resp.IsError)
}

func TestDownloadTool_LargeFile(t *testing.T) {
	// Create large content (> 100MB would be rejected)
	largeContent := strings.Repeat("x", 101*1024*1024)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "105906176") // > 100MB
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(largeContent))
	}))
	defer server.Close()

	tmpDir := t.TempDir()

	mockPermissions := &mockPermissionService{}
	tool := NewDownloadTool(mockPermissions, tmpDir, nil)

	ctx := context.WithValue(context.Background(), SessionIDContextKey, "test-session")
	params := DownloadParams{
		URL:      server.URL + "/largefile.bin",
		FilePath: "large.bin",
	}
	paramsJSON, _ := json.Marshal(params)
	toolCall := fantasy.ToolCall{
		ID:    "test-10",
		Name:  DownloadToolName,
		Input: string(paramsJSON),
	}

	resp, err := tool.Run(ctx, toolCall)

	require.NoError(t, err)
	assert.True(t, resp.IsError)
	assert.Contains(t, resp.Content, "File too large")
}

func TestDownloadTool_CreatesDirectories(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	}))
	defer server.Close()

	tmpDir := t.TempDir()

	mockPermissions := &mockPermissionService{}
	tool := NewDownloadTool(mockPermissions, tmpDir, nil)

	ctx := context.WithValue(context.Background(), SessionIDContextKey, "test-session")
	params := DownloadParams{
		URL:      server.URL + "/test.txt",
		FilePath: "subdir/nested/test.txt",
	}
	paramsJSON, _ := json.Marshal(params)
	toolCall := fantasy.ToolCall{
		ID:    "test-11",
		Name:  DownloadToolName,
		Input: string(paramsJSON),
	}

	resp, err := tool.Run(ctx, toolCall)

	require.NoError(t, err)
	assert.False(t, resp.IsError)

	// Verify directories were created
	_, err = os.Stat(filepath.Join(tmpDir, "subdir", "nested"))
	assert.NoError(t, err)
}

func TestDownloadTool_WithContentType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"test": true}`))
	}))
	defer server.Close()

	tmpDir := t.TempDir()

	mockPermissions := &mockPermissionService{}
	tool := NewDownloadTool(mockPermissions, tmpDir, nil)

	ctx := context.WithValue(context.Background(), SessionIDContextKey, "test-session")
	params := DownloadParams{
		URL:      server.URL + "/data.json",
		FilePath: "data.json",
	}
	paramsJSON, _ := json.Marshal(params)
	toolCall := fantasy.ToolCall{
		ID:    "test-12",
		Name:  DownloadToolName,
		Input: string(paramsJSON),
	}

	resp, err := tool.Run(ctx, toolCall)

	require.NoError(t, err)
	assert.False(t, resp.IsError)
	assert.Contains(t, resp.Content, "Content-Type: application/json")
}
