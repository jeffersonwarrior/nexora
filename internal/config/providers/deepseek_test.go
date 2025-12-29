package providers

import (
	"testing"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeepSeekProvider(t *testing.T) {
	t.Run("creates provider with correct basic info", func(t *testing.T) {
		provider := DeepSeekProvider([]catwalk.Provider{})

		require.NotNil(t, provider)
		assert.Equal(t, "deepseek", string(provider.ID))
		assert.Equal(t, "DeepSeek", provider.Name)
		assert.Equal(t, "openai-compat", string(provider.Type))
		assert.Equal(t, "$DEEPSEEK_API_KEY", provider.APIKey)
		assert.Equal(t, "https://api.deepseek.com", provider.APIEndpoint)
	})

	t.Run("creates provider with DeepSeek models", func(t *testing.T) {
		provider := DeepSeekProvider([]catwalk.Provider{})

		require.NotNil(t, provider)
		require.Len(t, provider.Models, 2)
		assert.Equal(t, "deepseek-chat", provider.DefaultLargeModelID)
	})

	t.Run("does not create if already exists", func(t *testing.T) {
		existingProvider := catwalk.Provider{ID: "deepseek"}
		provider := DeepSeekProvider([]catwalk.Provider{existingProvider})

		assert.Equal(t, "", string(provider.ID))
	})

	t.Run("has deepseek-chat flagship model", func(t *testing.T) {
		provider := DeepSeekProvider([]catwalk.Provider{})

		var chatModel *catwalk.Model
		for i := range provider.Models {
			if provider.Models[i].ID == "deepseek-chat" {
				chatModel = &provider.Models[i]
				break
			}
		}

		require.NotNil(t, chatModel)
		assert.Equal(t, "DeepSeek V3.2 Chat", chatModel.Name)
		assert.Equal(t, 0.28, chatModel.CostPer1MIn)
		assert.Equal(t, 0.42, chatModel.CostPer1MOut)
		assert.Equal(t, int64(131072), chatModel.ContextWindow)
		assert.False(t, chatModel.CanReason)
	})

	t.Run("has deepseek-reasoner with reasoning", func(t *testing.T) {
		provider := DeepSeekProvider([]catwalk.Provider{})

		var reasonerModel *catwalk.Model
		for i := range provider.Models {
			if provider.Models[i].ID == "deepseek-reasoner" {
				reasonerModel = &provider.Models[i]
				break
			}
		}

		require.NotNil(t, reasonerModel)
		assert.Equal(t, "DeepSeek V3.2 Reasoner", reasonerModel.Name)
		assert.True(t, reasonerModel.CanReason)
		assert.Equal(t, 0.28, reasonerModel.CostPer1MIn)
		assert.Equal(t, 0.42, reasonerModel.CostPer1MOut)
		assert.Equal(t, int64(131072), reasonerModel.ContextWindow)
		assert.Equal(t, int64(32768), reasonerModel.DefaultMaxTokens)
	})
}
