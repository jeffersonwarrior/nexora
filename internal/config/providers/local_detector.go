package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// LocalModel represents a detected local model with metadata
type LocalModel struct {
	ID      string `json:"id"`
	Name    string `json:"name,omitempty"` // Friendly display name
	Context int    `json:"context"`
	Matched string `json:"matched"`
}

// LocalProvider represents detected local provider config
type LocalProvider struct {
	Type     string       `json:"type"`
	Endpoint string       `json:"endpoint"`
	APIKey   string       `json:"api_key,omitempty"`
	Models   []LocalModel `json:"models"`
	Name     string       `json:"name"`
	BaseURL  string       `json:"base_url"`
}

// LocalDetector handles auto-detection of local model servers
type LocalDetector struct {
	endpoint string
	client   *http.Client
}

// NewLocalDetector creates a new detector for the given endpoint
func NewLocalDetector(endpoint string) *LocalDetector {
	return &LocalDetector{
		endpoint: strings.TrimSuffix(endpoint, "/"),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Detect attempts to identify the server type and available models
func (d *LocalDetector) Detect(serverType, apiKey string) (*LocalProvider, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	provider := &LocalProvider{
		Type:     "local",
		Endpoint: d.endpoint,
		APIKey:   apiKey,
		Name:     "Local Models",
		BaseURL:  d.endpoint,
	}

	switch serverType {
	case "ollama":
		return d.detectOllama(ctx, provider)
	case "openai", "vllm", "lm-studio":
		return d.detectOpenAICompatible(ctx, provider)
	default:
		// Auto-detect by trying common endpoints in priority order
		return d.autoDetect(ctx, provider)
	}
}

func (d *LocalDetector) detectOllama(ctx context.Context, provider *LocalProvider) (*LocalProvider, error) {
	// Test /api/tags for Ollama
	req, err := http.NewRequestWithContext(ctx, "GET", d.endpoint+"/api/tags", nil)
	if err != nil {
		return nil, err
	}

	if provider.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+provider.APIKey)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama detection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return nil, fmt.Errorf("ollama requires API key")
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("ollama /api/tags returned %d", resp.StatusCode)
	}

	var ollamaResp struct {
		Models []struct {
			Name string `json:"name"`
			Size int64  `json:"size"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to parse ollama response: %w", err)
	}

	for _, m := range ollamaResp.Models {
		// Try to get actual context window from model details
		context := d.getOllamaModelDetails(ctx, provider, m.Name)
		if context == 0 {
			// Fall back to estimation if API query fails
			context = d.getOllamaContext(m.Name)
		}
		matched := MatchModelToLibrary(m.Name)

		// Generate friendly name
		name := m.Name
		if matched != "" {
			name = matched
		}

		provider.Models = append(provider.Models, LocalModel{
			ID:      m.Name,
			Name:    name,
			Context: context,
			Matched: matched,
		})
	}

	return provider, nil
}

func (d *LocalDetector) detectOpenAICompatible(ctx context.Context, provider *LocalProvider) (*LocalProvider, error) {
	// Test /v1/models for OpenAI-compatible servers
	req, err := http.NewRequestWithContext(ctx, "GET", d.endpoint+"/v1/models", nil)
	if err != nil {
		return nil, err
	}

	if provider.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+provider.APIKey)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openai-compatible detection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return nil, fmt.Errorf("server requires API key")
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("v1/models returned %d", resp.StatusCode)
	}

	var openaiResp struct {
		Data []struct {
			ID           string `json:"id"`
			Object       string `json:"object"`
			Created      int64  `json:"created"`
			OwnedBy      string `json:"owned_by"`
			Context      int    `json:"context_window,omitempty"` // Some providers
			MaxModelLen  int    `json:"max_model_len,omitempty"`  // vLLM uses this
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&openaiResp); err != nil {
		return nil, fmt.Errorf("failed to parse openai response: %w", err)
	}

	for _, m := range openaiResp.Data {
		// Try different context window fields
		context := m.Context
		if context == 0 && m.MaxModelLen > 0 {
			// vLLM returns max_model_len instead of context_window
			context = m.MaxModelLen
		}
		if context == 0 {
			// Estimate context based on model name if not provided
			context = d.estimateContextFromName(m.ID)
		}

		matched := MatchModelToLibrary(m.ID)

		// Generate friendly name
		name := m.ID
		if matched != "" {
			name = matched
		}

		provider.Models = append(provider.Models, LocalModel{
			ID:      m.ID,
			Name:    name,
			Context: context,
			Matched: matched,
		})
	}

	return provider, nil
}

func (d *LocalDetector) autoDetect(ctx context.Context, provider *LocalProvider) (*LocalProvider, error) {
	// Try Ollama first
	if p, err := d.detectOllama(ctx, provider); err == nil {
		return p, nil
	}

	// Try OpenAI compatible
	if p, err := d.detectOpenAICompatible(ctx, provider); err == nil {
		return p, nil
	}

	return nil, fmt.Errorf("unable to detect server type at %s", d.endpoint)
}

// getOllamaModelDetails queries Ollama's /api/show endpoint to get actual context window
func (d *LocalDetector) getOllamaModelDetails(ctx context.Context, provider *LocalProvider, modelName string) int {
	reqBody := fmt.Sprintf(`{"name":"%s"}`, modelName)
	req, err := http.NewRequestWithContext(ctx, "POST", d.endpoint+"/api/show", strings.NewReader(reqBody))
	if err != nil {
		return 0
	}

	req.Header.Set("Content-Type", "application/json")
	if provider.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+provider.APIKey)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0
	}

	var showResp struct {
		ModelInfo struct {
			// Ollama returns model parameters in different fields
			ContextLength  int `json:"context_length"`
			NumCtx         int `json:"num_ctx"`
		} `json:"model_info"`
		Details struct {
			// Alternative location for context info
			ContextLength int `json:"context_length"`
		} `json:"details"`
		// Sometimes it's at the top level
		ContextLength int `json:"context_length"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&showResp); err != nil {
		return 0
	}

	// Try different fields where context length might be
	if showResp.ContextLength > 0 {
		return showResp.ContextLength
	}
	if showResp.ModelInfo.ContextLength > 0 {
		return showResp.ModelInfo.ContextLength
	}
	if showResp.ModelInfo.NumCtx > 0 {
		return showResp.ModelInfo.NumCtx
	}
	if showResp.Details.ContextLength > 0 {
		return showResp.Details.ContextLength
	}

	return 0
}

// EstimateContext estimates the context window size for a model based on its name
// This is a public method used by tests and potentially other callers
func (d *LocalDetector) EstimateContext(modelName string) int {
	return d.getOllamaContext(modelName)
}

// getOllamaContext estimates context window size based on model name patterns
func (d *LocalDetector) getOllamaContext(modelName string) int {
	name := strings.ToLower(modelName)

	// Check for explicit context window indicators first (e.g., "128k", "32k")
	if strings.Contains(name, "128k") || strings.Contains(name, ":128k") {
		return 131072 // 128k tokens
	}
	if strings.Contains(name, "32k") || strings.Contains(name, ":32k") {
		return 32768 // 32k tokens
	}
	if strings.Contains(name, "16k") || strings.Contains(name, ":16k") {
		return 16384 // 16k tokens
	}

	// Known Ollama model context windows based on parameter count
	// Check specific patterns only (e.g., "405b" but NOT just "405")
	if strings.Contains(name, "405b") {
		return 131072 // 128k
	}
	if strings.Contains(name, "120b") {
		return 131072 // Large models like GPT-OSS-120B
	}
	if strings.Contains(name, "72b") || strings.Contains(name, "70b") {
		return 131072 // 128k
	}
	if strings.Contains(name, "34b") || strings.Contains(name, "33b") {
		return 131072 // 128k
	}
	if strings.Contains(name, "22b") {
		return 65536 // 64k
	}
	if strings.Contains(name, "13b") {
		return 8192
	}
	if strings.Contains(name, "8b") {
		return 8192
	}
	if strings.Contains(name, "7b") {
		return 8192
	}

	// Default assumption for unknown models
	return 4096
}

func (d *LocalDetector) estimateContextFromName(modelID string) int {
	name := strings.ToLower(modelID)

	// Check for explicit context window indicators first (e.g., "128k", "32k")
	if strings.Contains(name, "128k") || strings.Contains(name, ":128k") {
		return 131072 // 128k tokens
	}
	if strings.Contains(name, "32k") || strings.Contains(name, ":32k") {
		return 32768 // 32k tokens
	}
	if strings.Contains(name, "16k") || strings.Contains(name, ":16k") {
		return 16384 // 16k tokens
	}

	// Model parameter count patterns (specific patterns only)
	if strings.Contains(name, "405b") {
		return 131072 // 128k
	}
	if strings.Contains(name, "120b") {
		return 131072 // Large models like GPT-OSS-120B
	}
	if strings.Contains(name, "70b") {
		return 131072 // 128k
	}
	if strings.Contains(name, "34b") {
		return 131072 // 128k
	}
	if strings.Contains(name, "22b") {
		return 65536 // 64k
	}
	if strings.Contains(name, "13b") {
		return 8192
	}
	if strings.Contains(name, "8b") {
		return 8192
	}
	if strings.Contains(name, "7b") {
		return 8192
	}

	// Default assumption for unknown models
	return 4096
}

// MatchModelToLibrary matches raw model names to Nexora library equivalents
func MatchModelToLibrary(rawModel string) string {
	name := strings.ToLower(rawModel)

	// Exact matches
	if strings.Contains(name, "llama3.1:405b") || strings.Contains(name, "llama-3.1-405b") {
		return "llama-3.1-405b-instruct"
	}
	if strings.Contains(name, "llama3.1:70b") || strings.Contains(name, "llama-3.1-70b") {
		return "llama-3.1-70b-instruct"
	}
	if strings.Contains(name, "llama3.1:8b") || strings.Contains(name, "llama-3.1-8b") {
		return "llama-3.1-8b-instruct"
	}
	if strings.Contains(name, "codellama:34b") || strings.Contains(name, "codellama-34b") {
		return "codellama-34b-instruct"
	}
	if strings.Contains(name, "mixtral:8x22b") || strings.Contains(name, "mixtral-8x22b") {
		return "mixtral-8x22b-instruct"
	}

	// Partial matches
	if strings.Contains(name, "70b") {
		return "llama-3-70b-instruct"
	}
	if strings.Contains(name, "8b") {
		return "llama-3-8b-instruct"
	}
	if strings.Contains(name, "7b") {
		return "llama-7b-instruct"
	}
	if strings.Contains(name, "codellama") {
		return "codellama-7b-instruct"
	}

	// Fallback
	return rawModel
}

// PrioritizeModels sorts models by context window priority
func PrioritizeModels(models []LocalModel) []LocalModel {
	prioritized := make([]LocalModel, len(models))
	copy(prioritized, models)

	// Sort: >64k first, then 32-64k, then 32k+
	for i := 0; i < len(prioritized)-1; i++ {
		for j := i + 1; j < len(prioritized); j++ {
			if compareContext(prioritized[j].Context, prioritized[i].Context) > 0 {
				prioritized[i], prioritized[j] = prioritized[j], prioritized[i]
			}
		}
	}

	return prioritized
}

func compareContext(a, b int) int {
	// >64k is highest priority
	isLargeA := a > 65536
	isLargeB := b > 65536

	if isLargeA && !isLargeB {
		return 1
	}
	if !isLargeA && isLargeB {
		return -1
	}

	// 32-64k is medium priority
	isMediumA := a >= 32768 && a <= 65536
	isMediumB := b >= 32768 && b <= 65536

	if isMediumA && !isMediumB {
		return 1
	}
	if !isMediumA && isMediumB {
		return -1
	}

	// Otherwise, larger context wins
	if a > b {
		return 1
	}
	if b > a {
		return -1
	}

	return 0
}
