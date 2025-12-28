package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveToolName_FetchAliases(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"curl to fetch", "curl", "fetch"},
		{"wget to fetch", "wget", "fetch"},
		{"http-get to fetch", "http-get", "fetch"},
		{"web_fetch to web_fetch", "web_fetch", "web_fetch"},
		{"uppercase curl", "CURL", "fetch"},
		{"mixed case Curl", "cUrL", "fetch"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ResolveToolName(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestResolveToolName_ViewAliases(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"read to view", "read", "view"},
		{"cat to view", "cat", "view"},
		{"open to view", "open", "view"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ResolveToolName(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestResolveToolName_ListAliases(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"list to ls", "list", "ls"},
		{"dir to ls", "dir", "ls"},
		{"directory to ls", "directory", "ls"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ResolveToolName(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestResolveToolName_GrepAliases(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"search to grep", "search", "grep"},
		{"find to find", "find", "find"},
		{"rg to grep", "rg", "grep"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ResolveToolName(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestResolveToolName_BashAliases(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"shell to bash", "shell", "bash"},
		{"exec to bash", "exec", "bash"},
		{"execute to bash", "execute", "bash"},
		{"run to bash", "run", "bash"},
		{"command to bash", "command", "bash"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ResolveToolName(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestResolveToolName_CanonicalReturnsUnchanged(t *testing.T) {
	// Canonical tool names should return themselves
	canonicalNames := []string{
		"fetch", "view", "edit", "write", "ls", "grep",
		"bash", "web_search", "sourcegraph", "download",
		"find", "glob", "job_kill", "job_output", "multiedit", "web_fetch",
	}

	for _, name := range canonicalNames {
		t.Run("canonical "+name, func(t *testing.T) {
			result := ResolveToolName(name)
			assert.Equal(t, name, result, "Canonical name should return itself")
		})
	}
}

func TestResolveToolName_EmptyString(t *testing.T) {
	result := ResolveToolName("")
	assert.Equal(t, "", result)
}

func TestResolveToolName_UnknownTool(t *testing.T) {
	result := ResolveToolName("unknown_tool_xyz")
	assert.Equal(t, "unknown_tool_xyz", result)
}

func TestIsAlias(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"curl is alias", "curl", true},
		{"wget is alias", "wget", true},
		{"fetch is not alias", "fetch", false},
		{"unknown is not alias", "unknown", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsAlias(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetCanonicalName(t *testing.T) {
	// GetCanonicalName is just an alias for ResolveToolName
	result := GetCanonicalName("curl")
	assert.Equal(t, "fetch", result)
}

func TestAddAndRemoveAlias(t *testing.T) {
	// Add a custom alias
	AddAlias("my_custom_tool", "fetch")
	assert.True(t, IsAlias("my_custom_tool"))
	assert.Equal(t, "fetch", ResolveToolName("my_custom_tool"))

	// Remove the alias
	RemoveAlias("my_custom_tool")
	assert.False(t, IsAlias("my_custom_tool"))
	assert.Equal(t, "my_custom_tool", ResolveToolName("my_custom_tool"))
}

func TestListAliases(t *testing.T) {
	// Test that we can list aliases for a canonical tool
	fetchAliases := ListAliases("fetch")
	assert.True(t, len(fetchAliases) > 0, "fetch should have aliases")
	assert.Contains(t, fetchAliases, "curl")
	assert.Contains(t, fetchAliases, "wget")
}

func TestToolNames(t *testing.T) {
	// Test that we can get all registered tool names
	names := ToolNames()
	assert.True(t, len(names) > 0, "should have some tool names")
	assert.Contains(t, names, "curl")
	assert.Contains(t, names, "wget")
}

func TestResolveToolName_WebSearchAliases(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"web-search to web_search", "web-search", "web_search"},
		{"websearch to web_search", "websearch", "web_search"},
		{"search-web to web_search", "search-web", "web_search"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ResolveToolName(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestResolveToolName_SourcegraphAliases(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"sg to sourcegraph", "sg", "sourcegraph"},
		{"code-search to sourcegraph", "code-search", "sourcegraph"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ResolveToolName(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
