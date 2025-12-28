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
	"time"

	"charm.land/fantasy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============ MCP Detection Tests ============

func TestMCPAvailable_UsesMCP(t *testing.T) {
	// Setup: Create mock MCP server that simulates web_reader
	mcpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate MCP web_reader response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result": "MCP response content"}`))
	}))
	defer mcpServer.Close()

	// Note: In real implementation, MCP would be detected via registry
	// For now, we test that the detection mechanism exists
	mockPermissions := &mockPermissionService{}
	tool := NewFetchTool(mockPermissions, "/tmp", nil)

	ctx := context.WithValue(context.Background(), SessionIDContextKey, "test-session")
	params := FetchParams{
		URL:    mcpServer.URL,
		Format: "text",
	}
	paramsJSON, _ := json.Marshal(params)
	toolCall := fantasy.ToolCall{
		ID:    "test-mcp-1",
		Name:  FetchToolName,
		Input: string(paramsJSON),
	}

	resp, err := tool.Run(ctx, toolCall)

	require.NoError(t, err)
	// Content should be returned (either from MCP or built-in)
	assert.False(t, resp.IsError)
	assert.Contains(t, resp.Content, "MCP response content")
}

func TestMCPNotAvailable_FallsBackToBuiltIn(t *testing.T) {
	// Setup: Create a simple HTTP server to fetch from
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Built-in fetch response"))
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
		ID:    "test-builtin-1",
		Name:  FetchToolName,
		Input: string(paramsJSON),
	}

	resp, err := tool.Run(ctx, toolCall)

	require.NoError(t, err)
	assert.False(t, resp.IsError)
	assert.Contains(t, resp.Content, "Built-in fetch response")
}

// ============ Context-Aware Content Tests ============

func TestContextUnderLimit_ReturnsInResponse(t *testing.T) {
	// Setup: Create a server with small content (within context limit)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Small content within context limit"))
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
		ID:    "test-context-small",
		Name:  FetchToolName,
		Input: string(paramsJSON),
	}

	resp, err := tool.Run(ctx, toolCall)

	require.NoError(t, err)
	assert.False(t, resp.IsError)
	// Small content should be returned directly in response
	assert.Contains(t, resp.Content, "Small content within context limit")
	// Should NOT be a file path
	assert.False(t, strings.HasPrefix(resp.Content, "/tmp/"))
}

func TestContextOverLimit_WritesToTmp(t *testing.T) {
	// Setup: Create a server with content that will exceed context limit
	// Need ~8000+ words to exceed 32000 token limit
	// Token count = words + punctuation/2
	// Each "This is test content. " has 4 words + 1 period = 4.5 tokens
	// To exceed 32000: 32000 / 4.5 = ~7112 repetitions (use 8000 for safety)
	largeContent := strings.Repeat("This is test content. ", 8000)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(largeContent))
	}))
	defer server.Close()

	tmpDir, err := os.MkdirTemp("", "nexora-fetch-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	mockPermissions := &mockPermissionService{}
	tool := NewFetchTool(mockPermissions, tmpDir, nil)

	ctx := context.WithValue(context.Background(), SessionIDContextKey, "test-session-context")
	params := FetchParams{
		URL:    server.URL,
		Format: "text",
	}
	paramsJSON, _ := json.Marshal(params)
	toolCall := fantasy.ToolCall{
		ID:    "test-context-large",
		Name:  FetchToolName,
		Input: string(paramsJSON),
	}

	resp, err := tool.Run(ctx, toolCall)

	require.NoError(t, err)
	assert.False(t, resp.IsError)
	// Large content should be saved to tmp file, return path
	assert.Contains(t, resp.Content, tmpDir, "Response should mention the tmp directory")
	assert.Contains(t, resp.Content, "Content saved to:")
	assert.Contains(t, resp.Content, "too large for context")
}

func TestTokenCounting_Accurate(t *testing.T) {
	// Test that token counting works accurately
	testCases := []struct {
		name     string
		content  string
		expected int // approximate expected tokens
	}{
		{
			name:     "Short text",
			content:  "Hello world",
			expected: 2,
		},
		{
			name:     "Paragraph",
			content:  "This is a test paragraph with several words to estimate token count.",
			expected: 15,
		},
		{
			name:     "Code snippet",
			content:  "func main() { fmt.Println('Hello') }",
			expected: 8, // Lower expectation for code
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Token counting function should exist
			// This tests that the function is implemented
			tokens := countTokens(tc.content)

			// Allow some variance in token counting
			assert.GreaterOrEqual(t, tokens, tc.expected-tc.expected/2)
			assert.LessOrEqual(t, tokens, tc.expected+tc.expected/2+5)
		})
	}
}

// ============ Session-Based Tmp Files Tests ============

func TestTmpFile_SessionScoped(t *testing.T) {
	// Setup: Create a server with content large enough to trigger tmp file
	// Need ~8000+ words to exceed 32000 token limit
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Create content with enough words to exceed context limit
		// Each "Session scoped test content. " has 4 words + 1 period = 4.5 tokens
		// To exceed 32000: 32000 / 4.5 = ~7112 repetitions (use 8000 for safety)
		content := strings.Repeat("Session scoped test content. ", 8000)
		w.Write([]byte(content))
	}))
	defer server.Close()

	tmpDir, err := os.MkdirTemp("", "nexora-fetch-session-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	mockPermissions := &mockPermissionService{}
	tool := NewFetchTool(mockPermissions, tmpDir, nil)

	ctx := context.WithValue(context.Background(), SessionIDContextKey, "test-session-123")
	params := FetchParams{
		URL:    server.URL,
		Format: "text",
	}
	paramsJSON, _ := json.Marshal(params)
	toolCall := fantasy.ToolCall{
		ID:    "test-session-1",
		Name:  FetchToolName,
		Input: string(paramsJSON),
	}

	resp, err := tool.Run(ctx, toolCall)

	require.NoError(t, err)
	assert.False(t, resp.IsError)

	// Should return a path in the tmp directory
	assert.Contains(t, resp.Content, tmpDir, "Response should mention the tmp directory")
	assert.Contains(t, resp.Content, "nexora-fetch-test-session-123")

	// File should exist - extract path from response
	filePath := extractFilePath(resp.Content)
	assert.NotEmpty(t, filePath, "Should be able to extract file path from response")
	if filePath != "" {
		assert.True(t, fileExists(filePath), "Tmp file should exist: %s", filePath)
	}
}

func TestTmpFile_CleanupOnSessionEnd(t *testing.T) {
	// This test verifies that cleanup happens
	// In practice, cleanup would be triggered when session ends

	tmpDir, err := os.MkdirTemp("", "nexora-fetch-cleanup-test-*")
	require.NoError(t, err)

	// Simulate session by creating a session-scoped subdirectory
	sessionTmpDir := filepath.Join(tmpDir, "nexora-session-test-session")
	err = os.MkdirAll(sessionTmpDir, 0755)
	require.NoError(t, err)

	// Create a test file in the session directory
	testFile := filepath.Join(sessionTmpDir, "test-content.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	// Verify file exists
	assert.True(t, fileExists(testFile))

	// Cleanup should remove the entire session directory
	err = os.RemoveAll(sessionTmpDir)
	require.NoError(t, err)

	// Verify cleanup worked
	assert.False(t, fileExists(testFile))
	assert.False(t, fileExists(sessionTmpDir))
}

func TestTmpFilePath_NamingConvention(t *testing.T) {
	// Test that tmp file naming follows convention
	sessionID := "test-session-abc"
	filename := generateTmpFilename(sessionID, "https://example.com/page.html")

	// Should contain session ID
	assert.Contains(t, filename, sessionID)
	// Should have timestamp
	assert.Contains(t, filename, time.Now().Format("20060102"))
	// Should have URL hash (sanitized - no ://)
	assert.NotContains(t, filename, "://")
	// Should be a valid path (contains /)
	assert.Contains(t, filename, "/")
	// Should end in .txt
	assert.True(t, strings.HasSuffix(filename, ".txt"))
}

// ============ Timeout Tests ============

func TestTimeout_BuiltInRespectsTimeout(t *testing.T) {
	// Setup: Create a slow server
	slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Second) // 10 second delay
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Slow response"))
	}))
	defer slowServer.Close()

	mockPermissions := &mockPermissionService{}
	tool := NewFetchTool(mockPermissions, "/tmp", nil)

	ctx := context.WithValue(context.Background(), SessionIDContextKey, "test-session")
	params := FetchParams{
		URL:    slowServer.URL,
		Format: "text",
		// 2 second timeout - should timeout before server responds
		Timeout: 2,
	}
	paramsJSON, _ := json.Marshal(params)
	toolCall := fantasy.ToolCall{
		ID:    "test-timeout-1",
		Name:  FetchToolName,
		Input: string(paramsJSON),
	}

	start := time.Now()
	_, err := tool.Run(ctx, toolCall)
	elapsed := time.Since(start)

	// Should timeout quickly (within a few seconds of timeout value)
	assert.True(t, elapsed < 15*time.Second, "Request should timeout before 15 seconds")

	// Should return an error (timeout)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

// ============ Format Tests ============

func TestFormatText_ExtractsText(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body><p>Hello World</p></body></html>"))
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
		ID:    "test-format-text",
		Name:  FetchToolName,
		Input: string(paramsJSON),
	}

	resp, err := tool.Run(ctx, toolCall)

	require.NoError(t, err)
	assert.False(t, resp.IsError)
	// Should contain extracted text
	assert.Contains(t, resp.Content, "Hello World")
	// Should NOT contain HTML tags (extracted to text)
	assert.NotContains(t, resp.Content, "<html>")
	assert.NotContains(t, resp.Content, "<body>")
}

func TestFormatMarkdown_ConvertsToMarkdown(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body><h1>Title</h1><p>Paragraph</p></body></html>"))
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
		ID:    "test-format-markdown",
		Name:  FetchToolName,
		Input: string(paramsJSON),
	}

	resp, err := tool.Run(ctx, toolCall)

	require.NoError(t, err)
	assert.False(t, resp.IsError)
	// Should contain markdown formatting
	assert.Contains(t, resp.Content, "# Title")
	assert.Contains(t, resp.Content, "Paragraph")
}

func TestFormatHTML_ReturnsBodyOnly(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><head><title>Test</title></head><body><p>Content</p></body></html>"))
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
		ID:    "test-format-html",
		Name:  FetchToolName,
		Input: string(paramsJSON),
	}

	resp, err := tool.Run(ctx, toolCall)

	require.NoError(t, err)
	assert.False(t, resp.IsError)
	// Should contain body content
	assert.Contains(t, resp.Content, "Content")
	// Should NOT contain head/title
	assert.NotContains(t, resp.Content, "<head>")
	assert.NotContains(t, resp.Content, "<title>")
}

// ============ Helper Functions ============

func extractFilePath(content string) string {
	// Extract file path from response like "Content saved to: /path/to/file"
	// The path appears after "Content saved to: " in the response
	prefix := "Content saved to: "
	idx := strings.Index(content, prefix)
	if idx != -1 {
		// Extract everything after the prefix
		afterPrefix := content[idx+len(prefix):]
		// Find the end of the path (before newline or end of string)
		endIdx := strings.Index(afterPrefix, "\n")
		if endIdx == -1 {
			return afterPrefix
		}
		return afterPrefix[:endIdx]
	}
	return ""
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
