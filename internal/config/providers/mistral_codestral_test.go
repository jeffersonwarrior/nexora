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
		assert.Equal(t, "codestral-latest", provider.DefaultLargeModelID)
	})

	t.Run("does not create if already exists", func(t *testing.T) {
		existingProvider := catwalk.Provider{ID: "mistral-codestral"}
		provider := MistralCodestralProvider([]catwalk.Provider{existingProvider})

		assert.Equal(t, "", string(provider.ID))
	})

	t.Run("has codestral latest model", func(t *testing.T) {
		provider := MistralCodestralProvider([]catwalk.Provider{})

		var codestralModel *catwalk.Model
		for i := range provider.Models {
			if provider.Models[i].ID == "codestral-latest" {
				codestralModel = &provider.Models[i]
				break
			}
		}

		require.NotNil(t, codestralModel)
		assert.Equal(t, "Codestral Latest", codestralModel.Name)
		assert.Equal(t, 1.0, codestralModel.CostPer1MIn)
		assert.Equal(t, 3.0, codestralModel.CostPer1MOut)
		assert.True(t, codestralModel.CanReason)
		assert.False(t, codestralModel.SupportsImages)
	})

	t.Run("has codestral 2501 model", func(t *testing.T) {
		provider := MistralCodestralProvider([]catwalk.Provider{})

		var codestral2501 *catwalk.Model
		for i := range provider.Models {
			if provider.Models[i].ID == "codestral-2501" {
				codestral2501 = &provider.Models[i]
				break
			}
		}

		require.NotNil(t, codestral2501)
		assert.Equal(t, "Codestral 2501", codestral2501.Name)
		assert.Equal(t, 1.0, codestral2501.CostPer1MIn)
		assert.Equal(t, 3.0, codestral2501.CostPer1MOut)
	})
}
