package providers

import (
	"cmp"
	"os"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
)

// AnthropicProvider creates Anthropic provider if it doesn't exist.
func AnthropicProvider(providers []catwalk.Provider) catwalk.Provider {
	// Check if anthropic already exists
	for _, provider := range providers {
		if provider.ID == "anthropic" {
			return catwalk.Provider{}
		}
	}

	anthropicProvider := catwalk.Provider{
		Name:                "Anthropic",
		ID:                  "anthropic",
		APIKey:              "$ANTHROPIC_API_KEY",
		APIEndpoint:         cmp.Or(os.Getenv("ANTHROPIC_API_ENDPOINT"), "https://api.anthropic.com/v1"),
		Type:                "anthropic",
		DefaultLargeModelID: "claude-opus-4-5-20251101",
		Models: []catwalk.Model{
			// Opus 4.5 - Flagship
			{
				ID:               "claude-opus-4-5-20251101",
				Name:             "Claude Opus 4.5",
				CostPer1MIn:      5.0,
				CostPer1MOut:     25.0,
				ContextWindow:    200000,
				DefaultMaxTokens: 16000,
				CanReason:        true,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			// Sonnet 4.5 - Balanced
			{
				ID:               "claude-sonnet-4-5-20250929",
				Name:             "Claude Sonnet 4.5",
				CostPer1MIn:      3.0,
				CostPer1MOut:     15.0,
				ContextWindow:    200000,
				DefaultMaxTokens: 16000,
				CanReason:        true,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			// Haiku 4.5 - Fast/Budget
			{
				ID:               "claude-haiku-4-5-20241022",
				Name:             "Claude Haiku 4.5",
				CostPer1MIn:      1.0,
				CostPer1MOut:     5.0,
				ContextWindow:    200000,
				DefaultMaxTokens: 8000,
				CanReason:        false,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			// Legacy Opus 4
			{
				ID:               "claude-opus-4-20250514",
				Name:             "Claude Opus 4",
				CostPer1MIn:      15.0,
				CostPer1MOut:     75.0,
				ContextWindow:    200000,
				DefaultMaxTokens: 4096,
				CanReason:        true,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			// Legacy Sonnet 3.5
			{
				ID:               "claude-sonnet-3-5-20241022",
				Name:             "Claude Sonnet 3.5",
				CostPer1MIn:      3.0,
				CostPer1MOut:     15.0,
				ContextWindow:    200000,
				DefaultMaxTokens: 4096,
				CanReason:        false,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
		},
		DefaultHeaders: map[string]string{},
	}

	return anthropicProvider
}
