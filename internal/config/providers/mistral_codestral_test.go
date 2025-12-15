package providers

import (
	"testing"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMistralCodestralProvider(t *testing.T) {
	t.Run("creates provider with correct basic info", func(t *testing.T) {
		provider := MistralCodestralProvider([]catwalk.Provider{})

		require.NotNil(t, provider)
		assert.Equal(t, "mistral-codestral", string(provider.ID))
		assert.Equal(t, "Mistral (Codestral)", provider.Name)
		assert.Equal(t, "openai-compat", string(provider.Type))
		assert.Equal(t, "$MISTRAL_API_KEY", provider.APIKey)
		assert.Equal(t, "https://api.mistral.ai/v1", provider.APIEndpoint)
	})

	t.Run("creates provider with codestral models", func(t *testing.T) {
		provider := MistralCodestralProvider([]catwalk.Provider{})

		require.NotNil(t, provider)
		require.Len(t, provider.Models, 2)
		assert.Equal(t, "codestral-25-08", provider.DefaultLargeModelID)
		assert.Equal(t, "codestral-25-08", provider.DefaultSmallModelID)
	})

	t.Run("does not create if already exists", func(t *testing.T) {
		existingProvider := catwalk.Provider{ID: "mistral-codestral"}
		provider := MistralCodestralProvider([]catwalk.Provider{existingProvider})

		assert.Equal(t, "", string(provider.ID))
	})

	t.Run("has codestral model", func(t *testing.T) {
		provider := MistralCodestralProvider([]catwalk.Provider{})

		var codestralModel *catwalk.Model
		for i := range provider.Models {
			if provider.Models[i].ID == "codestral-25-08" {
				codestralModel = &provider.Models[i]
				break
			}
		}

		require.NotNil(t, codestralModel)
		assert.Equal(t, "Codestral", codestralModel.Name)
		assert.Equal(t, 0.3, codestralModel.CostPer1MIn)
		assert.Equal(t, 0.9, codestralModel.CostPer1MOut)
		assert.True(t, codestralModel.CanReason)
		assert.False(t, codestralModel.SupportsImages)
	})

	t.Run("has codestral embed model", func(t *testing.T) {
		provider := MistralCodestralProvider([]catwalk.Provider{})

		var embedModel *catwalk.Model
		for i := range provider.Models {
			if provider.Models[i].ID == "codestral-embed-25-05" {
				embedModel = &provider.Models[i]
				break
			}
		}

		require.NotNil(t, embedModel)
		assert.Equal(t, "Codestral Embed", embedModel.Name)
		assert.False(t, embedModel.CanReason)
	})
}
