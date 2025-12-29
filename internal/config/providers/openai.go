package providers

import (
	"cmp"
	"os"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
)

// OpenAIProvider creates OpenAI provider if it doesn't exist.
func OpenAIProvider(providers []catwalk.Provider) catwalk.Provider {
	// Check if openai already exists
	for _, provider := range providers {
		if provider.ID == "openai" {
			return catwalk.Provider{}
		}
	}

	openaiProvider := catwalk.Provider{
		Name:                "OpenAI",
		ID:                  "openai",
		APIKey:              "$OPENAI_API_KEY",
		APIEndpoint:         cmp.Or(os.Getenv("OPENAI_API_ENDPOINT"), "https://api.openai.com/v1"),
		Type:                "openai-compat",
		DefaultLargeModelID: "gpt-4o",
		Models: []catwalk.Model{
			// Reasoning models (O1 series)
			{
				ID:               "o1",
				Name:             "O1",
				CostPer1MIn:      15.0,
				CostPer1MOut:     60.0,
				ContextWindow:    128000,
				DefaultMaxTokens: 32000,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "o1-mini",
				Name:             "O1 Mini",
				CostPer1MIn:      3.0,
				CostPer1MOut:     12.0,
				ContextWindow:    128000,
				DefaultMaxTokens: 16000,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			// GPT-5.2 (latest flagship)
			{
				ID:               "gpt-5-2",
				Name:             "GPT-5.2",
				CostPer1MIn:      1.75,
				CostPer1MOut:     7.0,
				ContextWindow:    131072,
				DefaultMaxTokens: 32000,
				CanReason:        true,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			// GPT-4o (latest standard)
			{
				ID:               "gpt-4o",
				Name:             "GPT-4o",
				CostPer1MIn:      2.5,
				CostPer1MOut:     10.0,
				ContextWindow:    131072,
				DefaultMaxTokens: 16000,
				CanReason:        true,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "gpt-4o-mini",
				Name:             "GPT-4o Mini",
				CostPer1MIn:      0.15,
				CostPer1MOut:     0.6,
				ContextWindow:    131072,
				DefaultMaxTokens: 16000,
				CanReason:        false,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			// GPT-4 Turbo
			{
				ID:               "gpt-4-turbo",
				Name:             "GPT-4 Turbo",
				CostPer1MIn:      10.0,
				CostPer1MOut:     30.0,
				ContextWindow:    131072,
				DefaultMaxTokens: 16000,
				CanReason:        false,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			// Legacy models
			{
				ID:               "gpt-4",
				Name:             "GPT-4",
				CostPer1MIn:      30.0,
				CostPer1MOut:     60.0,
				ContextWindow:    8192,
				DefaultMaxTokens: 4096,
				CanReason:        false,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "gpt-3.5-turbo",
				Name:             "GPT-3.5 Turbo",
				CostPer1MIn:      0.5,
				CostPer1MOut:     1.5,
				ContextWindow:    16384,
				DefaultMaxTokens: 4096,
				CanReason:        false,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
		},
		DefaultHeaders: map[string]string{},
	}

	return openaiProvider
}
