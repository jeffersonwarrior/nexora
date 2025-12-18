package native_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/nexora/nexora/internal/agent/native"
	"github.com/nexora/nexora/internal/permission"
	"github.com/nexora/nexora/internal/pubsub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAgentTool(t *testing.T) {
	info := native.ToolInfo{
		Name:        "test_tool",
		Description: "A test tool",
		Parameters:  nil,
	}

	handler := func(ctx context.Context, params any, call native.ToolCall) (native.ToolResponse, error) {
		return native.NewTextResponse("test response"), nil
	}

	tool := native.NewAgentTool(info, handler)

	require.NotNil(t, tool)
	assert.Equal(t, info.Name, tool.Info().Name)
	assert.Equal(t, info.Description, tool.Info().Description)
}

func TestBasicTool_Call(t *testing.T) {
	info := native.ToolInfo{
		Name:        "test_tool",
		Description: "A test tool",
	}

	handler := func(ctx context.Context, params any, call native.ToolCall) (native.ToolResponse, error) {
		return native.NewTextResponse("test response"), nil
	}

	tool := native.NewAgentTool(info, handler)

	ctx := context.Background()
	params := map[string]any{"param": "value"}

	response, err := tool.Call(ctx, params)
	require.NoError(t, err)
	assert.Equal(t, "test response", response.Content)
	assert.False(t, response.IsError)
}

func TestNewTextResponse(t *testing.T) {
	content := "Test content"
	response := native.NewTextResponse(content)

	assert.Equal(t, content, response.Content)
	assert.False(t, response.IsError)
	assert.Nil(t, response.Metadata)
}

func TestNewTextErrorResponse(t *testing.T) {
	content := "Error content"
	response := native.NewTextErrorResponse(content)

	assert.Equal(t, content, response.Content)
	assert.True(t, response.IsError)
	assert.Nil(t, response.Metadata)
}

func TestNewImageResponse(t *testing.T) {
	data := []byte("fake image data")
	mimeType := "image/png"

	response := native.NewImageResponse(data, mimeType)

	assert.Equal(t, string(data), response.Content)
	assert.False(t, response.IsError)
	assert.NotNil(t, response.Metadata)
	assert.Equal(t, mimeType, response.Metadata["mime_type"])
	assert.Equal(t, len(data), response.Metadata["size"])
}

func TestWithResponseMetadata(t *testing.T) {
	response := native.NewTextResponse("test content")
	metadata := map[string]any{
		"key1": "value1",
		"key2": 42,
	}

	enhanced := native.WithResponseMetadata(response, metadata)

	assert.Equal(t, response.Content, enhanced.Content)
	assert.Equal(t, false, enhanced.IsError)
	assert.Equal(t, "value1", enhanced.Metadata["key1"])
	assert.Equal(t, 42, enhanced.Metadata["key2"])
}

func TestWithResponseMetadata_Existing(t *testing.T) {
	response := native.NewTextResponse("test content")
	response = native.WithResponseMetadata(response, map[string]any{"existing": "value"})

	metadata := map[string]any{"new": "value"}
	enhanced := native.WithResponseMetadata(response, metadata)

	assert.Equal(t, "value", enhanced.Metadata["existing"])
	assert.Equal(t, "value", enhanced.Metadata["new"])
}

func TestNewParallelAgentTool(t *testing.T) {
	info := native.ToolInfo{
		Name:        "parallel_tool",
		Description: "A parallel test tool",
	}

	handler := func(ctx context.Context, params any, call native.ToolCall) (native.ToolResponse, error) {
		return native.NewTextResponse("parallel response"), nil
	}

	tool := native.NewParallelAgentTool(info, handler)

	require.NotNil(t, tool)
	assert.Equal(t, info.Name, tool.Info().Name)
	assert.Equal(t, info.Description, tool.Info().Description)

	// Test that it can be called like a regular tool
	ctx := context.Background()
	params := map[string]any{"param": "value"}

	response, err := tool.Call(ctx, params)
	require.NoError(t, err)
	assert.Equal(t, "parallel response", response.Content)
	assert.False(t, response.IsError)
}

func TestNewAgentToolWithParams(t *testing.T) {
	name := "param_tool"
	description := "A tool with parameters"
	params := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"input": map[string]any{
				"type": "string",
			},
		},
	}

	handler := func(ctx context.Context, params any, call native.ToolCall) (native.ToolResponse, error) {
		return native.NewTextResponse("parameterized response"), nil
	}

	tool := native.NewAgentToolWithParams(name, description, params, handler)

	require.NotNil(t, tool)
	info := tool.Info()
	assert.Equal(t, name, info.Name)
	assert.Equal(t, description, info.Description)
	assert.Equal(t, params, info.Parameters)

	// Test that it can be called
	ctx := context.Background()
	callParams := map[string]any{"input": "test"}

	response, err := tool.Call(ctx, callParams)
	require.NoError(t, err)
	assert.Equal(t, "parameterized response", response.Content)
	assert.False(t, response.IsError)
}

func TestProviderOptions(t *testing.T) {
	temp := 0.7
	topP := 0.9
	maxTokens := int64(1000)

	options := native.ProviderOptions{
		Temperature: &temp,
		TopP:        &topP,
		MaxTokens:   &maxTokens,
	}

	require.NotNil(t, options.Temperature)
	assert.Equal(t, temp, *options.Temperature)
	require.NotNil(t, options.TopP)
	assert.Equal(t, topP, *options.TopP)
	require.NotNil(t, options.MaxTokens)
	assert.Equal(t, maxTokens, *options.MaxTokens)
}

func TestProviderOptions_Empty(t *testing.T) {
	options := native.ProviderOptions{}

	assert.Nil(t, options.Temperature)
	assert.Nil(t, options.TopP)
	assert.Nil(t, options.MaxTokens)
}

func TestToolCall(t *testing.T) {
	call := native.ToolCall{
		ID:        "call_123",
		Name:      "test_tool",
		Arguments: json.RawMessage(`{"param": "value"}`),
	}

	assert.Equal(t, "call_123", call.ID)
	assert.Equal(t, "test_tool", call.Name)
	assert.Equal(t, json.RawMessage(`{"param": "value"}`), call.Arguments)
}

func TestResult(t *testing.T) {
	toolCall := native.ToolCall{
		ID:   "call_456",
		Name: "another_tool",
	}

	result := native.Result{
		Content:   "Test result content",
		ToolCalls: []native.ToolCall{toolCall},
	}

	assert.Equal(t, "Test result content", result.Content)
	require.Len(t, result.ToolCalls, 1)
	assert.Equal(t, toolCall.ID, result.ToolCalls[0].ID)
	assert.Equal(t, toolCall.Name, result.ToolCalls[0].Name)
}

func TestResult_Empty(t *testing.T) {
	result := native.Result{
		Content: "Simple content",
	}

	assert.Equal(t, "Simple content", result.Content)
	assert.Empty(t, result.ToolCalls)
}

// Mock provider for testing
type mockProvider struct {
	response *native.Result
	err      error
}

func (m *mockProvider) Call(ctx context.Context, messages []any, options native.ProviderOptions) (*native.Result, error) {
	return m.response, m.err
}

func TestProvider_Interface(t *testing.T) {
	mockResult := &native.Result{
		Content: "Mock response",
	}

	provider := &mockProvider{
		response: mockResult,
		err:      nil,
	}

	// Test that the mock implements the Provider interface
	var _ native.Provider = provider

	ctx := context.Background()
	messages := []any{"message1", "message2"}
	options := native.ProviderOptions{Temperature: ptr(0.5)}

	result, err := provider.Call(ctx, messages, options)

	require.NoError(t, err)
	assert.Equal(t, mockResult.Content, result.Content)
}

func TestProvider_Error(t *testing.T) {
	provider := &mockProvider{
		response: nil,
		err:      assert.AnError,
	}

	ctx := context.Background()
	messages := []any{"message1"}
	options := native.ProviderOptions{}

	result, err := provider.Call(ctx, messages, options)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestGetSessionFromContext(t *testing.T) {
	tests := []struct {
		name       string
		ctx        context.Context
		expectedID string
	}{
		{
			name:       "context with session ID",
			ctx:        context.WithValue(context.Background(), "session_id", "test-session-123"),
			expectedID: "test-session-123",
		},
		{
			name:       "context without session ID",
			ctx:        context.Background(),
			expectedID: "",
		},
		{
			name:       "context with wrong type",
			ctx:        context.WithValue(context.Background(), "session_id", 123),
			expectedID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := native.GetSessionFromContext(tt.ctx)
			assert.Equal(t, tt.expectedID, result)
		})
	}
}

func TestNewBashTool(t *testing.T) {
	// Create a mock permission service
	mockPermissions := &mockPermissionService{
		allowPermission: true,
	}

	// Test creation of bash tool
	workingDir := "/test/dir"
	description := "Test bash tool"

	tool := native.NewBashTool(mockPermissions, workingDir, description)

	require.NotNil(t, tool)

	// Check tool info
	info := tool.Info()
	assert.Equal(t, "bash", info.Name)
	assert.Equal(t, description, info.Description)
	assert.NotNil(t, info.Parameters)
}

func TestNewBashTool_CommandExecution(t *testing.T) {
	mockPermissions := &mockPermissionService{
		allowPermission: true,
	}

	workingDir := "/test/dir"
	description := "Test bash tool"

	tool := native.NewBashTool(mockPermissions, workingDir, description)
	require.NotNil(t, tool)

	// Test successful command execution with session context
	ctx := context.WithValue(context.Background(), "session_id", "test-session")

	params := native.BashParams{
		Description:     "Test command",
		Command:         "echo hello",
		WorkingDir:      "",
		RunInBackground: false,
	}

	response, err := tool.Call(ctx, params)
	require.NoError(t, err)
	assert.Contains(t, response.Content, "echo hello")
	assert.NotNil(t, response.Metadata)
}

func TestNewBashTool_UnsafeCommandDeniedPermission(t *testing.T) {
	mockPermissions := &mockPermissionService{
		allowPermission: false,
	}

	workingDir := "/test/dir"
	description := "Test bash tool"

	tool := native.NewBashTool(mockPermissions, workingDir, description)
	require.NotNil(t, tool)

	// Test unsafe command with denied permission
	ctx := context.WithValue(context.Background(), "session_id", "test-session")

	params := native.BashParams{
		Description: "Test unsafe command",
		Command:     "rm -rf /",
	}

	response, err := tool.Call(ctx, params)
	assert.Error(t, err)
	assert.Equal(t, native.ToolResponse{}, response)
}

func TestNewBashTool_MissingSessionID(t *testing.T) {
	mockPermissions := &mockPermissionService{
		allowPermission: true,
	}

	tool := native.NewBashTool(mockPermissions, "/test", "test")

	// Test command without session context
	ctx := context.Background()

	params := native.BashParams{
		Description: "Test command",
		Command:     "echo hello",
	}

	response, err := tool.Call(ctx, params)
	assert.Error(t, err)
	assert.Equal(t, native.ToolResponse{}, response)
}

func TestNewBashTool_InvalidParameters(t *testing.T) {
	mockPermissions := &mockPermissionService{
		allowPermission: true,
	}

	tool := native.NewBashTool(mockPermissions, "/test", "test")

	ctx := context.WithValue(context.Background(), "session_id", "test-session")

	// Test with invalid parameters
	response, err := tool.Call(ctx, "not a valid params struct")
	require.NoError(t, err)
	assert.Contains(t, response.Content, "invalid parameters")
}

func TestNewBashTool_EmptyCommand(t *testing.T) {
	mockPermissions := &mockPermissionService{
		allowPermission: true,
	}

	tool := native.NewBashTool(mockPermissions, "/test", "test")

	ctx := context.WithValue(context.Background(), "session_id", "test-session")

	// Test with empty command
	params := native.BashParams{
		Description: "Empty command test",
		Command:     "",
	}

	response, err := tool.Call(ctx, params)
	require.NoError(t, err)
	assert.Contains(t, response.Content, "missing command")
}

// Mock permission service for testing
type mockPermissionService struct {
	allowPermission bool
}

func (m *mockPermissionService) Request(opts permission.CreatePermissionRequest) bool {
	return m.allowPermission
}

func (m *mockPermissionService) Subscribe(ctx context.Context) <-chan pubsub.Event[permission.PermissionRequest] {
	ch := make(chan pubsub.Event[permission.PermissionRequest], 1)
	// Return closed channel to avoid blocking
	close(ch)
	return ch
}

func (m *mockPermissionService) SubscribeNotifications(ctx context.Context) <-chan pubsub.Event[permission.PermissionNotification] {
	ch := make(chan pubsub.Event[permission.PermissionNotification], 1)
	// Return closed channel to avoid blocking
	close(ch)
	return ch
}

func (m *mockPermissionService) AutoApproveSession(sessionID string) {}

func (m *mockPermissionService) SetSkipRequests(skip bool) {}

func (m *mockPermissionService) SkipRequests() bool {
	return false
}

func (m *mockPermissionService) GrantPersistent(permission.PermissionRequest) {}

func (m *mockPermissionService) Grant(permission.PermissionRequest) {}

func (m *mockPermissionService) Deny(permission.PermissionRequest) {}

// Helper function to create pointers for tests
func ptr[T any](v T) *T {
	return &v
}
