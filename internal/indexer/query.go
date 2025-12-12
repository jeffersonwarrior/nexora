package indexer

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

// QueryEngine performs multi-modal search combining semantic, text, and graph queries
type QueryEngine struct {
	storage    *Indexer
	embeddings *EmbeddingEngine
	graph      *Graph
}

// NewQueryEngine creates a new QueryEngine
func NewQueryEngine(storage *Indexer, embeddings *EmbeddingEngine) *QueryEngine {
	return &QueryEngine{
		storage:    storage,
		embeddings: embeddings,
	}
}

// SetGraph sets the graph for graph-based queries
func (qe *QueryEngine) SetGraph(graph *Graph) {
	qe.graph = graph
}

// QueryType specifies the type of search to perform
type QueryType int

const (
	QueryTypeAll QueryType = iota
	QueryTypeSemantic
	QueryTypeText
	QueryTypeGraph
)

// QueryRequest represents a search request
type QueryRequest struct {
	Query       string    `json:"query"`
	Type        QueryType `json:"type"`
	Limit       int       `json:"limit"`
	Context     string    `json:"context,omitempty"`      // package, file, or type context
	SymbolTypes []string  `json:"symbol_types,omitempty"` // filter by symbol types
	IncludeDocs bool      `json:"include_docs"`           // include documentation in search
}

// QueryResult represents a search result
type QueryResult struct {
	Symbol    *Symbol `json:"symbol"`
	Score     float64 `json:"score"`
	MatchType string  `json:"match_type"` // "semantic", "text", "graph", "hybrid"
	Location  string  `json:"location"`
	Reason    string  `json:"reason"` // why this result matched
}

// SearchResult represents the complete search response
type SearchResult struct {
	Query    string        `json:"query"`
	Type     QueryType     `json:"type"`
	Results  []QueryResult `json:"results"`
	Total    int           `json:"total"`
	Duration string        `json:"duration"`
}

// Search performs a multi-modal search
func (qe *QueryEngine) Search(ctx context.Context, req *QueryRequest) (*SearchResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if req.Limit <= 0 {
		req.Limit = 20
	}

	// Start timing
	startTime := time.Now()
	var allResults []QueryResult

	// Perform semantic search
	if req.Type == QueryTypeAll || req.Type == QueryTypeSemantic {
		semanticResults, err := qe.semanticSearch(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("semantic search failed: %w", err)
		}
		allResults = append(allResults, semanticResults...)
	}

	// Perform text search
	if req.Type == QueryTypeAll || req.Type == QueryTypeText {
		textResults, err := qe.textSearch(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("text search failed: %w", err)
		}
		allResults = append(allResults, textResults...)
	}

	// Perform graph search
	if (req.Type == QueryTypeAll || req.Type == QueryTypeGraph) && qe.graph != nil {
		graphResults, err := qe.graphSearch(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("graph search failed: %w", err)
		}
		allResults = append(allResults, graphResults...)
	}

	// Merge and rank results
	finalResults := qe.mergeAndRankResults(allResults, req)

	// Calculate duration
	duration := time.Since(startTime)
	durationStr := fmt.Sprintf("%.2fms", float64(duration.Nanoseconds())/1_000_000)

	return &SearchResult{
		Query:    req.Query,
		Type:     req.Type,
		Results:  finalResults,
		Total:    len(finalResults),
		Duration: durationStr,
	}, nil
}

// semanticSearch performs semantic search using embeddings
func (qe *QueryEngine) semanticSearch(ctx context.Context, req *QueryRequest) ([]QueryResult, error) {
	// Generate embedding for the query
	queryEmbedding, err := qe.embeddings.GenerateEmbedding(ctx, req.Query)
	if err != nil {
		return nil, err
	}

	// Search for similar embeddings
	similarEmbeddings, err := qe.embeddings.SearchSimilar(ctx, req.Query, req.Limit*2)
	if err != nil {
		return nil, err
	}

	var results []QueryResult
	for _, embResult := range similarEmbeddings {
		// Get symbol for this embedding
		symbol, err := qe.storage.GetSymbol(ctx, embResult.ID)
		if err != nil {
			continue
		}
		if symbol == nil {
			continue
		}

		// Apply filters
		if !qe.matchesFilters(symbol, req) {
			continue
		}

		result := QueryResult{
			Symbol:    symbol,
			Score:     float64(qe.embeddings.cosineSimilarity(queryEmbedding, embResult.Vector)),
			MatchType: "semantic",
			Location:  fmt.Sprintf("%s:%d", symbol.File, symbol.Line),
			Reason:    fmt.Sprintf("semantic similarity: %.3f", qe.embeddings.cosineSimilarity(queryEmbedding, embResult.Vector)),
		}

		results = append(results, result)
	}

	return results, nil
}

// getAllSymbolsWithFilters returns all symbols with applied filters
func (qe *QueryEngine) getAllSymbolsWithFilters(ctx context.Context, req *QueryRequest) ([]QueryResult, error) {
	symbols, err := qe.storage.SearchSymbols(ctx, "", req.Limit*2)
	if err != nil {
		return nil, err
	}

	var results []QueryResult
	for i := range symbols {
		symbol := &symbols[i] // Get pointer to the symbol

		// Apply filters
		if !qe.matchesFilters(symbol, req) {
			continue
		}

		result := QueryResult{
			Symbol:    symbol, // Already a pointer
			Score:     1.0,
			MatchType: "text",
			Location:  fmt.Sprintf("%s:%d", symbol.File, symbol.Line),
			Reason:    "exact match",
		}

		results = append(results, result)
	}

	return results, nil
}

// textSearch performs full-text search using SQLite FTS
func (qe *QueryEngine) textSearch(ctx context.Context, req *QueryRequest) ([]QueryResult, error) {
	// Handle empty query case
	if req.Query == "" {
		return qe.getAllSymbolsWithFilters(ctx, req)
	}

	// Build FTS query
	ftsQuery := qe.buildFTSQuery(req.Query, req.IncludeDocs)

	rows, err := qe.storage.db.QueryContext(ctx, `
		SELECT s.name, s.type, s.package, s.file, s.line, s.column, s.signature, s.doc, s.imports, s.callers, s.calls, s.public, s.params, s.returns, s.fields, s.methods
		FROM symbols_fts fts
		JOIN symbols s ON s.id = fts.id
		WHERE symbols_fts MATCH ?
		LIMIT ?`, ftsQuery, req.Limit*2)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []QueryResult
	for rows.Next() {
		var symbol Symbol
		var id string
		var importsJSON, callersJSON, callsJSON, paramsJSON, returnsJSON, fieldsJSON, methodsJSON string
		var createdAt time.Time

		err := rows.Scan(
			&id, &symbol.Name, &symbol.Type, &symbol.Package,
			&symbol.File, &symbol.Line, &symbol.Column, &symbol.Signature,
			&symbol.Doc, &importsJSON, &callersJSON, &callsJSON,
			&symbol.Public, &paramsJSON, &returnsJSON, &fieldsJSON,
			&methodsJSON, &createdAt,
		)
		if err != nil {
			continue
		}

		// Parse JSON fields
		json.Unmarshal([]byte(importsJSON), &symbol.Imports)
		json.Unmarshal([]byte(callersJSON), &symbol.Callers)
		json.Unmarshal([]byte(callsJSON), &symbol.Calls)
		json.Unmarshal([]byte(paramsJSON), &symbol.Params)
		json.Unmarshal([]byte(returnsJSON), &symbol.Returns)
		json.Unmarshal([]byte(fieldsJSON), &symbol.Fields)
		json.Unmarshal([]byte(methodsJSON), &symbol.Methods)

		// Apply filters
		if !qe.matchesFilters(&symbol, req) {
			continue
		}

		result := QueryResult{
			Symbol:    &symbol,
			Score:     1.0, // Default score since we're not using rank
			MatchType: "text",
			Location:  fmt.Sprintf("%s:%d", symbol.File, symbol.Line),
			Reason:    "text match",
		}

		results = append(results, result)
	}

	return results, nil
}

// graphSearch performs graph-based search using relationships
func (qe *QueryEngine) graphSearch(ctx context.Context, req *QueryRequest) ([]QueryResult, error) {
	// This is a simplified graph search - in practice, this would be more sophisticated
	// For now, we'll search for symbols that are related to symbols matching the text query

	// First, find symbols that match the query text
	textResults, err := qe.textSearch(ctx, req)
	if err != nil {
		return nil, err
	}

	// Find related symbols through the graph
	var results []QueryResult
	processed := make(map[string]bool)

	for _, result := range textResults {
		symbolID := result.Symbol.Name
		if processed[symbolID] {
			continue
		}
		processed[symbolID] = true

		// Add the original result
		results = append(results, result)

		// Add related symbols
		if qe.graph != nil {
			related := qe.getRelatedSymbols(symbolID, 2) // depth 2
			for _, relatedID := range related {
				if processed[relatedID] {
					continue
				}
				processed[relatedID] = true

				symbols, err := qe.storage.GetSymbol(ctx, relatedID)
				if err != nil || symbols == nil {
					continue
				}

				symbol := symbols
				if !qe.matchesFilters(symbol, req) {
					continue
				}

				// Calculate a simple graph-based score
				graphScore := qe.calculateGraphScore(symbolID, relatedID)

				graphResult := QueryResult{
					Symbol:    symbol,
					Score:     graphScore,
					MatchType: "graph",
					Location:  fmt.Sprintf("%s:%d", symbol.File, symbol.Line),
					Reason:    fmt.Sprintf("graph relationship to: %s", symbolID),
				}

				results = append(results, graphResult)
			}
		}
	}

	return results, nil
}

// getRelatedSymbols finds symbols related to the given symbol within the graph
func (qe *QueryEngine) getRelatedSymbols(symbolID string, maxDepth int) []string {
	if qe.graph == nil {
		return nil
	}

	related := make([]string, 0)
	visited := make(map[string]bool)

	var traverse func(currentID string, depth int)

	traverse = func(currentID string, depth int) {
		if depth <= 0 || visited[currentID] {
			return
		}
		visited[currentID] = true

		// Get upstream and downstream dependencies
		upstream := qe.graph.GetUpstreamDependencies(currentID)
		downstream := qe.graph.GetDownstreamDependencies(currentID)

		for _, dep := range upstream {
			if !visited[dep] {
				related = append(related, dep)
				traverse(dep, depth-1)
			}
		}

		for _, dep := range downstream {
			if !visited[dep] {
				related = append(related, dep)
				traverse(dep, depth-1)
			}
		}
	}

	traverse(symbolID, maxDepth)
	return related
}

// calculateGraphScore calculates a score for graph relationships
func (qe *QueryEngine) calculateGraphScore(fromID, toID string) float64 {
	if qe.graph == nil {
		return 0.1
	}

	// Simple scoring: stronger relationships have higher scores
	if edges, exists := qe.graph.Edges[fromID]; exists {
		for _, edge := range edges {
			if edge.To == toID {
				return float64(edge.Weight) / 10.0
			}
		}
	}

	return 0.1 // Default low score for indirect relationships
}

// buildFTSQuery constructs a SQLite FTS query from a natural language query
func (qe *QueryEngine) buildFTSQuery(query string, includeDocs bool) string {
	// Sanitize input to prevent SQL injection
	// Replace SQLite FTS special characters with escaped versions
	sanitized := strings.ReplaceAll(query, `"`, `""`)
	sanitized = strings.ReplaceAll(sanitized, `'`, `''`)

	// Remove any other potentially harmful characters
	sanitized = strings.ReplaceAll(sanitized, ";", "")
	sanitized = strings.ReplaceAll(sanitized, "--", "")
	sanitized = strings.ReplaceAll(sanitized, "/*", "")
	sanitized = strings.ReplaceAll(sanitized, "*/", "")

	terms := strings.Fields(sanitized)
	var ftsTerms []string

	for _, term := range terms {
		// Skip empty terms and control characters
		if term == "" || strings.ContainsAny(term, "\x00\x1a\x0d\x0a") {
			continue
		}

		// Quote terms that contain spaces or special characters
		if strings.ContainsAny(term, " -:()[]{}") {
			term = `"` + term + `"`
		}

		if includeDocs && !strings.HasPrefix(term, "-") {
			// Include documentation in search
			ftsTerms = append(ftsTerms, term)
		} else {
			// Search only in name and signature
			ftsTerms = append(ftsTerms, term)
		}
	}

	if len(ftsTerms) == 0 {
		return `""` // Return empty search if no valid terms
	}

	if includeDocs {
		return strings.Join(ftsTerms, " OR ")
	}

	// SQLite FTS syntax for column-specific search - use parameterized approach
	joined := strings.Join(ftsTerms, " OR ")
	return fmt.Sprintf("(name:%s OR signature:%s)", joined, joined)
}

// matchesFilters checks if a symbol matches the query filters
func (qe *QueryEngine) matchesFilters(symbol *Symbol, req *QueryRequest) bool {
	// Filter by symbol type
	if len(req.SymbolTypes) > 0 {
		found := false
		for _, allowedType := range req.SymbolTypes {
			if symbol.Type == allowedType {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Filter by context (package, file)
	if req.Context != "" {
		// Check if context matches package or file
		if !strings.Contains(symbol.Package, req.Context) && !strings.Contains(symbol.File, req.Context) {
			return false
		}
	}

	return true
}

// mergeAndRankResults combines results from different search types and ranks them
func (qe *QueryEngine) mergeAndRankResults(results []QueryResult, req *QueryRequest) []QueryResult {
	// Group by symbol ID
	symbolResults := make(map[string]*QueryResult)

	for _, result := range results {
		symbolID := result.Symbol.Name + "@" + result.Symbol.Package

		if existing, exists := symbolResults[symbolID]; exists {
			// Boost score if found by multiple methods
			combinedScore := existing.Score + result.Score*0.5

			// Update match type to hybrid if we have multiple types
			if existing.MatchType != result.MatchType {
				result.MatchType = "hybrid"
				result.Score = combinedScore
				symbolResults[symbolID] = &result
			} else {
				existing.Score = combinedScore
			}
		} else {
			symbolResults[symbolID] = &result
		}
	}

	// Convert back to slice and sort by score
	var finalResults []QueryResult
	for _, result := range symbolResults {
		finalResults = append(finalResults, *result)
	}

	// Sort by score (descending)
	sort.Slice(finalResults, func(i, j int) bool {
		return finalResults[i].Score > finalResults[j].Score
	})

	// Apply limit
	if len(finalResults) > req.Limit {
		finalResults = finalResults[:req.Limit]
	}

	return finalResults
}

// AdvancedQuery performs more complex queries like "find all functions that X calls"
func (qe *QueryEngine) AdvancedQuery(ctx context.Context, queryPattern string) ([]QueryResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Parse common query patterns
	if strings.HasPrefix(queryPattern, "find all ") {
		return qe.parseFindQuery(ctx, queryPattern)
	} else if strings.Contains(queryPattern, " that ") {
		return qe.parseRelationshipQuery(ctx, queryPattern)
	} else if strings.Contains(queryPattern, " calls ") {
		return qe.parseCallQuery(ctx, queryPattern)
	}

	// Fall back to regular search
	req := &QueryRequest{
		Query: queryPattern,
		Type:  QueryTypeAll,
		Limit: 20,
	}

	result, err := qe.Search(ctx, req)
	if err != nil {
		return nil, err
	}

	return result.Results, nil
}

// parseFindQuery handles queries like "find all functions in package X"
func (qe *QueryEngine) parseFindQuery(ctx context.Context, query string) ([]QueryResult, error) {
	// Simple pattern parsing - in practice this would be more sophisticated
	query = strings.TrimPrefix(query, "find all ")

	var symbolType string
	var context string

	terms := strings.Fields(query)
	for i, term := range terms {
		if term == "functions" {
			symbolType = "function"
		} else if term == "function" {
			symbolType = "function"
		} else if term == "structs" {
			symbolType = "struct"
		} else if term == "interfaces" {
			symbolType = "interface"
		} else if term == "variables" {
			symbolType = "var"
		}
		if term == "in" && i+1 < len(terms) {
			context = terms[i+1]
		}
	}

	req := &QueryRequest{
		Query:       "",
		Type:        QueryTypeText,
		Limit:       50,
		Context:     context,
		SymbolTypes: []string{symbolType},
	}

	result, err := qe.Search(ctx, req)
	if err != nil {
		return nil, err
	}

	return result.Results, nil
}

// parseRelationshipQuery handles queries like "functions that implement interface X"
func (qe *QueryEngine) parseRelationshipQuery(ctx context.Context, query string) ([]QueryResult, error) {
	// This would use the graph to find relationships
	// For now, fall back to text search
	req := &QueryRequest{
		Query: query,
		Type:  QueryTypeAll,
		Limit: 20,
	}

	result, err := qe.Search(ctx, req)
	if err != nil {
		return nil, err
	}

	return result.Results, nil
}

// parseCallQuery handles queries like "function X calls" or "who calls X"
func (qe *QueryEngine) parseCallQuery(ctx context.Context, query string) ([]QueryResult, error) {
	if qe.graph == nil {
		return nil, fmt.Errorf("graph search not available")
	}

	// Simple parsing for call relationships
	var functionName string
	var direction string // "outgoing" or "incoming"

	if strings.Contains(query, " calls ") {
		// Find functions that are called by X
		parts := strings.Split(query, " calls ")
		if len(parts) >= 1 {
			functionName = strings.TrimSpace(parts[0])
			direction = "outgoing"
		}
	} else if strings.Contains(query, " who calls ") {
		// Find functions that call X
		parts := strings.Split(query, " who calls ")
		if len(parts) >= 2 {
			functionName = strings.TrimSpace(parts[1])
			direction = "incoming"
		}
	}

	if functionName == "" {
		return nil, fmt.Errorf("unable to parse call query: %s", query)
	}

	var relatedSymbols []string
	if direction == "outgoing" {
		relatedSymbols = qe.graph.FindCallees(functionName)
	} else {
		relatedSymbols = qe.graph.FindCallers(functionName)
	}

	var results []QueryResult
	for _, symbolID := range relatedSymbols {
		symbol, err := qe.storage.GetSymbol(ctx, symbolID)
		if err != nil || symbol == nil {
			continue
		}

		result := QueryResult{
			Symbol:    symbol,
			Score:     0.8, // Fixed score for graph relationships
			MatchType: "graph",
			Location:  fmt.Sprintf("%s:%d", symbol.File, symbol.Line),
			Reason:    fmt.Sprintf("%s relationship to %s", direction, functionName),
		}

		results = append(results, result)
	}

	return results, nil
}

// ExplainQuery explains why certain results were returned
func (qe *QueryEngine) ExplainQuery(ctx context.Context, query string) (string, error) {
	req := &QueryRequest{
		Query: query,
		Type:  QueryTypeAll,
		Limit: 5,
	}

	result, err := qe.Search(ctx, req)
	if err != nil {
		return "", err
	}

	var explanation strings.Builder
	explanation.WriteString(fmt.Sprintf("Query: '%s'\n", query))
	explanation.WriteString(fmt.Sprintf("Found %d results\n\n", len(result.Results)))

	for i, res := range result.Results {
		explanation.WriteString(fmt.Sprintf("%d. %s (%s) - score: %.3f\n", i+1, res.Symbol.Name, res.Symbol.Type, res.Score))
		explanation.WriteString(fmt.Sprintf("   Location: %s\n", res.Location))
		explanation.WriteString(fmt.Sprintf("   Match type: %s\n", res.MatchType))
		explanation.WriteString(fmt.Sprintf("   Reason: %s\n", res.Reason))

		if res.Symbol.Doc != "" {
			explanation.WriteString(fmt.Sprintf("   Documentation: %s\n", res.Symbol.Doc))
		}
		explanation.WriteString("\n")
	}

	return explanation.String(), nil
}
