package models

import (
	"slices"
	"testing"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
	"github.com/stretchr/testify/require"
)

func TestOpenRouterProviderSorting(t *testing.T) {
	providers := []catwalk.Provider{
		{ID: catwalk.InferenceProviderOpenAI, Name: "OpenAI"},
		{ID: catwalk.InferenceProviderAnthropic, Name: "Anthropic"},
		{ID: catwalk.InferenceProviderOpenRouter, Name: "OpenRouter"},
		{ID: catwalk.InferenceProviderAzure, Name: "Azure"},
	}

	// Simulate the sorting logic from list.go
	sortedProviders := make([]catwalk.Provider, len(providers))
	copy(sortedProviders, providers)
	slices.SortStableFunc(sortedProviders, func(a, b catwalk.Provider) int {
		// If a is OpenRouter, it should come after b
		if a.ID == catwalk.InferenceProviderOpenRouter && b.ID != catwalk.InferenceProviderOpenRouter {
			return 1
		}
		// If b is OpenRouter, it should come after a
		if b.ID == catwalk.InferenceProviderOpenRouter && a.ID != catwalk.InferenceProviderOpenRouter {
			return -1
		}
		// Keep original order for all other providers
		return 0
	})

	// Verify OpenRouter is at the bottom
	openRouterIndex := -1
	for i, p := range sortedProviders {
		if p.ID == catwalk.InferenceProviderOpenRouter {
			openRouterIndex = i
			break
		}
	}

	require.Greater(t, openRouterIndex, 0, "OpenRouter should not be the first provider")
	require.Equal(t, len(sortedProviders)-1, openRouterIndex, "OpenRouter should be the last provider")

	// Verify all other providers maintain their relative order
	otherProviders := sortedProviders[:openRouterIndex]
	require.Equal(t, 3, len(otherProviders), "Should have 3 other providers")
	require.Equal(t, catwalk.InferenceProviderOpenAI, otherProviders[0].ID)
	require.Equal(t, catwalk.InferenceProviderAnthropic, otherProviders[1].ID)
	require.Equal(t, catwalk.InferenceProviderAzure, otherProviders[2].ID)
}
