package providers

import (
	"cmp"
	"os"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
)

// MistralCodestralProvider creates Mistral Codestral provider if it doesn't exist.
func MistralCodestralProvider(providers []catwalk.Provider) catwalk.Provider {
	// Check if mistral-codestral already exists
	for _, provider := range providers {
		if provider.ID == "mistral-codestral" {
			return catwalk.Provider{}
		}
	}

	mistralProvider := catwalk.Provider{
		Name:                "Mistral (Codestral)",
		ID:                  "mistral-codestral",
		APIKey:              "$MISTRAL_API_KEY",
		APIEndpoint:         cmp.Or(os.Getenv("MISTRAL_API_ENDPOINT"), "https://api.mistral.ai/v1"),
		Type:                "openai-compat",
		DefaultLargeModelID: "codestral-25-08",
		DefaultSmallModelID: "codestral-25-08",
		Models: []catwalk.Model{
			// Codestral - Code generation and code-specific tasks
			{
				ID:               "codestral-25-08",
				Name:             "Codestral",
				CostPer1MIn:      0.3,
				CostPer1MOut:     0.9,
				ContextWindow:    131072,
				DefaultMaxTokens: 32000,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			// Codestral Embed - Code embeddings
			{
				ID:               "codestral-embed-25-05",
				Name:             "Codestral Embed",
				CostPer1MIn:      0.1,
				CostPer1MOut:     0.1,
				ContextWindow:    8192,
				DefaultMaxTokens: 8192,
				CanReason:        false,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
		},
		DefaultHeaders: map[string]string{},
	}

	return mistralProvider
}
