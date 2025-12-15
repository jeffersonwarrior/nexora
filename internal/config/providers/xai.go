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
		DefaultLargeModelID: "grok-beta",
		DefaultSmallModelID: "grok-beta",
		Models: []catwalk.Model{
			{
				ID:               "grok-beta",
				Name:             "Grok Beta",
				CostPer1MIn:      5.0,
				CostPer1MOut:     15.0,
				ContextWindow:    131072,
				DefaultMaxTokens: 8192,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
		},
		DefaultHeaders: map[string]string{},
	}

	return xaiProvider
}
