package providers

import (
	"cmp"
	"os"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
)

// MiniMaxProvider creates MiniMax provider if it doesn't exist.
func MiniMaxProvider(providers []catwalk.Provider) catwalk.Provider {
	// Check if minimax already exists
	for _, provider := range providers {
		if provider.ID == "minimax" {
			return catwalk.Provider{}
		}
	}

	minimaxProvider := catwalk.Provider{
		Name:                "MiniMax",
		ID:                  "minimax",
		APIKey:              "$MINIMAX_API_KEY",
		APIEndpoint:         cmp.Or(os.Getenv("MINIMAX_API_ENDPOINT"), "https://api.minimax.io/anthropic"),
		Type:                "anthropic",
		DefaultLargeModelID: "MiniMax-M2.1",
		Models: []catwalk.Model{
			{
				ID:               "MiniMax-M2.1",
				Name:             "MiniMax M2.1",
				CostPer1MIn:      0.3,
				CostPer1MOut:     1.2,
				ContextWindow:    204800,
				DefaultMaxTokens: 8000,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "MiniMax-M2",
				Name:             "MiniMax M2",
				CostPer1MIn:      0.5,
				CostPer1MOut:     1.5,
				ContextWindow:    200000,
				DefaultMaxTokens: 8000,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "MiniMax-M2-Stable",
				Name:             "MiniMax M2 Stable",
				CostPer1MIn:      0.4,
				CostPer1MOut:     1.2,
				ContextWindow:    200000,
				DefaultMaxTokens: 8000,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
		},
		DefaultHeaders: map[string]string{
			"Authorization": "Bearer $MINIMAX_API_KEY",
		},
	}

	return minimaxProvider
}
