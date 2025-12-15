package config

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
	"github.com/stretchr/testify/require"
)

type emptyProviderClient struct{}

func (m *emptyProviderClient) GetProviders(ctx context.Context, githubURL string) ([]catwalk.Provider, error) {
	return []catwalk.Provider{}, nil
}

func TestProvider_loadProvidersEmptyResult(t *testing.T) {
	client := &emptyProviderClient{}
	tmpPath := t.TempDir() + "/providers.json"

	providers, err := loadProviders(false, client, tmpPath)
	require.NoError(t, err)
	// Should get embedded providers + injected providers
	require.NotEmpty(t, providers)
	// Should have embedded providers (mock) + custom providers
	require.Greater(t, len(providers), 0)

	// Check that no cache file was created for embedded fallback
	require.NoFileExists(t, tmpPath, "Cache file should not exist when using embedded providers as fallback")
}

func TestProvider_loadProvidersEmptyCache(t *testing.T) {
	client := &mockProviderClient{shouldFail: false}
	tmpPath := t.TempDir() + "/providers.json"

	// Create an empty cache file
	emptyProviders := []catwalk.Provider{}
	data, err := json.Marshal(emptyProviders)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(tmpPath, data, 0o644))

	// Should refresh and get real providers instead of using empty cache
	providers, err := loadProviders(false, client, tmpPath)
	require.NoError(t, err)
	require.NotNil(t, providers)
	// Mock + mistral + nexora + xai + minimax
	require.Equal(t, "Mock", providers[0].Name)
	require.Equal(t, "Mistral", providers[1].Name)
	require.Equal(t, "Nexora", providers[2].Name)
	require.Equal(t, "xAI", providers[3].Name)
	require.Equal(t, "MiniMax", providers[4].Name)
}
