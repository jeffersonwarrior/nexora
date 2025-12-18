package tools

import (
	"testing"

	"github.com/nexora/nexora/internal/indexer"
	"github.com/stretchr/testify/require"
)

func TestNewSearchIndexedTool(t *testing.T) {
	// Test with nil query engine - tool should still be created
	tool := NewSearchIndexedTool(nil)
	require.NotNil(t, tool)
}

func TestSearchParams(t *testing.T) {
	// Test parameter validation
	params := SearchIndexedParams{
		Query: "test query",
		Type:  "invalid_type",
		Limit: 10,
	}

	// Test parseQueryType function indirectly
	_, err := parseQueryType(params.Type)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown query type")
}

func TestParseQueryType(t *testing.T) {
	tests := []struct {
		name        string
		queryType   string
		expected    indexer.QueryType
		expectError bool
	}{
		{"Valid all type", "all", indexer.QueryTypeAll, false},
		{"Valid semantic type", "semantic", indexer.QueryTypeSemantic, false},
		{"Valid text type", "text", indexer.QueryTypeText, false},
		{"Valid graph type", "graph", indexer.QueryTypeGraph, false},
		{"Invalid type", "invalid", indexer.QueryTypeAll, true},
		{"Empty type", "", indexer.QueryTypeAll, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseQueryType(tt.queryType)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, result)
			}
		})
	}
}
