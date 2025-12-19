package providers

import (
	"testing"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
	"github.com/stretchr/testify/require"
)

func TestMistralNativeProvider(t *testing.T) {
	t.Run("creates provider", func(t *testing.T) {
		provider := MistralNativeProvider([]catwalk.Provider{})

		require.NotNil(t, provider)
		require.Equal(t, catwalk.InferenceProvider("mistral"), provider.ID)
		require.Equal(t, "Mistral AI (Native)", provider.Name)
		require.Equal(t, catwalk.TypeOpenAICompat, provider.Type)
		require.Equal(t, "$MISTRAL_API_KEY", provider.APIKey)
		require.Equal(t, "https://api.mistral.ai/v1", provider.APIEndpoint)
		require.Len(t, provider.Models, 2)
		require.Equal(t, "devstral-2512", provider.DefaultLargeModelID)
		require.Equal(t, "devstral-small-2512", provider.DefaultSmallModelID)

		var devstralModel *catwalk.Model
		for i := range provider.Models {
			if provider.Models[i].ID == "devstral-2512" {
				devstralModel = &provider.Models[i]
				break
			}
		}
		require.NotNil(t, devstralModel)
		require.Equal(t, "Devstral 2 (123B)", devstralModel.Name)
		require.Equal(t, int64(262144), devstralModel.ContextWindow)

		var smallModel *catwalk.Model
		for i := range provider.Models {
			if provider.Models[i].ID == "devstral-small-2512" {
				smallModel = &provider.Models[i]
				break
			}
		}
		require.NotNil(t, smallModel)
		require.Equal(t, "Devstral Small 2 (24B)", smallModel.Name)
	})

	t.Run("does not create if already exists", func(t *testing.T) {
		existing := catwalk.Provider{ID: "mistral"}
		provider := MistralNativeProvider([]catwalk.Provider{existing})

		require.Equal(t, catwalk.InferenceProvider(""), provider.ID)
	})
}
