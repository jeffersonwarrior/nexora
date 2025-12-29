package providers

import (
	"testing"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
	"github.com/stretchr/testify/require"
)

func TestMistralGeneralProvider_ReturnsValidProvider(t *testing.T) {
	t.Parallel()
	provider := MistralGeneralProvider(nil)

	require.NotEmpty(t, provider.ID)
	require.Equal(t, "mistral-general", string(provider.ID))
	require.Equal(t, "Mistral (General)", provider.Name)
	require.Equal(t, "openai-compat", string(provider.Type))
	requireValidProvider(t, provider)
}

func TestMistralGeneralProvider_HasCorrectModelCount(t *testing.T) {
	t.Parallel()
	provider := MistralGeneralProvider(nil)

	require.Len(t, provider.Models, 11)
}

func TestMistralGeneralProvider_ModelIDsUnique(t *testing.T) {
	t.Parallel()
	provider := MistralGeneralProvider(nil)

	requireUniqueModelIDs(t, provider)
}

func TestMistralGeneralProvider_AllModelsHaveMetadata(t *testing.T) {
	t.Parallel()
	provider := MistralGeneralProvider(nil)

	requireAllModelsHaveMetadata(t, provider)
}

func TestMistralGeneralProvider_SkipsIfAlreadyExists(t *testing.T) {
	t.Parallel()
	provider := MistralGeneralProvider(nil)
	require.NotEmpty(t, provider.ID)

	// Now pass provider list with mistral-general already in it
	mistralInList := []catwalk.Provider{provider}
	result := MistralGeneralProvider(mistralInList)
	require.Empty(t, result.ID, "Should return empty provider if mistral-general already exists")
}

func TestMistralGeneralProvider_DefaultAPIEndpoint(t *testing.T) {
	// Can't use t.Parallel with t.Setenv
	t.Setenv("MISTRAL_API_ENDPOINT", "")

	provider := MistralGeneralProvider(nil)
	require.Equal(t, "https://api.mistral.ai/v1", provider.APIEndpoint)
}

func TestMistralGeneralProvider_RespectEnvVarEndpoint(t *testing.T) {
	// Can't use t.Parallel with t.Setenv
	customEndpoint := "https://custom.mistral.endpoint/v1"
	t.Setenv("MISTRAL_API_ENDPOINT", customEndpoint)

	provider := MistralGeneralProvider(nil)
	require.Equal(t, customEndpoint, provider.APIEndpoint)
}

func TestMistralGeneralProvider_HasDefaultModelIDs(t *testing.T) {
	t.Parallel()
	provider := MistralGeneralProvider(nil)

	require.Equal(t, "mistral-large-2512", string(provider.DefaultLargeModelID))
}

func TestMistralGeneralProvider_LargeModelIsExpensive(t *testing.T) {
	t.Parallel()
	provider := MistralGeneralProvider(nil)

	// Find large model
	var largeModelCost struct {
		in  float64
		out float64
	}
	for _, m := range provider.Models {
		if string(m.ID) == "mistral-large-2512" {
			largeModelCost.in = m.CostPer1MIn
			largeModelCost.out = m.CostPer1MOut
			break
		}
	}

	require.Equal(t, 2.0, largeModelCost.in)
	require.Equal(t, 6.0, largeModelCost.out)
}
