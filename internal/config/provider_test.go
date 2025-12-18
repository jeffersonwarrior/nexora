package config

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
	"github.com/stretchr/testify/require"
)

type mockProviderClient struct {
	shouldFail bool
}

func (m *mockProviderClient) GetProviders(ctx context.Context, githubURL string) ([]catwalk.Provider, error) {
	if m.shouldFail {
		return nil, errors.New("failed to load providers")
	}
	return []catwalk.Provider{
		{
			Name: "Mock",
		},
	}, nil
}

func TestProvider_loadProvidersNoIssues(t *testing.T) {
	client := &mockProviderClient{shouldFail: false}
	tmpPath := t.TempDir() + "/providers.json"
	providers, err := loadProviders(false, client, tmpPath)
	require.NoError(t, err)
	require.NotNil(t, providers)
	// Check for required providers instead of exact count
	providerMap := make(map[string]bool)
	providerNames := make(map[string]bool)
	for _, p := range providers {
		providerMap[string(p.ID)] = true
		providerNames[p.Name] = true
	}
	// Mock provider from client has no ID, just check by name
	require.True(t, providerNames["Mock"], "Expected mock provider from client")
	// Check that injected providers are included
	require.True(t, providerMap["mistral-general"], "Expected mistral-general provider injected")

	// check if file got saved
	fileInfo, err := os.Stat(tmpPath)
	require.NoError(t, err)
	require.False(t, fileInfo.IsDir(), "Expected a file, not a directory")
}

func TestProvider_loadProvidersWithIssues(t *testing.T) {
	client := &mockProviderClient{shouldFail: true}
	tmpPath := t.TempDir() + "/providers.json"
	// store providers to a temporary file
	oldProviders := []catwalk.Provider{
		{
			Name: "OldProvider",
		},
	}
	data, err := json.Marshal(oldProviders)
	if err != nil {
		t.Fatalf("Failed to marshal old providers: %v", err)
	}

	err = os.WriteFile(tmpPath, data, 0o644)
	if err != nil {
		t.Fatalf("Failed to write old providers to file: %v", err)
	}
	providers, err := loadProviders(true, client, tmpPath)
	require.NoError(t, err)
	require.NotNil(t, providers)
	// Check that old provider is there + injected providers
	providerMap := make(map[string]bool)
	providerNames := make(map[string]bool)
	for _, p := range providers {
		providerMap[string(p.ID)] = true
		providerNames[p.Name] = true
	}
	require.True(t, providerNames["OldProvider"], "Expected to keep old provider when loading fails")
	require.True(t, providerMap["mistral-general"], "Expected to have mistral-general provider injected")
	require.True(t, providerMap["xai"], "Expected to have xai provider injected")
}

func TestProvider_loadProvidersWithIssuesAndNoCache(t *testing.T) {
	client := &mockProviderClient{shouldFail: true}
	tmpPath := t.TempDir() + "/providers.json"
	providers, err := loadProviders(false, client, tmpPath)
	// When Catwalk fails and no cache, fallback to embedded providers
	require.NoError(t, err)
	require.NotNil(t, providers)
	require.Greater(t, len(providers), 0, "Expected embedded providers as fallback")
}
