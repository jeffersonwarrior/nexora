package providers

import (
	"testing"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMoonshotProvider(t *testing.T) {
	t.Run("creates provider with correct basic info", func(t *testing.T) {
		provider := MoonshotProvider([]catwalk.Provider{})

		require.NotNil(t, provider)
		assert.Equal(t, "moonshot", string(provider.ID))
		assert.Equal(t, "Moonshot AI", provider.Name)
		assert.Equal(t, "openai-compat", string(provider.Type))
		assert.Equal(t, "$MOONSHOT_API_KEY", provider.APIKey)
		assert.Equal(t, "https://api.moonshot.ai/v1", provider.APIEndpoint)
	})

	t.Run("creates provider with K2 models", func(t *testing.T) {
		provider := MoonshotProvider([]catwalk.Provider{})

		require.NotNil(t, provider)
		require.Len(t, provider.Models, 5)
		assert.Equal(t, "kimi-k2-turbo-preview", provider.DefaultLargeModelID)
	})

	t.Run("does not create if already exists", func(t *testing.T) {
		existingProvider := catwalk.Provider{ID: "moonshot"}
		provider := MoonshotProvider([]catwalk.Provider{existingProvider})

		assert.Equal(t, "", string(provider.ID))
	})

	t.Run("has kimi-k2-turbo-preview flagship model", func(t *testing.T) {
		provider := MoonshotProvider([]catwalk.Provider{})

		var k2Model *catwalk.Model
		for i := range provider.Models {
			if provider.Models[i].ID == "kimi-k2-turbo-preview" {
				k2Model = &provider.Models[i]
				break
			}
		}

		require.NotNil(t, k2Model)
		assert.Equal(t, "Kimi K2 Turbo", k2Model.Name)
		assert.Equal(t, 0.15, k2Model.CostPer1MIn)
		assert.Equal(t, 2.5, k2Model.CostPer1MOut)
		assert.Equal(t, int64(256000), k2Model.ContextWindow)
	})

	t.Run("has kimi-k2-thinking with reasoning", func(t *testing.T) {
		provider := MoonshotProvider([]catwalk.Provider{})

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

	t.Run("has moonshot-v1-128k model", func(t *testing.T) {
		provider := MoonshotProvider([]catwalk.Provider{})

		var v1Model *catwalk.Model
		for i := range provider.Models {
			if provider.Models[i].ID == "moonshot-v1-128k" {
				v1Model = &provider.Models[i]
				break
			}
		}

		require.NotNil(t, v1Model)
		assert.Equal(t, int64(128000), v1Model.ContextWindow)
	})
}
