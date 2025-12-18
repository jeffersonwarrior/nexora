package mistral

import (
	"time"
)

// Request/Response Types for Mistral API

// Base Response
type BaseResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
}

// Usage information for API responses
type Usage struct {
	PromptTokens       int  `json:"prompt_tokens"`
	CompletionTokens   int  `json:"completion_tokens"`
	TotalTokens        int  `json:"total_tokens"`
	PromptAudioSeconds *int `json:"prompt_audio_seconds,omitempty"`
}

// System Message
type SystemMessage struct {
	Content interface{} `json:"content"` // string or []ContentChunk
	Role    string      `json:"role"`
}

// User Message
type UserMessage struct {
	Content interface{} `json:"content"` // string or []TextChunk
	Role    string      `json:"role"`
}

// Assistant Message
type AssistantMessage struct {
	Content   *string    `json:"content,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	Prefix    bool       `json:"prefix"`
	Role      string     `json:"role"`
}

// Tool Message
type ToolMessage struct {
	Content    string  `json:"content"`
	ToolCallID *string `json:"tool_call_id,omitempty"`
	Name       *string `json:"name,omitempty"`
	Role       string  `json:"role"`
}

// Text Chunk for content
type TextChunk struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Content Chunk for richer content
type ContentChunk struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Tool definition
type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

// ToolCall from assistant
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

// Function definition for tool
type Function struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// Function call in tool use
type FunctionCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolChoice options
type ToolChoice struct {
	Type     string   `json:"type"`
	Function ToolCall `json:"function,omitempty"`
}

// ResponseFormat for structured output
type ResponseFormat struct {
	Type       string      `json:"type"`
	JSONSchema *JSONSchema `json:"json_schema,omitempty"`
}

// JSONSchema for JSON schema mode
type JSONSchema struct {
	Name        string                 `json:"name"`
	Schema      map[string]interface{} `json:"schema"`
	Strict      bool                   `json:"strict"`
	Description string                 `json:"description"`
}

// CompletionEvent for streaming
type CompletionEvent struct {
	Data CompletionChunk `json:"data"`
}

// CompletionChunk for streaming response
type CompletionChunk struct {
	ID      string                           `json:"id"`
	Object  string                           `json:"object"`
	Created int64                            `json:"created"`
	Model   string                           `json:"model"`
	Usage   *Usage                           `json:"usage,omitempty"`
	Choices []CompletionResponseStreamChoice `json:"choices"`
}

// CompletionResponseStreamChoice for streaming
type CompletionResponseStreamChoice struct {
	Index        int          `json:"index"`
	Delta        DeltaMessage `json:"delta"`
	FinishReason *string      `json:"finish_reason"`
}

// DeltaMessage for streaming updates
type DeltaMessage struct {
	Role      *string    `json:"role,omitempty"`
	Content   *string    `json:"content,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// Chat completion request
type ChatCompletionRequest struct {
	Model            string          `json:"model"`
	Messages         []interface{}   `json:"messages"`
	Temperature      *float64        `json:"temperature,omitempty"`
	TopP             *float64        `json:"top_p,omitempty"`
	MaxTokens        *int            `json:"max_tokens,omitempty"`
	MinTokens        *int            `json:"min_tokens,omitempty"`
	Stream           *bool           `json:"stream,omitempty"`
	Stop             interface{}     `json:"stop,omitempty"` // string or []string
	RandomSeed       *int            `json:"random_seed,omitempty"`
	ResponseFormat   *ResponseFormat `json:"response_format,omitempty"`
	Tools            []Tool          `json:"tools,omitempty"`
	ToolChoice       interface{}     `json:"tool_choice,omitempty"` // string or ToolChoice
	SafePrompt       *bool           `json:"safe_prompt,omitempty"`
	PresencePenalty  *float64        `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float64        `json:"frequency_penalty,omitempty"`
	N                *int            `json:"n,omitempty"`
}

// Chat completion response
type ChatCompletionResponse struct {
	BaseResponse
	Choices []ChatCompletionChoice `json:"choices"`
	Usage   Usage                  `json:"usage"`
}

// ChatCompletionChoice
type ChatCompletionChoice struct {
	Index        int              `json:"index"`
	Message      AssistantMessage `json:"message"`
	FinishReason string           `json:"finish_reason"`
}

// FIM completion request
type FIMCompletionRequest struct {
	Model       string      `json:"model"`
	Prompt      string      `json:"prompt"`
	Suffix      *string     `json:"suffix,omitempty"`
	Temperature *float64    `json:"temperature,omitempty"`
	TopP        *float64    `json:"top_p,omitempty"`
	MaxTokens   *int        `json:"max_tokens,omitempty"`
	MinTokens   *int        `json:"min_tokens,omitempty"`
	Stream      *bool       `json:"stream,omitempty"`
	Stop        interface{} `json:"stop,omitempty"` // string or []string
	RandomSeed  *int        `json:"random_seed,omitempty"`
}

// FIM completion response
type FIMCompletionResponse struct {
	BaseResponse
	Choices []FIMCompletionChoice `json:"choices"`
	Usage   Usage                 `json:"usage"`
}

// FIMCompletionChoice
type FIMCompletionChoice struct {
	Index        int    `json:"index"`
	Message      string `json:"text"`
	FinishReason string `json:"finish_reason"`
}

// Embedding request
type EmbeddingRequest struct {
	Model           string      `json:"model"`
	Input           interface{} `json:"input"` // string or []string
	EncodingFormat  *string     `json:"encoding_format,omitempty"`
	OutputDimension *int        `json:"output_dimension,omitempty"`
	OutputDtype     *string     `json:"output_dtype,omitempty"`
}

// Embedding response
type EmbeddingResponse struct {
	BaseResponse
	Data  []EmbeddingResponseData `json:"data"`
	Usage Usage                   `json:"usage"`
}

// EmbeddingResponseData
type EmbeddingResponseData struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

// Agent types
type Agent struct {
	ID             string          `json:"id"`
	Object         string          `json:"object"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	Name           string          `json:"name"`
	Model          string          `json:"model"`
	Description    *string         `json:"description,omitempty"`
	Instructions   *string         `json:"instructions,omitempty"`
	Tools          []interface{}   `json:"tools,omitempty"`
	CompletionArgs *CompletionArgs `json:"completion_args,omitempty"`
	Handoffs       []string        `json:"handoffs,omitempty"`
	Version        int             `json:"version"`
}

type CompletionArgs struct {
	Temperature *float64    `json:"temperature,omitempty"`
	TopP        *float64    `json:"top_p,omitempty"`
	MaxTokens   *int        `json:"max_tokens,omitempty"`
	Stop        interface{} `json:"stop,omitempty"`
	RandomSeed  *int        `json:"random_seed,omitempty"`
	SafePrompt  *bool       `json:"safe_prompt,omitempty"`
}

type AgentCompletionRequest struct {
	AgentID        string          `json:"agent_id"`
	Messages       []interface{}   `json:"messages"`
	MaxTokens      *int            `json:"max_tokens,omitempty"`
	MinTokens      *int            `json:"min_tokens,omitempty"`
	Stream         *bool           `json:"stream,omitempty"`
	Stop           interface{}     `json:"stop,omitempty"`
	RandomSeed     *int            `json:"random_seed,omitempty"`
	ResponseFormat *ResponseFormat `json:"response_format,omitempty"`
	Tools          []Tool          `json:"tools,omitempty"`
	ToolChoice     interface{}     `json:"tool_choice,omitempty"`
}

// Model types
type Model struct {
	ID               string            `json:"id"`
	Object           string            `json:"object"`
	Created          int64             `json:"created"`
	OwnedBy          string            `json:"owned_by"`
	Root             *string           `json:"root,omitempty"`
	Archived         bool              `json:"archived"`
	Name             *string           `json:"name,omitempty"`
	Description      *string           `json:"description,omitempty"`
	MaxContextLength int               `json:"max_context_length"`
	Capabilities     ModelCapabilities `json:"capabilities"`
	Aliases          []string          `json:"aliases"`
	Deprecation      *time.Time        `json:"deprecation,omitempty"`
}

type ModelCapabilities struct {
	CompletionChat  bool `json:"completion_chat"`
	CompletionFIM   bool `json:"completion_fim"`
	FunctionCalling bool `json:"function_calling"`
	FineTuning      bool `json:"fine_tuning"`
	Vision          bool `json:"vision"`
	Classification  bool `json:"classification"`
	Audio           bool `json:"audio"`
}

// File types
type File struct {
	ID         string     `json:"id"`
	Object     string     `json:"object"`
	Bytes      int        `json:"bytes"`
	CreatedAt  int64      `json:"created_at"`
	Filename   string     `json:"filename"`
	Purpose    string     `json:"purpose"`
	SampleType SampleType `json:"sample_type"`
	NumLines   *int       `json:"num_lines,omitempty"`
	Source     Source     `json:"source"`
	Status     string     `json:"status,omitempty"`
}

type SampleType string

const (
	SampleTypePretrain SampleType = "pretrain"
	SampleTypeInstruct SampleType = "instruct"
)

type Source string

const (
	SourceUpload     Source = "upload"
	SourceRepository Source = "repository"
)

// Fine-tuning types
type Job struct {
	ID              string             `json:"id"`
	Object          string             `json:"object"`
	JobType         string             `json:"job_type"`
	Status          JobStatus          `json:"status"`
	CreatedAt       int64              `json:"created_at"`
	ModifiedAt      int64              `json:"modified_at"`
	TrainingFiles   []string           `json:"training_files"`
	ValidationFiles []string           `json:"validation_files"`
	Hyperparameters TrainingParameters `json:"hyperparameters"`
	Model           FineTuneableModel  `json:"model"`
	FineTunedModel  *string            `json:"fine_tuned_model,omitempty"`
	Suffix          *string            `json:"suffix,omitempty"`
	AutoStart       bool               `json:"auto_start"`
	TrainedTokens   *int               `json:"trained_tokens,omitempty"`
	Metadata        *JobMetadata       `json:"metadata,omitempty"`
	Repositories    []GithubRepository `json:"repositories,omitempty"`
	Integrations    []interface{}      `json:"integrations,omitempty"`
	Events          []Event            `json:"events,omitempty"`
	Checkpoints     []Checkpoint       `json:"checkpoints,omitempty"`
}

type JobStatus string

const (
	JobStatusQueued                JobStatus = "QUEUED"
	JobStatusStarted               JobStatus = "STARTED"
	JobStatusValidating            JobStatus = "VALIDATING"
	JobStatusValidated             JobStatus = "VALIDATED"
	JobStatusRunning               JobStatus = "RUNNING"
	JobStatusFailedValidation      JobStatus = "FAILED_VALIDATION"
	JobStatusFailed                JobStatus = "FAILED"
	JobStatusSuccess               JobStatus = "SUCCESS"
	JobStatusCancelled             JobStatus = "CANCELLED"
	JobStatusCancellationRequested JobStatus = "CANCELLATION_REQUESTED"
)

type TrainingParameters struct {
	TrainingSteps  *int     `json:"training_steps,omitempty"`
	LearningRate   *float64 `json:"learning_rate,omitempty"`
	WeightDecay    *float64 `json:"weight_decay,omitempty"`
	WarmupFraction *float64 `json:"warmup_fraction,omitempty"`
	Epochs         *float64 `json:"epochs,omitempty"`
	FIMRatio       *float64 `json:"fim_ratio,omitempty"`
}

type FineTuneableModel string

const (
	OpenMistral7B      FineTuneableModel = "open-mistral-7b"
	MistralSmallLatest FineTuneableModel = "mistral-small-latest"
	CodestralLatest    FineTuneableModel = "codestral-latest"
	OpenMistralNemo    FineTuneableModel = "open-mistral-nemo"
	MistralLargeLatest FineTuneableModel = "mistral-large-latest"
)

type GithubRepository struct {
	Type     string  `json:"type"`
	Name     string  `json:"name"`
	Owner    string  `json:"owner"`
	Ref      *string `json:"ref,omitempty"`
	Weight   float64 `json:"weight"`
	Token    string  `json:"token,omitempty"`
	CommitID string  `json:"commit_id"`
}

type JobMetadata struct {
	ExpectedDurationSeconds *int     `json:"expected_duration_seconds,omitempty"`
	Cost                    *float64 `json:"cost,omitempty"`
	CostCurrency            *string  `json:"cost_currency,omitempty"`
	TrainTokensPerStep      *int     `json:"train_tokens_per_step,omitempty"`
	TrainTokens             *int     `json:"train_tokens,omitempty"`
	DataTokens              *int     `json:"data_tokens,omitempty"`
	EstimatedStartTime      *int64   `json:"estimated_start_time,omitempty"`
}

type Event struct {
	Name      string                 `json:"name"`
	Data      map[string]interface{} `json:"data,omitempty"`
	CreatedAt int64                  `json:"created_at"`
}

type Checkpoint struct {
	Metrics    Metric `json:"metrics"`
	StepNumber int    `json:"step_number"`
	CreatedAt  int64  `json:"created_at"`
}

type Metric struct {
	TrainLoss              *float64 `json:"train_loss,omitempty"`
	ValidLoss              *float64 `json:"valid_loss,omitempty"`
	ValidMeanTokenAccuracy *float64 `json:"valid_mean_token_accuracy,omitempty"`
}
