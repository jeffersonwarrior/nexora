package providers

import (
	"testing"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKimiCodingProvider(t *testing.T) {
	t.Run("creates provider with correct basic info", func(t *testing.T) {
		provider := KimiCodingProvider([]catwalk.Provider{})

		require.NotNil(t, provider)
		assert.Equal(t, "kimi", string(provider.ID))
		assert.Equal(t, "Kimi for Coding", provider.Name)
		assert.Equal(t, "openai-compat", string(provider.Type))
		assert.Equal(t, "$KIMI_API_KEY", provider.APIKey)
		assert.Equal(t, "https://api.kimi.com/coding/v1", provider.APIEndpoint)
	})

	t.Run("creates provider with kimi-for-coding model", func(t *testing.T) {
		provider := KimiCodingProvider([]catwalk.Provider{})

		require.NotNil(t, provider)
		require.Len(t, provider.Models, 1)
		assert.Equal(t, "kimi-for-coding", provider.DefaultLargeModelID)
	})

	t.Run("does not create if already exists", func(t *testing.T) {
		existingProvider := catwalk.Provider{ID: "kimi"}
		provider := KimiCodingProvider([]catwalk.Provider{existingProvider})

		assert.Equal(t, "", string(provider.ID))
	})

	t.Run("has kimi-for-coding model with reasoning", func(t *testing.T) {
		provider := KimiCodingProvider([]catwalk.Provider{})

		var codingModel *catwalk.Model
		for i := range provider.Models {
			if provider.Models[i].ID == "kimi-for-coding" {
				codingModel = &provider.Models[i]
				break
			}
		}

		require.NotNil(t, codingModel)
		assert.Equal(t, "Kimi for Coding", codingModel.Name)
		assert.True(t, codingModel.CanReason)
		assert.Equal(t, int64(262144), codingModel.ContextWindow)
		assert.Equal(t, int64(32768), codingModel.DefaultMaxTokens)
	})
}
