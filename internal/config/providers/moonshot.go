package providers

import (
	"cmp"
	"os"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
)

// MoonshotProvider creates Moonshot AI provider if it doesn't exist.
// API docs: https://platform.moonshot.ai/docs/
func MoonshotProvider(providers []catwalk.Provider) catwalk.Provider {
	// Check if moonshot already exists
	for _, provider := range providers {
		if provider.ID == "moonshot" {
			return catwalk.Provider{}
		}
	}

	moonshotProvider := catwalk.Provider{
		Name:                "Moonshot AI",
		ID:                  "moonshot",
		APIKey:              "$MOONSHOT_API_KEY",
		APIEndpoint:         cmp.Or(os.Getenv("MOONSHOT_API_ENDPOINT"), "https://api.moonshot.ai/v1"),
		Type:                "openai-compat",
		DefaultLargeModelID: "kimi-k2-turbo-preview",
		Models: []catwalk.Model{
			// K2 Series (Latest)
			{
				ID:               "kimi-k2-turbo-preview",
				Name:             "Kimi K2 Turbo",
				CostPer1MIn:      0.15,
				CostPer1MOut:     2.5,
				ContextWindow:    256000,
				DefaultMaxTokens: 32000,
				CanReason:        false,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "kimi-k2-thinking",
				Name:             "Kimi K2 Thinking",
				CostPer1MIn:      0.6,
				CostPer1MOut:     2.5,
				ContextWindow:    256000,
				DefaultMaxTokens: 32000,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			// V1 Series (Legacy)
			{
				ID:               "moonshot-v1-8k",
				Name:             "Moonshot V1 8K",
				CostPer1MIn:      0.2,
				CostPer1MOut:     2.0,
				ContextWindow:    8000,
				DefaultMaxTokens: 4000,
				CanReason:        false,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "moonshot-v1-32k",
				Name:             "Moonshot V1 32K",
				CostPer1MIn:      1.0,
				CostPer1MOut:     3.0,
				ContextWindow:    32000,
				DefaultMaxTokens: 16000,
				CanReason:        false,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "moonshot-v1-128k",
				Name:             "Moonshot V1 128K",
				CostPer1MIn:      2.0,
				CostPer1MOut:     5.0,
				ContextWindow:    128000,
				DefaultMaxTokens: 32000,
				CanReason:        false,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
		},
		DefaultHeaders: map[string]string{},
	}

	return moonshotProvider
}
