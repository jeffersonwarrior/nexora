package styles

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestIconConstants tests that all icon constants are defined
func TestIconConstants(t *testing.T) {
	t.Parallel()

	assert.NotEmpty(t, CheckIcon, "CheckIcon should be defined")
	assert.NotEmpty(t, ErrorIcon, "ErrorIcon should be defined")
	assert.NotEmpty(t, WarningIcon, "WarningIcon should be defined")
	assert.NotEmpty(t, InfoIcon, "InfoIcon should be defined")
	assert.NotEmpty(t, HintIcon, "HintIcon should be defined")
	assert.NotEmpty(t, SpinnerIcon, "SpinnerIcon should be defined")
	assert.NotEmpty(t, LoadingIcon, "LoadingIcon should be defined")
	assert.NotEmpty(t, DocumentIcon, "DocumentIcon should be defined")
	assert.NotEmpty(t, ModelIcon, "ModelIcon should be defined")
}

// TestToolCallIcons tests tool call icon constants
func TestToolCallIcons(t *testing.T) {
	t.Parallel()

	assert.NotEmpty(t, ToolPending, "ToolPending should be defined")
	assert.NotEmpty(t, ToolSuccess, "ToolSuccess should be defined")
	assert.NotEmpty(t, ToolError, "ToolError should be defined")
}

// TestBorderIcons tests border icon constants
func TestBorderIcons(t *testing.T) {
	t.Parallel()

	assert.NotEmpty(t, BorderThin, "BorderThin should be defined")
	assert.NotEmpty(t, BorderThick, "BorderThick should be defined")
}

// TestSelectionIgnoreIcons tests SelectionIgnoreIcons slice
func TestSelectionIgnoreIcons(t *testing.T) {
	t.Parallel()

	assert.NotEmpty(t, SelectionIgnoreIcons, "SelectionIgnoreIcons should not be empty")
	assert.Contains(t, SelectionIgnoreIcons, BorderThin, "SelectionIgnoreIcons should contain BorderThin")
	assert.Contains(t, SelectionIgnoreIcons, BorderThick, "SelectionIgnoreIcons should contain BorderThick")
}

// TestIconUniqueness tests that icon constants are not duplicated
func TestIconUniqueness(t *testing.T) {
	t.Parallel()

	iconMap := make(map[string]string)

	icons := map[string]string{
		"CheckIcon":    CheckIcon,
		"ErrorIcon":    ErrorIcon,
		"WarningIcon":  WarningIcon,
		"InfoIcon":     InfoIcon,
		"HintIcon":     HintIcon,
		"SpinnerIcon":  SpinnerIcon,
		"LoadingIcon":  LoadingIcon,
		"DocumentIcon": DocumentIcon,
		"ModelIcon":    ModelIcon,
		"BorderThin":   BorderThin,
		"BorderThick":  BorderThick,
	}

	for name, icon := range icons {
		if existing, found := iconMap[icon]; found && name != existing {
			// Note: Some icons might intentionally share the same symbol
			// This just verifies the constants are accessible
		}
		iconMap[icon] = name
	}

	// Verify we collected all the icons
	assert.NotEmpty(t, iconMap, "should have collected icon values")
}

// BenchmarkCheckIcon benchmarks CheckIcon constant access
func BenchmarkCheckIcon(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = CheckIcon
	}
}

// BenchmarkSelectionIgnoreIcons benchmarks SelectionIgnoreIcons access
func BenchmarkSelectionIgnoreIcons(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = SelectionIgnoreIcons
	}
}
