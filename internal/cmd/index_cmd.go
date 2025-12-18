package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/nexora/nexora/internal/indexer"
)

// indexCmd represents the index command
var indexCmd = &cobra.Command{
	Use:   "index [path]",
	Short: "Index Go code for AI-accelerated search",
	Long: `Index Go source code to enable semantic, text, and graph-based search.

This command parses Go source files, extracts symbols, generates embeddings,
and builds dependency graphs to enable powerful code search across your codebase.

Examples:
  nexora index .                    # Index current directory
  nexora index ./pkg/               # Index specific directory
  nexora index . --recursive        # Index all subdirectories
  nexora index . --embeddings       # Include semantic embeddings
  nexora index . --output db.sqlite  # Specify output database`,
	Args: cobra.MaximumNArgs(1),
	RunE: runIndex,
}

var (
	indexRecursive    bool
	indexEmbeddings   bool
	indexOutput       string
	indexIncludeTests bool
	indexWorkers      int
)

func init() {
	rootCmd.AddCommand(indexCmd)

	indexCmd.Flags().BoolVarP(&indexRecursive, "recursive", "r", true, "Index directories recursively")
	indexCmd.Flags().BoolVarP(&indexEmbeddings, "embeddings", "e", true, "Generate semantic embeddings")
	indexCmd.Flags().StringVarP(&indexOutput, "output", "o", "nexora_index.db", "Output database path")
	indexCmd.Flags().BoolVarP(&indexIncludeTests, "include-tests", "t", false, "Include test files")
	indexCmd.Flags().IntVarP(&indexWorkers, "workers", "w", 4, "Number of parallel workers")
}

func runIndex(cmd *cobra.Command, args []string) error {
	start := time.Now()

	// Determine path to index
	var path string
	if len(args) == 0 {
		path = "."
	} else {
		path = args[0]
	}

	// Resolve path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve path %s: %w", path, err)
	}

	// Check if path exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", absPath)
	}

	// Create progress reporter
	progress := &IndexProgress{}

	// Index the codebase
	ctx := context.Background()
	err = indexCodebase(ctx, absPath, progress)
	if err != nil {
		return fmt.Errorf("indexing failed: %w", err)
	}

	duration := time.Since(start)
	fmt.Printf("✅ Indexing complete symbols=%d embeddings=%d duration=%s database=%s\n",
		progress.SymbolsFound, progress.EmbeddingsGenerated, duration.Round(time.Millisecond), indexOutput)

	return nil
}

// IndexProgress tracks indexing progress
type IndexProgress struct {
	SymbolsFound        int
	EmbeddingsGenerated int
	FilesProcessed      int
	CurrentFile         string
	DirectoriesIndexed  int
}

func (p *IndexProgress) UpdateStats() {
	// Suppress output during tests
	if flag.Lookup("test.v") == nil {
		fmt.Printf("📊 Progress: files=%d, symbols=%d, embeddings=%d, current=%s\n",
			p.FilesProcessed, p.SymbolsFound, p.EmbeddingsGenerated, filepath.Base(p.CurrentFile))
	}
}

// indexCodebase handles the complete indexing pipeline
func indexCodebase(ctx context.Context, path string, progress *IndexProgress) error {
	// Create output database
	if err := ensureOutputDir(indexOutput); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Initialize indexer
	storage, err := indexer.NewIndexer(indexOutput)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	defer storage.Close()

	// Initialize embedding engine if requested
	var embeddingEngine *indexer.EmbeddingEngine
	if indexEmbeddings {
		// For now, use mock provider - in production would have configurable provider
		provider := indexer.NewLocalProvider("mock", "/tmp")
		embeddingEngine = indexer.NewEmbeddingEngine(provider, storage)
	}

	// Create parser
	parser := indexer.NewASTParser()

	// Find all Go files
	goFiles, err := findGoFiles(path, indexRecursive)
	if err != nil {
		return fmt.Errorf("failed to find Go files: %w", err)
	}

	if len(goFiles) == 0 {
		return fmt.Errorf("no Go files found in %s", path)
	}

	// Process files - use directory parsing since ParseFile isn't exposed
	symbolMap := make(map[string]*indexer.Symbol)
	allSymbols := []indexer.Symbol{}

	// Group files by directory and parse each directory
	dirFiles := make(map[string][]string)
	for _, file := range goFiles {
		dir := filepath.Dir(file)
		dirFiles[dir] = append(dirFiles[dir], file)
	}

	fileCount := 0
	for dir := range dirFiles {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// Parse directory
		symbols, err := parser.ParseDirectory(ctx, dir)
		if err != nil {
			fmt.Printf("Failed to parse directory %s: %v\n", dir, err)
			continue
		}

		// Filter to only include files we want
		for _, symbol := range symbols {
			symbolMap[symbol.Name+"@"+symbol.Package] = &symbol
			allSymbols = append(allSymbols, symbol)
			progress.SymbolsFound++
			fileCount++

			if fileCount%10 == 0 {
				progress.CurrentFile = symbol.File
				progress.FilesProcessed = fileCount
				progress.UpdateStats()
			}
		}
		progress.DirectoriesIndexed++
	}

	progress.FilesProcessed = len(goFiles)
	progress.UpdateStats()

	// Store symbols
	err = storage.StoreSymbols(ctx, allSymbols)
	if err != nil {
		return fmt.Errorf("failed to store symbols: %w", err)
	}

	// Generate and store embeddings if requested
	if indexEmbeddings && embeddingEngine != nil {

		embeddings, err := embeddingEngine.GenerateSymbolEmbeddings(ctx, allSymbols)
		if err != nil {
			return fmt.Errorf("failed to generate embeddings: %w", err)
		}

		progress.EmbeddingsGenerated = len(embeddings)

		err = storage.StoreEmbeddings(ctx, embeddings)
		if err != nil {
			return fmt.Errorf("failed to store embeddings: %w", err)
		}
	}

	// Build and store graph
	graphBuilder := indexer.NewGraphBuilder()
	graph, err := graphBuilder.BuildGraph(ctx, symbolMap)
	if err != nil {
		return fmt.Errorf("failed to build graph: %w", err)
	}

	// For now, just keep graph in memory - could be serialized to JSON later
	_ = graph

	return nil
}

// findGoFiles finds all Go files in the specified path
func findGoFiles(root string, recursive bool) ([]string, error) {
	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip if not a Go file
		if filepath.Ext(path) != ".go" {
			return nil
		}

		// Skip test files unless requested
		if !indexIncludeTests && (strings.Contains(filepath.Base(path), "_test.go")) {
			return nil
		}

		// Skip hidden files and directories
		if strings.HasPrefix(filepath.Base(path), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip vendor directory
		if strings.Contains(path, "vendor"+string(filepath.Separator)) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip .git directory
		if strings.Contains(path, ".git"+string(filepath.Separator)) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		files = append(files, path)
		return nil
	})

	return files, err
}

// ensureOutputDir ensures the output directory exists
func ensureOutputDir(dbPath string) error {
	dir := filepath.Dir(dbPath)
	return os.MkdirAll(dir, 0o755)
}
