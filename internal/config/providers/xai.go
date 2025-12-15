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
		DefaultLargeModelID: "grok-2",
		DefaultSmallModelID: "grok-3-mini",
		Models: []catwalk.Model{
			{
				ID:               "grok-2",
				Name:             "Grok 2",
				CostPer1MIn:      0.2,
				CostPer1MOut:     0.5,
				ContextWindow:    131072,
				DefaultMaxTokens: 16000,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "grok-3",
				Name:             "Grok 3",
				CostPer1MIn:      0.2,
				CostPer1MOut:     0.5,
				ContextWindow:    131072,
				DefaultMaxTokens: 16000,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "grok-3-mini",
				Name:             "Grok 3 Mini",
				CostPer1MIn:      0.3,
				CostPer1MOut:     0.5,
				ContextWindow:    1048576,
				DefaultMaxTokens: 8192,
				CanReason:        false,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "grok-vision-beta",
				Name:             "Grok Vision (Beta)",
				CostPer1MIn:      2.0,
				CostPer1MOut:     10.0,
				ContextWindow:    131072,
				DefaultMaxTokens: 4096,
				CanReason:        false,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "grok-beta",
				Name:             "Grok Beta",
				CostPer1MIn:      0.2,
				CostPer1MOut:     0.5,
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
