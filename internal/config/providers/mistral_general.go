package providers

import (
	"cmp"
	"os"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
)

// MistralGeneralProvider creates Mistral General provider if it doesn't exist.
func MistralGeneralProvider(providers []catwalk.Provider) catwalk.Provider {
	// Check if mistral-general already exists
	for _, provider := range providers {
		if provider.ID == "mistral-general" {
			return catwalk.Provider{}
		}
	}

	mistralProvider := catwalk.Provider{
		Name:                "Mistral (General)",
		ID:                  "mistral-general",
		APIKey:              "$MISTRAL_API_KEY",
		APIEndpoint:         cmp.Or(os.Getenv("MISTRAL_API_ENDPOINT"), "https://api.mistral.ai/v1"),
		Type:                "openai-compat",
		DefaultLargeModelID: "mistral-large-2512",
		Models: []catwalk.Model{
			// Large reasoning models
			{
				ID:               "mistral-large-2512",
				Name:             "Mistral Large 3",
				CostPer1MIn:      2.0,
				CostPer1MOut:     6.0,
				ContextWindow:    256000,
				DefaultMaxTokens: 32000,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "mistral-large-2411",
				Name:             "Mistral Large 2.1",
				CostPer1MIn:      3.0,
				CostPer1MOut:     9.0,
				ContextWindow:    128000,
				DefaultMaxTokens: 32000,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			// Multimodal Models
			{
				ID:               "pixtral-large-2411",
				Name:             "Pixtral Large",
				CostPer1MIn:      2.0,
				CostPer1MOut:     6.0,
				ContextWindow:    128000,
				DefaultMaxTokens: 32000,
				CanReason:        true,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "pixtral-12b-2409",
				Name:             "Pixtral 12B",
				CostPer1MIn:      0.15,
				CostPer1MOut:     0.15,
				ContextWindow:    128000,
				DefaultMaxTokens: 16000,
				CanReason:        false,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			// Medium/Small Models
			{
				ID:               "mistral-medium-2508",
				Name:             "Mistral Medium 3.1",
				CostPer1MIn:      0.4,
				CostPer1MOut:     2.0,
				ContextWindow:    128000,
				DefaultMaxTokens: 32000,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "mistral-small-2501",
				Name:             "Mistral Small 3",
				CostPer1MIn:      0.2,
				CostPer1MOut:     0.6,
				ContextWindow:    128000,
				DefaultMaxTokens: 16000,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "open-mistral-nemo",
				Name:             "Mistral Nemo",
				CostPer1MIn:      0.3,
				CostPer1MOut:     0.3,
				ContextWindow:    128000,
				DefaultMaxTokens: 16000,
				CanReason:        false,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			// Ministral Series
			{
				ID:               "ministral-8b-2512",
				Name:             "Ministral 3 8B",
				CostPer1MIn:      0.1,
				CostPer1MOut:     0.1,
				ContextWindow:    128000,
				DefaultMaxTokens: 8000,
				CanReason:        false,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "ministral-3b-2512",
				Name:             "Ministral 3 3B",
				CostPer1MIn:      0.04,
				CostPer1MOut:     0.04,
				ContextWindow:    128000,
				DefaultMaxTokens: 8000,
				CanReason:        false,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			// Open Source Models
			{
				ID:               "open-mistral-7b",
				Name:             "Mistral 7B",
				CostPer1MIn:      0.25,
				CostPer1MOut:     0.25,
				ContextWindow:    32000,
				DefaultMaxTokens: 8000,
				CanReason:        false,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "open-mixtral-8x22b",
				Name:             "Mixtral 8x22B",
				CostPer1MIn:      2.0,
				CostPer1MOut:     6.0,
				ContextWindow:    64000,
				DefaultMaxTokens: 16000,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
		},
		DefaultHeaders: map[string]string{},
	}

	return mistralProvider
}
