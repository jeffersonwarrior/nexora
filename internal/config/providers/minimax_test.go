package providers

import (
	"testing"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
	"github.com/stretchr/testify/require"
)

func TestMiniMaxProvider_ReturnsValidProvider(t *testing.T) {
	t.Parallel()
	provider := MiniMaxProvider(nil)

	require.NotEmpty(t, provider.ID)
	require.Equal(t, "minimax", string(provider.ID))
	require.Equal(t, "MiniMax", provider.Name)
	requireValidProvider(t, provider)
}

func TestMiniMaxProvider_HasThreeModels(t *testing.T) {
	t.Parallel()
	provider := MiniMaxProvider(nil)

	require.Len(t, provider.Models, 3)
}

func TestMiniMaxProvider_UsesAnthropicType(t *testing.T) {
	t.Parallel()
	provider := MiniMaxProvider(nil)

	require.Equal(t, "anthropic", string(provider.Type))
}

func TestMiniMaxProvider_SkipsIfAlreadyExists(t *testing.T) {
	t.Parallel()
	provider := MiniMaxProvider(nil)
	require.NotEmpty(t, provider.ID)

	minimaxInList := []catwalk.Provider{provider}
	result := MiniMaxProvider(minimaxInList)
	require.Empty(t, result.ID)
}

func TestMiniMaxProvider_ModelIDs(t *testing.T) {
	t.Parallel()
	provider := MiniMaxProvider(nil)

	modelIDs := make(map[string]bool)
	for _, m := range provider.Models {
		modelIDs[string(m.ID)] = true
	}

	require.True(t, modelIDs["MiniMax-M2.1"])
	require.True(t, modelIDs["MiniMax-M2"])
	require.True(t, modelIDs["MiniMax-M2-Stable"])
}

func TestMiniMaxProvider_DefaultAPIEndpoint(t *testing.T) {
	t.Setenv("MINIMAX_API_ENDPOINT", "")

	provider := MiniMaxProvider(nil)
	require.Equal(t, "https://api.minimax.io/anthropic", provider.APIEndpoint)
}

func TestMiniMaxProvider_RespectEnvVarEndpoint(t *testing.T) {
	customEndpoint := "https://custom.minimax.endpoint/anthropic"
	t.Setenv("MINIMAX_API_ENDPOINT", customEndpoint)

	provider := MiniMaxProvider(nil)
	require.Equal(t, customEndpoint, provider.APIEndpoint)
}
