package providers

import (
	"testing"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
	"github.com/stretchr/testify/require"
)

// requireValidProvider ensures a provider has required fields.
func requireValidProvider(t *testing.T, p catwalk.Provider) {
	t.Helper()
	require.NotEmpty(t, p.ID, "Provider must have ID")
	require.NotEmpty(t, p.Name, "Provider must have Name")
	require.NotEmpty(t, p.Type, "Provider must have Type")
}

// requireUniqueModelIDs ensures all model IDs are unique within provider.
func requireUniqueModelIDs(t *testing.T, p catwalk.Provider) {
	t.Helper()
	seen := make(map[string]bool)
	for _, m := range p.Models {
		require.False(t, seen[m.ID], "Duplicate model ID: %s", m.ID)
		seen[m.ID] = true
	}
}

// requireAllModelsHaveMetadata ensures each model has required fields.
func requireAllModelsHaveMetadata(t *testing.T, p catwalk.Provider) {
	t.Helper()
	for _, m := range p.Models {
		require.NotEmpty(t, m.ID, "Model must have ID")
		require.NotEmpty(t, m.Name, "Model must have Name")
		require.NotZero(t, m.ContextWindow, "Model must have ContextWindow")
	}
}
