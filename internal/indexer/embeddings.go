package indexer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"os"
	"strings"
	"time"
)

// Embedding represents a vector embedding of code text
type Embedding struct {
	ID       string    `json:"id"`
	Type     string    `json:"type"` // function, struct, interface, variable, comment
	Text     string    `json:"text"`
	Vector   []float32 `json:"vector"`
	Metadata MetaData  `json:"metadata"`
	Created  time.Time `json:"created"`
}

type MetaData struct {
	Package    string `json:"package"`
	File       string `json:"file"`
	Line       int    `json:"line"`
	Complexity int    `json:"complexity"`
	Public     bool   `json:"public"`
}

// EmbeddingProvider interface for different embedding models
type EmbeddingProvider interface {
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
	Name() string
}

// OpenAIProvider uses OpenAI's embedding API
type OpenAIProvider struct {
	apiKey  string
	model   string
	baseURL string
}

func NewOpenAIProvider(apiKey, baseURL, model string) *OpenAIProvider {
	if model == "" {
		model = "text-embedding-3-small"
	}
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	return &OpenAIProvider{
		apiKey:  apiKey,
		model:   model,
		baseURL: baseURL,
	}
}

func (p *OpenAIProvider) Name() string {
	return "openai-" + p.model
}

func (p *OpenAIProvider) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	// Mock implementation for now - in production, would call OpenAI API
	// Return a simple hash-based embedding for demo purposes
	return generateMockEmbedding(text), nil
}

// LocalProvider uses local embedding models
type LocalProvider struct {
	model string
	path  string
}

func NewLocalProvider(model, path string) *LocalProvider {
	if model == "" {
		model = "all-minilm:l6-v2"
	}
	return &LocalProvider{
		model: model,
		path:  path,
	}
}

func (p *LocalProvider) Name() string {
	return "local-" + p.model
}

func (p *LocalProvider) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	// Mock implementation - would call local model via exec or HTTP
	return generateMockEmbedding(text), nil
}

// MistralProvider uses Mistral AI's embedding API with advanced features
type MistralProvider struct {
	apiKey      string
	model       string
	baseURL     string
	client      *http.Client
	modelConfig MistralModelConfig
}

// MistralModelConfig holds configuration specific to each model
type MistralModelConfig struct {
	MaxTokens     int      `json:"max_tokens"`
	EmbeddingDims int      `json:"embedding_dims"`
	InputPrice    float64  `json:"input_price_per_1k"`
	Capabilities  []string `json:"capabilities"`
}

// Model configurations for all supported Mistral models
var mistralModelConfigs = map[string]MistralModelConfig{
	MistralModelMistralLarge3: {
		MaxTokens:     131072,
		EmbeddingDims: 1536,
		InputPrice:    0.3,
		Capabilities:  []string{"reasoning", "analysis", "embeddings", "large-context"},
	},
	MistralModelMinistral3: {
		MaxTokens:     65536,
		EmbeddingDims: 1024,
		InputPrice:    0.025,
		Capabilities:  []string{"fast", "efficient", "embeddings"},
	},
	MistralModelEmbed: {
		MaxTokens:     8000,
		EmbeddingDims: 1024,
		InputPrice:    0.01,
		Capabilities:  []string{"embeddings-only", "fast"},
	},
}

func NewMistralProvider(apiKey, model string) *MistralProvider {
	if model == "" {
		model = MistralModelEmbed // Use embedding-optimized model by default
	}
	if apiKey == "" {
		apiKey = os.Getenv("MISTRAL_API_KEY")
	}

	config, exists := mistralModelConfigs[model]
	if !exists {
		slog.Warn("Unknown Mistral model, using default config", "model", model)
		config = mistralModelConfigs[MistralModelEmbed]
	}

	return &MistralProvider{
		apiKey:  apiKey,
		model:   model,
		baseURL: "https://api.mistral.ai/v1",
		client: &http.Client{
			Timeout: 60 * time.Second, // Increased timeout for large models
		},
		modelConfig: config,
	}
}

func (p *MistralProvider) Name() string {
	return "mistral-" + p.model
}

func (p *MistralProvider) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	if p.apiKey == "" {
		slog.Warn("No Mistral API key provided, falling back to mock embeddings")
		return generateMockEmbedding(text), nil
	}

	// Validate model capabilities
	if !p.modelSupportsEmbeddings() {
		return nil, fmt.Errorf("model %s does not support embeddings", p.model)
	}

	// Prepare request with model-specific optimizations
	reqBody := p.buildEmbeddingRequest(text)

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request with retry logic
	embedding, err := p.makeEmbeddingRequestWithRetry(ctx, jsonData, 3)
	if err != nil {
		return nil, err
	}

	// Validate embedding dimensions
	expectedDims := p.modelConfig.EmbeddingDims
	if len(embedding) != expectedDims {
		slog.Warn("Embedding dimension mismatch",
			"expected", expectedDims,
			"actual", len(embedding),
			"model", p.model)
	}

	return embedding, nil
}

func (p *MistralProvider) modelSupportsEmbeddings() bool {
	capabilities := p.modelConfig.Capabilities
	for _, cap := range capabilities {
		if cap == "embeddings" || cap == "embeddings-only" {
			return true
		}
	}
	return false
}

func (p *MistralProvider) buildEmbeddingRequest(text string) map[string]any {
	reqBody := map[string]any{
		"input": text,
		"model": p.model,
	}

	// Add model-specific optimizations
	switch p.model {
	case MistralModelMistralLarge3:
		// For large models, we can request higher precision
		reqBody["encoding_format"] = "float"
		reqBody["output_dimension"] = p.modelConfig.EmbeddingDims
	default:
		reqBody["encoding_format"] = "float"
	}

	return reqBody
}

func (p *MistralProvider) makeEmbeddingRequestWithRetry(ctx context.Context, jsonData []byte, maxRetries int) ([]float32, error) {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			slog.Debug("Retrying Mistral API request", "attempt", attempt, "backoff", backoff)

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}

		embedding, err := p.makeEmbeddingRequest(ctx, jsonData)
		if err == nil {
			return embedding, nil
		}

		lastErr = err

		// Don't retry on certain errors
		if strings.Contains(err.Error(), "401") || // Unauthorized
			strings.Contains(err.Error(), "400") || // Bad request
			strings.Contains(err.Error(), "404") { // Model not found
			break
		}
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", maxRetries+1, lastErr)
}

func (p *MistralProvider) makeEmbeddingRequest(ctx context.Context, jsonData []byte) ([]float32, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("User-Agent", "Nexora-Indexer/1.0")

	// Send request
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
			Index     int       `json:"index"`
			Object    string    `json:"object"`
		} `json:"data"`
		Model  string `json:"model"`
		Object string `json:"object"`
		Usage  struct {
			PromptTokens     int     `json:"prompt_tokens"`
			CompletionTokens int     `json:"completion_tokens"`
			TotalTokens      int     `json:"total_tokens"`
			PromptAudioSec   float64 `json:"prompt_audio_seconds"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(result.Data) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	// Log usage for cost tracking
	if result.Usage.TotalTokens > 0 {
		cost := float64(result.Usage.PromptTokens) / 1000.0 * p.modelConfig.InputPrice
		slog.Debug("Mistral API usage",
			"model", p.model,
			"tokens", result.Usage.PromptTokens,
			"cost", fmt.Sprintf("$%.4f", cost))
	}

	return result.Data[0].Embedding, nil
}

// GetModelInfo returns information about the current model
func (p *MistralProvider) GetModelInfo() MistralModelConfig {
	return p.modelConfig
}

// SupportsBatchEmbeddings checks if the model supports batch processing
func (p *MistralProvider) SupportsBatchEmbeddings() bool {
	// All Mistral models support batch embeddings
	return true
}

// GenerateBatchEmbeddings generates embeddings for multiple texts efficiently
func (p *MistralProvider) GenerateBatchEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	if p.apiKey == "" {
		// Fallback to individual mock embeddings
		embeddings := make([][]float32, len(texts))
		for i, text := range texts {
			embeddings[i] = generateMockEmbedding(text)
		}
		return embeddings, nil
	}

	// For batch requests, we can't use the single embedding response
	// Instead, we should make individual requests for each text as fallback
	slog.Warn("Batch embedding request not properly implemented, falling back to individual requests")
	embeddings := make([][]float32, len(texts))
	for i, text := range texts {
		singleEmbedding, err := p.GenerateEmbedding(ctx, text)
		if err != nil {
			return nil, fmt.Errorf("failed to generate embedding for text %d: %w", i, err)
		}
		embeddings[i] = singleEmbedding
	}

	return embeddings, nil
}

// ValidateAPIKey checks if the provided API key is valid
func (p *MistralProvider) ValidateAPIKey(ctx context.Context) bool {
	if p.apiKey == "" {
		return false
	}

	// Make a minimal request to validate the API key
	testText := "test"
	reqBody := map[string]any{
		"input": testText,
		"model": p.model,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return false
	}

	_, err = p.makeEmbeddingRequest(ctx, jsonData)
	return err == nil
}

// GetPricingInfo returns pricing information
func (p *MistralProvider) GetPricingInfo() map[string]any {
	return map[string]any{
		"model":              p.model,
		"input_price_per_1k": p.modelConfig.InputPrice,
		"max_tokens":         p.modelConfig.MaxTokens,
		"embedding_dims":     p.modelConfig.EmbeddingDims,
		"capabilities":       p.modelConfig.Capabilities,
	}
}

// Mistral model constants
const (
	MistralModelDevstral2 = "mistral-embed"

	MistralModelMistralLarge3 = "mistral-large-3-25-12"
	MistralModelMinistral3    = "ministral-3-14b-25-12"
	MistralModelEmbed         = "mistral-embed"
)

// Optimization helpers for different use cases
func OptimizeForCodeSearch() string {
	return MistralModelMinistral3
}

func OptimizeForSpeed() string {
	return MistralModelMinistral3
}

func OptimizeForQuality() string {
	return MistralModelMistralLarge3
}

func OptimizeForCost() string {
	return MistralModelEmbed
}

func GetAllSupportedModels() []string {
	return []string{
		MistralModelDevstral2,
		MistralModelMinistral3,
		MistralModelMistralLarge3,
		MistralModelMinistral3,
		MistralModelEmbed,
	}
}

// EmbeddingEngine handles the creation and storage of embeddings
type EmbeddingEngine struct {
	provider EmbeddingProvider
	indexer  *Indexer
	ctx      context.Context
}

func NewEmbeddingEngine(provider EmbeddingProvider, indexer *Indexer) *EmbeddingEngine {
	return &EmbeddingEngine{
		provider: provider,
		indexer:  indexer,
		ctx:      context.Background(),
	}
}

// GenerateEmbedding creates an embedding for the given text
func (e *EmbeddingEngine) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	return e.provider.GenerateEmbedding(ctx, text)
}

// GenerateSymbolEmbeddings creates embeddings for all symbols
func (e *EmbeddingEngine) GenerateSymbolEmbeddings(ctx context.Context, symbols []Symbol) ([]Embedding, error) {
	embeddings := make([]Embedding, 0, len(symbols))

	for _, symbol := range symbols {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Create rich text for embedding
		text := e.createEmbeddingText(symbol)
		if text == "" {
			continue
		}

		vector, err := e.provider.GenerateEmbedding(ctx, text)
		if err != nil {
			slog.Warn("Failed to generate embedding", "symbol", symbol.Name, "error", err)
			continue
		}

		embedding := Embedding{
			ID:     symbol.Name,
			Type:   symbol.Type,
			Text:   text,
			Vector: vector,
			Metadata: MetaData{
				Package:    symbol.Package,
				File:       symbol.File,
				Line:       symbol.Line,
				Complexity: e.calculateComplexity(symbol),
				Public:     symbol.Public,
			},
			Created: time.Now(),
		}

		embeddings = append(embeddings, embedding)
	}

	return embeddings, nil
}

// createEmbeddingText builds rich text for better semantic understanding
func (e *EmbeddingEngine) createEmbeddingText(symbol Symbol) string {
	var parts []string

	// Add documentation
	if symbol.Doc != "" {
		parts = append(parts, "Documentation: "+symbol.Doc)
	}

	// Add signature
	if symbol.Signature != "" {
		parts = append(parts, "Signature: "+symbol.Signature)
	}

	// Add function body context for functions
	if symbol.Type == "func" || symbol.Type == "method" {
		parts = append(parts, "Type: Function")
		if symbol.Params != nil {
			var paramDesc []string
			for _, param := range symbol.Params {
				if param.Name != "" {
					paramDesc = append(paramDesc, fmt.Sprintf("%s %s", param.Name, param.Type))
				} else {
					paramDesc = append(paramDesc, param.Type)
				}
			}
			parts = append(parts, "Parameters: "+strings.Join(paramDesc, ", "))
		}
		if symbol.Returns != nil {
			parts = append(parts, "Returns: "+strings.Join(symbol.Returns, ", "))
		}
	}

	// Add type-specific information
	switch symbol.Type {
	case "struct":
		parts = append(parts, "Type: Struct")
		if symbol.Fields != nil {
			var fieldDesc []string
			for _, field := range symbol.Fields {
				fieldDesc = append(fieldDesc, fmt.Sprintf("%s %s", field.Name, field.Type))
			}
			parts = append(parts, "Fields: "+strings.Join(fieldDesc, ", "))
		}
	case "interface":
		parts = append(parts, "Type: Interface")
		if symbol.Methods != nil {
			parts = append(parts, "Methods: "+strings.Join(symbol.Methods, ", "))
		}
	case "var", "const":
		parts = append(parts, fmt.Sprintf("Type: %s", strings.Title(symbol.Type)))
	}

	// Add package context
	parts = append(parts, fmt.Sprintf("Package: %s", symbol.Package))
	parts = append(parts, fmt.Sprintf("File: %s:%d", symbol.File, symbol.Line))

	return strings.Join(parts, "\n")
}

// calculateComplexity estimates code complexity
func (e *EmbeddingEngine) calculateComplexity(symbol Symbol) int {
	complexity := 1 // base complexity

	switch symbol.Type {
	case "func", "method":
		// Complexity based on number of parameters and return values
		complexity += len(symbol.Params) + len(symbol.Returns)
		// If we have function calls, add complexity
		complexity += len(symbol.Calls)
	case "struct":
		complexity += len(symbol.Fields)
	case "interface":
		complexity += len(symbol.Methods)
	}

	// Add complexity for public symbols
	if symbol.Public {
		complexity++
	}

	// Add complexity for detailed documentation
	if symbol.Doc != "" && len(strings.Fields(symbol.Doc)) > 10 {
		complexity += 2
	}

	return complexity
}

// SearchSimilar finds similar code using vector similarity
func (e *EmbeddingEngine) SearchSimilar(ctx context.Context, query string, limit int) ([]Embedding, error) {
	// Generate embedding for query
	queryVector, err := e.provider.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Get all embeddings from storage
	allEmbeddings, err := e.indexer.GetAllEmbeddings(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get embeddings: %w", err)
	}

	// Calculate similarity scores
	type scoredEmbedding struct {
		embedding Embedding
		score     float32
	}

	scored := make([]scoredEmbedding, 0, len(allEmbeddings))
	for _, emb := range allEmbeddings {
		score := e.cosineSimilarity(queryVector, emb.Vector)
		if score > 0.1 { // Threshold for relevance
			scored = append(scored, scoredEmbedding{
				embedding: emb,
				score:     score,
			})
		}
	}

	// Sort by score
	for i := 0; i < len(scored)-1; i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[i].score < scored[j].score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	// Return top results
	result := make([]Embedding, 0, limit)
	for i, s := range scored {
		if i >= limit {
			break
		}
		result = append(result, s.embedding)
	}

	return result, nil
}

// cosineSimilarity calculates the cosine similarity between two vectors
func (e *EmbeddingEngine) cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float32
	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (float32(math.Sqrt(float64(normA)) * math.Sqrt(float64(normB))))
}

// Mock embedding generation for development
func generateMockEmbedding(text string) []float32 {
	// Simple hash-based embedding for demo purposes
	// In production, this would use actual embedding models
	dimensions := 384 // Standard embedding size
	vector := make([]float32, dimensions)

	// Generate deterministic pseudo-random vector based on text
	hash := 0
	for _, char := range text {
		hash = hash*31 + int(char)
	}
	seed := uint64(hash)

	for i := 0; i < dimensions; i++ {
		seed = seed*1103515245 + 12345
		vector[i] = float32((seed>>16)&0xFF) / 255.0
	}

	return vector
}
