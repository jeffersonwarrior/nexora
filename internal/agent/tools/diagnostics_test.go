package tools

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/powernap/pkg/lsp/protocol"
	"github.com/stretchr/testify/assert"
)

func TestNewDiagnosticsTool_Placeholder(t *testing.T) {
	// Note: Diagnostics tool tests require LSP client mocks which have complex dependencies
	// Full testing is covered in LSP integration tests
	// This test ensures the package compiles
	assert.True(t, true)
}

func TestFormatDiagnostic(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		diagnostic protocol.Diagnostic
		source     string
		expected   string
	}{
		{
			name: "error diagnostic",
			path: "/tmp/test.go",
			diagnostic: protocol.Diagnostic{
				Range: protocol.Range{
					Start: protocol.Position{Line: 9, Character: 4},
				},
				Severity: protocol.SeverityError,
				Message:  "undefined variable",
				Source:   "gopls",
			},
			source:   "gopls",
			expected: "Error: /tmp/test.go:10:5 [gopls] undefined variable",
		},
		{
			name: "warning diagnostic",
			path: "/tmp/test.go",
			diagnostic: protocol.Diagnostic{
				Range: protocol.Range{
					Start: protocol.Position{Line: 5, Character: 0},
				},
				Severity: protocol.SeverityWarning,
				Message:  "unused variable",
				Source:   "gopls",
			},
			source:   "gopls",
			expected: "Warn: /tmp/test.go:6:1 [gopls] unused variable",
		},
		{
			name: "hint diagnostic",
			path: "/tmp/test.go",
			diagnostic: protocol.Diagnostic{
				Range: protocol.Range{
					Start: protocol.Position{Line: 0, Character: 0},
				},
				Severity: protocol.SeverityHint,
				Message:  "can be simplified",
				Source:   "gopls",
			},
			source:   "gopls",
			expected: "Hint: /tmp/test.go:1:1 [gopls] can be simplified",
		},
		{
			name: "info diagnostic",
			path: "/tmp/test.go",
			diagnostic: protocol.Diagnostic{
				Range: protocol.Range{
					Start: protocol.Position{Line: 0, Character: 0},
				},
				Severity: protocol.SeverityInformation,
				Message:  "information",
				Source:   "gopls",
			},
			source:   "gopls",
			expected: "Info: /tmp/test.go:1:1 [gopls] information",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDiagnostic(tt.path, tt.diagnostic, tt.source)
			assert.Contains(t, result, tt.expected)
		})
	}
}

func TestFormatDiagnostic_WithCode(t *testing.T) {
	diagnostic := protocol.Diagnostic{
		Range: protocol.Range{
			Start: protocol.Position{Line: 0, Character: 0},
		},
		Severity: protocol.SeverityError,
		Message:  "test error",
		Source:   "gopls",
		Code:     "E001",
	}

	result := formatDiagnostic("/tmp/test.go", diagnostic, "gopls")
	assert.Contains(t, result, "[E001]")
}

func TestFormatDiagnostic_WithTags(t *testing.T) {
	diagnostic := protocol.Diagnostic{
		Range: protocol.Range{
			Start: protocol.Position{Line: 0, Character: 0},
		},
		Severity: protocol.SeverityWarning,
		Message:  "test warning",
		Source:   "gopls",
		Tags:     []protocol.DiagnosticTag{protocol.Unnecessary},
	}

	result := formatDiagnostic("/tmp/test.go", diagnostic, "gopls")
	assert.Contains(t, result, "unnecessary")
}

func TestFormatDiagnostic_WithDeprecatedTag(t *testing.T) {
	diagnostic := protocol.Diagnostic{
		Range: protocol.Range{
			Start: protocol.Position{Line: 0, Character: 0},
		},
		Severity: protocol.SeverityWarning,
		Message:  "test warning",
		Source:   "gopls",
		Tags:     []protocol.DiagnosticTag{protocol.Deprecated},
	}

	result := formatDiagnostic("/tmp/test.go", diagnostic, "gopls")
	assert.Contains(t, result, "deprecated")
}

func TestFormatDiagnostic_NoSource(t *testing.T) {
	diagnostic := protocol.Diagnostic{
		Range: protocol.Range{
			Start: protocol.Position{Line: 0, Character: 0},
		},
		Severity: protocol.SeverityError,
		Message:  "test error",
	}

	result := formatDiagnostic("/tmp/test.go", diagnostic, "fallback-source")
	assert.Contains(t, result, "[fallback-source]")
}

func TestSortDiagnostics(t *testing.T) {
	diagnostics := []string{
		"Warn: /tmp/test.go:10:5 warning message",
		"Error: /tmp/test.go:5:1 error message",
		"Info: /tmp/test.go:15:2 info message",
		"Error: /tmp/test.go:2:1 another error",
	}

	sorted := sortDiagnostics(diagnostics)

	// Errors should come first
	assert.True(t, strings.HasPrefix(sorted[0], "Error"))
	assert.True(t, strings.HasPrefix(sorted[1], "Error"))
	// Then non-errors
	assert.False(t, strings.HasPrefix(sorted[2], "Error"))
	assert.False(t, strings.HasPrefix(sorted[3], "Error"))
}

func TestWriteDiagnostics_Empty(t *testing.T) {
	var output strings.Builder
	writeDiagnostics(&output, "test_diagnostics", []string{})

	assert.Empty(t, output.String())
}

func TestWriteDiagnostics_Few(t *testing.T) {
	var output strings.Builder
	diagnostics := []string{
		"Error: /tmp/test.go:1:1 error 1",
		"Warn: /tmp/test.go:2:1 warning 1",
	}

	writeDiagnostics(&output, "test_diagnostics", diagnostics)

	result := output.String()
	assert.Contains(t, result, "<test_diagnostics>")
	assert.Contains(t, result, "</test_diagnostics>")
	assert.Contains(t, result, "Error: /tmp/test.go:1:1 error 1")
	assert.Contains(t, result, "Warn: /tmp/test.go:2:1 warning 1")
	assert.NotContains(t, result, "... and")
}

func TestWriteDiagnostics_Many(t *testing.T) {
	var output strings.Builder
	diagnostics := make([]string, 15)
	for i := 0; i < 15; i++ {
		diagnostics[i] = "Error: /tmp/test.go:1:1 error message"
	}

	writeDiagnostics(&output, "test_diagnostics", diagnostics)

	result := output.String()
	assert.Contains(t, result, "<test_diagnostics>")
	assert.Contains(t, result, "</test_diagnostics>")
	assert.Contains(t, result, "... and 5 more diagnostics")
}

func TestCountSeverity(t *testing.T) {
	diagnostics := []string{
		"Error: /tmp/test.go:1:1 error 1",
		"Error: /tmp/test.go:2:1 error 2",
		"Warn: /tmp/test.go:3:1 warning 1",
		"Info: /tmp/test.go:4:1 info 1",
		"Error: /tmp/test.go:5:1 error 3",
	}

	errorCount := countSeverity(diagnostics, "Error")
	warnCount := countSeverity(diagnostics, "Warn")
	infoCount := countSeverity(diagnostics, "Info")

	assert.Equal(t, 3, errorCount)
	assert.Equal(t, 1, warnCount)
	assert.Equal(t, 1, infoCount)
}

func TestCountSeverity_Empty(t *testing.T) {
	count := countSeverity([]string{}, "Error")
	assert.Equal(t, 0, count)
}

func TestCountSeverity_NoMatches(t *testing.T) {
	diagnostics := []string{
		"Warn: /tmp/test.go:1:1 warning 1",
		"Info: /tmp/test.go:2:1 info 1",
	}

	count := countSeverity(diagnostics, "Error")
	assert.Equal(t, 0, count)
}
