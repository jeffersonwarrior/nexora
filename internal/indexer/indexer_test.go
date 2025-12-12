package indexer

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/ncruces/go-sqlite3"
	_ "github.com/ncruces/go-sqlite3/embed"
	"github.com/stretchr/testify/require"
)

func TestIndexerIntegration(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "nexora-indexer-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test Go files
	testFile := filepath.Join(tempDir, "test.go")
	testContent := `package main

import "fmt"

// TestFunction demonstrates a simple function
func TestFunction(name string) string {
	return "Hello, " + name
}

// AnotherFunction calls TestFunction
func AnotherFunction() {
	TestFunction("world")
}

type TestStruct struct {
	Name string
}

func (t TestStruct) Method() string {
	return TestFunction(t.Name)
}

const TestConstant = "test"

var TestVariable = "variable"
`
	err = os.WriteFile(testFile, []byte(testContent), 0o644)
	require.NoError(t, err)

	// Test AST Parser
	t.Run("AST Parser", func(t *testing.T) {
		ctx := context.Background()
		parser := NewASTParser()
		symbols, err := parser.ParseDirectory(ctx, tempDir)
		require.NoError(t, err)
		require.NotEmpty(t, symbols)

		// Check that we found expected symbols
		symbolNames := make(map[string]bool)
		for _, symbol := range symbols {
			symbolNames[symbol.Name] = true
		}

		require.True(t, symbolNames["TestFunction"], "Should find TestFunction")
		require.True(t, symbolNames["AnotherFunction"], "Should find AnotherFunction")
		require.True(t, symbolNames["TestStruct"], "Should find TestStruct")
		require.True(t, symbolNames["TestConstant"], "Should find TestConstant")
		require.True(t, symbolNames["TestVariable"], "Should find TestVariable")
	})

	// Test Storage
	t.Run("Storage", func(t *testing.T) {
		ctx := context.Background()

		// Create indexer
		dbPath := filepath.Join(tempDir, "test.db")
		indexer, err := NewIndexer(dbPath)
		require.NoError(t, err)
		defer indexer.Close()

		// Store symbols
		ctx = context.Background()
		parser := NewASTParser()
		symbols, err := parser.ParseDirectory(ctx, tempDir)
		require.NoError(t, err)
		require.NotEmpty(t, symbols)

		// Store symbols directly as slice
		err = indexer.StoreSymbols(ctx, symbols)
		require.NoError(t, err)

		// Retrieve symbols
		retrievedSymbols, err := indexer.SearchSymbols(ctx, "", 100)
		require.NoError(t, err)
		require.NotEmpty(t, retrievedSymbols)

		// Check that we can find a specific symbol
		testFunc, err := indexer.GetSymbol(ctx, "TestFunction")
		require.NoError(t, err)
		require.NotNil(t, testFunc)
		require.Equal(t, "TestFunction", testFunc.Name)
		require.Equal(t, "function", testFunc.Type)
		require.Contains(t, testFunc.Doc, "TestFunction demonstrates")
	})

	// Test Embeddings
	t.Run("Embeddings", func(t *testing.T) {
		ctx := context.Background()

		dbPath := filepath.Join(tempDir, "test.db")
		indexer, err := NewIndexer(dbPath)
		require.NoError(t, err)
		defer indexer.Close()

		provider := NewLocalProvider("mock", "/tmp")
		embeddingEngine := NewEmbeddingEngine(provider, indexer)

		// Create test symbols
		symbols := []Symbol{
			{
				Name:      "TestFunction",
				Type:      "function",
				Package:   "main",
				File:      testFile,
				Line:      5,
				Signature: "func TestFunction(name string) string",
				Doc:       "TestFunction demonstrates a simple function",
			},
		}

		// Generate embeddings
		embeddings, err := embeddingEngine.GenerateSymbolEmbeddings(ctx, symbols)
		require.NoError(t, err)
		require.Len(t, embeddings, 1)

		// Check embedding properties
		embedding := embeddings[0]
		require.Equal(t, "TestFunction", embedding.ID)
		require.Equal(t, "function", embedding.Type)
		require.NotEmpty(t, embedding.Vector)
		require.Len(t, embedding.Vector, 384) // Standard embedding size

		// Store embeddings
		err = indexer.StoreEmbeddings(ctx, embeddings)
		require.NoError(t, err)

		// Retrieve embeddings
		retrievedEmbeddings, err := indexer.GetAllEmbeddings(ctx)
		require.NoError(t, err)
		require.Len(t, retrievedEmbeddings, 1)
		require.Equal(t, embedding.ID, retrievedEmbeddings[0].ID)
	})

	// Test Graph
	t.Run("Graph", func(t *testing.T) {
		ctx := context.Background()

		dbPath := filepath.Join(tempDir, "test.db")
		indexer, err := NewIndexer(dbPath)
		require.NoError(t, err)
		defer indexer.Close()

		ctx = context.Background()
		parser := NewASTParser()
		symbols, err := parser.ParseDirectory(ctx, tempDir)
		require.NoError(t, err)

		// Convert to symbol map
		symbolMap := make(map[string]*Symbol)
		for i := range symbols {
			s := &symbols[i]
			symbolMap[s.Name] = s
		}

		// Build graph
		builder := NewGraphBuilder()
		graph, err := builder.BuildGraph(ctx, symbolMap)
		require.NoError(t, err)
		require.NotNil(t, graph)
		require.NotEmpty(t, graph.Nodes)

		// Check that we have call relationships
		anotherFunc, exists := graph.Nodes["AnotherFunction"]
		require.True(t, exists, "Should find AnotherFunction in graph")
		require.Greater(t, anotherFunc.CallCount, 0, "AnotherFunction should call other functions")

		// Test graph traversal
		callees := graph.FindCallees("AnotherFunction")
		require.NotEmpty(t, callees, "AnotherFunction should have callees")

		callers := graph.FindCallers("TestFunction")
		require.NotEmpty(t, callers, "TestFunction should have callers")

		// Test impact analysis
		analysis := graph.GetImpactAnalysis("TestFunction", 2)
		require.Equal(t, "TestFunction", analysis.SymbolID)
		require.NotNil(t, analysis.DirectCalls)
		require.NotNil(t, analysis.DirectCallers)
	})

	// Test Query Engine
	t.Run("Query Engine", func(t *testing.T) {
		ctx := context.Background()

		dbPath := filepath.Join(tempDir, "test.db")
		storage, err := NewIndexer(dbPath)
		require.NoError(t, err)
		defer storage.Close()

		provider := NewLocalProvider("mock", "/tmp")
		embeddingEngine := NewEmbeddingEngine(provider, storage)

		// Setup data
		ctx = context.Background()
		parser := NewASTParser()
		symbols := []Symbol{}
		parsedSymbols, err := parser.ParseDirectory(ctx, tempDir)
		require.NoError(t, err)

		symbolMap := make(map[string]*Symbol)
		for i := range parsedSymbols {
			s := &parsedSymbols[i]
			symbolMap[s.Name] = s
			symbols = append(symbols, *s)
		}

		// Store symbols
		err = storage.StoreSymbols(ctx, symbols)
		require.NoError(t, err)

		// Debug: Check what's in the database
		countSymbols, err := storage.db.QueryContext(ctx, "SELECT COUNT(*) FROM symbols")
		require.NoError(t, err)
		var numSymbols int
		countSymbols.Next()
		countSymbols.Scan(&numSymbols)
		countSymbols.Close()
		t.Logf("Found %d symbols in database", numSymbols)

		countFts, err := storage.db.QueryContext(ctx, "SELECT COUNT(*) FROM symbols_fts")
		require.NoError(t, err)
		var numFts int
		countFts.Next()
		countFts.Scan(&numFts)
		countFts.Close()
		t.Logf("Found %d entries in symbols_fts", numFts)

		embeddings, err := embeddingEngine.GenerateSymbolEmbeddings(ctx, symbols)
		require.NoError(t, err)
		err = storage.StoreEmbeddings(ctx, embeddings)
		require.NoError(t, err)

		builder := NewGraphBuilder()
		graph, err := builder.BuildGraph(ctx, symbolMap)
		require.NoError(t, err)

		// Create query engine
		queryEngine := NewQueryEngine(storage, embeddingEngine)
		queryEngine.SetGraph(graph)

		// Test semantic search
		req := &QueryRequest{
			Query: "function that returns a greeting",
			Type:  QueryTypeSemantic,
			Limit: 5,
		}

		result, err := queryEngine.Search(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, req.Query, result.Query)

		// Test text search
		req.Type = QueryTypeText
		result, err = queryEngine.Search(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Test advanced query
		results, err := queryEngine.AdvancedQuery(ctx, "find all functions")
		require.NoError(t, err)
		require.NotEmpty(t, results)
	})
}

func TestMockEmbedding(t *testing.T) {
	provider := NewLocalProvider("mock", "/tmp")

	ctx := context.Background()
	embedding1, err := provider.GenerateEmbedding(ctx, "test text")
	require.NoError(t, err)
	require.NotEmpty(t, embedding1)
	require.Len(t, embedding1, 384)

	embedding2, err := provider.GenerateEmbedding(ctx, "test text")
	require.NoError(t, err)
	require.NotEmpty(t, embedding2)
	require.Len(t, embedding2, 384)

	// Same input should produce same embedding for mock
	require.Equal(t, embedding1, embedding2)

	// Different input should produce different embedding
	embedding3, err := provider.GenerateEmbedding(ctx, "different text")
	require.NoError(t, err)
	require.NotEqual(t, embedding1, embedding3)
}
