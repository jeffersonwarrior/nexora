package providers

import (
	"testing"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenAIProvider(t *testing.T) {
	t.Run("creates provider with correct basic info", func(t *testing.T) {
		provider := OpenAIProvider([]catwalk.Provider{})

		require.NotNil(t, provider)
		assert.Equal(t, "openai", string(provider.ID))
		assert.Equal(t, "OpenAI", provider.Name)
		assert.Equal(t, "openai-compat", string(provider.Type))
		assert.Equal(t, "$OPENAI_API_KEY", provider.APIKey)
		assert.Equal(t, "https://api.openai.com/v1", provider.APIEndpoint)
	})

	t.Run("creates provider with default models", func(t *testing.T) {
		provider := OpenAIProvider([]catwalk.Provider{})

		require.NotNil(t, provider)
		assert.NotEmpty(t, provider.Models)
		assert.Equal(t, "gpt-4o", provider.DefaultLargeModelID)
	})

	t.Run("does not create if already exists", func(t *testing.T) {
		existingProvider := catwalk.Provider{ID: "openai"}
		provider := OpenAIProvider([]catwalk.Provider{existingProvider})

		assert.Equal(t, "", string(provider.ID))
	})

	t.Run("has gpt-4o model", func(t *testing.T) {
		provider := OpenAIProvider([]catwalk.Provider{})

		var gpt4oModel *catwalk.Model
		for i := range provider.Models {
			if provider.Models[i].ID == "gpt-4o" {
				gpt4oModel = &provider.Models[i]
				break
			}
		}

		require.NotNil(t, gpt4oModel)
		assert.Equal(t, "GPT-4o", gpt4oModel.Name)
		assert.Equal(t, 5.0, gpt4oModel.CostPer1MIn)
		assert.Equal(t, 15.0, gpt4oModel.CostPer1MOut)
		assert.True(t, gpt4oModel.SupportsImages)
	})

	t.Run("has gpt-4o-mini model", func(t *testing.T) {
		provider := OpenAIProvider([]catwalk.Provider{})

		var miniModel *catwalk.Model
		for i := range provider.Models {
			if provider.Models[i].ID == "gpt-4o-mini" {
				miniModel = &provider.Models[i]
				break
			}
		}

		require.NotNil(t, miniModel)
		assert.Equal(t, "GPT-4o Mini", miniModel.Name)
		assert.Equal(t, 0.15, miniModel.CostPer1MIn)
		assert.Equal(t, 0.6, miniModel.CostPer1MOut)
	})

	t.Run("has gpt-5 model", func(t *testing.T) {
		provider := OpenAIProvider([]catwalk.Provider{})

		var gpt5Model *catwalk.Model
		for i := range provider.Models {
			if provider.Models[i].ID == "gpt-5" {
				gpt5Model = &provider.Models[i]
				break
			}
		}

		require.NotNil(t, gpt5Model)
		assert.Equal(t, "GPT-5", gpt5Model.Name)
		assert.Equal(t, 1.25, gpt5Model.CostPer1MIn)
		assert.Equal(t, 10.0, gpt5Model.CostPer1MOut)
	})

	t.Run("has o1 reasoning model", func(t *testing.T) {
		provider := OpenAIProvider([]catwalk.Provider{})

		var o1Model *catwalk.Model
		for i := range provider.Models {
			if provider.Models[i].ID == "o1" {
				o1Model = &provider.Models[i]
				break
			}
		}

		require.NotNil(t, o1Model)
		assert.True(t, o1Model.CanReason)
		assert.Equal(t, 15.0, o1Model.CostPer1MIn)
	})
}
