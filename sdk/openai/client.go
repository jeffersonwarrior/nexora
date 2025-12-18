package openai

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/nexora/sdk/base"
)

// Client represents the OpenAI API client extending the base client
type Client struct {
	*base.Client
}

// NewClient creates a new OpenAI client
func NewClient(apiKey string, opts ...base.ClientOption) *Client {
	// Create base configuration
	config := base.DefaultClientConfig
	config.APIKey = apiKey
	config.BaseURL = "https://api.openai.com/v1"

	// Apply user options
	for _, opt := range opts {
		opt(&config)
	}

	baseClient := base.NewClient(config)

	return &Client{
		Client: baseClient,
	}
}

// CreateChatCompletion creates a chat completion
func (c *Client) CreateChatCompletion(ctx context.Context, req base.ChatCompletionRequest) (base.ChatCompletionResponse, error) {
	var response interface{}
	err := c.DoRequest(ctx, "POST", "/chat/completions", req, &response)
	if err != nil {
		return base.ChatCompletionResponse{}, err
	}

	// Convert interface{} back to response structure
	responseBytes, _ := json.Marshal(response)
	var chatResponse base.ChatCompletionResponse
	json.Unmarshal(responseBytes, &chatResponse)

	return chatResponse, nil
}

// CreateEmbedding creates embeddings
func (c *Client) CreateEmbedding(ctx context.Context, req base.EmbeddingRequest) (base.EmbeddingResponse, error) {
	var response interface{}
	err := c.DoRequest(ctx, "POST", "/embeddings", req, &response)
	if err != nil {
		return base.EmbeddingResponse{}, err
	}

	// Convert interface{} back to response structure
	responseBytes, _ := json.Marshal(response)
	var embedResponse base.EmbeddingResponse
	json.Unmarshal(responseBytes, &embedResponse)

	return embedResponse, nil
}

// ListModels lists available models with pagination support
func (c *Client) ListModels(ctx context.Context) (*base.PaginatedResponse, error) {
	// For OpenAI, models endpoint returns all models in one request
	paginatedResp, err := c.DoRequestWithPagination(ctx, "GET", "/models", nil, nil)
	if err != nil {
		return nil, err
	}

	// Convert the response
	responseBytes, _ := json.Marshal(paginatedResp.Data)
	var modelsPage []struct {
		ID      string `json:"id"`
		Object  string `json:"object"`
		Created int64  `json:"created"`
		OwnedBy string `json:"owned_by"`
	}
	json.Unmarshal(responseBytes, &modelsPage)

	// Convert to BaseModel interfaces
	models := make([]base.BaseModel, len(modelsPage))
	for i, model := range modelsPage {
		models[i] = model
	}

	// Return as paginated response
	result := &base.PaginatedResponse{
		Data:    models,
		HasMore: false, // OpenAI returns all in one page
		Object:  paginatedResp.Object,
		Page:    1,
		Size:    len(modelsPage),
	}

	return result, nil
}

// GetModel retrieves details for a specific model
func (c *Client) GetModel(ctx context.Context, modelID string) (base.BaseModel, error) {
	// First try to find it in the model list
	listResp, err := c.ListModels(ctx)
	if err != nil {
		return nil, err
	}

	models, ok := listResp.Data.([]base.BaseModel)
	if !ok {
		return nil, fmt.Errorf("models data is not in expected format")
	}

	for _, model := range models {
		if model.GetID() == modelID {
			return model, nil
		}
	}

	return nil, &base.APIError{
		Code:    base.ErrModelNotFound,
		Message: fmt.Sprintf("model not found: %s", modelID),
		Type:    "model_not_found",
	}
}
