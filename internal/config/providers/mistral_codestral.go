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
		DefaultLargeModelID: "codestral-latest",
		Models: []catwalk.Model{
			// Codestral Latest - Code generation and code-specific tasks
			{
				ID:               "codestral-latest",
				Name:             "Codestral Latest",
				CostPer1MIn:      1.0,
				CostPer1MOut:     3.0,
				ContextWindow:    256000,
				DefaultMaxTokens: 32000,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			// Codestral 2501 - Specific version
			{
				ID:               "codestral-2501",
				Name:             "Codestral 2501",
				CostPer1MIn:      1.0,
				CostPer1MOut:     3.0,
				ContextWindow:    256000,
				DefaultMaxTokens: 32000,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
		},
		DefaultHeaders: map[string]string{},
	}

	return mistralProvider
}
