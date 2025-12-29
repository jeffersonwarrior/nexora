package providers

import (
	"testing"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnthropicProvider(t *testing.T) {
	t.Run("creates provider with correct basic info", func(t *testing.T) {
		provider := AnthropicProvider([]catwalk.Provider{})

		require.NotNil(t, provider)
		assert.Equal(t, "anthropic", string(provider.ID))
		assert.Equal(t, "Anthropic", provider.Name)
		assert.Equal(t, "anthropic", string(provider.Type))
		assert.Equal(t, "$ANTHROPIC_API_KEY", provider.APIKey)
		assert.Equal(t, "https://api.anthropic.com/v1", provider.APIEndpoint)
	})

	t.Run("creates provider with default models", func(t *testing.T) {
		provider := AnthropicProvider([]catwalk.Provider{})

		require.NotNil(t, provider)
		assert.NotEmpty(t, provider.Models)
		assert.Equal(t, "claude-opus-4-5-20251101", provider.DefaultLargeModelID)
	})

	t.Run("does not create if already exists", func(t *testing.T) {
		existingProvider := catwalk.Provider{ID: "anthropic"}
		provider := AnthropicProvider([]catwalk.Provider{existingProvider})

		assert.Equal(t, "", string(provider.ID))
	})

	t.Run("has claude-opus-4-5 model", func(t *testing.T) {
		provider := AnthropicProvider([]catwalk.Provider{})

		var opusModel *catwalk.Model
		for i := range provider.Models {
			if provider.Models[i].ID == "claude-opus-4-5-20251101" {
				opusModel = &provider.Models[i]
				break
			}
		}

		require.NotNil(t, opusModel)
		assert.Equal(t, "Claude Opus 4.5", opusModel.Name)
		assert.Equal(t, 5.0, opusModel.CostPer1MIn)
		assert.Equal(t, 25.0, opusModel.CostPer1MOut)
		assert.True(t, opusModel.SupportsImages)
		assert.True(t, opusModel.CanReason)
	})

	t.Run("has claude-sonnet-4-5 model", func(t *testing.T) {
		provider := AnthropicProvider([]catwalk.Provider{})

		var sonnetModel *catwalk.Model
		for i := range provider.Models {
			if provider.Models[i].ID == "claude-sonnet-4-5-20250929" {
				sonnetModel = &provider.Models[i]
				break
			}
		}

		require.NotNil(t, sonnetModel)
		assert.Equal(t, "Claude Sonnet 4.5", sonnetModel.Name)
		assert.Equal(t, 3.0, sonnetModel.CostPer1MIn)
		assert.Equal(t, 15.0, sonnetModel.CostPer1MOut)
	})

	t.Run("has claude-haiku-4-5 model", func(t *testing.T) {
		provider := AnthropicProvider([]catwalk.Provider{})

		var haikuModel *catwalk.Model
		for i := range provider.Models {
			if provider.Models[i].ID == "claude-haiku-4-5-20241022" {
				haikuModel = &provider.Models[i]
				break
			}
		}

		require.NotNil(t, haikuModel)
		assert.Equal(t, "Claude Haiku 4.5", haikuModel.Name)
		assert.Equal(t, 1.0, haikuModel.CostPer1MIn)
		assert.Equal(t, 5.0, haikuModel.CostPer1MOut)
		assert.False(t, haikuModel.CanReason)
	})
}
