package tools

import (
	"strings"

	"github.com/nexora/nexora/internal/agent/utils"
)

// Tool aliases map alternative names to canonical tool names
// Key: alias (case-insensitive), Value: canonical tool name
//
// Canonical tools (not aliased):
// - bash: Execute shell commands
// - download: Download files
// - find: Find files by name or pattern
// - glob: Match files by glob pattern
// - job_kill: Kill background jobs
// - job_output: Get background job output
// - multiedit: Edit multiple files
// - smart_edit: Smart editing with AI
// - agentic_fetch: Fetch with agentic retry logic
// - lsp_diagnostics: Language server protocol diagnostics
// - lsp_references: Language server protocol references
var toolAliases = map[string]string{
	// Fetch tool aliases
	"curl":          "fetch",
	"wget":          "fetch",
	"get":           "fetch",
	"http-get":      "fetch",
	"http_get":      "fetch",
	"web-fetch":     "web_fetch",
	"webfetch":      "web_fetch",
	"http":          "fetch",

	// File tools
	"read":          "view",
	"cat":           "view",
	"open":          "view",
	"list":          "ls",
	"dir":           "ls",
	"directory":     "ls",

	// Edit tools
	"modify":        "edit",
	"change":        "edit",
	"replace":       "edit",
	"update":        "edit",

	// Write tools
	"create":        "write",
	"make":          "write",
	"new":           "write",

	// Search tools
	"search":        "grep",
	"rg":            "grep",

	// Bash/shell tools
	"shell":         "bash",
	"exec":          "bash",
	"execute":       "bash",
	"run":           "bash",
	"command":       "bash",

	// Web tools
	"web-search":    "web_search",
	"websearch":     "web_search",
	"search-web":    "web_search",

	// Sourcegraph
	"sg":            "sourcegraph",
	"code-search":   "sourcegraph",
}

// ResolveToolName resolves a tool name alias to its canonical name
// Returns the canonical name if an alias is found, otherwise returns the original name
// Also sanitizes tool names to remove XML/JSON artifacts that some models leak
func ResolveToolName(name string) string {
	if name == "" {
		return name
	}

	// Sanitize first to handle models that leak serialization format
	// Example: "grep_path_pattern</arg_key><arg_value>..." -> "grep_path_pattern"
	name = utils.SanitizeToolName(name)

	// Try exact match (case-insensitive)
	if canonical, ok := toolAliases[strings.ToLower(name)]; ok {
		return canonical
	}

	// No alias found, return sanitized name
	return name
}

// IsAlias checks if a name is an alias for another tool
func IsAlias(name string) bool {
	_, ok := toolAliases[strings.ToLower(name)]
	return ok
}

// GetCanonicalName returns the canonical name for a tool or alias
// This is an alias for ResolveToolName for readability
func GetCanonicalName(name string) string {
	return ResolveToolName(name)
}

// AddAlias adds a new alias to the alias map
// This allows dynamic alias registration at runtime
func AddAlias(alias, canonical string) {
	if alias != "" && canonical != "" {
		toolAliases[strings.ToLower(alias)] = canonical
	}
}

// RemoveAlias removes an alias from the alias map
func RemoveAlias(alias string) {
	delete(toolAliases, strings.ToLower(alias))
}

// ListAliases returns all aliases for a given canonical tool name
func ListAliases(canonical string) []string {
	var aliases []string
	for alias, canon := range toolAliases {
		if canon == canonical {
			aliases = append(aliases, alias)
		}
	}
	return aliases
}

// ToolNames returns all registered tool names (both aliases and canonical)
func ToolNames() []string {
	names := make([]string, 0, len(toolAliases))
	for name := range toolAliases {
		names = append(names, name)
	}
	return names
}
