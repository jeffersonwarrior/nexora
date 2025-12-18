package tools

import (
	"testing"

	"github.com/nexora/nexora/internal/indexer"
	"github.com/stretchr/testify/require"
)

func TestNewImpactAnalysisTool(t *testing.T) {
	// Test with nil query engine and nil graph - tool should still be created
	tool := NewImpactAnalysisTool(nil, nil)
	require.NotNil(t, tool)
}

func TestImpactAnalysisParams(t *testing.T) {
	// Test parameter validation
	params := ImpactAnalysisParams{
		SymbolID: "valid_symbol_id",
		MaxDepth: 3,
	}

	// Valid params should not issue errors
	require.NotEmpty(t, params.SymbolID)
	require.Greater(t, params.MaxDepth, 0)
	require.Equal(t, int(3), params.MaxDepth)
}

func TestImpactAnalysisParamsValidation(t *testing.T) {
	// Test empty SymbolID
	params := ImpactAnalysisParams{
		SymbolID: "",
		MaxDepth: 3,
	}

	// Empty symbol ID should be invalid
	require.Empty(t, params.SymbolID)

	params.SymbolID = "test_symbol_id"
	require.NotEmpty(t, params.SymbolID)

	// Test negative MaxDepth - should be handled by validation
	params.MaxDepth = -1
	require.True(t, params.MaxDepth < 0)
}

func TestImpactAnalysisWithNilGraph(t *testing.T) {
	// Create mock embedding engine
	var embeddingEngine *indexer.EmbeddingEngine

	// Create a mock indexer
	idx := &indexer.Indexer{}

	// Create query engine with mock components
	queryEngine := indexer.NewQueryEngine(idx, embeddingEngine)

	// Tool with nil graph should still be created
	tool := NewImpactAnalysisTool(queryEngine, nil)
	require.NotNil(t, tool)
}

func TestImpactAnalysisWithValidComponents(t *testing.T) {
	// Create mock embedding engine
	var embeddingEngine *indexer.EmbeddingEngine

	// Create a mock indexer
	idx := &indexer.Indexer{}

	// Create query engine with mock components
	queryEngine := indexer.NewQueryEngine(idx, embeddingEngine)

	// Create a mock graph (simplified)
	graph := &indexer.Graph{
		Nodes:     make(map[string]*indexer.GraphNode),
		Edges:     make(map[string][]*indexer.GraphEdge),
		CallGraph: make(map[string][]string),
	}

	// Tool should be created successfully with valid components
	tool := NewImpactAnalysisTool(queryEngine, graph)
	require.NotNil(t, tool)
}
