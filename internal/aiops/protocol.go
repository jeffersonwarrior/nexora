package aiops

import "errors"

// ErrNotEnabled is returned when AIOPS is not enabled.
var ErrNotEnabled = errors.New("aiops not enabled")

// ErrorResponse is the error response from the AIOPS service.
type ErrorResponse struct {
	Error string `json:"error"`
}

// ResolveEditRequest is the request body for /resolve-edit.
type ResolveEditRequest struct {
	FileContent string `json:"file_content"`
	OldString   string `json:"old_string"`
	NewString   string `json:"new_string"`
}

// DetectLoopRequest is the request body for /detect-loop.
type DetectLoopRequest struct {
	Calls []ToolCall `json:"calls"`
}

// DetectDriftRequest is the request body for /detect-drift.
type DetectDriftRequest struct {
	Task          string   `json:"task"`
	RecentActions []Action `json:"recent_actions"`
}

// CompressRequest is the request body for /compress.
type CompressRequest struct {
	Content   string `json:"content"`
	MaxTokens int    `json:"max_tokens"`
}

// CompressResponse is the response body for /compress.
type CompressResponse struct {
	Compressed string `json:"compressed"`
	TokenCount int    `json:"token_count,omitempty"`
}

// ValidateToolCallRequest is the request body for /validate-tool.
type ValidateToolCallRequest struct {
	Tool   string         `json:"tool"`
	Params map[string]any `json:"params"`
}

// DetectPatternRequest is the request body for /scriptor/detect-pattern.
type DetectPatternRequest struct {
	Calls []ToolCall `json:"calls"`
}
