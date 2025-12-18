package openai

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nexora/sdk/base"
)

// ChatCompletionRequest wraps base.ChatCompletionRequest with OpenAI-specific fields
type ChatCompletionRequest struct {
	base.ChatCompletionRequest
	// OpenAI-specific fields can be added here
}

// CreateChatCompletionRequest creates a new chat completion request
func CreateChatCompletionRequest(model string, messages []base.Message) *ChatCompletionRequest {
	return &ChatCompletionRequest{
		base.ChatCompletionRequest{
			Model:    model,
			Messages: messages,
		},
	}
}

// CreateChatCompletionResponse creates a response object
func CreateChatCompletionResponse() *ChatCompletionResponse {
	return &ChatCompletionResponse{}
}

// ChatCompletionResponse represents OpenAI's chat completion response
type ChatCompletionResponse struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int64                  `json:"created"`
	Model   string                 `json:"model"`
	Choices []ChatCompletionChoice `json:"choices"`
	Usage   base.Usage             `json:"usage"`
}

// ChatCompletionChoice represents a single completion choice
type ChatCompletionChoice struct {
	Index        int              `json:"index"`
	Message      AssistantMessage `json:"message"`
	FinishReason string           `json:"finish_reason"`
}

// AssistantMessage is a message from the assistant
type AssistantMessage struct {
	Content *string `json:"content"`
	Refusal *string `json:"refusal"`
	Role    string  `json:"role"`
}

// DeltaMessage represents a delta message in streaming
type DeltaMessage struct {
	Role      *string    `json:"role,omitempty"`
	Content   *string    `json:"content,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// ToolCall represents a tool call by the model
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

// FunctionCall is the actual function call
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// CompletionChunk represents a chunk of a streaming response
type CompletionChunk struct {
	ID      string                   `json:"id"`
	Object  string                   `json:"object"`
	Created int64                    `json:"created"`
	Model   string                   `json:"model"`
	Choices []CompletionStreamChoice `json:"choices"`
	Usage   *base.Usage              `json:"usage,omitempty"`
}

// CompletionStreamChoice is a choice in a streaming response
type CompletionStreamChoice struct {
	Index        int          `json:"index"`
	Delta        DeltaMessage `json:"delta"`
	FinishReason *string      `json:"finish_reason"`
}

// CreateChatCompletionStream creates a streaming chat completion
func (c *Client) CreateChatCompletionStream(
	ctx context.Context,
	req *ChatCompletionRequest,
) (<-chan *base.CompletionChunk, error) {
	stream := true
	req.Stream = &stream

	resp, err := c.DoRawRequest(ctx, "POST", "/chat/completions", req)
	if err != nil {
		return nil, err
	}

	ch := make(chan *base.CompletionChunk)
	go func() {
		defer resp.Body.Close()
		defer close(ch)

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()

			// Skip empty lines and SSE format prefixes
			if line == "" || !strings.HasPrefix(line, "data: ") {
				continue
			}

			// Remove "data: " prefix
			data := strings.TrimPrefix(line, "data: ")

			// Check for [DONE] marker
			if data == "[DONE]" {
				break
			}

			var chunk base.CompletionEvent
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				fmt.Printf("Error parsing chunk: %v\n", err)
				continue
			}

			ch <- &chunk.Data
		}
	}()

	return ch, nil
}

// AddSystemMessage adds a system message to the request
func (req *ChatCompletionRequest) AddSystemMessage(content string) {
	msg := base.SystemMessage{
		Content: content,
		Role:    "system",
	}
	req.Messages = append(req.Messages, msg)
}

// AddUserMessage adds a user message to the request
func (req *ChatCompletionRequest) AddUserMessage(content string) {
	msg := base.UserMessage{
		Content: content,
		Role:    "user",
	}
	req.Messages = append(req.Messages, msg)
}

// AddAssistantMessage adds an assistant message to the request
func (req *ChatCompletionRequest) AddAssistantMessage(content string) {
	msg := base.AssistantMessage{
		Content: &content,
		Role:    "assistant",
	}
	req.Messages = append(req.Messages, msg)
}

// AddToolMessage adds a tool message to the request
func (req *ChatCompletionRequest) AddToolMessage(content, toolCallID, toolName string) {
	msg := base.ToolMessage{
		Content:    content,
		ToolCallID: toolCallID,
		Name:       &toolName,
		Role:       "tool",
	}
	req.Messages = append(req.Messages, msg)
}

// EnableJSONMode enables JSON output mode
func (req *ChatCompletionRequest) EnableJSONMode() {
	responseFormat := base.ResponseFormat{
		Type: "json_object",
	}
	req.ResponseFormat = &responseFormat
}

// EnableJSONSchemaMode enables JSON schema mode
func (req *ChatCompletionRequest) EnableJSONSchemaMode(name string, description string, schema map[string]interface{}, strict bool) {
	responseFormat := base.ResponseFormat{
		Type: "json_schema",
		JSONSchema: &base.JSONSchema{
			Name:        name,
			Description: description,
			Schema:      schema,
			Strict:      strict,
		},
	}
	req.ResponseFormat = &responseFormat
}

// SetTemperature sets the sampling temperature
func (req *ChatCompletionRequest) SetTemperature(temp float64) {
	req.Temperature = &temp
}

// SetMaxTokens sets the maximum number of tokens
func (req *ChatCompletionRequest) SetMaxTokens(maxTokens int) {
	req.MaxTokens = &maxTokens
}

// SetStop sets stop sequences
func (req *ChatCompletionRequest) SetStop(stop []string) {
	req.Stop = stop
}

// EnableStreaming enables streaming responses
func (req *ChatCompletionRequest) EnableStreaming() {
	stream := true
	req.Stream = &stream
}

// AddTool adds a tool to the request
func (req *ChatCompletionRequest) AddTool(tool base.Tool) {
	if req.Tools == nil {
		req.Tools = []base.Tool{}
	}
	req.Tools = append(req.Tools, tool)
}

// SetToolChoice sets the tool choice
func (req *ChatCompletionRequest) SetToolChoice(choice interface{}) {
	req.ToolChoice = choice
}
