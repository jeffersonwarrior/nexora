package providers

import (
	"cmp"
	"os"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
)

// MistralProvider creates Mistral provider if it doesn't exist.
func MistralProvider(providers []catwalk.Provider) catwalk.Provider {
	// Check if mistral already exists
	for _, provider := range providers {
		if provider.ID == "mistral" {
			return catwalk.Provider{}
		}
	}

	mistralProvider := catwalk.Provider{
		Name:                "Mistral",
		ID:                  "mistral",
		APIKey:              "$MISTRAL_API_KEY",
		APIEndpoint:         cmp.Or(os.Getenv("MISTRAL_API_ENDPOINT"), "https://api.mistral.ai/v1"),
		Type:                "openai-compat",
		DefaultLargeModelID: "mistral-large-3-25-12",
		DefaultSmallModelID: "ministral-3-8b-25-12",
		Models: []catwalk.Model{
			// Large reasoning models
			{
				ID:               "mistral-large-3-25-12",
				Name:             "Mistral Large 3",
				CostPer1MIn:      2.0,
				CostPer1MOut:     6.0,
				ContextWindow:    131072,
				DefaultMaxTokens: 32000,
				CanReason:        true,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "mistral-medium-3-1-25-08",
				Name:             "Mistral Medium 3.1",
				CostPer1MIn:      1.5,
				CostPer1MOut:     4.5,
				ContextWindow:    131072,
				DefaultMaxTokens: 32000,
				CanReason:        true,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "mistral-small-3-2-25-06",
				Name:             "Mistral Small 3.2",
				CostPer1MIn:      0.2,
				CostPer1MOut:     0.6,
				ContextWindow:    128000,
				DefaultMaxTokens: 16000,
				CanReason:        true,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			// Ministral models
			{
				ID:               "ministral-3-14b-25-12",
				Name:             "Ministral 3 14B",
				CostPer1MIn:      0.15,
				CostPer1MOut:     0.45,
				ContextWindow:    128000,
				DefaultMaxTokens: 16000,
				CanReason:        true,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "ministral-3-8b-25-12",
				Name:             "Ministral 3 8B",
				CostPer1MIn:      0.1,
				CostPer1MOut:     0.3,
				ContextWindow:    128000,
				DefaultMaxTokens: 8000,
				CanReason:        false,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "ministral-3-3b-25-12",
				Name:             "Ministral 3 3B",
				CostPer1MIn:      0.05,
				CostPer1MOut:     0.15,
				ContextWindow:    128000,
				DefaultMaxTokens: 4000,
				CanReason:        false,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			// Devstral models
			{
				ID:               "devstral-2512",
				Name:             "Devstral 2",
				CostPer1MIn:      0.3,
				CostPer1MOut:     0.9,
				ContextWindow:    262144,
				DefaultMaxTokens: 16000,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
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
			// Embedding models
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
			{
				ID:               "mistral-embed",
				Name:             "Mistral Embed",
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
