package indexer

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	_ "github.com/ncruces/go-sqlite3"
)

// PerformanceBenchmark runs performance tests on the indexer
type PerformanceBenchmark struct {
	ctx        context.Context
	indexer    *Indexer
	parser     *ASTParser
	embeddings *EmbeddingEngine
}

// NewPerformanceBenchmark creates a new performance benchmark
func NewPerformanceBenchmark(ctx context.Context) (*PerformanceBenchmark, error) {
	// Create in-memory indexer for testing
	indexer, err := NewIndexer(":memory:")
	if err != nil {
		return nil, fmt.Errorf("failed to create indexer: %w", err)
	}

	parser := NewASTParser()
	provider := NewLocalProvider("mock", "")
	embeddings := NewEmbeddingEngine(provider, indexer)

	return &PerformanceBenchmark{
		ctx:        ctx,
		indexer:    indexer,
		parser:     parser,
		embeddings: embeddings,
	}, nil
}

// BenchmarkResult holds performance benchmark results
type BenchmarkResult struct {
	Name           string        `json:"name"`
	OperationCount int           `json:"operation_count"`
	Duration       time.Duration `json:"duration"`
	Throughput     float64       `json:"throughput"` // Operations per second
	MemoryUsage    int64         `json:"memory_usage_bytes"`
	Success        bool          `json:"success"`
	Error          string        `json:"error,omitempty"`
}

// BenchmarkSuite holds multiple benchmark results
type BenchmarkSuite struct {
	Benchmarks []BenchmarkResult `json:"benchmarks"`
	TotalTime  time.Duration     `json:"total_time"`
	Completed  time.Time         `json:"completed_at"`
}

// RunAllBenchmarks executes the complete performance test suite
func (pb *PerformanceBenchmark) RunAllBenchmarks() *BenchmarkSuite {
	slog.Info("Starting performance benchmark suite")

	var results []BenchmarkResult
	start := time.Now()

	// Benchmark 1: Parsing performance
	results = append(results, pb.benchmarkParsing())

	// Benchmark 2: Indexing performance
	results = append(results, pb.benchmarkIndexing())

	// Benchmark 3: Query performance
	results = append(results, pb.benchmarkQuerying())

	// Benchmark 4: Embedding performance
	results = append(results, pb.benchmarkEmbeddings())

	// Benchmark 5: Concurrency performance
	results = append(results, pb.benchmarkConcurrency())

	suite := &BenchmarkSuite{
		Benchmarks: results,
		TotalTime:  time.Since(start),
		Completed:  time.Now(),
	}

	slog.Info("Performance benchmark suite completed",
		"benchmarks", len(results),
		"total_duration", suite.TotalTime)

	return suite
}

// benchmarkParsing tests AST parsing speed
func (pb *PerformanceBenchmark) benchmarkParsing() BenchmarkResult {
	start := time.Now()

	// Parse the indexer directory multiple times
	iterations := 10
	totalFiles := 0

	for range iterations {
		symbols, err := pb.parser.ParseDirectory(pb.ctx, "/home/nexora/internal/indexer")
		if err != nil {
			return BenchmarkResult{
				Name:     "Parser_AST",
				Error:    fmt.Sprintf("Parse failed: %v", err),
				Duration: time.Since(start),
				Success:  false,
			}
		}
		totalFiles += len(symbols)
	}

	duration := time.Since(start)
	throughput := float64(totalFiles) / duration.Seconds()

	return BenchmarkResult{
		Name:           "Parser_AST",
		OperationCount: totalFiles,
		Duration:       duration,
		Throughput:     throughput,
		Success:        true,
	}
}

// benchmarkIndexing tests symbol storage speed
func (pb *PerformanceBenchmark) benchmarkIndexing() BenchmarkResult {
	start := time.Now()

	// Parse symbols once
	symbols, err := pb.parser.ParseDirectory(pb.ctx, "/home/nexora/internal/indexer")
	if err != nil {
		return BenchmarkResult{
			Name:     "Indexing_Storage",
			Error:    fmt.Sprintf("Parse failed: %v", err),
			Duration: time.Since(start),
			Success:  false,
		}
	}

	// Store symbols multiple times
	iterations := 5
	for range iterations {
		err := pb.indexer.StoreSymbols(pb.ctx, symbols)
		if err != nil {
			return BenchmarkResult{
				Name:     "Indexing_Storage",
				Error:    fmt.Sprintf("Store failed: %v", err),
				Duration: time.Since(start),
				Success:  false,
			}
		}
	}

	duration := time.Since(start)
	throughput := float64(len(symbols)*iterations) / duration.Seconds()

	return BenchmarkResult{
		Name:           "Indexing_Storage",
		OperationCount: len(symbols) * iterations,
		Duration:       duration,
		Throughput:     throughput,
		Success:        true,
	}
}

// benchmarkQuerying tests search speed
func (pb *PerformanceBenchmark) benchmarkQuerying() BenchmarkResult {
	start := time.Now()

	// Index some data first
	symbols, err := pb.parser.ParseDirectory(pb.ctx, "/home/nexora/internal/indexer")
	if err != nil {
		return BenchmarkResult{
			Name:     "Query_Search",
			Error:    fmt.Sprintf("Parse failed: %v", err),
			Duration: time.Since(start),
			Success:  false,
		}
	}

	if err := pb.indexer.StoreSymbols(pb.ctx, symbols); err != nil {
		return BenchmarkResult{
			Name:     "Query_Search",
			Error:    fmt.Sprintf("Store failed: %v", err),
			Duration: time.Since(start),
			Success:  false,
		}
	}

	// Perform many searches
	queryStart := time.Now()
	iterations := 100
	for i := range iterations {
		_, err := pb.indexer.SearchSymbols(pb.ctx, fmt.Sprintf("test_query_%d", i), 10)
		if err != nil {
			return BenchmarkResult{
				Name:     "Query_Search",
				Error:    fmt.Sprintf("Query failed: %v", err),
				Duration: time.Since(start),
				Success:  false,
			}
		}
	}

	duration := time.Since(queryStart)
	throughput := float64(iterations) / duration.Seconds()

	return BenchmarkResult{
		Name:           "Query_Search",
		OperationCount: iterations,
		Duration:       duration,
		Throughput:     throughput,
		Success:        true,
	}
}

// benchmarkEmbeddings tests embedding generation speed
func (pb *PerformanceBenchmark) benchmarkEmbeddings() BenchmarkResult {
	start := time.Now()

	// Get some symbols
	symbols, err := pb.parser.ParseDirectory(pb.ctx, "/home/nexora/internal/indexer")
	if err != nil {
		return BenchmarkResult{
			Name:     "Embeddings_Generation",
			Error:    fmt.Sprintf("Parse failed: %v", err),
			Duration: time.Since(start),
			Success:  false,
		}
	}

	// Limit symbols for testing
	if len(symbols) > 50 {
		symbols = symbols[:50]
	}

	// Generate embeddings
	embeddings, err := pb.embeddings.GenerateSymbolEmbeddings(pb.ctx, symbols)
	if err != nil {
		return BenchmarkResult{
			Name:     "Embeddings_Generation",
			Error:    fmt.Sprintf("Embedding failed: %v", err),
			Duration: time.Since(start),
			Success:  false,
		}
	}

	duration := time.Since(start)
	throughput := float64(len(embeddings)) / duration.Seconds()

	return BenchmarkResult{
		Name:           "Embeddings_Generation",
		OperationCount: len(embeddings),
		Duration:       duration,
		Throughput:     throughput,
		Success:        true,
	}
}

// benchmarkConcurrency tests concurrent operations
func (pb *PerformanceBenchmark) benchmarkConcurrency() BenchmarkResult {
	start := time.Now()

	// Index some data first
	symbols, err := pb.parser.ParseDirectory(pb.ctx, "/home/nexora/internal/indexer")
	if err != nil {
		return BenchmarkResult{
			Name:     "Concurrency_Operations",
			Error:    fmt.Sprintf("Parse failed: %v", err),
			Duration: time.Since(start),
			Success:  false,
		}
	}

	if err := pb.indexer.StoreSymbols(pb.ctx, symbols); err != nil {
		return BenchmarkResult{
			Name:     "Concurrency_Operations",
			Error:    fmt.Sprintf("Store failed: %v", err),
			Duration: time.Since(start),
			Success:  false,
		}
	}

	// Test concurrent operations
	var wg sync.WaitGroup
	errors := make(chan error, 20)

	// Concurrent writes
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			err := pb.indexer.StoreSymbols(pb.ctx, symbols[:10])
			if err != nil {
				errors <- fmt.Errorf("concurrent write %d failed: %w", id, err)
			}
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_, err := pb.indexer.SearchSymbols(pb.ctx, fmt.Sprintf("query_%d", id), 5)
			if err != nil {
				errors <- fmt.Errorf("concurrent read %d failed: %w", id, err)
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Count errors
	errorCount := 0
	for err := range errors {
		errorCount++
		slog.Warn("Concurrency error", "error", err)
	}

	duration := time.Since(start)
	throughput := float64(15) / duration.Seconds() // 5 writes + 10 reads

	success := errorCount == 0

	return BenchmarkResult{
		Name:           "Concurrency_Operations",
		OperationCount: 15,
		Duration:       duration,
		Throughput:     throughput,
		Success:        success,
		Error:          fmt.Sprintf("%d errors occurred", errorCount),
	}
}

// PrintReport formats and prints a performance report
func (suite *BenchmarkSuite) PrintReport() {
	fmt.Println("\n⚡ PERFORMANCE BENCHMARK REPORT")
	fmt.Println("==============================")
	fmt.Printf("Total Benchmark Time: %v\n", suite.TotalTime)
	fmt.Printf("Completed At: %s\n", suite.Completed.Format(time.RFC3339))
	fmt.Printf("Benchmarks Run: %d\n", len(suite.Benchmarks))

	fmt.Println("\nDetailed Results:")
	for _, result := range suite.Benchmarks {
		status := "✅ PASS"
		if !result.Success {
			status = "❌ FAIL"
		}

		fmt.Printf("  %s %s\n", status, result.Name)
		fmt.Printf("    Operations: %d\n", result.OperationCount)
		fmt.Printf("    Duration: %v\n", result.Duration)
		fmt.Printf("    Throughput: %.2f ops/sec\n", result.Throughput)

		if result.Error != "" {
			fmt.Printf("    Error: %s\n", result.Error)
		}
		fmt.Println()
	}

	// Performance summary
	var avgThroughput float64
	var successfulTests int

	for _, result := range suite.Benchmarks {
		if result.Success {
			successfulTests++
			avgThroughput += result.Throughput
		}
	}

	if successfulTests > 0 {
		avgThroughput /= float64(successfulTests)
	}

	fmt.Println("📊 Performance Summary:")
	fmt.Printf("  Success Rate: %d/%d (%.1f%%)\n",
		successfulTests, len(suite.Benchmarks),
		float64(successfulTests)/float64(len(suite.Benchmarks))*100)
	fmt.Printf("  Average Throughput: %.2f ops/sec\n", avgThroughput)
}

// Close cleans up resources
func (pb *PerformanceBenchmark) Close() {
	if pb.indexer != nil {
		pb.indexer.Close()
	}
}

// RunPerformanceTests is the main entry point for performance testing
func RunPerformanceTests(ctx context.Context) error {
	benchmark, err := NewPerformanceBenchmark(ctx)
	if err != nil {
		return fmt.Errorf("failed to create benchmark: %w", err)
	}
	defer benchmark.Close()

	// Run benchmarks
	suite := benchmark.RunAllBenchmarks()

	// Print report
	suite.PrintReport()

	// Return error if any benchmarks failed
	for _, result := range suite.Benchmarks {
		if !result.Success {
			return fmt.Errorf("performance benchmarks failed")
		}
	}

	return nil
}
