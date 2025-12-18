package tools

import (
	"context"
	"fmt"
	"strings"

	"charm.land/fantasy"
	"github.com/nexora/cli/internal/indexer"
)

type SearchIndexedParams struct {
	Query       string   `json:"query" description:"Search query or question about the codebase"`
	Type        string   `json:"type,omitempty" description:"Search type: all (default), semantic, text, or graph"`
	Limit       int      `json:"limit,omitempty" description:"Maximum number of results to return (default: 20)"`
	Context     string   `json:"context,omitempty" description:"Filter by package, file, or type context"`
	SymbolTypes []string `json:"symbol_types,omitempty" description:"Filter by symbol types: func, struct, interface, var, const"`
	IncludeDocs bool     `json:"include_docs,omitempty" description:"Include documentation in search results"`
}

// NewSearchIndexedTool creates a new search indexed tool
func NewSearchIndexedTool(queryEngine *indexer.QueryEngine) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		"search_indexed",
		"Search code using AI-accelerated indexing with semantic, text, and graph search capabilities",
		func(ctx context.Context, params SearchIndexedParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if queryEngine == nil {
				return fantasy.NewTextErrorResponse("search indexed tool is not available - indexer not initialized"), nil
			}

			if params.Query == "" {
				return fantasy.NewTextErrorResponse("query parameter is required"), nil
			}

			// Set defaults
			if params.Type == "" {
				params.Type = "all"
			}
			if params.Limit == 0 {
				params.Limit = 20
			}

			// Convert to indexer QueryRequest
			queryType, err := parseQueryType(params.Type)
			if err != nil {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("Error: %v", err)), nil
			}

			req := &indexer.QueryRequest{
				Query:       params.Query,
				Type:        queryType,
				Limit:       params.Limit,
				Context:     params.Context,
				SymbolTypes: params.SymbolTypes,
				IncludeDocs: params.IncludeDocs,
			}

			// Execute search
			result, err := queryEngine.Search(ctx, req)
			if err != nil {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("Search failed: %v", err)), nil
			}

			// Format results
			output := formatResults(result, params)
			return fantasy.NewTextResponse(output), nil
		},
	)
}

// parseQueryType converts string query type to indexer.QueryType
func parseQueryType(typeStr string) (indexer.QueryType, error) {
	switch strings.ToLower(typeStr) {
	case "all":
		return indexer.QueryTypeAll, nil
	case "semantic":
		return indexer.QueryTypeSemantic, nil
	case "text":
		return indexer.QueryTypeText, nil
	case "graph":
		return indexer.QueryTypeGraph, nil
	default:
		return indexer.QueryTypeAll, fmt.Errorf("unknown query type: %s", typeStr)
	}
}

// formatResults formats search results for display
func formatResults(result *indexer.SearchResult, params SearchIndexedParams) string {
	var output strings.Builder

	// Header
	output.WriteString(fmt.Sprintf("🔍 Search Results for: %s\n", result.Query))
	output.WriteString(fmt.Sprintf("📊 Type: %v | Total: %d results | Duration: %s\n\n",
		result.Type, result.Total, result.Duration))

	if len(result.Results) == 0 {
		output.WriteString("No results found.\n")
		return output.String()
	}

	// Results
	for i, res := range result.Results {
		output.WriteString(fmt.Sprintf("%d. **%s** `%s`\n", i+1, res.Symbol.Name, res.Symbol.Type))
		output.WriteString(fmt.Sprintf("   📍 Location: %s\n", res.Location))
		output.WriteString(fmt.Sprintf("   🎯 Score: %.3f (%s)\n", res.Score, res.MatchType))
		output.WriteString(fmt.Sprintf("   💡 Reason: %s\n", res.Reason))

		// Package context
		if res.Symbol.Package != "" {
			output.WriteString(fmt.Sprintf("   📦 Package: %s\n", res.Symbol.Package))
		}

		// Documentation (if available and not too long)
		if res.Symbol.Doc != "" && len(res.Symbol.Doc) < 200 {
			output.WriteString(fmt.Sprintf("   📖 Documentation: %s\n", res.Symbol.Doc))
		}

		// Signature (if available and not too long)
		if res.Symbol.Signature != "" && len(res.Symbol.Signature) < 200 {
			output.WriteString(fmt.Sprintf("   🔍 Signature: %s\n", res.Symbol.Signature))
		} else if res.Symbol.Signature != "" {
			// truncated signature for very long ones
			truncated := res.Symbol.Signature
			if len(truncated) > 200 {
				truncated = truncated[:197] + "..."
			}
			output.WriteString(fmt.Sprintf("   🔍 Signature: %s\n", truncated))
		}

		output.WriteString("\n")
	}

	// Footer with tips
	output.WriteString("💡 **Tips:**\n")
	output.WriteString("   • Use semantic search for finding code by intent\n")
	output.WriteString("   • Use text search for exact name or pattern matching\n")
	output.WriteString("   • Use graph search for finding related or dependent code\n")
	output.WriteString("   • Filter results with `type:`, `context:`, or `symbol_types:` parameters\n")

	return output.String()
}
