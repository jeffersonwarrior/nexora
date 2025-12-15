package providers

import (
	"cmp"
	"os"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
)

// GeminiProvider creates Google Gemini provider if it doesn't exist.
func GeminiProvider(providers []catwalk.Provider) catwalk.Provider {
	// Check if gemini already exists
	for _, provider := range providers {
		if provider.ID == "gemini" {
			return catwalk.Provider{}
		}
	}

	geminiProvider := catwalk.Provider{
		Name:                "Google Gemini",
		ID:                  "gemini",
		APIKey:              "$GEMINI_API_KEY",
		APIEndpoint:         cmp.Or(os.Getenv("GEMINI_API_ENDPOINT"), "https://generativelanguage.googleapis.com/v1beta/models"),
		Type:                "openai-compat",
		DefaultLargeModelID: "gemini-3-pro",
		DefaultSmallModelID: "gemini-2-5-flash",
		Models: []catwalk.Model{
			// Gemini 3 Pro - Latest flagship (Note: tiered pricing, using lowest tier)
			{
				ID:               "gemini-3-pro",
				Name:             "Gemini 3 Pro",
				CostPer1MIn:      2.0,
				CostPer1MOut:     12.0,
				ContextWindow:    200000,
				DefaultMaxTokens: 32000,
				CanReason:        true,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			// Gemini 2.5 Flash - Fast and cheap
			{
				ID:               "gemini-2-5-flash",
				Name:             "Gemini 2.5 Flash",
				CostPer1MIn:      0.075,
				CostPer1MOut:     0.3,
				ContextWindow:    1000000,
				DefaultMaxTokens: 16000,
				CanReason:        false,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			// Gemini 2.0 Flash Experimental
			{
				ID:               "gemini-2-0-flash-exp",
				Name:             "Gemini 2.0 Flash (Experimental)",
				CostPer1MIn:      0.075,
				CostPer1MOut:     0.3,
				ContextWindow:    1000000,
				DefaultMaxTokens: 16000,
				CanReason:        false,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			// Gemini 2.0 Flash Thinking (Extended Thinking)
			{
				ID:               "gemini-2-0-flash-thinking-exp",
				Name:             "Gemini 2.0 Flash (Thinking)",
				CostPer1MIn:      5.0,
				CostPer1MOut:     20.0,
				ContextWindow:    100000,
				DefaultMaxTokens: 16000,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			// Gemini 1.5 Pro
			{
				ID:               "gemini-1-5-pro",
				Name:             "Gemini 1.5 Pro",
				CostPer1MIn:      1.25,
				CostPer1MOut:     5.0,
				ContextWindow:    2000000,
				DefaultMaxTokens: 8000,
				CanReason:        false,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			// Gemini 1.5 Flash
			{
				ID:               "gemini-1-5-flash",
				Name:             "Gemini 1.5 Flash",
				CostPer1MIn:      0.075,
				CostPer1MOut:     0.3,
				ContextWindow:    1000000,
				DefaultMaxTokens: 8000,
				CanReason:        false,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
		},
		DefaultHeaders: map[string]string{},
	}

	return geminiProvider
}
