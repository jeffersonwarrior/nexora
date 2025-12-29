package providers

import (
	"cmp"
	"os"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
)

// KimiCodingProvider creates Kimi for Coding provider if it doesn't exist.
// API docs: https://docs.kimi.com/coding/
func KimiCodingProvider(providers []catwalk.Provider) catwalk.Provider {
	// Check if kimi already exists
	for _, provider := range providers {
		if provider.ID == "kimi" {
			return catwalk.Provider{}
		}
	}

	kimiProvider := catwalk.Provider{
		Name:                "Kimi for Coding",
		ID:                  "kimi",
		APIKey:              "$KIMI_API_KEY",
		APIEndpoint:         cmp.Or(os.Getenv("KIMI_API_ENDPOINT"), "https://api.kimi.com/coding/v1"),
		Type:                "openai-compat",
		DefaultLargeModelID: "kimi-for-coding",
		Models: []catwalk.Model{
			{
				ID:               "kimi-for-coding",
				Name:             "Kimi for Coding",
				CostPer1MIn:      0.0, // Subscription-based
				CostPer1MOut:     0.0,
				ContextWindow:    262144,
				DefaultMaxTokens: 32768,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
		},
		DefaultHeaders: map[string]string{},
	}

	return kimiProvider
}
