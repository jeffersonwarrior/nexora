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
		DefaultLargeModelID: "gemini-2.5-pro",
		Models: []catwalk.Model{
			// Gemini 2.5 Series
			{
				ID:               "gemini-2.5-pro",
				Name:             "Gemini 2.5 Pro",
				CostPer1MIn:      1.25,
				CostPer1MOut:     10.0,
				ContextWindow:    1000000,
				DefaultMaxTokens: 32000,
				CanReason:        true,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "gemini-2.5-flash",
				Name:             "Gemini 2.5 Flash",
				CostPer1MIn:      0.3,
				CostPer1MOut:     2.5,
				ContextWindow:    1000000,
				DefaultMaxTokens: 16000,
				CanReason:        true,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "gemini-2.5-flash-lite",
				Name:             "Gemini 2.5 Flash-Lite",
				CostPer1MIn:      0.1,
				CostPer1MOut:     0.4,
				ContextWindow:    1000000,
				DefaultMaxTokens: 16000,
				CanReason:        false,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			// Gemini 2.0 Series
			{
				ID:               "gemini-2.0-flash",
				Name:             "Gemini 2.0 Flash",
				CostPer1MIn:      0.1,
				CostPer1MOut:     0.3,
				ContextWindow:    1000000,
				DefaultMaxTokens: 16000,
				CanReason:        false,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "gemini-2.0-flash-thinking-exp",
				Name:             "Gemini 2.0 Flash (Thinking)",
				CostPer1MIn:      0.0,
				CostPer1MOut:     0.0,
				ContextWindow:    1000000,
				DefaultMaxTokens: 32000,
				CanReason:        true,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "gemini-2.0-pro-exp",
				Name:             "Gemini 2.0 Pro (Experimental)",
				CostPer1MIn:      0.0,
				CostPer1MOut:     0.0,
				ContextWindow:    2000000,
				DefaultMaxTokens: 32000,
				CanReason:        true,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			// Gemini 3.0 Series (Preview)
			{
				ID:               "gemini-3-pro-preview",
				Name:             "Gemini 3 Pro (Preview)",
				CostPer1MIn:      2.0,
				CostPer1MOut:     12.0,
				ContextWindow:    200000,
				DefaultMaxTokens: 32000,
				CanReason:        true,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "gemini-3-flash-preview",
				Name:             "Gemini 3 Flash (Preview)",
				CostPer1MIn:      0.3,
				CostPer1MOut:     2.5,
				ContextWindow:    1000000,
				DefaultMaxTokens: 16000,
				CanReason:        true,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			// Gemini 1.5 Series (Legacy)
			{
				ID:               "gemini-1.5-pro-002",
				Name:             "Gemini 1.5 Pro",
				CostPer1MIn:      1.25,
				CostPer1MOut:     5.0,
				ContextWindow:    2000000,
				DefaultMaxTokens: 8000,
				CanReason:        false,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "gemini-1.5-flash-002",
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
