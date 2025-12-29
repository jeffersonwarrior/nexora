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
			// GPT-5 Series (Latest Flagship)
			{
				ID:               "gpt-5",
				Name:             "GPT-5",
				CostPer1MIn:      1.25,
				CostPer1MOut:     10.0,
				ContextWindow:    272000,
				DefaultMaxTokens: 16000,
				CanReason:        true,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "gpt-5-mini",
				Name:             "GPT-5 Mini",
				CostPer1MIn:      0.25,
				CostPer1MOut:     2.0,
				ContextWindow:    272000,
				DefaultMaxTokens: 16000,
				CanReason:        true,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "gpt-5-nano",
				Name:             "GPT-5 Nano",
				CostPer1MIn:      0.05,
				CostPer1MOut:     0.4,
				ContextWindow:    272000,
				DefaultMaxTokens: 8000,
				CanReason:        true,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			// GPT-4.1 Series (Million-Token Context)
			{
				ID:               "gpt-4.1",
				Name:             "GPT-4.1",
				CostPer1MIn:      2.0,
				CostPer1MOut:     8.0,
				ContextWindow:    1000000,
				DefaultMaxTokens: 16000,
				CanReason:        true,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "gpt-4.1-mini",
				Name:             "GPT-4.1 Mini",
				CostPer1MIn:      0.4,
				CostPer1MOut:     1.6,
				ContextWindow:    1000000,
				DefaultMaxTokens: 16000,
				CanReason:        true,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "gpt-4.1-nano",
				Name:             "GPT-4.1 Nano",
				CostPer1MIn:      0.1,
				CostPer1MOut:     0.4,
				ContextWindow:    1000000,
				DefaultMaxTokens: 8000,
				CanReason:        false,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			// o3/o4 Reasoning Models
			{
				ID:               "o3",
				Name:             "O3",
				CostPer1MIn:      0.4,
				CostPer1MOut:     1.6,
				ContextWindow:    200000,
				DefaultMaxTokens: 32000,
				CanReason:        true,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "o4-mini",
				Name:             "O4 Mini",
				CostPer1MIn:      1.1,
				CostPer1MOut:     4.4,
				ContextWindow:    200000,
				DefaultMaxTokens: 16000,
				CanReason:        true,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "o3-mini",
				Name:             "O3 Mini",
				CostPer1MIn:      1.1,
				CostPer1MOut:     4.4,
				ContextWindow:    200000,
				DefaultMaxTokens: 16000,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			// Legacy o1 series
			{
				ID:               "o1",
				Name:             "O1",
				CostPer1MIn:      15.0,
				CostPer1MOut:     60.0,
				ContextWindow:    200000,
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
			// GPT-4o Series
			{
				ID:               "gpt-4o",
				Name:             "GPT-4o",
				CostPer1MIn:      5.0,
				CostPer1MOut:     15.0,
				ContextWindow:    128000,
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
				ContextWindow:    128000,
				DefaultMaxTokens: 16000,
				CanReason:        false,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
		},
		DefaultHeaders: map[string]string{},
	}

	return openaiProvider
}
