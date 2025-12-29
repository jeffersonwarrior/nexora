package providers

import (
	"cmp"
	"os"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
)

// KimiProvider creates Kimi (Moonshot) provider if it doesn't exist.
func KimiProvider(providers []catwalk.Provider) catwalk.Provider {
	// Check if kimi already exists
	for _, provider := range providers {
		if provider.ID == "kimi" {
			return catwalk.Provider{}
		}
	}

	kimiProvider := catwalk.Provider{
		Name:                "Kimi",
		ID:                  "kimi",
		APIKey:              "$KIMI_API_KEY",
		APIEndpoint:         cmp.Or(os.Getenv("KIMI_API_ENDPOINT"), "https://api.kimi.com/coding/"),
		Type:                "openai-compat",
		DefaultLargeModelID: "kimi-k2",
		Models: []catwalk.Model{
			// Kimi K2 - Latest flagship
			{
				ID:               "kimi-k2",
				Name:             "Kimi K2",
				CostPer1MIn:      0.15,
				CostPer1MOut:     2.5,
				ContextWindow:    1000000,
				DefaultMaxTokens: 32000,
				CanReason:        false,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			// Kimi K2 Thinking - With reasoning capability
			{
				ID:               "kimi-k2-thinking",
				Name:             "Kimi K2 (Thinking)",
				CostPer1MIn:      0.6,
				CostPer1MOut:     2.5,
				ContextWindow:    1000000,
				DefaultMaxTokens: 32000,
				CanReason:        true,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			// Kimi K2 Turbo - Fast and affordable
			{
				ID:               "kimi-k2-turbo",
				Name:             "Kimi K2 Turbo",
				CostPer1MIn:      0.05,
				CostPer1MOut:     1.5,
				ContextWindow:    1000000,
				DefaultMaxTokens: 16000,
				CanReason:        false,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			// Kimi K2 Vision - Optimized for images
			{
				ID:               "kimi-k2-vision",
				Name:             "Kimi K2 Vision",
				CostPer1MIn:      0.2,
				CostPer1MOut:     2.5,
				ContextWindow:    500000,
				DefaultMaxTokens: 16000,
				CanReason:        false,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
		},
		DefaultHeaders: map[string]string{},
	}

	return kimiProvider
}
