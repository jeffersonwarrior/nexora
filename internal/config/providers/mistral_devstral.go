package providers

import (
	"cmp"
	"os"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
)

// MistralDevstralProvider creates Mistral Devstral provider if it doesn't exist.
func MistralDevstralProvider(providers []catwalk.Provider) catwalk.Provider {
	// Check if mistral-devstral already exists
	for _, provider := range providers {
		if provider.ID == "mistral-devstral" {
			return catwalk.Provider{}
		}
	}

	mistralProvider := catwalk.Provider{
		Name:                "Mistral (Devstral)",
		ID:                  "mistral-devstral",
		APIKey:              "$MISTRAL_API_KEY",
		APIEndpoint:         cmp.Or(os.Getenv("MISTRAL_API_ENDPOINT"), "https://api.mistral.ai/v1"),
		Type:                "openai-compat",
		DefaultLargeModelID: "devstral-2-2512",
		DefaultSmallModelID: "devstral-small-2-2512",
		Models: []catwalk.Model{
			// Devstral 2 - Code understanding and agentic tasks
			{
				ID:               "devstral-2-2512",
				Name:             "Devstral 2 (123B)",
				CostPer1MIn:      0.0,  // FREE during beta
				CostPer1MOut:     0.0,  // FREE during beta
				ContextWindow:    262144,
				DefaultMaxTokens: 16000,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			// Devstral Small 2 - Lightweight code tasks
			{
				ID:               "devstral-small-2-2512",
				Name:             "Devstral Small 2 (24B)",
				CostPer1MIn:      0.0,  // FREE during beta
				CostPer1MOut:     0.0,  // FREE during beta
				ContextWindow:    262144,
				DefaultMaxTokens: 8000,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
		},
		DefaultHeaders: map[string]string{},
	}

	return mistralProvider
}
