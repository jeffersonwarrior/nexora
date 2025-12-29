package providers

import (
	"testing"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKimiProvider(t *testing.T) {
	t.Run("creates provider with correct basic info", func(t *testing.T) {
		provider := KimiProvider([]catwalk.Provider{})

		require.NotNil(t, provider)
		assert.Equal(t, "kimi", string(provider.ID))
		assert.Equal(t, "Kimi", provider.Name)
		assert.Equal(t, "openai-compat", string(provider.Type))
		assert.Equal(t, "$KIMI_API_KEY", provider.APIKey)
		assert.Equal(t, "https://api.kimi.com/coding/", provider.APIEndpoint)
	})

	t.Run("creates provider with Kimi K2 models", func(t *testing.T) {
		provider := KimiProvider([]catwalk.Provider{})

		require.NotNil(t, provider)
		require.Len(t, provider.Models, 4)
		assert.Equal(t, "kimi-k2", provider.DefaultLargeModelID)
	})

	t.Run("does not create if already exists", func(t *testing.T) {
		existingProvider := catwalk.Provider{ID: "kimi"}
		provider := KimiProvider([]catwalk.Provider{existingProvider})

		assert.Equal(t, "", string(provider.ID))
	})

	t.Run("has kimi-k2 flagship model", func(t *testing.T) {
		provider := KimiProvider([]catwalk.Provider{})

		var k2Model *catwalk.Model
		for i := range provider.Models {
			if provider.Models[i].ID == "kimi-k2" {
				k2Model = &provider.Models[i]
				break
			}
		}

		require.NotNil(t, k2Model)
		assert.Equal(t, "Kimi K2", k2Model.Name)
		assert.Equal(t, 0.15, k2Model.CostPer1MIn)
		assert.Equal(t, 2.5, k2Model.CostPer1MOut)
		assert.Equal(t, int64(1000000), k2Model.ContextWindow)
	})

	t.Run("has kimi-k2-thinking with reasoning", func(t *testing.T) {
		provider := KimiProvider([]catwalk.Provider{})

		var thinkingModel *catwalk.Model
		for i := range provider.Models {
			if provider.Models[i].ID == "kimi-k2-thinking" {
				thinkingModel = &provider.Models[i]
				break
			}
		}

		require.NotNil(t, thinkingModel)
		assert.True(t, thinkingModel.CanReason)
		assert.Equal(t, 0.6, thinkingModel.CostPer1MIn)
	})

	t.Run("has kimi-k2-turbo budget model", func(t *testing.T) {
		provider := KimiProvider([]catwalk.Provider{})

		var turboModel *catwalk.Model
		for i := range provider.Models {
			if provider.Models[i].ID == "kimi-k2-turbo" {
				turboModel = &provider.Models[i]
				break
			}
		}

		require.NotNil(t, turboModel)
		assert.Equal(t, 0.05, turboModel.CostPer1MIn)
		assert.False(t, turboModel.SupportsImages)
	})

	t.Run("has kimi-k2-vision with image support", func(t *testing.T) {
		provider := KimiProvider([]catwalk.Provider{})

		var visionModel *catwalk.Model
		for i := range provider.Models {
			if provider.Models[i].ID == "kimi-k2-vision" {
				visionModel = &provider.Models[i]
				break
			}
		}

		require.NotNil(t, visionModel)
		assert.True(t, visionModel.SupportsImages)
		assert.Equal(t, int64(500000), visionModel.ContextWindow)
	})
}
