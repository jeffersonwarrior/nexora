package providers

import (
	"testing"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
	"github.com/stretchr/testify/require"
)

func TestMistralProvider_ReturnsValidProvider(t *testing.T) {
	t.Parallel()
	provider := MistralProvider(nil)

	require.NotEmpty(t, provider.ID)
	require.Equal(t, "mistral", string(provider.ID))
	require.Equal(t, "Mistral", provider.Name)
	require.Equal(t, "openai-compat", string(provider.Type))
	requireValidProvider(t, provider)
}

func TestMistralProvider_HasCorrectModelCount(t *testing.T) {
	t.Parallel()
	provider := MistralProvider(nil)

	require.Len(t, provider.Models, 10)
}

func TestMistralProvider_ModelIDsUnique(t *testing.T) {
	t.Parallel()
	provider := MistralProvider(nil)

	requireUniqueModelIDs(t, provider)
}

func TestMistralProvider_AllModelsHaveMetadata(t *testing.T) {
	t.Parallel()
	provider := MistralProvider(nil)

	requireAllModelsHaveMetadata(t, provider)
}

func TestMistralProvider_SkipsIfAlreadyExists(t *testing.T) {
	t.Parallel()
	provider := MistralProvider(nil)
	require.NotEmpty(t, provider.ID)

	// Now pass provider list with mistral already in it
	mistralInList := []catwalk.Provider{provider}
	result := MistralProvider(mistralInList)
	require.Empty(t, result.ID, "Should return empty provider if mistral already exists")
}

func TestMistralProvider_DefaultAPIEndpoint(t *testing.T) {
	// Can't use t.Parallel with t.Setenv
	t.Setenv("MISTRAL_API_ENDPOINT", "")

	provider := MistralProvider(nil)
	require.Equal(t, "https://api.mistral.ai/v1", provider.APIEndpoint)
}

func TestMistralProvider_RespectEnvVarEndpoint(t *testing.T) {
	// Can't use t.Parallel with t.Setenv
	customEndpoint := "https://custom.mistral.endpoint/v1"
	t.Setenv("MISTRAL_API_ENDPOINT", customEndpoint)

	provider := MistralProvider(nil)
	require.Equal(t, customEndpoint, provider.APIEndpoint)
}

func TestMistralProvider_HasDefaultModelIDs(t *testing.T) {
	t.Parallel()
	provider := MistralProvider(nil)

	require.Equal(t, "mistral-large-3-25-12", string(provider.DefaultLargeModelID))
	require.Equal(t, "ministral-3-8b-25-12", string(provider.DefaultSmallModelID))
}

func TestMistralProvider_LargeModelIsExpensive(t *testing.T) {
	t.Parallel()
	provider := MistralProvider(nil)

	// Find large model
	var largeModelCost struct {
		in  float64
		out float64
	}
	for _, m := range provider.Models {
		if string(m.ID) == "mistral-large-3-25-12" {
			largeModelCost.in = m.CostPer1MIn
			largeModelCost.out = m.CostPer1MOut
			break
		}
	}

	require.Equal(t, 2.0, largeModelCost.in)
	require.Equal(t, 6.0, largeModelCost.out)
}
