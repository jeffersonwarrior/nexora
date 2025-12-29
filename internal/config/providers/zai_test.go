package providers

import (
	"testing"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestZAIProvider(t *testing.T) {
	t.Run("creates provider with correct basic info", func(t *testing.T) {
		provider := ZAIProvider([]catwalk.Provider{})

		require.NotNil(t, provider)
		assert.Equal(t, "zai", string(provider.ID))
		assert.Equal(t, "Z.AI", provider.Name)
		assert.Equal(t, "openai-compat", string(provider.Type))
		assert.Equal(t, "$ZAI_API_KEY", provider.APIKey)
		assert.Equal(t, "https://api.z.ai/api/paas/v4", provider.APIEndpoint)
	})

	t.Run("creates provider with GLM models", func(t *testing.T) {
		provider := ZAIProvider([]catwalk.Provider{})

		require.NotNil(t, provider)
		require.Len(t, provider.Models, 8)
		assert.Equal(t, "glm-4.7", provider.DefaultLargeModelID)
	})

	t.Run("does not create if already exists", func(t *testing.T) {
		existingProvider := catwalk.Provider{ID: "zai"}
		provider := ZAIProvider([]catwalk.Provider{existingProvider})

		assert.Equal(t, "", string(provider.ID))
	})

	t.Run("has glm-4.6 flagship model", func(t *testing.T) {
		provider := ZAIProvider([]catwalk.Provider{})

		var glm46Model *catwalk.Model
		for i := range provider.Models {
			if provider.Models[i].ID == "glm-4.6" {
				glm46Model = &provider.Models[i]
				break
			}
		}

		require.NotNil(t, glm46Model)
		assert.Equal(t, "GLM-4.6", glm46Model.Name)
		assert.Equal(t, 0.6, glm46Model.CostPer1MIn)
		assert.Equal(t, 2.2, glm46Model.CostPer1MOut)
		assert.Equal(t, int64(200000), glm46Model.ContextWindow)
	})

	t.Run("has free flash models", func(t *testing.T) {
		provider := ZAIProvider([]catwalk.Provider{})

		var flashModel *catwalk.Model
		for i := range provider.Models {
			if provider.Models[i].ID == "glm-4.5-flash" {
				flashModel = &provider.Models[i]
				break
			}
		}

		require.NotNil(t, flashModel)
		assert.Equal(t, 0.0, flashModel.CostPer1MIn)
		assert.Equal(t, 0.0, flashModel.CostPer1MOut)
	})

	t.Run("has vision models", func(t *testing.T) {
		provider := ZAIProvider([]catwalk.Provider{})

		var visionModel *catwalk.Model
		for i := range provider.Models {
			if provider.Models[i].ID == "glm-4.6v" {
				visionModel = &provider.Models[i]
				break
			}
		}

		require.NotNil(t, visionModel)
		assert.True(t, visionModel.SupportsImages)
	})
}
