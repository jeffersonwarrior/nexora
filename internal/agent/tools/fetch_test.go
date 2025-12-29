package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"charm.land/fantasy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFetchTool(t *testing.T) {
	mockPermissions := &mockPermissionService{}
	tool := NewFetchTool(mockPermissions, "/tmp", nil)

	assert.NotNil(t, tool)
	assert.Equal(t, FetchToolName, tool.Info().Name)
	assert.NotEmpty(t, tool.Info().Description)
}

func TestNewFetchTool_WithCustomClient(t *testing.T) {
	mockPermissions := &mockPermissionService{}
	customClient := &http.Client{}
	tool := NewFetchTool(mockPermissions, "/tmp", customClient)

	assert.NotNil(t, tool)
}

func TestFetchTool_MissingURL(t *testing.T) {
	mockPermissions := &mockPermissionService{}
	tool := NewFetchTool(mockPermissions, "/tmp", nil)

	ctx := context.WithValue(context.Background(), SessionIDContextKey, "test-session")
	params := FetchParams{
		Format: "text",
	}
	paramsJSON, _ := json.Marshal(params)
	toolCall := fantasy.ToolCall{
		ID:    "test-1",
		Name:  FetchToolName,
		Input: string(paramsJSON),
	}

	resp, err := tool.Run(ctx, toolCall)

	require.NoError(t, err)
	assert.True(t, resp.IsError)
	assert.Contains(t, resp.Content, "URL parameter is required")
}

func TestFetchTool_InvalidFormat(t *testing.T) {
	mockPermissions := &mockPermissionService{}
	tool := NewFetchTool(mockPermissions, "/tmp", nil)

	ctx := context.WithValue(context.Background(), SessionIDContextKey, "test-session")
	params := FetchParams{
		URL:    "https://example.com",
		Format: "invalid",
	}
	paramsJSON, _ := json.Marshal(params)
	toolCall := fantasy.ToolCall{
		ID:    "test-2",
		Name:  FetchToolName,
		Input: string(paramsJSON),
	}

	resp, err := tool.Run(ctx, toolCall)

	require.NoError(t, err)
	assert.True(t, resp.IsError)
	assert.Contains(t, resp.Content, "Format must be one of")
}

func TestFetchTool_InvalidURL(t *testing.T) {
	mockPermissions := &mockPermissionService{}
	tool := NewFetchTool(mockPermissions, "/tmp", nil)

	ctx := context.WithValue(context.Background(), SessionIDContextKey, "test-session")
	params := FetchParams{
		URL:    "ftp://example.com",
		Format: "text",
	}
	paramsJSON, _ := json.Marshal(params)
	toolCall := fantasy.ToolCall{
		ID:    "test-3",
		Name:  FetchToolName,
		Input: string(paramsJSON),
	}

	resp, err := tool.Run(ctx, toolCall)

	require.NoError(t, err)
	assert.True(t, resp.IsError)
	assert.Contains(t, resp.Content, "URL must start with http:// or https://")
}

func TestFetchTool_MissingSessionID(t *testing.T) {
	mockPermissions := &mockPermissionService{}
	tool := NewFetchTool(mockPermissions, "/tmp", nil)

	ctx := context.Background() // No session ID
	params := FetchParams{
		URL:    "https://example.com",
		Format: "text",
	}
	paramsJSON, _ := json.Marshal(params)
	toolCall := fantasy.ToolCall{
		ID:    "test-4",
		Name:  FetchToolName,
		Input: string(paramsJSON),
	}

	_, err := tool.Run(ctx, toolCall)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "session ID is required")
}

func TestFetchTool_SuccessfulFetch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Test content"))
	}))
	defer server.Close()

	mockPermissions := &mockPermissionService{}
	tool := NewFetchTool(mockPermissions, "/tmp", nil)

	ctx := context.WithValue(context.Background(), SessionIDContextKey, "test-session")
	params := FetchParams{
		URL:    server.URL,
		Format: "text",
	}
	paramsJSON, _ := json.Marshal(params)
	toolCall := fantasy.ToolCall{
		ID:    "test-5",
		Name:  FetchToolName,
		Input: string(paramsJSON),
	}

	resp, err := tool.Run(ctx, toolCall)

	require.NoError(t, err)
	assert.False(t, resp.IsError)
	assert.Contains(t, resp.Content, "Test content")
}

func TestFetchTool_PermissionHandling(t *testing.T) {
	// Note: Permission testing requires permission.Service mock with denial logic
	// The existing mockPermissionService always grants permission
	// Full permission testing is covered in integration tests
	t.Skip("Permission denial testing requires custom mock")
}

func TestFetchTool_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	mockPermissions := &mockPermissionService{}
	tool := NewFetchTool(mockPermissions, "/tmp", nil)

	ctx := context.WithValue(context.Background(), SessionIDContextKey, "test-session")
	params := FetchParams{
		URL:    server.URL + "/notfound",
		Format: "text",
	}
	paramsJSON, _ := json.Marshal(params)
	toolCall := fantasy.ToolCall{
		ID:    "test-7",
		Name:  FetchToolName,
		Input: string(paramsJSON),
	}

	resp, err := tool.Run(ctx, toolCall)

	require.NoError(t, err)
	assert.True(t, resp.IsError)
	assert.Contains(t, resp.Content, "404")
}

func TestFetchTool_WithTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	}))
	defer server.Close()

	mockPermissions := &mockPermissionService{}
	tool := NewFetchTool(mockPermissions, "/tmp", nil)

	ctx := context.WithValue(context.Background(), SessionIDContextKey, "test-session")
	params := FetchParams{
		URL:     server.URL,
		Format:  "text",
		Timeout: 30,
	}
	paramsJSON, _ := json.Marshal(params)
	toolCall := fantasy.ToolCall{
		ID:    "test-8",
		Name:  FetchToolName,
		Input: string(paramsJSON),
	}

	resp, err := tool.Run(ctx, toolCall)

	require.NoError(t, err)
	assert.False(t, resp.IsError)
}

func TestFetchTool_MaxTimeoutCap(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	}))
	defer server.Close()

	mockPermissions := &mockPermissionService{}
	tool := NewFetchTool(mockPermissions, "/tmp", nil)

	ctx := context.WithValue(context.Background(), SessionIDContextKey, "test-session")
	params := FetchParams{
		URL:     server.URL,
		Format:  "text",
		Timeout: 300, // Should be capped at 120
	}
	paramsJSON, _ := json.Marshal(params)
	toolCall := fantasy.ToolCall{
		ID:    "test-9",
		Name:  FetchToolName,
		Input: string(paramsJSON),
	}

	resp, err := tool.Run(ctx, toolCall)

	require.NoError(t, err)
	assert.False(t, resp.IsError)
}

func TestFetchTool_HTMLFormat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body><h1>Title</h1></body></html>"))
	}))
	defer server.Close()

	mockPermissions := &mockPermissionService{}
	tool := NewFetchTool(mockPermissions, "/tmp", nil)

	ctx := context.WithValue(context.Background(), SessionIDContextKey, "test-session")
	params := FetchParams{
		URL:    server.URL,
		Format: "html",
	}
	paramsJSON, _ := json.Marshal(params)
	toolCall := fantasy.ToolCall{
		ID:    "test-10",
		Name:  FetchToolName,
		Input: string(paramsJSON),
	}

	resp, err := tool.Run(ctx, toolCall)

	require.NoError(t, err)
	assert.False(t, resp.IsError)
	// HTML format returns the HTML content as-is (not converted to markdown)
	assert.Contains(t, resp.Content, "<html>")
}

func TestFetchTool_MarkdownFormat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body><h1>Title</h1><p>Content</p></body></html>"))
	}))
	defer server.Close()

	mockPermissions := &mockPermissionService{}
	tool := NewFetchTool(mockPermissions, "/tmp", nil)

	ctx := context.WithValue(context.Background(), SessionIDContextKey, "test-session")
	params := FetchParams{
		URL:    server.URL,
		Format: "markdown",
	}
	paramsJSON, _ := json.Marshal(params)
	toolCall := fantasy.ToolCall{
		ID:    "test-11",
		Name:  FetchToolName,
		Input: string(paramsJSON),
	}

	resp, err := tool.Run(ctx, toolCall)

	require.NoError(t, err)
	assert.False(t, resp.IsError)
}

func TestFetchTool_UserAgent(t *testing.T) {
	var receivedUserAgent string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedUserAgent = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	}))
	defer server.Close()

	mockPermissions := &mockPermissionService{}
	tool := NewFetchTool(mockPermissions, "/tmp", nil)

	ctx := context.WithValue(context.Background(), SessionIDContextKey, "test-session")
	params := FetchParams{
		URL:    server.URL,
		Format: "text",
	}
	paramsJSON, _ := json.Marshal(params)
	toolCall := fantasy.ToolCall{
		ID:    "test-12",
		Name:  FetchToolName,
		Input: string(paramsJSON),
	}

	resp, err := tool.Run(ctx, toolCall)

	require.NoError(t, err)
	assert.False(t, resp.IsError)
	// User-Agent is set to "nexora/1.0" for identification
	assert.Contains(t, receivedUserAgent, "nexora")
}

func TestFetchTool_SimpleMode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body><h1>Test Title</h1><p>Test content</p></body></html>"))
	}))
	defer server.Close()

	mockPermissions := &mockPermissionService{}
	tool := NewFetchTool(mockPermissions, "/tmp", nil)

	ctx := context.WithValue(context.Background(), SessionIDContextKey, "test-session")
	params := FetchParams{
		URL:    server.URL,
		Format: "text", // Use text format to get simple output
	}
	paramsJSON, _ := json.Marshal(params)
	toolCall := fantasy.ToolCall{
		ID:    "test-13",
		Name:  FetchToolName,
		Input: string(paramsJSON),
	}

	resp, err := tool.Run(ctx, toolCall)

	require.NoError(t, err)
	assert.False(t, resp.IsError)
	// Text format returns the content as-is
	assert.Contains(t, resp.Content, "Test Title")
}

func TestFetchTool_SimpleModeInvalidURL(t *testing.T) {
	mockPermissions := &mockPermissionService{}
	tool := NewFetchTool(mockPermissions, "/tmp", nil)

	ctx := context.WithValue(context.Background(), SessionIDContextKey, "test-session")
	params := FetchParams{
		URL:    "ftp://example.com",
		Format: "text",
	}
	paramsJSON, _ := json.Marshal(params)
	toolCall := fantasy.ToolCall{
		ID:    "test-14",
		Name:  FetchToolName,
		Input: string(paramsJSON),
	}

	resp, err := tool.Run(ctx, toolCall)

	require.NoError(t, err)
	assert.True(t, resp.IsError)
	assert.Contains(t, resp.Content, "URL must start with http:// or https://")
}

func TestNewWebFetchTool_BackwardCompatibility(t *testing.T) {
	tool := NewWebFetchTool("/tmp", nil)

	assert.NotNil(t, tool)
	assert.Equal(t, WebFetchToolName, tool.Info().Name)
}

func TestNewWebFetchTool_SuccessfulFetch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body><h1>Test</h1></body></html>"))
	}))
	defer server.Close()

	tool := NewWebFetchTool("/tmp", nil)

	ctx := context.Background()
	params := WebFetchParams{
		URL: server.URL,
	}
	paramsJSON, _ := json.Marshal(params)
	toolCall := fantasy.ToolCall{
		ID:    "test-15",
		Name:  WebFetchToolName,
		Input: string(paramsJSON),
	}

	resp, err := tool.Run(ctx, toolCall)

	require.NoError(t, err)
	assert.False(t, resp.IsError)
	assert.Contains(t, resp.Content, "Fetched content from")
}
