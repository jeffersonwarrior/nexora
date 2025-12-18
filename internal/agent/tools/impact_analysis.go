package tools

import (
	"context"
	"fmt"
	"strings"

	"charm.land/fantasy"
	"github.com/nexora/nexora/internal/indexer"
)

type ImpactAnalysisParams struct {
	SymbolID string `json:"symbol_id" description:"ID of the symbol to analyze"`
	MaxDepth int    `json:"max_depth,omitempty" description:"Maximum depth for transitive analysis (default: 3)"`
}

// NewImpactAnalysisTool creates a new impact analysis tool
func NewImpactAnalysisTool(queryEngine *indexer.QueryEngine, graph *indexer.Graph) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		"impact_analysis",
		"Analyze the impact of changes to a symbol using dependency graphs and call relationships",
		func(ctx context.Context, params ImpactAnalysisParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if graph == nil {
				return fantasy.NewTextErrorResponse("impact analysis tool is not available - graph not initialized"), nil
			}

			if params.SymbolID == "" {
				return fantasy.NewTextErrorResponse("symbol_id parameter is required"), nil
			}

			// Set defaults
			if params.MaxDepth == 0 {
				params.MaxDepth = 3
			}

			// Get the symbol information
			symbol, err := getSymbolInfo(ctx, params.SymbolID)
			if err != nil {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("Failed to get symbol info: %v", err)), nil
			}
			if symbol == nil {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("Symbol '%s' not found", params.SymbolID)), nil
			}

			// Perform impact analysis
			analysis := graph.GetImpactAnalysis(params.SymbolID, params.MaxDepth)

			// Format results
			output := formatImpactAnalysis(symbol, analysis, params.MaxDepth)
			return fantasy.NewTextResponse(output), nil
		},
	)
}

// getSymbolInfo retrieves symbol information from storage
func getSymbolInfo(ctx context.Context, symbolID string) (*indexer.Symbol, error) {
	// This is a simplified implementation - in practice, this would use the storage layer
	// For now, we'll return a mock symbol structure
	return &indexer.Symbol{
		Name:    symbolID,
		Type:    "function",
		Package: "main",
		File:    fmt.Sprintf("%s.go", symbolID),
		Line:    1,
	}, nil
}

// formatImpactAnalysis formats the impact analysis results for display
func formatImpactAnalysis(symbol *indexer.Symbol, analysis indexer.ImpactAnalysis, maxDepth int) string {
	var output strings.Builder

	// Header
	output.WriteString(fmt.Sprintf("🔬 Impact Analysis for: **%s** `%s`\n", symbol.Name, symbol.Type))
	output.WriteString(fmt.Sprintf("📍 Location: %s:%d\n\n", symbol.File, symbol.Line))

	// Analysis depth
	output.WriteString(fmt.Sprintf("🔍 Analysis depth: %d levels\n\n", maxDepth))

	// Direct relationships
	output.WriteString("## Direct Relationships\n\n")

	if len(analysis.DirectCalls) > 0 {
		output.WriteString("### 📞 Functions Called (Downstream)\n")
		for i, call := range analysis.DirectCalls {
			output.WriteString(fmt.Sprintf("%d. `%s`\n", i+1, call))
		}
		output.WriteString("\n")
	}

	if len(analysis.DirectCallers) > 0 {
		output.WriteString("### 📞 Functions Called By (Upstream)\n")
		for i, caller := range analysis.DirectCallers {
			output.WriteString(fmt.Sprintf("%d. `%s`\n", i+1, caller))
		}
		output.WriteString("\n")
	}

	// Dependencies
	output.WriteString("## Dependencies\n\n")

	if len(analysis.UpstreamDeps) > 0 {
		output.WriteString("### ⬆️ Upstream Dependencies\n")
		output.WriteString("Functions/packages this symbol depends on:\n")
		for i, dep := range analysis.UpstreamDeps {
			output.WriteString(fmt.Sprintf("%d. `%s`\n", i+1, dep))
		}
		output.WriteString("\n")
	}

	if len(analysis.DownstreamDeps) > 0 {
		output.WriteString("### ⬇️ Downstream Dependencies\n")
		output.WriteString("Functions/packages that depend on this symbol:\n")
		for i, dep := range analysis.DownstreamDeps {
			output.WriteString(fmt.Sprintf("%d. `%s`\n", i+1, dep))
		}
		output.WriteString("\n")
	}

	// Transitive analysis
	output.WriteString("## Transitive Impact\n\n")

	if len(analysis.TransitiveDownstream) > 0 {
		output.WriteString(fmt.Sprintf("### 🌊 Transitive Downstream Impact (%d levels)\n", maxDepth))
		output.WriteString("All functions that might be affected by changes to this symbol:\n")
		for i, dep := range analysis.TransitiveDownstream {
			output.WriteString(fmt.Sprintf("%d. `", i+1))
			// Show hierarchy with indentation for transitive relationships
			output.WriteString(fmt.Sprintf("%s`\n", dep))
		}
		output.WriteString("\n")
	}

	if len(analysis.TransitiveUpstream) > 0 {
		output.WriteString(fmt.Sprintf("### ⬆️ Transitive Upstream Impact (%d levels)\n", maxDepth))
		output.WriteString("All functions that could affect this symbol:\n")
		for i, dep := range analysis.TransitiveUpstream {
			output.WriteString(fmt.Sprintf("%d. `%s`\n", i+1, dep))
		}
		output.WriteString("\n")
	}

	// Risk assessment
	output.WriteString("## 🎯 Risk Assessment\n\n")

	riskLevel := assessRisk(analysis)
	output.WriteString(fmt.Sprintf("**Risk Level:** %s\n\n", riskLevel))

	// Test recommendations
	output.WriteString("## 🧪 Test Recommendations\n\n")
	output.WriteString("Based on this analysis, you should test:\n\n")

	if len(analysis.DownstreamDeps) > 0 {
		output.WriteString("• **Downstream consumers** - functions that call this symbol\n")
	}
	if len(analysis.UpstreamDeps) > 0 {
		output.WriteString("• **Upstream dependencies** - functions this symbol calls\n")
	}
	if len(analysis.TransitiveDownstream) > 0 {
		output.WriteString("• **Transitive impact** - functions indirectly affected\n")
	}

	// Add specific testing suggestions
	if len(analysis.DirectCallers) > 10 {
		output.WriteString("⚠️  High impact - consider integration tests\n")
	}
	if len(analysis.DirectCalls) > 20 {
		output.WriteString("⚠️  Complex dependencies - unit test each interface\n")
	}

	output.WriteString("\n")

	// Footer with tips
	output.WriteString("💡 **Tips:**\n")
	output.WriteString("   • Run integration tests on all downstream dependencies\n")
	output.WriteString("   • Check for breaking changes in function signatures\n")
	output.WriteString("   • Consider version compatibility for public APIs\n")
	output.WriteString("   • Mock upstream dependencies during testing\n")

	return output.String()
}

// assessRisk provides a simple risk assessment based on the impact analysis
func assessRisk(analysis indexer.ImpactAnalysis) string {
	totalImpact := len(analysis.DirectCallers) +
		len(analysis.DownstreamDeps) +
		len(analysis.TransitiveDownstream)

	if totalImpact == 0 {
		return "🟢 Low (no dependencies)"
	} else if totalImpact < 5 {
		return "🟡 Medium (limited impact)"
	} else if totalImpact < 20 {
		return "🟠 High (significant impact)"
	} else {
		return "🔴 Critical (widespread impact)"
	}
}
