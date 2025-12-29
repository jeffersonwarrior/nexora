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
		APIEndpoint:         cmp.Or(os.Getenv("MINIMAX_API_ENDPOINT"), "https://api.minimax.io/v1"),
		Type:                "openai-compat",
		DefaultLargeModelID: "MiniMax-M2.1",
		Models: []catwalk.Model{
			// M2 Series (Latest)
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
				CostPer1MIn:      0.3,
				CostPer1MOut:     1.2,
				ContextWindow:    204800,
				DefaultMaxTokens: 8000,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			// Text-01 Series (Open Source)
			{
				ID:               "MiniMax-Text-01",
				Name:             "MiniMax-Text-01",
				CostPer1MIn:      0.2,
				CostPer1MOut:     1.1,
				ContextWindow:    4000000,
				DefaultMaxTokens: 16000,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "MiniMax-VL-01",
				Name:             "MiniMax-VL-01 (Vision)",
				CostPer1MIn:      0.2,
				CostPer1MOut:     1.1,
				ContextWindow:    4000000,
				DefaultMaxTokens: 16000,
				CanReason:        true,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			// abab Series
			{
				ID:               "abab7-chat-preview",
				Name:             "abab7 Preview",
				CostPer1MIn:      10.0,
				CostPer1MOut:     10.0,
				ContextWindow:    200000,
				DefaultMaxTokens: 8000,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "abab6.5g-chat",
				Name:             "abab6.5 General",
				CostPer1MIn:      5.0,
				CostPer1MOut:     5.0,
				ContextWindow:    200000,
				DefaultMaxTokens: 8000,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "abab6.5s-chat",
				Name:             "abab6.5 Speed",
				CostPer1MIn:      1.0,
				CostPer1MOut:     1.0,
				ContextWindow:    200000,
				DefaultMaxTokens: 8000,
				CanReason:        false,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
		},
	}

	return minimaxProvider
}
