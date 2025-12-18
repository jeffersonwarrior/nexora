package base

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"go.uber.org/ratelimit"
	"golang.org/x/time/rate"
)

// Client is the base LLM API client that all provider clients extend
type Client struct {
	config  ClientConfig
	limiter *rate.Limiter
}

// NewClient creates a new base client with the given configuration
func NewClient(config ClientConfig) *Client {
	// Apply defaults
	if config.UserAgent == "" {
		config.UserAgent = DefaultClientConfig.UserAgent
	}
	if config.Timeout == 0 {
		config.Timeout = DefaultClientConfig.Timeout
	}
	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{
			Timeout: config.Timeout,
		}
	}
	if config.RetryConfig.MaxRetries == 0 {
		config.RetryConfig = DefaultRetryConfig
	}

	// Set up rate limiter if not provided
	if config.RateLimiter == nil {
		// Default: 60 requests per minute
		config.RateLimiter = rate.NewLimiter(rate.Every(time.Minute, 60))
	}

	client := &Client{
		config:  config,
		limiter: config.RateLimiter,
	}

	return client
}

// DoRequest performs an HTTP request with retry logic
func (c *Client) DoRequest(
	ctx context.Context,
	method string,
	endpoint string,
	body interface{},
	result interface{},
) error {
	// Apply rate limiting
	if err := c.limiter.Wait(ctx); err != nil {
		return &APIError{
			Code:    ErrRateLimited,
			Message: fmt.Sprintf("rate limit exceeded: %v", err),
			Type:    "rate_limit_exceeded",
		}
	}

	// Prepare request
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return &APIError{
				Code:    ErrInvalidRequest,
				Message: fmt.Sprintf("failed to marshal request body: %w", err),
				Type:    "json_marshal_error",
			}
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.config.BaseURL+endpoint, reqBody)
	if err != nil {
		return &APIError{
			Code:    ErrProviderError,
			Message: fmt.Sprintf("failed to create request: %w", err),
			Type:    "request_creation_error",
		}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.config.UserAgent)

	// Add authorization header if API key is provided
	if c.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	}

	// Execute with retry logic
	var lastErr error
	for attempt := 0; attempt <= c.config.RetryConfig.MaxRetries; attempt++ {
		resp, err := c.config.HTTPClient.Do(req)
		if err != nil {
			lastErr = &APIError{
				Code:    ErrProviderError,
				Message: fmt.Sprintf("request failed: %w", err),
				Type:    "request_error",
			}

			if !IsRetryableError(lastErr) || attempt == c.config.RetryConfig.MaxRetries {
				break
			}

			// Calculate delay for next attempt
			delay := BackoffDelay(attempt,
				c.config.RetryConfig.BaseDelay,
				c.config.RetryConfig.BackoffFactor,
				c.config.RetryConfig.MaxDelay,
			)
			SleepWithJitter(delay)
			continue
		}

		// Check response status
		if resp.StatusCode < 200 || resp.StatusCode >= 500 {
			// Don't retry for 4xx errors (client errors) or 5xx errors (server errors)
			lastErr = &APIError{
				Code:    ErrProviderError,
				Message: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, resp.Status),
				Type:    "http_error",
			}
			if attempt == c.config.RetryConfig.MaxRetries {
				break
			}

			// Calculate delay and retry
			delay := BackoffDelay(attempt,
				c.config.RetryConfig.BaseDelay,
				c.config.RetryConfig.BackoffFactor,
				c.config.RetryConfig.MaxDelay,
			)
			SleepWithJitter(delay)
			continue
		}

		// Success - parse response
		if result != nil {
			if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
				lastErr = &APIError{
					Code:    ErrProviderError,
					Message: fmt.Sprintf("failed to decode response: %w", err),
					Type:    "json_decode_error",
				}
			}
		}

		// Close body and return
		resp.Body.Close()
		return nil
	}

	return lastErr
}

// DoRequestWithPagination performs a request that may return paginated results
func (c *Client) DoRequestWithPagination(
	ctx context.Context,
	method string,
	endpoint string,
	body interface{},
	result interface{},
) (*PaginatedResponse, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, &APIError{
			Code:    ErrRateLimited,
			Message: fmt.Sprintf("rate limit exceeded: %v", err),
			Type:    "rate_limit_exceeded",
		}
	}

	// Prepare request
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, &APIError{
				Code:    ErrInvalidRequest,
				Message: fmt.Sprintf("failed to marshal request body: %w", err),
				Type:    "json_marshal_error",
			}
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.config.BaseURL+endpoint, reqBody)
	if err != nil {
		return nil, &APIError{
			Code:    ErrProviderError,
			Message: fmt.Sprintf("failed to create request: %w", err),
			Type:    "request_creation_error",
		}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.config.UserAgent)

	if c.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	}

	resp, err := c.config.HTTPClient.Do(req)
	if err != nil {
		return nil, &APIError{
			Code:    ErrProviderError,
			Message: fmt.Sprintf("request failed: %w", err),
			Type:    "request_error",
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, &APIError{
			Code:    ErrProviderError,
			Message: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, resp.Status),
			Type:    "http_error",
			Param:   string(body),
		}
	}

	var paginatedResp PaginatedResponse
	if err := json.NewDecoder(resp.Body).Decode(&paginatedResp); err != nil {
		return nil, &APIError{
			Code:    ErrProviderError,
			Message: fmt.Sprintf("failed to decode paginated response: %w", err),
			Type:    "json_decode_error",
		}
	}

	return &paginatedRespResp, nil
}

// DoRawRequest performs an HTTP request without processing the response body
func (c *Client) DoRawRequest(
	ctx context.Context,
	method string,
	endpoint string,
	body interface{},
) (*http.Response, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, &APIError{
			Code:    ErrRateLimited,
			Message: fmt.Sprintf("rate limit exceeded: %v", err),
			Type:    "RateLimitExceeded",
		}
	}

	// Prepare request
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, &APIError{
				Code:    ErrInvalidRequest,
				Message: fmt.Sprintf("failed to marshal request body: %w", err),
				Type:    "json_marshal_error",
			}
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.config.BaseURL+endpoint, reqBody)
	if err != nil {
		return nil, &APIError{
			Code:    ErrProviderError,
			Message: fmt.Sprintf("failed to request: %w", err),
			Type:    "request_creation_error",
		}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.config.UserAgent)

	if c.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	}

	return c.config.HTTPClient.Do(req)
}

// PaginatedResponse contains paginated response data
type PaginatedResponse struct {
	Data    interface{} `json:"data"`
	HasMore bool        `json:"has_more"`
	Object  string      `json:"object"`
	Page    int         `json:"page"`
	Size    int         `json:"size"`
}

// SimpleHTTPClient is a minimal HTTP client implementation for testing
type SimpleHTTPClient struct {
	*http.Client
}

// NewSimpleHTTPClient creates a simple HTTP client
func NewSimpleHTTPClient() *SimpleHTTPClient {
	return &SimpleHTTPClient{
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// TestHealth tests if the client can reach the API
func (c *Client) TestHealth(ctx context.Context, healthEndpoint string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", healthEndpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	if c.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	}
	req.Header.Set("User-Agent", c.config.UserAgent)

	resp, err := c.config.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status %d", resp.StatusCode)
	}

	return nil
}
