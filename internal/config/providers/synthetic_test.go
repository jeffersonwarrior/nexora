package providers

import (
	"testing"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
	"github.com/stretchr/testify/require"
)

func TestSyntheticProvider_ReturnsValidProvider(t *testing.T) {
	t.Parallel()
	provider := SyntheticProvider(nil)

	require.NotEmpty(t, provider.ID)
	require.Equal(t, "synthetic", string(provider.ID))
	require.Equal(t, "Synthetic", provider.Name)
	requireValidProvider(t, provider)
}

func TestSyntheticProvider_HasTwoModels(t *testing.T) {
	t.Parallel()
	provider := SyntheticProvider(nil)

	require.Len(t, provider.Models, 2)
}

func TestSyntheticProvider_UsesOpenAICompatType(t *testing.T) {
	t.Parallel()
	provider := SyntheticProvider(nil)

	require.Equal(t, "openai-compat", string(provider.Type))
}

func TestSyntheticProvider_OverridesExisting(t *testing.T) {
	t.Parallel()

	// Create a provider list with an existing synthetic provider (simulating embedded providers)
	existingSynthetic := catwalk.Provider{
		ID:                  "synthetic",
		Name:                "Old Synthetic",
		DefaultLargeModelID: "glm-4.6", // Old default that causes issues
	}
	syntheticInList := []catwalk.Provider{existingSynthetic}

	// Our provider should always override, not skip
	result := SyntheticProvider(syntheticInList)
	require.NotEmpty(t, result.ID)
	require.Equal(t, "synthetic", string(result.ID))
	require.Equal(t, "minimax/minimax-m2.1", string(result.DefaultLargeModelID))
}

func TestSyntheticProvider_ModelIDs(t *testing.T) {
	t.Parallel()
	provider := SyntheticProvider(nil)

	modelIDs := make(map[string]bool)
	for _, m := range provider.Models {
		modelIDs[string(m.ID)] = true
	}

	require.True(t, modelIDs["minimax/minimax-m2.1"])
	require.True(t, modelIDs["glm-4.7"])
}

func TestSyntheticProvider_DefaultAPIEndpoint(t *testing.T) {
	t.Setenv("SYNTHETIC_API_ENDPOINT", "")

	provider := SyntheticProvider(nil)
	require.Equal(t, "https://api.synthetic.new/v1", provider.APIEndpoint)
}

func TestSyntheticProvider_RespectEnvVarEndpoint(t *testing.T) {
	customEndpoint := "https://custom.synthetic.endpoint/v1"
	t.Setenv("SYNTHETIC_API_ENDPOINT", customEndpoint)

	provider := SyntheticProvider(nil)
	require.Equal(t, customEndpoint, provider.APIEndpoint)
}

func TestSyntheticProvider_DefaultModelExists(t *testing.T) {
	t.Parallel()
	provider := SyntheticProvider(nil)

	// Verify that the default large model actually exists in the models list
	defaultLargeFound := false
	for _, model := range provider.Models {
		if model.ID == provider.DefaultLargeModelID {
			defaultLargeFound = true
			break
		}
	}

	require.True(t, defaultLargeFound, "DefaultLargeModelID %s not found in models list", provider.DefaultLargeModelID)
}
