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

func TestSyntheticProvider_SkipsIfAlreadyExists(t *testing.T) {
	t.Parallel()
	provider := SyntheticProvider(nil)
	require.NotEmpty(t, provider.ID)

	syntheticInList := []catwalk.Provider{provider}
	result := SyntheticProvider(syntheticInList)
	require.Empty(t, result.ID)
}

func TestSyntheticProvider_ModelIDs(t *testing.T) {
	t.Parallel()
	provider := SyntheticProvider(nil)

	modelIDs := make(map[string]bool)
	for _, m := range provider.Models {
		modelIDs[string(m.ID)] = true
	}

	require.True(t, modelIDs["minimax/minimax-m2.1"])
	require.True(t, modelIDs["hf:zai-org/GLM-4.6"])
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
