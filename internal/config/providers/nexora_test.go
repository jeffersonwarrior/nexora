package providers

import (
	"testing"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
	"github.com/stretchr/testify/require"
)

func TestNexoraProvider_ReturnsValidProvider(t *testing.T) {
	t.Parallel()
	provider := NexoraProvider(nil)

	require.NotEmpty(t, provider.ID)
	require.Equal(t, "nexora", string(provider.ID))
	require.Equal(t, "Nexora", provider.Name)
	requireValidProvider(t, provider)
}

func TestNexoraProvider_UsesLocalhostEndpoint(t *testing.T) {
	t.Parallel()
	provider := NexoraProvider(nil)

	require.Equal(t, "http://localhost:9000/v1", provider.APIEndpoint)
}

func TestNexoraProvider_SkipsIfAlreadyExists(t *testing.T) {
	t.Parallel()
	provider := NexoraProvider(nil)
	require.NotEmpty(t, provider.ID)

	nexoraInList := []catwalk.Provider{provider}
	result := NexoraProvider(nexoraInList)
	require.Empty(t, result.ID)
}

func TestNexoraProvider_EmptyAPIKey(t *testing.T) {
	t.Parallel()
	provider := NexoraProvider(nil)

	require.Empty(t, provider.APIKey)
}

func TestNexoraProvider_HasDevstralModel(t *testing.T) {
	t.Parallel()
	provider := NexoraProvider(nil)

	require.Len(t, provider.Models, 1)
	require.Equal(t, "devstral-small-2", string(provider.Models[0].ID))
}
