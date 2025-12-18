// Package aiops provides a client for the remote AIOPS service.
//
// AIOPS runs on a separate machine with a GPU and handles operational tasks
// like edit resolution, loop detection, task drift detection, context
// compression, and script generation using a local 3B model.
package aiops

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client connects to a remote AIOPS service.
type Client struct {
	endpoint string
	client   *http.Client
	enabled  bool
}

// NewClient creates a new AIOPS client.
func NewClient(cfg Config) *Client {
	if !cfg.Enabled || cfg.Endpoint == "" {
		return &Client{enabled: false}
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	return &Client{
		endpoint: cfg.Endpoint,
		enabled:  true,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// Available returns true if the AIOPS service is reachable.
func (c *Client) Available() bool {
	if !c.enabled {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", c.endpoint+"/health", nil)
	if err != nil {
		return false
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// ModelInfo returns information about the remote model.
func (c *Client) ModelInfo() ModelInfo {
	if !c.enabled {
		return ModelInfo{Available: false}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", c.endpoint+"/info", nil)
	if err != nil {
		return ModelInfo{Available: false}
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return ModelInfo{Available: false}
	}
	defer resp.Body.Close()

	var info ModelInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return ModelInfo{Available: false}
	}

	return info
}

// ResolveEdit resolves a fuzzy old_string to an exact match.
func (c *Client) ResolveEdit(ctx context.Context, fileContent, oldString, newString string) (*EditResolution, error) {
	if !c.enabled {
		return nil, ErrNotEnabled
	}

	req := ResolveEditRequest{
		FileContent: fileContent,
		OldString:   oldString,
		NewString:   newString,
	}

	var resp EditResolution
	if err := c.post(ctx, "/resolve-edit", req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// DetectLoop analyzes recent tool calls for loop patterns.
func (c *Client) DetectLoop(ctx context.Context, calls []ToolCall) (*LoopDetection, error) {
	if !c.enabled {
		return nil, ErrNotEnabled
	}

	req := DetectLoopRequest{Calls: calls}

	var resp LoopDetection
	if err := c.post(ctx, "/detect-loop", req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// DetectDrift checks if recent actions have drifted from the original task.
func (c *Client) DetectDrift(ctx context.Context, task string, recentActions []Action) (*DriftDetection, error) {
	if !c.enabled {
		return nil, ErrNotEnabled
	}

	req := DetectDriftRequest{
		Task:          task,
		RecentActions: recentActions,
	}

	var resp DriftDetection
	if err := c.post(ctx, "/detect-drift", req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// Compress summarizes content to fit within token budget.
func (c *Client) Compress(ctx context.Context, content string, maxTokens int) (string, error) {
	if !c.enabled {
		return "", ErrNotEnabled
	}

	req := CompressRequest{
		Content:   content,
		MaxTokens: maxTokens,
	}

	var resp CompressResponse
	if err := c.post(ctx, "/compress", req, &resp); err != nil {
		return "", err
	}

	return resp.Compressed, nil
}

// ValidateToolCall validates and optionally fixes tool call parameters.
func (c *Client) ValidateToolCall(ctx context.Context, tool string, params map[string]any) (*ValidationResult, error) {
	if !c.enabled {
		return nil, ErrNotEnabled
	}

	req := ValidateToolCallRequest{
		Tool:   tool,
		Params: params,
	}

	var resp ValidationResult
	if err := c.post(ctx, "/validate-tool", req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// Scriptor returns a scriptor client for batch operations.
func (c *Client) Scriptor() Scriptor {
	return &scriptorClient{client: c}
}

func (c *Client) post(ctx context.Context, path string, reqBody, respBody any) error {
	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.endpoint+path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil && errResp.Error != "" {
			return fmt.Errorf("aiops error: %s", errResp.Error)
		}
		return fmt.Errorf("aiops returned status %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

// scriptorClient implements Scriptor over HTTP.
type scriptorClient struct {
	client *Client
}

func (s *scriptorClient) DetectPattern(ctx context.Context, calls []ToolCall) (*ScriptPlan, error) {
	if !s.client.enabled {
		return nil, ErrNotEnabled
	}

	req := DetectPatternRequest{Calls: calls}

	var resp ScriptPlan
	if err := s.client.post(ctx, "/scriptor/detect-pattern", req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (s *scriptorClient) Compile(ctx context.Context, plan *ScriptPlan) (*Script, error) {
	if !s.client.enabled {
		return nil, ErrNotEnabled
	}

	var resp Script
	if err := s.client.post(ctx, "/scriptor/compile", plan, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (s *scriptorClient) Execute(ctx context.Context, script *Script) (*ScriptResult, error) {
	if !s.client.enabled {
		return nil, ErrNotEnabled
	}

	var resp ScriptResult
	if err := s.client.post(ctx, "/scriptor/execute", script, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (s *scriptorClient) Report(result *ScriptResult) string {
	if result == nil {
		return ""
	}

	if result.Success {
		return fmt.Sprintf("Script executed successfully (%dms):\n%s", result.Duration, result.Output)
	}
	return fmt.Sprintf("Script failed (exit %d):\n%s\n%s", result.ExitCode, result.Output, result.Error)
}
