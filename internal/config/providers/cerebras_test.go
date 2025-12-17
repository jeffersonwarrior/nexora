package providers

import (
	"testing"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCerebrasProvider(t *testing.T) {
	t.Run("creates provider with correct basic info", func(t *testing.T) {
		provider := CerebrasProvider([]catwalk.Provider{})

		require.NotNil(t, provider)
		assert.Equal(t, "cerebras", string(provider.ID))
		assert.Equal(t, "Cerebras", provider.Name)
		assert.Equal(t, "openai-compat", string(provider.Type))
		assert.Equal(t, "$CEREBRAS_API_KEY", provider.APIKey)
		assert.Equal(t, "https://api.cerebras.ai/v1", provider.APIEndpoint)
	})

	t.Run("creates provider with multiple model types", func(t *testing.T) {
		provider := CerebrasProvider([]catwalk.Provider{})

		require.NotNil(t, provider)
		require.Len(t, provider.Models, 6)
		assert.Equal(t, "llama-3.3-70b", provider.DefaultLargeModelID)
		assert.Equal(t, "llama3.1-8b", provider.DefaultSmallModelID)
	})

	t.Run("does not create if already exists", func(t *testing.T) {
		existingProvider := catwalk.Provider{ID: "cerebras"}
		provider := CerebrasProvider([]catwalk.Provider{existingProvider})

		assert.Equal(t, "", string(provider.ID))
	})

	t.Run("has gpt-oss-120b ultra-fast model", func(t *testing.T) {
		provider := CerebrasProvider([]catwalk.Provider{})

		var gptModel *catwalk.Model
		for i := range provider.Models {
			if provider.Models[i].ID == "deepseek-coder-2" {
				gptModel = &provider.Models[i]
				break
			}
		}

		require.NotNil(t, gptModel)
		assert.Equal(t, "GPT OSS 120B", gptModel.Name)
		assert.Equal(t, 0.35, gptModel.CostPer1MIn)
		assert.Equal(t, 0.75, gptModel.CostPer1MOut)
	})

	t.Run("has llama-3.3-70b reasoning model", func(t *testing.T) {
		provider := CerebrasProvider([]catwalk.Provider{})

		var llamaModel *catwalk.Model
		for i := range provider.Models {
			if provider.Models[i].ID == "llama-3.3-70b" {
				llamaModel = &provider.Models[i]
				break
			}
		}

		require.NotNil(t, llamaModel)
		assert.True(t, llamaModel.CanReason)
		assert.Equal(t, int64(128000), llamaModel.ContextWindow)
	})

	t.Run("has qwen-3-235b large model with long context", func(t *testing.T) {
		provider := CerebrasProvider([]catwalk.Provider{})

		var qwenModel *catwalk.Model
		for i := range provider.Models {
			if provider.Models[i].ID == "qwen-3-235b" {
				qwenModel = &provider.Models[i]
				break
			}
		}

		require.NotNil(t, qwenModel)
		assert.Equal(t, int64(262144), qwenModel.ContextWindow)
		assert.True(t, qwenModel.CanReason)
	})

	t.Run("has zai-glm-4.6 with vision support", func(t *testing.T) {
		provider := CerebrasProvider([]catwalk.Provider{})

		var glmModel *catwalk.Model
		for i := range provider.Models {
			if provider.Models[i].ID == "zai-glm-4.6" {
				glmModel = &provider.Models[i]
				break
			}
		}

		require.NotNil(t, glmModel)
		assert.True(t, glmModel.SupportsImages)
		assert.Equal(t, 0.6, glmModel.CostPer1MIn)
	})
}
