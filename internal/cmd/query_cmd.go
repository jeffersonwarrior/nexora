package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/nexora/nexora/internal/indexer"
)

// queryCmd represents the query command
var queryCmd = &cobra.Command{
	Use:   "query [search terms]",
	Short: "Search indexed code using AI-accelerated search",
	Long: `Search indexed Go code using semantic, text, and graph-based search.

This command enables powerful code search combining:
• Semantic similarity (AI understanding of code intent)  
• Full-text search (exact pattern matching)
• Graph relationships (dependencies and call analysis)

Examples:
  nexora query "functions that handle user authentication"
  nexora query --type semantic "error handling patterns"
  nexora query --type text "func.*Error" 
  nexora query --x main "database connection"
  nexora query --limit 10 "HTTP middleware"
  nexora query --advanced "who calls NewServer"
  nexora query --explain "API validation"`,
	Args: cobra.MinimumNArgs(1),
	RunE: runQuery,
}

var (
	queryType        string
	queryLimit       int
	queryContext     string
	querySymbolTypes []string
	queryIncludeDocs bool
	queryAdvanced    bool
	queryExplain     bool
	queryDatabase    string
)

func init() {
	rootCmd.AddCommand(queryCmd)

	queryCmd.Flags().StringVarP(&queryType, "type", "t", "all",
		"Search type: all, semantic, text, or graph")
	queryCmd.Flags().IntVarP(&queryLimit, "limit", "l", 20,
		"Maximum number of results to return")
	queryCmd.Flags().StringVarP(&queryContext, "context", "x", "",
		"Filter by package, file, or type context")
	queryCmd.Flags().StringSliceVarP(&querySymbolTypes, "symbol-types", "s", []string{},
		"Filter by symbol types: func, struct, interface, var, const")
	queryCmd.Flags().BoolVarP(&queryIncludeDocs, "include-docs", "i", true,
		"Include documentation in search")
	queryCmd.Flags().BoolVarP(&queryAdvanced, "advanced", "a", false,
		"Use advanced query parsing for natural language queries")
	queryCmd.Flags().BoolVarP(&queryExplain, "explain", "e", false,
		"Explain why results were returned")
	queryCmd.Flags().StringVarP(&queryDatabase, "database", "b", "nexora_index.db",
		"Path to index database")
}

func runQuery(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	start := time.Now()

	// Check if database exists
	if _, err := os.Stat(queryDatabase); os.IsNotExist(err) {
		return fmt.Errorf("index database not found: %s\nRun 'nexora index' first to create an index", queryDatabase)
	}

	// Build query string
	query := strings.Join(args, " ")

	// Initialize indexer
	storage, err := indexer.NewIndexer(queryDatabase)
	if err != nil {
		return fmt.Errorf("failed to open index database: %w", err)
	}
	defer storage.Close()

	// Initialize components
	provider := indexer.NewLocalProvider("mock", "/tmp")
	embeddingEngine := indexer.NewEmbeddingEngine(provider, storage)
	queryEngine := indexer.NewQueryEngine(storage, embeddingEngine)

	// Build graph for graph-based search
	symbols, err := storage.SearchSymbols(ctx, "", 10000)
	if err != nil {
		return fmt.Errorf("failed to load symbols: %w", err)
	}

	symbolMap := make(map[string]*indexer.Symbol)
	for i := range symbols {
		s := &symbols[i]
		symbolMap[s.Name+"@"+s.Package] = s
	}

	graphBuilder := indexer.NewGraphBuilder()
	graph, err := graphBuilder.BuildGraph(ctx, symbolMap)
	if err != nil {
		return fmt.Errorf("failed to build graph: %w", err)
	}
	queryEngine.SetGraph(graph)

	fmt.Printf("🔍 Searching index: %s\n", query)

	var results string
	if queryAdvanced {
		// Use advanced query parsing
		queryResults, err := queryEngine.AdvancedQuery(ctx, query)
		if err != nil {
			return fmt.Errorf("advanced query failed: %w", err)
		}
		results = formatAdvancedResults(query, queryResults)
	} else if queryExplain {
		// Use explain mode
		explanation, err := queryEngine.ExplainQuery(ctx, query)
		if err != nil {
			return fmt.Errorf("explain query failed: %w", err)
		}
		results = explanation
	} else {
		// Use standard query
		results, err = performStandardQuery(ctx, queryEngine, query)
		if err != nil {
			return fmt.Errorf("query failed: %w", err)
		}
	}

	// Print results
	fmt.Print(results)

	duration := time.Since(start)
	fmt.Printf("⏱️  Query completed in %s\n", duration.Round(time.Millisecond))

	return nil
}

// performStandardQuery handles regular search queries
func performStandardQuery(ctx context.Context, engine *indexer.QueryEngine, query string) (string, error) {
	// Parse query type
	queryType, err := parseQueryType(queryType)
	if err != nil {
		return "", fmt.Errorf("invalid query type: %w", err)
	}

	// Build request
	req := &indexer.QueryRequest{
		Query:       query,
		Type:        queryType,
		Limit:       queryLimit,
		Context:     queryContext,
		SymbolTypes: querySymbolTypes,
		IncludeDocs: queryIncludeDocs,
	}

	// Execute search
	result, err := engine.Search(ctx, req)
	if err != nil {
		return "", err
	}

	return formatSearchResults(result), nil
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

// formatSearchResults formats search results for display
func formatSearchResults(result *indexer.SearchResult) string {
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
	if queryType == "semantic" || queryType == "all" {
		output.WriteString("💡 **Semantic Search Tips:**\n")
		output.WriteString("   • Try describing what you want in natural language\n")
		output.WriteString("   • Include concepts like \"authentication\", \"validation\", \"error handling\"\n")
		output.WriteString("   • Results are ranked by semantic similarity\n")
	} else if queryType == "text" {
		output.WriteString("💡 **Text Search Tips:**\n")
		output.WriteString("   • Use exact names, patterns, and keywords\n")
		output.WriteString("   • Supports regex patterns for function names\n")
		output.WriteString("   • Results are ranked by text relevance\n")
	} else if queryType == "graph" {
		output.WriteString("💡 **Graph Search Tips:**\n")
		output.WriteString("   • Finds related symbols through dependencies\n")
		output.WriteString("   • Useful for impact analysis and function relationships\n")
		output.WriteString("   • Results include callers and callees\n")
	}

	return output.String()
}

// formatAdvancedResults formats advanced query results
func formatAdvancedResults(query string, results []indexer.QueryResult) string {
	var output strings.Builder

	output.WriteString(fmt.Sprintf("🔍 Advanced Search Results for: %s\n", query))
	output.WriteString(fmt.Sprintf("📊 Found %d results\n\n", len(results)))

	if len(results) == 0 {
		output.WriteString("No results found.\n")
		return output.String()
	}

	// Results
	for i, res := range results {
		output.WriteString(fmt.Sprintf("%d. **%s** `%s`\n", i+1, res.Symbol.Name, res.Symbol.Type))
		output.WriteString(fmt.Sprintf("   📍 Location: %s\n", res.Location))
		output.WriteString(fmt.Sprintf("   🎯 Score: %.3f (%s)\n", res.Score, res.MatchType))
		output.WriteString(fmt.Sprintf("   💡 Reason: %s\n", res.Reason))

		if res.Symbol.Package != "" {
			output.WriteString(fmt.Sprintf("   📦 Package: %s\n", res.Symbol.Package))
		}

		output.WriteString("\n")
	}

	return output.String()
}
