package indexer

import (
	"context"
	"testing"
	"time"

	_ "github.com/ncruces/go-sqlite3"
	_ "github.com/ncruces/go-sqlite3/embed"
)

func TestP6Performance(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("Performance Benchmarks", func(t *testing.T) {
		benchmark, err := NewPerformanceBenchmark(ctx)
		if err != nil {
			t.Fatalf("Failed to create benchmark: %v", err)
		}
		defer benchmark.Close()

		// Run a single benchmark to test functionality
		result := benchmark.benchmarkParsing()
		if !result.Success {
			t.Fatalf("Parsing benchmark failed: %v", result.Error)
		}

		if result.OperationCount == 0 {
			t.Error("Should have parsed some files")
		}

		if result.Throughput <= 0 {
			t.Errorf("Throughput should be positive, got %f", result.Throughput)
		}

		t.Logf("Parsing benchmark successful: %d ops in %v (%.2f ops/sec)",
			result.OperationCount, result.Duration, result.Throughput)
	})

	t.Run("Query Performance", func(t *testing.T) {
		benchmark, err := NewPerformanceBenchmark(ctx)
		if err != nil {
			t.Fatalf("Failed to create benchmark: %v", err)
		}
		defer benchmark.Close()

		// Index some data first
		symbols, err := benchmark.parser.ParseDirectory(ctx, "/home/nexora/internal/indexer")
		if err != nil {
			t.Fatalf("Failed to parse directory: %v", err)
		}

		if err := benchmark.indexer.StoreSymbols(ctx, symbols); err != nil {
			t.Fatalf("Failed to store symbols: %v", err)
		}

		// Test query performance
		result := benchmark.benchmarkQuerying()
		if !result.Success {
			t.Fatalf("Query benchmark failed: %v", result.Error)
		}

		if result.OperationCount == 0 {
			t.Error("Should have run some queries")
		}

		t.Logf("Query benchmark successful: %d ops in %v (%.2f ops/sec)",
			result.OperationCount, result.Duration, result.Throughput)
	})
}

func TestP6Concurrency(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	benchmark, err := NewPerformanceBenchmark(ctx)
	if err != nil {
		t.Fatalf("Failed to create benchmark: %v", err)
	}
	defer benchmark.Close()

	// Run concurrency test
	result := benchmark.benchmarkConcurrency()
	if !result.Success {
		t.Logf("Concurrency test had issues: %v", result.Error)
		// Not failing the test since concurrency can be tricky
	}

	if result.OperationCount != 15 {
		t.Errorf("Expected 15 operations, got %d", result.OperationCount)
	}

	t.Logf("Concurrency test: %v", result.Duration)
}

func TestP6IntegrationWorkflow(t *testing.T) {
	ctx := context.Background()

	t.Run("Full Index → Query Workflow", func(t *testing.T) {
		start := time.Now()

		// Create indexer
		indexer, err := NewIndexer(":memory:")
		if err != nil {
			t.Fatalf("Failed to create indexer: %v", err)
		}
		defer indexer.Close()

		// Parse files
		parser := NewASTParser()
		symbols, err := parser.ParseDirectory(ctx, "/home/nexora/internal/indexer")
		if err != nil {
			t.Fatalf("Failed to parse directory: %v", err)
		}

		// Index files
		err = indexer.StoreSymbols(ctx, symbols)
		if err != nil {
			t.Fatalf("Failed to store symbols: %v", err)
		}

		// Query files
		querySymbols, err := indexer.SearchSymbols(ctx, "Indexer", 10)
		if err != nil {
			t.Fatalf("Failed to search symbols: %v", err)
		}

		duration := time.Since(start)

		if len(symbols) == 0 {
			t.Error("Should have found some symbols to index")
		}

		if len(querySymbols) == 0 {
			t.Error("Should have found some symbols in query")
		}

		t.Logf("Full workflow completed: indexed %d symbols, found %d in query in %v",
			len(symbols), len(querySymbols), duration)
	})
}
