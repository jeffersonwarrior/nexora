package providers

import (
	"cmp"
	"os"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
)

// XAIProvider creates x.ai provider if it doesn't exist.
func XAIProvider(providers []catwalk.Provider) catwalk.Provider {
	// Check if xai already exists
	for _, provider := range providers {
		if provider.ID == "xai" {
			return catwalk.Provider{}
		}
	}

	xaiProvider := catwalk.Provider{
		Name:                "xAI",
		ID:                  "xai",
		APIKey:              "$XAI_API_KEY",
		APIEndpoint:         cmp.Or(os.Getenv("XAI_API_ENDPOINT"), "https://api.x.ai/v1"),
		Type:                "openai-compat",
		DefaultLargeModelID: "grok-4",
		DefaultSmallModelID: "grok-3-mini",
		Models: []catwalk.Model{
			{
				ID:               "grok-4",
				Name:             "Grok 4",
				CostPer1MIn:      0.2,
				CostPer1MOut:     0.5,
				ContextWindow:    131072,
				DefaultMaxTokens: 16000,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "grok-4-fast",
				Name:             "Grok 4 Fast",
				CostPer1MIn:      0.15,
				CostPer1MOut:     0.4,
				ContextWindow:    131072,
				DefaultMaxTokens: 16000,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "grok-4-heavy",
				Name:             "Grok 4 Heavy",
				CostPer1MIn:      0.3,
				CostPer1MOut:     0.7,
				ContextWindow:    131072,
				DefaultMaxTokens: 16000,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "grok-3",
				Name:             "Grok 3",
				CostPer1MIn:      0.15,
				CostPer1MOut:     0.4,
				ContextWindow:    131072,
				DefaultMaxTokens: 16000,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "grok-3-mini",
				Name:             "Grok 3 Mini",
				CostPer1MIn:      0.1,
				CostPer1MOut:     0.25,
				ContextWindow:    131072,
				DefaultMaxTokens: 8192,
				CanReason:        false,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
		},
		DefaultHeaders: map[string]string{},
	}

	return xaiProvider
}
