package mistral

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// CreateChatCompletion creates a chat completion response
func (c *Client) CreateChatCompletion(ctx context.Context, req *ChatCompletionRequest) (*ChatCompletionResponse, error) {
	if req.Stream != nil && *req.Stream {
		return nil, fmt.Errorf("use CreateChatCompletionStream for streaming requests")
	}

	stream := false
	req.Stream = &stream

	var resp ChatCompletionResponse
	if err := c.doRequest(ctx, "POST", "/chat/completions", req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// CreateChatCompletionStream creates a streaming chat completion
func (c *Client) CreateChatCompletionStream(ctx context.Context, req *ChatCompletionRequest) (<-chan *CompletionChunk, error) {
	stream := true
	req.Stream = &stream

	resp, err := c.doRawRequest(ctx, "POST", "/chat/completions", req)
	if err != nil {
		return nil, err
	}

	ch := make(chan *CompletionChunk)
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

			var chunk CompletionEvent
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				fmt.Printf("Error parsing chunk: %v\n", err)
				continue
			}

			ch <- &chunk.Data
		}
	}()

	return ch, nil
}

// ValidateChatCompletionRequest validates a chat completion request
func ValidateChatCompletionRequest(req *ChatCompletionRequest) error {
	if req.Model == "" {
		return fmt.Errorf("model is required")
	}

	if len(req.Messages) == 0 {
		return fmt.Errorf("at least one message is required")
	}

	if req.MaxTokens != nil && *req.MaxTokens < 0 {
		return fmt.Errorf("max_tokens must be non-negative")
	}

	if req.MinTokens != nil && *req.MinTokens < 0 {
		return fmt.Errorf("min_tokens must be non-negative")
	}

	if req.Temperature != nil {
		if *req.Temperature < 0 || *req.Temperature > 1.5 {
			return fmt.Errorf("temperature must be between 0 and 1.5")
		}
	}

	if req.TopP != nil {
		if *req.TopP < 0 || *req.TopP > 1 {
			return fmt.Errorf("top_p must be between 0 and 1")
		}
	}

	if req.RandomSeed != nil && *req.RandomSeed < 0 {
		return fmt.Errorf("random_seed must be non-negative")
	}

	// Validate messages
	for i, msg := range req.Messages {
		switch m := msg.(type) {
		case string:
			continue
		case map[string]interface{}:
			var role string
			if r, ok := m["role"].(string); !ok {
				return fmt.Errorf("message %d: role is required", i)
			} else {
				role = r
			}

			if role != "system" && role != "user" && role != "assistant" && role != "tool" {
				return fmt.Errorf("message %d: invalid role: %s", i, role)
			}

			if _, ok := m["content"]; !ok {
				return fmt.Errorf("message %d: content is required", i)
			}

			// Tool messages must have tool_call_id
			if role == "tool" {
				if _, ok := m["tool_call_id"]; !ok {
					return fmt.Errorf("message %d: tool messages must have tool_call_id", i)
				}
			}
		default:
			return fmt.Errorf("message %d: invalid message type", i)
		}
	}

	return nil
}

// CreateChatCompletionRequestWithDefaults creates a new chat completion request with sensible defaults
func CreateChatCompletionRequestWithDefaults(model string, messages []interface{}) *ChatCompletionRequest {
	temperature := 0.7
	topP := 1.0
	stream := false
	safePrompt := false

	return &ChatCompletionRequest{
		Model:       model,
		Messages:    messages,
		Temperature: &temperature,
		TopP:        &topP,
		Stream:      &stream,
		SafePrompt:  &safePrompt,
	}
}

// AddSystemMessage adds a system message to the request
func (req *ChatCompletionRequest) AddSystemMessage(content string) {
	msg := SystemMessage{
		Content: content,
		Role:    "system",
	}
	req.Messages = append(req.Messages, msg)
}

// AddUserMessage adds a user message to the request
func (req *ChatCompletionRequest) AddUserMessage(content string) {
	msg := UserMessage{
		Content: content,
		Role:    "user",
	}
	req.Messages = append(req.Messages, msg)
}

// AddAssistantMessage adds an assistant message to the request
func (req *ChatCompletionRequest) AddAssistantMessage(content string) {
	msg := AssistantMessage{
		Content: &content,
		Role:    "assistant",
		Prefix:  false,
	}
	req.Messages = append(req.Messages, msg)
}

// AddToolMessage adds a tool message to the request
func (req *ChatCompletionRequest) AddToolMessage(content, toolCallID string) {
	msg := ToolMessage{
		Content:    content,
		ToolCallID: &toolCallID,
		Role:       "tool",
	}
	req.Messages = append(req.Messages, msg)
}

// EnableJSONMode enables JSON output mode
func (req *ChatCompletionRequest) EnableJSONMode() {
	textFormat := "text"
	responseFormat := ResponseFormat{
		Type: textFormat,
	}
	req.ResponseFormat = &responseFormat
}

// EnableJSONSchemaMode enables JSON schema mode
func (req *ChatCompletionRequest) EnableJSONSchemaMode(name string, description string, schema map[string]interface{}, strict bool) {
	responseFormat := ResponseFormat{
		Type: "json_schema",
		JSONSchema: &JSONSchema{
			Name:        name,
			Description: description,
			Schema:      schema,
			Strict:      strict,
		},
	}
	req.ResponseFormat = &responseFormat
}

func (req *ChatCompletionRequest) setMaxTokens(max int) {
	req.MaxTokens = &max
}

func (req *ChatCompletionRequest) setMinTokens(min int) {
	req.MinTokens = &min
}

func (req *ChatCompletionRequest) setTemperature(temp float64) {
	req.Temperature = &temp
}

func (req *ChatCompletionRequest) setTopP(p float64) {
	req.TopP = &p
}

func (req *ChatCompletionRequest) enableStream() {
	stream := true
	req.Stream = &stream
}

func (req *ChatCompletionRequest) setSafePrompt(enable bool) {
	req.SafePrompt = &enable
}

func (req *ChatCompletionRequest) setRandomSeed(seed int) {
	req.RandomSeed = &seed
}

func (req *ChatCompletionRequest) setStop(stop []string) {
	req.Stop = stop
}

func (req *ChatCompletionRequest) addTool(tool Tool) {
	if req.Tools == nil {
		req.Tools = []Tool{}
	}
	req.Tools = append(req.Tools, tool)
}

func (req *ChatCompletionRequest) setToolChoice(choice interface{}) {
	req.ToolChoice = choice
}

func (req *ChatCompletionRequest) setPresencePenalty(penalty float64) {
	req.PresencePenalty = &penalty
}

func (req *ChatCompletionRequest) setFrequencyPenalty(penalty float64) {
	req.FrequencyPenalty = &penalty
}

func (req *ChatCompletionRequest) setN(n int) {
	req.N = &n
}
