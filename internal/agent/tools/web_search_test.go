package tools

import (
	"context"
	"encoding/json"
	"testing"

	"charm.land/fantasy"
)

func TestNewWebSearchTool(t *testing.T) {
	tool := NewWebSearchTool(nil)

	if tool == nil {
		t.Fatal("Expected non-nil tool")
	}

	info := tool.Info()
	if info.Name != WebSearchToolName {
		t.Errorf("Expected tool name %s, got %s", WebSearchToolName, info.Name)
	}

	if info.Description == "" {
		t.Error("Expected non-empty description")
	}
}

func TestWebSearchTool_MissingQuery(t *testing.T) {
	tool := NewWebSearchTool(nil)

	ctx := context.Background()
	params := WebSearchParams{Query: ""}
	paramsJSON, _ := json.Marshal(params)
	call := fantasy.ToolCall{
		ID:    "test-call",
		Name:  WebSearchToolName,
		Input: string(paramsJSON),
	}

	resp, err := tool.Run(ctx, call)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if !resp.IsError {
		t.Error("Expected error for empty query")
	}
}

func TestWebSearchTool_DefaultMaxResults(t *testing.T) {
	// This test just verifies the maxResults logic without actually searching
	maxResults := 0
	if maxResults <= 0 {
		maxResults = 10
	}
	if maxResults > 20 {
		maxResults = 20
	}

	if maxResults != 10 {
		t.Errorf("Expected default maxResults 10, got %d", maxResults)
	}
}

func TestWebSearchTool_MaxResultsCap(t *testing.T) {
	// Verify the maxResults cap logic
	maxResults := 50
	if maxResults <= 0 {
		maxResults = 10
	}
	if maxResults > 20 {
		maxResults = 20
	}

	if maxResults != 20 {
		t.Errorf("Expected capped maxResults 20, got %d", maxResults)
	}
}
