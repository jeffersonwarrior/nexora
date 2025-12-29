package providers

import (
	"testing"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCerebrasCodeProvider(t *testing.T) {
	t.Run("creates provider with correct basic info", func(t *testing.T) {
		provider := CerebrasCodeProvider([]catwalk.Provider{})

		require.NotNil(t, provider)
		assert.Equal(t, "cerebras-code", string(provider.ID))
		assert.Equal(t, "Cerebras Code", provider.Name)
		assert.Equal(t, "openai-compat", string(provider.Type))
		assert.Equal(t, "$CEREBRAS_API_KEY", provider.APIKey)
		assert.Equal(t, "https://api.cerebras.ai/v1", provider.APIEndpoint)
	})

	t.Run("has only GLM 4.6 model", func(t *testing.T) {
		provider := CerebrasCodeProvider([]catwalk.Provider{})

		require.NotNil(t, provider)
		require.Len(t, provider.Models, 1)
		assert.Equal(t, "zai-glm-4.6", provider.DefaultLargeModelID)
		assert.Equal(t, "zai-glm-4.6", provider.Models[0].ID)
	})

	t.Run("does not create if already exists", func(t *testing.T) {
		existingProvider := catwalk.Provider{ID: "cerebras-code"}
		provider := CerebrasCodeProvider([]catwalk.Provider{existingProvider})

		assert.Equal(t, "", string(provider.ID))
	})

	t.Run("GLM 4.6 has correct specs", func(t *testing.T) {
		provider := CerebrasCodeProvider([]catwalk.Provider{})

		model := provider.Models[0]
		assert.Equal(t, "Z.ai GLM 4.6 (Cerebras Code)", model.Name)
		assert.Equal(t, 2.25, model.CostPer1MIn)
		assert.Equal(t, 2.75, model.CostPer1MOut)
		assert.Equal(t, int64(131000), model.ContextWindow)
		assert.True(t, model.CanReason)
		assert.True(t, model.SupportsImages)
	})
}
