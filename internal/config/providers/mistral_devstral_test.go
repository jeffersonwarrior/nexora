package providers

import (
	"testing"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMistralDevstralProvider(t *testing.T) {
	t.Run("creates provider with correct basic info", func(t *testing.T) {
		provider := MistralDevstralProvider([]catwalk.Provider{})

		require.NotNil(t, provider)
		assert.Equal(t, "mistral-devstral", string(provider.ID))
		assert.Equal(t, "Mistral (Devstral)", provider.Name)
		assert.Equal(t, "openai-compat", string(provider.Type))
		assert.Equal(t, "$MISTRAL_API_KEY", provider.APIKey)
		assert.Equal(t, "https://api.mistral.ai/v1", provider.APIEndpoint)
	})

	t.Run("creates provider with devstral models", func(t *testing.T) {
		provider := MistralDevstralProvider([]catwalk.Provider{})

		require.NotNil(t, provider)
		require.Len(t, provider.Models, 2)
		assert.Equal(t, "devstral-2512", provider.DefaultLargeModelID)
		assert.Equal(t, "devstral-small-2512", provider.DefaultSmallModelID)
	})

	t.Run("does not create if already exists", func(t *testing.T) {
		existingProvider := catwalk.Provider{ID: "mistral-devstral"}
		provider := MistralDevstralProvider([]catwalk.Provider{existingProvider})

		assert.Equal(t, "", string(provider.ID))
	})

	t.Run("has devstral-2 model marked as FREE", func(t *testing.T) {
		provider := MistralDevstralProvider([]catwalk.Provider{})

		var devstral2Model *catwalk.Model
		for i := range provider.Models {
			if provider.Models[i].ID == "devstral-2512" {
				devstral2Model = &provider.Models[i]
				break
			}
		}

		require.NotNil(t, devstral2Model)
		assert.Equal(t, "Devstral 2 (123B)", devstral2Model.Name)
		assert.Equal(t, 0.0, devstral2Model.CostPer1MIn)
		assert.Equal(t, 0.0, devstral2Model.CostPer1MOut)
		assert.True(t, devstral2Model.CanReason)
		assert.False(t, devstral2Model.SupportsImages)
	})

	t.Run("has devstral-small-2 model", func(t *testing.T) {
		provider := MistralDevstralProvider([]catwalk.Provider{})

		var smallModel *catwalk.Model
		for i := range provider.Models {
			if provider.Models[i].ID == "devstral-small-2512" {
				smallModel = &provider.Models[i]
				break
			}
		}

		require.NotNil(t, smallModel)
		assert.Equal(t, "Devstral Small 2 (24B)", smallModel.Name)
		assert.Equal(t, 0.0, smallModel.CostPer1MIn)
		assert.Equal(t, 0.0, smallModel.CostPer1MOut)
	})
}
