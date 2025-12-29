package base

import (
	"context"
	"fmt"
	"math"
	"math/rand/v2"
	"net/http"
	"time"
)

// Common types shared across all LLM providers

// BaseModel represents a language model available from a provider
type BaseModel interface {
	GetID() string
	GetName() string
	GetProvider() string
	GetCapabilities() ModelCapabilities
	GetPricing() ModelPricing
	GetContextWindow() int
}

// ModelCapabilities describes what a model supports
type ModelCapabilities struct {
	SupportsChat        bool     `json:"supports_chat"`
	SupportsFIM         bool     `json:"supports_fim"`
	SupportsEmbeddings  bool     `json:"supports_embeddings"`
	SupportsFineTuning  bool     `json:"supports_fine_tuning"`
	SupportsAgents      bool     `json:"supports_agents"`
	SupportsFileUpload  bool     `json:"supports_file_upload"`
	SupportsStreaming   bool     `json:"supports_streaming"`
	SupportsJSONMode    bool     `json:"supports_json_mode"`
	SupportsVision      bool     `json:"supports_vision"`
	SupportsAudio       bool     `json:"supports_audio"`
	SupportsTools       bool     `json:"supports_tools"`
	CanReason           bool     `json:"can_reason"`
	MaxTokens           int      `json:"max_tokens,omitempty"`
	SupportedParameters []string `json:"supported_parameters"`
	SecurityFeatures    []string `json:"security_features"`
}

// ModelPricing contains pricing information
type ModelPricing struct {
	CostPer1MIn  float64 `json:"cost_per_1m_in"`
	CostPer1MOut float64 `json:"cost_per_1m_out"`
	Currency     string  `json:"currency"`
	Unit         string  `json:"unit"`
}

// Message represents a chat message
type Message interface {
	GetRole() string
	GetContent() interface{}
}

// SystemMessage is a system message
type SystemMessage struct {
	Content interface{}
	Role    string
}

func (m SystemMessage) GetRole() string         { return m.Role }
func (m SystemMessage) GetContent() interface{} { return m.Content }

// UserMessage is a user message
type UserMessage struct {
	Content interface{}
	Role    string
}

func (m UserMessage) GetRole() string         { return m.Role }
func (m UserMessage) GetContent() interface{} { return m.Content }

// AssistantMessage is an assistant message
type AssistantMessage struct {
	Content   *string
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	Role      string
}

func (m AssistantMessage) GetRole() string { return m.Role }
func (m AssistantMessage) GetContent() interface{} {
	if m.Content != nil {
		return m.Content
	}
	return ""
}

// ToolMessage is a tool execution result message
type ToolMessage struct {
	Content    string
	ToolCallID string
	Name       *string
	Role       string
}

func (m ToolMessage) GetRole() string         { return m.Role }
func (m ToolMessage) GetContent() interface{} { return m.Content }

// Tool defines a tool/message function that the model can call
type Tool struct {
	Type     string   `json:"type"` // always "function" for OpenAI compatibility
	Function Function `json:"function"`
}

// Function defines a specific function/tool
type Function struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	Required    []string               `json:"required,omitempty"`
}

// ToolCall is a call to a tool/function from the model
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

// FunctionCall represents the actual function call with arguments
type FunctionCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ChatCompletionRequest is a request to create a chat completion
type ChatCompletionRequest struct {
	Model            string          `json:"model"`
	Messages         []Message       `json:"messages"`
	MaxTokens        *int            `json:"max_tokens,omitempty"`
	Temperature      *float64        `json:"temperature,omitempty"`
	TopP             *float64        `json:"top_p,omitempty"`
	Stream           *bool           `json:"stream,omitempty"`
	Stop             interface{}     `json:"stop,omitempty"` // string or []string
	FrequencyPenalty *float64        `json:"frequency_penalty,omitempty"`
	PresencePenalty  *float64        `json:"presence_penalty,omitempty"`
	Tools            []Tool          `json:"tools,omitempty"`
	ToolChoice       interface{}     `json:"tool_choice,omitempty"` // string or object
	ResponseFormat   *ResponseFormat `json:"response_format,omitempty"`
	N                *int            `json:"n,omitempty"`
	Seed             *int64          `json:"seed,omitempty"`
	User             string          `json:"user,omitempty"`
}

// ResponseFormat for structured output
type ResponseFormat struct {
	Type       string      `json:"type"` // "text", "json_object", or "json_schema"
	JSONSchema *JSONSchema `json:"json_schema,omitempty"`
}

// JSONSchema for JSON schema mode
type JSONSchema struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Schema      map[string]interface{} `json:"schema"`
	Strict      bool                   `json:"strict"`
}

// EmbeddingRequest creates embeddings for the provided input texts
type EmbeddingRequest struct {
	Model          string      `json:"model"`
	Input          interface{} `json:"input"` // string or []string
	EncodingFormat string      `json:"encoding_format,omitempty"`
	Dimensions     *int        `json:"dimensions,omitempty"`
	User           string      `json:"user,omitempty"`
}

// EmbeddingResponse contains the embeddings
type EmbeddingResponse struct {
	Object string          `json:"object"`
	Data   []EmbeddingData `json:"data"`
	Model  string          `json:"model"`
	Usage  EmbeddingUsage  `json:"usage"`
}

// EmbeddingData represents a single embedding
type EmbeddingData struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

// EmbeddingUsage token usage information
type EmbeddingUsage struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// Usage information for API responses
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatCompletionResponse is a response to a chat completion request
type ChatCompletionResponse interface {
	GetID() string
	GetObject() string
	GetCreated() int64
	GetModel() string
	GetChoices() []ChatCompletionChoice
	GetUsage() Usage
}

// ChatCompletionChoice is a single completion choice
type ChatCompletionChoice struct {
	Index        int              `json:"index"`
	Message      AssistantMessage `json:"message"`
	FinishReason string           `json:"finish_reason"`
}

// CompletionChunk represents a chunk of a streaming response
type CompletionChunk struct {
	ID      string                   `json:"id"`
	Object  string                   `json:"object"`
	Created int64                    `json:"created"`
	Model   string                   `json:"model"`
	Choices []CompletionStreamChoice `json:"choices"`
	Usage   *Usage                   `json:"usage,omitempty"`
}

// CompletionStreamChoice is a choice in a streaming response
type CompletionStreamChoice struct {
	Index        int          `json:"index"`
	Delta        DeltaMessage `json:"delta"`
	FinishReason *string      `json:"finish_reason"`
}

// DeltaMessage represents a delta in a streaming response
type DeltaMessage struct {
	Role      *string    `json:"role,omitempty"`
	Content   *string    `json:"content,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// ErrorCode represents different types of errors
type ErrorCode string

const (
	ErrInvalidRequest    ErrorCode = "invalid_request"
	ErrInvalidAPIKey     ErrorCode = "invalid_api_key"
	ErrInsufficientQuota ErrorCode = "insufficient_quota"
	ErrModelNotFound     ErrorCode = "model_not_found"
	ErrRateLimited       ErrorCode = "rate_limited"
	ErrServerError       ErrorCode = "server_error"
	ErrProviderError     ErrorCode = "provider_error"
)

// APIError represents an error returned by the API
type APIError struct {
	Code    ErrorCode `json:"code,omitempty"`
	Message string    `json:"message"`
	Type    string    `json:"type"`
	Param   string    `json:"param,omitempty"`
}

func (e *APIError) Error() string {
	if e.Type != "" {
		return fmt.Sprintf("%s: %s (type: %s, param: %s)", e.Code, e.Message, e.Type, e.Param)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// ClientConfig holds configuration for API clients
type ClientConfig struct {
	APIKey      string
	BaseURL     string
	HTTPClient  *http.Client
	UserAgent   string
	Timeout     time.Duration
	RateLimiter RateLimiter
	RetryConfig RetryConfig
}

// RateLimiter interface for rate limiting
type RateLimiter interface {
	Wait(ctx context.Context) error
	Reset()
}

// RetryConfig for API call retries
type RetryConfig struct {
	MaxRetries    int
	BaseDelay     time.Duration
	BackoffFactor float64
	MaxDelay      time.Duration
}

// DefaultRetryConfig provides sensible defaults
var DefaultRetryConfig = RetryConfig{
	MaxRetries:    3,
	BaseDelay:     1 * time.Second,
	BackoffFactor: 2,
	MaxDelay:      60 * time.Second,
}

// DefaultClientConfig provides sensible defaults
var DefaultClientConfig = ClientConfig{
	UserAgent:   "nexora-sdk/1.0",
	Timeout:     300 * time.Second, // 5 minutes for large generations
	RetryConfig: DefaultRetryConfig,
}

// Helper functions

// Contains checks if a string contains any of the substrings
func contains(s string, substrings []string) bool {
	for _, substr := range substrings {
		if len(s) >= len(substr) && s[:len(substr)] == substr {
			return true
		}
	}
	return false
}

// ValidateChatRequest validates a chat completion request
func ValidateChatRequest(req ChatCompletionRequest) error {
	if req.Model == "" {
		return &APIError{
			Code:    ErrInvalidRequest,
			Message: "model is required",
		}
	}

	if len(req.Messages) == 0 {
		return &APIError{
			Code:    ErrInvalidRequest,
			Message: "at least one message is required",
		}
	}

	if req.MaxTokens != nil && *req.MaxTokens < 0 {
		return &APIError{
			Code:    ErrInvalidRequest,
			Message: "max_tokens must be non-negative",
		}
	}

	if req.Temperature != nil {
		if *req.Temperature < 0 || *req.Temperature > 2 {
			return &APIError{
				Code:    ErrInvalidRequest,
				Message: "temperature must be between 0 and 2",
			}
		}
	}

	if req.TopP != nil {
		if *req.TopP < 0 || *req.TopP > 1 {
			return &APIError{
				Code:    ErrInvalidRequest,
				Message: "top_p must be between 0 and 1",
			}
		}
	}

	return nil
}

// IsRetryableError checks if an error should be retried
func IsRetryableError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		switch apiErr.Code {
		case ErrRateLimited, ErrServerError:
			return true
		default:
			return false
		}
	}
	return false
}

// SleepWithJitter sleeps for the given duration with jitter
func SleepWithJitter(duration time.Duration) {
	jitter := time.Duration(float64(duration) * (0.8 + 0.4*rand.Float64()))
	time.Sleep(jitter)
}

// BackoffDelay calculates exponential backoff delay
func BackoffDelay(attempt int, base time.Duration, factor float64, max time.Duration) time.Duration {
	delay := time.Duration(float64(base) * math.Pow(factor, float64(attempt)))
	if delay > max {
		return max
	}
	return delay
}
