package providers

import (
	"testing"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeminiProvider(t *testing.T) {
	t.Run("creates provider with correct basic info", func(t *testing.T) {
		provider := GeminiProvider([]catwalk.Provider{})

		require.NotNil(t, provider)
		assert.Equal(t, "gemini", string(provider.ID))
		assert.Equal(t, "Google Gemini", provider.Name)
		assert.Equal(t, "openai-compat", string(provider.Type))
		assert.Equal(t, "$GEMINI_API_KEY", provider.APIKey)
		assert.Equal(t, "https://generativelanguage.googleapis.com/v1beta/models", provider.APIEndpoint)
	})

	t.Run("creates provider with default models", func(t *testing.T) {
		provider := GeminiProvider([]catwalk.Provider{})

		require.NotNil(t, provider)
		assert.NotEmpty(t, provider.Models)
		assert.Equal(t, "gemini-3-pro", provider.DefaultLargeModelID)
		assert.Equal(t, "gemini-2-5-flash", provider.DefaultSmallModelID)
	})

	t.Run("does not create if already exists", func(t *testing.T) {
		existingProvider := catwalk.Provider{ID: "gemini"}
		provider := GeminiProvider([]catwalk.Provider{existingProvider})

		assert.Equal(t, "", string(provider.ID))
	})

	t.Run("has gemini-3-pro model", func(t *testing.T) {
		provider := GeminiProvider([]catwalk.Provider{})

		var gemini3Model *catwalk.Model
		for i := range provider.Models {
			if provider.Models[i].ID == "gemini-3-pro" {
				gemini3Model = &provider.Models[i]
				break
			}
		}

		require.NotNil(t, gemini3Model)
		assert.Equal(t, "Gemini 3 Pro", gemini3Model.Name)
		assert.Equal(t, 2.0, gemini3Model.CostPer1MIn)
		assert.Equal(t, 12.0, gemini3Model.CostPer1MOut)
		assert.True(t, gemini3Model.SupportsImages)
		assert.True(t, gemini3Model.CanReason)
	})

	t.Run("has gemini-2-5-flash model", func(t *testing.T) {
		provider := GeminiProvider([]catwalk.Provider{})

		var flashModel *catwalk.Model
		for i := range provider.Models {
			if provider.Models[i].ID == "gemini-2-5-flash" {
				flashModel = &provider.Models[i]
				break
			}
		}

		require.NotNil(t, flashModel)
		assert.Equal(t, "Gemini 2.5 Flash", flashModel.Name)
		assert.Equal(t, 0.075, flashModel.CostPer1MIn)
		assert.Equal(t, 0.3, flashModel.CostPer1MOut)
		assert.False(t, flashModel.CanReason)
	})

	t.Run("has extended thinking model", func(t *testing.T) {
		provider := GeminiProvider([]catwalk.Provider{})

		var thinkingModel *catwalk.Model
		for i := range provider.Models {
			if provider.Models[i].ID == "gemini-2-0-flash-thinking-exp" {
				thinkingModel = &provider.Models[i]
				break
			}
		}

		require.NotNil(t, thinkingModel)
		assert.True(t, thinkingModel.CanReason)
		assert.Equal(t, 5.0, thinkingModel.CostPer1MIn)
		assert.Equal(t, 20.0, thinkingModel.CostPer1MOut)
	})
}
