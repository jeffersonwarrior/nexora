package providers

import (
	"testing"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
	"github.com/stretchr/testify/require"
)

func TestXAIProvider_ReturnsValidProvider(t *testing.T) {
	t.Parallel()
	provider := XAIProvider(nil)

	require.NotEmpty(t, provider.ID)
	require.Equal(t, "xai", string(provider.ID))
	require.Equal(t, "xAI", provider.Name)
	requireValidProvider(t, provider)
}

func TestXAIProvider_HasGrokModels(t *testing.T) {
	t.Parallel()
	provider := XAIProvider(nil)

	require.Len(t, provider.Models, 9)
	require.Equal(t, "grok-4", string(provider.Models[0].ID))
	require.Equal(t, "Grok 4", provider.Models[0].Name)
}

func TestXAIProvider_SkipsIfAlreadyExists(t *testing.T) {
	t.Parallel()
	provider := XAIProvider(nil)
	require.NotEmpty(t, provider.ID)

	xaiInList := []catwalk.Provider{provider}
	result := XAIProvider(xaiInList)
	require.Empty(t, result.ID)
}

func TestXAIProvider_DefaultAPIEndpoint(t *testing.T) {
	t.Setenv("XAI_API_ENDPOINT", "")

	provider := XAIProvider(nil)
	require.Equal(t, "https://api.x.ai/v1", provider.APIEndpoint)
}

func TestXAIProvider_RespectEnvVarEndpoint(t *testing.T) {
	customEndpoint := "https://custom.xai.endpoint/v1"
	t.Setenv("XAI_API_ENDPOINT", customEndpoint)

	provider := XAIProvider(nil)
	require.Equal(t, customEndpoint, provider.APIEndpoint)
}

func TestXAIProvider_OpenAICompat(t *testing.T) {
	t.Parallel()
	provider := XAIProvider(nil)

	require.Equal(t, "openai-compat", string(provider.Type))
}
