// Package aiops provides a local AI operations agent for Nexora.
//
// It acts as middleware between the CLI and large language models, handling
// operational tasks like edit resolution, loop detection, task drift detection,
// context compression, and script generation.
package aiops

import (
	"context"
	"time"
)

// ToolCall represents a single tool invocation for pattern analysis.
type ToolCall struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Params    map[string]any `json:"params"`
	Result    string         `json:"result,omitempty"`
	Error     string         `json:"error,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
}

// Action represents a high-level action taken by the agent.
type Action struct {
	Description string     `json:"description"`
	ToolCalls   []ToolCall `json:"tool_calls,omitempty"`
	Timestamp   time.Time  `json:"timestamp"`
}

// EditResolution contains the result of resolving a fuzzy edit.
type EditResolution struct {
	ExactOldString string  `json:"exact_old_string"`
	Confidence     float64 `json:"confidence"`
	LineNumber     int     `json:"line_number,omitempty"`
	Context        string  `json:"context,omitempty"`
}

// LoopDetection contains the result of loop detection analysis.
type LoopDetection struct {
	IsLooping   bool   `json:"is_looping"`
	Reason      string `json:"reason,omitempty"`
	Suggestion  string `json:"suggestion,omitempty"`
	Iterations  int    `json:"iterations"`
	PatternType string `json:"pattern_type,omitempty"` // "repetition", "oscillation", "no_progress"
}

// DriftDetection contains the result of task drift analysis.
type DriftDetection struct {
	IsDrifting   bool    `json:"is_drifting"`
	Reason       string  `json:"reason,omitempty"`
	Suggestion   string  `json:"suggestion,omitempty"`
	DriftScore   float64 `json:"drift_score"` // 0.0 = on track, 1.0 = completely off
	OriginalTask string  `json:"original_task,omitempty"`
}

// ValidationResult contains the result of tool call validation.
type ValidationResult struct {
	Valid       bool           `json:"valid"`
	Errors      []string       `json:"errors,omitempty"`
	Warnings    []string       `json:"warnings,omitempty"`
	FixedParams map[string]any `json:"fixed_params,omitempty"`
}

// ModelInfo contains information about the local model.
type ModelInfo struct {
	Name      string        `json:"name"`
	Runtime   string        `json:"runtime"` // "ollama", "llamacpp", "mlx"
	Available bool          `json:"available"`
	Latency   time.Duration `json:"latency,omitempty"`
}

// ScriptPlan describes a batched operation plan.
type ScriptPlan struct {
	PatternType string     `json:"pattern_type"` // "multi_grep", "multi_view", "multi_edit", "pipeline"
	ToolCalls   []ToolCall `json:"tool_calls"`
	Description string     `json:"description"`
	Estimated   struct {
		Speedup    float64 `json:"speedup"`     // e.g., 5.0 = 5x faster
		ToolsSaved int     `json:"tools_saved"` // number of tool calls saved
	} `json:"estimated"`
}

// Script is a generated executable script.
type Script struct {
	Language string      `json:"language"` // "bash", "python"
	Code     string      `json:"code"`
	Plan     *ScriptPlan `json:"plan,omitempty"`
}

// ScriptResult contains the result of script execution.
type ScriptResult struct {
	Success  bool    `json:"success"`
	Output   string  `json:"output"`
	Error    string  `json:"error,omitempty"`
	ExitCode int     `json:"exit_code"`
	Duration int64   `json:"duration_ms"`
	Script   *Script `json:"script,omitempty"`
}

// Pattern types for Scriptor.
const (
	PatternMultiGrep = "multi_grep" // Multiple grep calls → single regex
	PatternMultiView = "multi_view" // Multiple view calls → single range
	PatternMultiEdit = "multi_edit" // Multiple edits same file → batch
	PatternPipeline  = "pipeline"   // Search/read cycles → pipeline
	PatternMultiLS   = "multi_ls"   // Multiple ls calls → find
	PatternMultiTest = "multi_test" // Multiple test runs → single json run
	PatternExistence = "existence"  // File existence checks → test -f batch
)

// Scriptor batches multiple tool operations into single scripts.
type Scriptor interface {
	// DetectPattern analyzes tool calls for batchable patterns.
	DetectPattern(ctx context.Context, calls []ToolCall) (*ScriptPlan, error)

	// Compile generates an executable script from a plan.
	Compile(ctx context.Context, plan *ScriptPlan) (*Script, error)

	// Execute runs a script and captures output.
	Execute(ctx context.Context, script *Script) (*ScriptResult, error)

	// Report formats a script result for the big model.
	Report(result *ScriptResult) string
}

// Ops is the main interface for the local AI ops agent.
type Ops interface {
	// ResolveEdit resolves a fuzzy old_string to an exact match in the file.
	ResolveEdit(ctx context.Context, fileContent, oldString, newString string) (*EditResolution, error)

	// DetectLoop analyzes recent tool calls for loop patterns.
	DetectLoop(ctx context.Context, calls []ToolCall) (*LoopDetection, error)

	// DetectDrift checks if recent actions have drifted from the original task.
	DetectDrift(ctx context.Context, task string, recentActions []Action) (*DriftDetection, error)

	// Compress summarizes content to fit within token budget.
	Compress(ctx context.Context, content string, maxTokens int) (string, error)

	// ValidateToolCall validates and optionally fixes tool call parameters.
	ValidateToolCall(ctx context.Context, tool string, params map[string]any) (*ValidationResult, error)

	// Scriptor returns the script batching interface.
	Scriptor() Scriptor

	// Available returns true if a local model is available.
	Available() bool

	// ModelInfo returns information about the local model.
	ModelInfo() ModelInfo
}

// Config holds the configuration for the aiops agent.
type Config struct {
	Enabled  bool          `json:"enabled" yaml:"enabled"`
	Endpoint string        `json:"endpoint" yaml:"endpoint"` // e.g., "http://gpu-box:8420"
	Timeout  time.Duration `json:"timeout" yaml:"timeout"`
	Fallback bool          `json:"fallback" yaml:"fallback"` // continue without aiops if unavailable

	Scriptor struct {
		Enabled      bool     `json:"enabled" yaml:"enabled"`
		MinBatchSize int      `json:"min_batch_size" yaml:"min_batch_size"`
		Languages    []string `json:"languages" yaml:"languages"`
	} `json:"scriptor" yaml:"scriptor"`

	LoopDetection struct {
		Enabled   bool `json:"enabled" yaml:"enabled"`
		Threshold int  `json:"threshold" yaml:"threshold"` // same tool N times
	} `json:"loop_detection" yaml:"loop_detection"`

	DriftDetection struct {
		Enabled       bool `json:"enabled" yaml:"enabled"`
		CheckInterval int  `json:"check_interval" yaml:"check_interval"` // check every N tool calls
	} `json:"drift_detection" yaml:"drift_detection"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	cfg := Config{
		Enabled:  false, // disabled by default until endpoint configured
		Endpoint: "",
		Timeout:  5 * time.Second,
		Fallback: true,
	}
	cfg.Scriptor.Enabled = true
	cfg.Scriptor.MinBatchSize = 3
	cfg.Scriptor.Languages = []string{"bash", "python"}
	cfg.LoopDetection.Enabled = true
	cfg.LoopDetection.Threshold = 5
	cfg.DriftDetection.Enabled = true
	cfg.DriftDetection.CheckInterval = 10
	return cfg
}
