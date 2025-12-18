package native

import (
	"context"
	"encoding/json"
)

// AgentTool defines the interface for all agent tools
type AgentTool interface {
	Info() ToolInfo
	Call(ctx context.Context, params any) (ToolResponse, error)
}

// ToolInfo contains metadata about a tool
type ToolInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  any    `json:"parameters,omitempty"`
}

// ToolResponse represents the response from a tool call
type ToolResponse struct {
	Content  string         `json:"content"`
	IsError  bool           `json:"is_error,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// NewTextResponse creates a successful text response
func NewTextResponse(content string) ToolResponse {
	return ToolResponse{
		Content: content,
		IsError: false,
	}
}

// NewTextErrorResponse creates an error response
func NewTextErrorResponse(content string) ToolResponse {
	return ToolResponse{
		Content: content,
		IsError: true,
	}
}

// NewImageResponse creates an image response
func NewImageResponse(data []byte, mimeType string) ToolResponse {
	metadata := map[string]any{
		"mime_type": mimeType,
		"size":      len(data),
	}
	return ToolResponse{
		Content:  string(data), // Store base64 encoded image data as string
		Metadata: metadata,
		IsError:  false,
	}
}

// WithResponseMetadata adds metadata to a response
func WithResponseMetadata(resp ToolResponse, metadata map[string]any) ToolResponse {
	if resp.Metadata == nil {
		resp.Metadata = make(map[string]any)
	}
	for k, v := range metadata {
		resp.Metadata[k] = v
	}
	return resp
}

// ToolCall represents a tool call from the agent
type ToolCall struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
}

// ToolHandler is a function that handles tool execution
type ToolHandler func(ctx context.Context, params any, call ToolCall) (ToolResponse, error)

// BasicTool provides a simple implementation of AgentTool
type BasicTool struct {
	info    ToolInfo
	handler ToolHandler
}

// NewAgentTool creates a new basic tool
func NewAgentTool(info ToolInfo, handler ToolHandler) AgentTool {
	return &BasicTool{
		info:    info,
		handler: handler,
	}
}

// Info returns the tool's metadata
func (t *BasicTool) Info() ToolInfo {
	return t.info
}

// Call executes the tool with the given parameters
func (t *BasicTool) Call(ctx context.Context, params any) (ToolResponse, error) {
	call := ToolCall{
		Name: t.info.Name,
	}
	return t.handler(ctx, params, call)
}

// ParallelTool is a tool that can run operations in parallel
type ParallelTool struct {
	BasicTool
}

// NewParallelAgentTool creates a new parallel tool
func NewParallelAgentTool(info ToolInfo, handler ToolHandler) AgentTool {
	return &ParallelTool{
		BasicTool: BasicTool{
			info:    info,
			handler: handler,
		},
	}
}

// ProviderOptions contains options for LLM providers
type ProviderOptions struct {
	Temperature *float64 `json:"temperature,omitempty"`
	TopP        *float64 `json:"top_p,omitempty"`
	MaxTokens   *int64   `json:"max_tokens,omitempty"`
}

// Result represents the result from an agent
type Result struct {
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// Provider defines the interface for LLM providers
type Provider interface {
	Call(ctx context.Context, messages []any, options ProviderOptions) (*Result, error)
}

// NewAgentToolWithParams creates a tool with JSON parameter schema
func NewAgentToolWithParams(name, description string, params any, handler ToolHandler) AgentTool {
	info := ToolInfo{
		Name:        name,
		Description: description,
		Parameters:  params,
	}
	return &BasicTool{
		info:    info,
		handler: handler,
	}
}
