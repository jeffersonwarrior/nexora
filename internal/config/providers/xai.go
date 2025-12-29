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
		Models: []catwalk.Model{
			// Grok 4 Series
			{
				ID:               "grok-4",
				Name:             "Grok 4",
				CostPer1MIn:      3.0,
				CostPer1MOut:     15.0,
				ContextWindow:    256000,
				DefaultMaxTokens: 16000,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "grok-4-fast",
				Name:             "Grok 4 Fast",
				CostPer1MIn:      0.2,
				CostPer1MOut:     0.5,
				ContextWindow:    2000000,
				DefaultMaxTokens: 16000,
				CanReason:        false,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "grok-4-1-fast",
				Name:             "Grok 4.1 Fast",
				CostPer1MIn:      0.2,
				CostPer1MOut:     0.5,
				ContextWindow:    2000000,
				DefaultMaxTokens: 16000,
				CanReason:        false,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "grok-4-1-fast-reasoning",
				Name:             "Grok 4.1 Fast (Reasoning)",
				CostPer1MIn:      0.2,
				CostPer1MOut:     0.5,
				ContextWindow:    2000000,
				DefaultMaxTokens: 4096,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "grok-code-fast",
				Name:             "Grok Code Fast",
				CostPer1MIn:      0.2,
				CostPer1MOut:     1.5,
				ContextWindow:    2000000,
				DefaultMaxTokens: 16000,
				CanReason:        false,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			// Grok 3 Series
			{
				ID:               "grok-3",
				Name:             "Grok 3",
				CostPer1MIn:      3.0,
				CostPer1MOut:     15.0,
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
				ContextWindow:    131072,
				DefaultMaxTokens: 8192,
				CanReason:        false,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			// Grok 2 Series
			{
				ID:               "grok-2-1212",
				Name:             "Grok 2",
				CostPer1MIn:      2.0,
				CostPer1MOut:     10.0,
				ContextWindow:    131072,
				DefaultMaxTokens: 16000,
				CanReason:        false,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "grok-2-vision-1212",
				Name:             "Grok 2 Vision",
				CostPer1MIn:      2.0,
				CostPer1MOut:     10.0,
				ContextWindow:    32000,
				DefaultMaxTokens: 8000,
				CanReason:        false,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
		},
		DefaultHeaders: map[string]string{},
	}

	return xaiProvider
}
