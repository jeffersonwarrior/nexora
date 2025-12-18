package indexer

import (
	"context"
	"fmt"
	"go/token"
	"sort"
	"strings"
	"sync"
	"time"
)

// Graph represents the call graph and dependency relationships between symbols
type Graph struct {
	Nodes       map[string]*GraphNode   `json:"nodes"`
	Edges       map[string][]*GraphEdge `json:"edges"`
	CallGraph   map[string][]string     `json:"call_graph"` // simple representation: node -> list of callers
	Mutex       sync.RWMutex            `json:"-"`          // protects concurrent access
	NodeMetrics map[string]NodeMetrics  `json:"node_metrics"`
}

// NodeMetrics represents computed metrics for graph nodes
type NodeMetrics struct {
	Degree       int       `json:"degree"`
	InDegree     int       `json:"in_degree"`
	OutDegree    int       `json:"out_degree"`
	LastModified time.Time `json:"last_modified"`
}

// GraphNode represents a node in the graph (a symbol)
type GraphNode struct {
	ID         string  `json:"id"`
	Symbol     *Symbol `json:"symbol"`
	NodeType   string  `json:"node_type"` // "function", "struct", "interface", "variable"
	Package    string  `json:"package"`
	File       string  `json:"file"`
	LineNumber int     `json:"line_number"`
	CallCount  int     `json:"call_count"` // how many functions this calls
	CalledBy   int     `json:"called_by"`  // how many functions call this
	Cyclomatic int     `json:"cyclomatic"` // cyclomatic complexity
}

// GraphEdge represents a relationship between nodes
type GraphEdge struct {
	ID       string `json:"id"`
	From     string `json:"from"`
	To       string `json:"to"`
	Type     string `json:"type"`     // "calls", "implements", "embeds", "references", "depends_on"
	Weight   int    `json:"weight"`   // strength of relationship
	Location string `json:"location"` // file:line where this relationship occurs
}

// uniqueKeys returns the keys of a map as a slice
func uniqueKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// GraphBuilder constructs call graphs and dependency graphs from symbols
type GraphBuilder struct {
	fs      *token.FileSet
	symbols map[string]*Symbol
}

// NewGraphBuilder creates a new GraphBuilder
func NewGraphBuilder() *GraphBuilder {
	return &GraphBuilder{
		fs:      token.NewFileSet(),
		symbols: make(map[string]*Symbol),
	}
}

// BuildGraph constructs a graph from a map of symbols
func (gb *GraphBuilder) BuildGraph(ctx context.Context, symbols map[string]*Symbol) (*Graph, error) {
	// Store symbols for use in resolveFunctionCall
	gb.symbols = symbols

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	graph := &Graph{
		Nodes:       make(map[string]*GraphNode),
		Edges:       make(map[string][]*GraphEdge),
		CallGraph:   make(map[string][]string),
		NodeMetrics: make(map[string]NodeMetrics),
	}

	// First pass: create nodes for all symbols
	for id, symbol := range symbols {
		node := &GraphNode{
			ID:         id,
			Symbol:     symbol,
			NodeType:   symbol.Type,
			Package:    symbol.Package,
			File:       symbol.File,
			LineNumber: symbol.Line,
		}

		// Calculate cyclomatic complexity for functions
		if symbol.Type == "func" {
			node.Cyclomatic = gb.calculateCyclomaticComplexity(symbol)
		}

		graph.Nodes[id] = node
	}

	// Second pass: analyze relationships and create edges
	for id, symbol := range symbols {
		// Build call relationships
		if symbol.Type == "function" || symbol.Type == "method" {
			gb.buildCallRelationships(graph, id, symbol)
		}

		// Build structural relationships
		gb.buildStructuralRelationships(graph, id, symbol)

		// Build import relationships
		gb.buildImportRelationships(graph, id, symbol)
	}

	// Build the simplified call graph
	gb.buildCallGraph(graph)

	// Calculate metrics
	gb.calculateNodeMetrics(graph)

	return graph, nil
}

// buildCallRelationships creates edges for function calls
func (gb *GraphBuilder) buildCallRelationships(graph *Graph, callerID string, symbol *Symbol) {
	for _, call := range symbol.Calls {
		// Try to resolve the called function
		calleeID := gb.resolveFunctionCall(call)
		if calleeID == "" {
			continue
		}

		// Check if both nodes exist
		callerNode, callerExists := graph.Nodes[callerID]
		calleeNode, calleeExists := graph.Nodes[calleeID]

		if !callerExists || !calleeExists {
			continue
		}

		// Create edge
		edge := &GraphEdge{
			ID:       fmt.Sprintf("%s->%s", callerID, calleeID),
			From:     callerID,
			To:       calleeID,
			Type:     "calls",
			Weight:   1, // Can be enhanced with call frequency
			Location: fmt.Sprintf("%s:%d", symbol.File, 0),
		}

		graph.Edges[callerID] = append(graph.Edges[callerID], edge)
		callerNode.CallCount++
		calleeNode.CalledBy++
	}
}

// buildStructuralRelationships creates edges for struct/interface relationships
func (gb *GraphBuilder) buildStructuralRelationships(graph *Graph, id string, symbol *Symbol) {
	structType := strings.TrimPrefix(symbol.Type, "struct ")

	switch structType {
	case "struct":
		// Handle embedded fields
		gb.buildEmbeddingRelationships(graph, id, symbol)
	case "interface":
		// Handle method implementations
		gb.buildImplementationRelationships(graph, id, symbol)
	}
}

// buildEmbeddingRelationships handles struct embedding
func (gb *GraphBuilder) buildEmbeddingRelationships(graph *Graph, id string, symbol *Symbol) {
	if symbol.Signature == "" {
		return
	}

	// Parse embedded fields from signature (simplified approach)
	signature := symbol.Signature
	fields := strings.Split(signature, ";")

	for _, field := range fields {
		field = strings.TrimSpace(field)
		if strings.Contains(field, "embedded:") {
			// Extract embedded type
			parts := strings.Split(field, "embedded:")
			if len(parts) < 2 {
				continue
			}
			embeddedType := strings.TrimSpace(parts[1])

			// Try to resolve embedded type
			embeddedID := gb.resolveType(embeddedType, symbol.Package)
			if embeddedID == "" {
				continue
			}

			if _, exists := graph.Nodes[embeddedID]; !exists {
				continue
			}

			edge := &GraphEdge{
				ID:       fmt.Sprintf("%s->%s", id, embeddedID),
				From:     id,
				To:       embeddedID,
				Type:     "embeds",
				Weight:   2, // Structural relationship
				Location: fmt.Sprintf("%s:%d", symbol.File, symbol.Line),
			}

			graph.Edges[id] = append(graph.Edges[id], edge)
		}
	}
}

// buildImplementationRelationships handles interface implementation
func (gb *GraphBuilder) buildImplementationRelationships(graph *Graph, id string, symbol *Symbol) {
	if symbol.Signature == "" {
		return
	}

	// Find structs that implement this interface
	for nodeID, node := range graph.Nodes {
		if node.NodeType != "struct" || nodeID == id {
			continue
		}

		if gb.structImplementsInterface(node.Symbol, symbol) {
			edge := &GraphEdge{
				ID:       fmt.Sprintf("%s->%s", nodeID, id),
				From:     nodeID,
				To:       id,
				Type:     "implements",
				Weight:   3, // Interface implementation is strong relationship
				Location: node.File,
			}

			graph.Edges[nodeID] = append(graph.Edges[nodeID], edge)
		}
	}
}

// buildImportRelationships handles package dependencies
func (gb *GraphBuilder) buildImportRelationships(graph *Graph, id string, symbol *Symbol) {
	for _, pkg := range symbol.Imports {
		// Create relationship to package symbols
		for nodeID, node := range graph.Nodes {
			if nodeID == id {
				continue
			}

			if strings.HasPrefix(node.Package, pkg) {
				edge := &GraphEdge{
					ID:       fmt.Sprintf("%s->%s", id, nodeID),
					From:     id,
					To:       nodeID,
					Type:     "depends_on",
					Weight:   1, // Import dependency
					Location: symbol.File,
				}

				graph.Edges[id] = append(graph.Edges[id], edge)
			}
		}
	}
}

// buildCallGraph creates a simplified call graph representation
func (gb *GraphBuilder) buildCallGraph(graph *Graph) {
	graph.Mutex.Lock()
	defer graph.Mutex.Unlock()

	// Initialize CallGraph if nil
	if graph.CallGraph == nil {
		graph.CallGraph = make(map[string][]string)
	}

	for fromID, edges := range graph.Edges {
		for _, edge := range edges {
			if edge.Type == "calls" {
				graph.CallGraph[fromID] = append(graph.CallGraph[fromID], edge.To)
			}
		}
	}
}

// calculateNodeMetrics computes various metrics for graph nodes
func (gb *GraphBuilder) calculateNodeMetrics(graph *Graph) {
	graph.Mutex.RLock()
	defer graph.Mutex.RUnlock()

	// Calculate degree centrality (sum of incoming and outgoing edges)
	for nodeID := range graph.Nodes {
		metrics := NodeMetrics{
			Degree:       0,
			InDegree:     0,
			OutDegree:    0,
			LastModified: time.Now(),
		}

		// Count outgoing edges
		if edges, exists := graph.Edges[nodeID]; exists {
			metrics.OutDegree = len(edges)
			metrics.Degree += len(edges)
		}

		// Count incoming edges
		for _, edges := range graph.Edges {
			for _, edge := range edges {
				if edge.To == nodeID {
					metrics.InDegree++
					metrics.Degree++
				}
			}
		}

		graph.NodeMetrics[nodeID] = metrics
	}
}

// resolveFunctionCall attempts to resolve a function call to a symbol ID
func (gb *GraphBuilder) resolveFunctionCall(callName string) string {
	// This is a simplified resolver - in practice, this would be more sophisticated
	// using type information and package resolution

	if gb.symbols == nil {
		return ""
	}

	// Clean the call name
	name := strings.TrimSpace(callName)
	if name == "" {
		return ""
	}

	// Try exact match first - return the map key (ID) when symbol name matches
	for id, symbol := range gb.symbols {
		if symbol.Name == name {
			return id
		}
	}

	// Try to match qualified names (e.g., "pkg.Func")
	parts := strings.Split(name, ".")
	if len(parts) > 1 {
		for id, symbol := range gb.symbols {
			if strings.HasSuffix(symbol.Name, "."+parts[len(parts)-1]) {
				return id
			}
		}
	}

	// Return empty string if no match found
	return ""
}

// resolveType attempts to resolve a type name to a symbol ID
func (gb *GraphBuilder) resolveType(typeName, currentPackage string) string {
	// Strip package qualifier if present
	if idx := strings.LastIndex(typeName, "."); idx != -1 {
		typeName = typeName[idx+1:]
	}

	return typeName
}

// structImplementsInterface checks if a struct implements an interface
func (gb *GraphBuilder) structImplementsInterface(structSymbol, interfaceSymbol *Symbol) bool {
	// Simplified implementation - checks for method name overlap
	// In practice, this would use type information to check actual method signatures

	if structSymbol.Type != "struct" || interfaceSymbol.Type != "interface" {
		return false
	}

	// Extract method names from struct signature
	structMethods := gb.extractMethods(structSymbol.Signature)
	interfaceMethods := gb.extractMethods(interfaceSymbol.Signature)

	// Check if all interface methods are present in struct
	for intMethod := range interfaceMethods {
		if _, found := structMethods[intMethod]; !found {
			return false
		}
	}

	return true
}

// extractMethods extracts method names from a signature
func (gb *GraphBuilder) extractMethods(signature string) map[string]bool {
	methods := make(map[string]bool)

	// This is a very simplified parser
	// In practice, this would properly parse the Go AST from the signature
	lines := strings.Split(signature, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "func ") {
			// Extract method name
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				methodName := parts[2]
				// Remove parenthesis from receiver methods
				if idx := strings.Index(methodName, "("); idx != -1 {
					methodName = strings.TrimSpace(methodName[:idx])
				}
				methods[methodName] = true
			}
		}
	}

	return methods
}

// calculateCyclomaticComplexity calculates cyclomatic complexity of a function
func (gb *GraphBuilder) calculateCyclomaticComplexity(symbol *Symbol) int {
	// Simplified complexity calculation from signature
	// In practice, this would parse the function body

	complexity := 1 // Base complexity

	signature := symbol.Signature

	// Count potential complexity indicators
	complexity += strings.Count(signature, "if") * 1
	complexity += strings.Count(signature, "for") * 1
	complexity += strings.Count(signature, "range") * 1
	complexity += strings.Count(signature, "switch") * 1
	complexity += strings.Count(signature, "select") * 1
	complexity += strings.Count(signature, "&&") * 1 // Rough approximation
	complexity += strings.Count(signature, "||") * 1 // Rough approximation

	return int(complexity)
}

// FindCallers returns all functions that call the given function
func (g *Graph) FindCallers(symbolID string) []string {
	var callers []string

	for callerID, callees := range g.CallGraph {
		for _, callee := range callees {
			if callee == symbolID {
				callers = append(callers, callerID)
				break
			}
		}
	}

	return callers
}

// FindCallees returns all functions called by the given function
func (g *Graph) FindCallees(symbolID string) []string {
	return g.CallGraph[symbolID]
}

// GetUpstreamDependencies returns all symbols that this symbol depends on
func (g *Graph) GetUpstreamDependencies(symbolID string) []string {
	var deps []string

	if edges, exists := g.Edges[symbolID]; exists {
		for _, edge := range edges {
			deps = append(deps, edge.To)
		}
	}

	// Remove duplicates
	unique := make(map[string]bool)
	for _, dep := range deps {
		unique[dep] = true
	}

	return uniqueKeys(unique)
}

// GetDownstreamDependencies returns all symbols that depend on this symbol
func (g *Graph) GetDownstreamDependencies(symbolID string) []string {
	var downstream []string

	for sourceID, edges := range g.Edges {
		for _, edge := range edges {
			if edge.To == symbolID {
				downstream = append(downstream, sourceID)
				break
			}
		}
	}

	return downstream
}

// GetImpactAnalysis performs impact analysis for a given symbol
func (g *Graph) GetImpactAnalysis(symbolID string, maxDepth int) ImpactAnalysis {
	directCalls := g.FindCallees(symbolID)
	if directCalls == nil {
		directCalls = []string{}
	}

	directCallers := g.FindCallers(symbolID)
	if directCallers == nil {
		directCallers = []string{}
	}

	analysis := ImpactAnalysis{
		SymbolID:       symbolID,
		DirectCalls:    directCalls,
		DirectCallers:  directCallers,
		DownstreamDeps: g.GetDownstreamDependencies(symbolID),
		UpstreamDeps:   g.GetUpstreamDependencies(symbolID),
	}

	// Calculate transitive dependencies
	if maxDepth > 0 {
		analysis.TransitiveDownstream = g.getTransitiveDependencies(symbolID, "downstream", maxDepth)
		analysis.TransitiveUpstream = g.getTransitiveDependencies(symbolID, "upstream", maxDepth)
	}

	return analysis
}

// getTransitiveDependencies recursively finds all dependencies up to maxDepth
func (g *Graph) getTransitiveDependencies(symbolID, direction string, maxDepth int) []string {
	visited := make(map[string]bool)
	var result []string

	var traverse func(currentID string, depth int)

	traverse = func(currentID string, depth int) {
		if depth <= 0 {
			return
		}

		var next []string
		if direction == "downstream" {
			next = g.GetDownstreamDependencies(currentID)
		} else {
			next = g.GetUpstreamDependencies(currentID)
		}

		for _, nextID := range next {
			if !visited[nextID] && nextID != symbolID {
				visited[nextID] = true
				result = append(result, nextID)
				traverse(nextID, depth-1)
			}
		}
	}

	traverse(symbolID, maxDepth)

	// Sort for consistent output
	sort.Strings(result)

	return result
}

// ImpactAnalysis contains the results of an impact analysis
type ImpactAnalysis struct {
	SymbolID             string   `json:"symbol_id"`
	DirectCalls          []string `json:"direct_calls"`
	DirectCallers        []string `json:"direct_callers"`
	DownstreamDeps       []string `json:"downstream_deps"`
	UpstreamDeps         []string `json:"upstream_deps"`
	TransitiveDownstream []string `json:"transitive_downstream"`
	TransitiveUpstream   []string `json:"transitive_upstream"`
}

// GetCriticalPath finds the critical path(s) through the codebase
func (g *Graph) GetCriticalPath() []string {
	// Simple implementation: find nodes with highest combined metrics
	var critical []string

	type nodeScore struct {
		id    string
		score float64
	}

	var scores []nodeScore

	for id, node := range g.Nodes {
		// Calculate a simple score based on call counts and complexity
		score := float64(node.CalledBy)*2 + float64(node.CallCount) + float64(node.Cyclomatic)*0.5
		scores = append(scores, nodeScore{id: id, score: score})
	}

	// Sort by score
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	// Return top 10%
	topCount := len(scores) / 10
	if topCount < 5 {
		topCount = 5
	}

	for i := 0; i < topCount && i < len(scores); i++ {
		critical = append(critical, scores[i].id)
	}

	return critical
}
