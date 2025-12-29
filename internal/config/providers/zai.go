package providers

import (
	"cmp"
	"os"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
)

// ZAIProvider creates Z.AI (GLM) provider if it doesn't exist.
func ZAIProvider(providers []catwalk.Provider) catwalk.Provider {
	// Check if zai already exists
	for _, provider := range providers {
		if provider.ID == "zai" {
			return catwalk.Provider{}
		}
	}

	zaiProvider := catwalk.Provider{
		Name:                "Z.AI",
		ID:                  "zai",
		APIKey:              "$ZAI_API_KEY",
		APIEndpoint:         cmp.Or(os.Getenv("ZAI_API_ENDPOINT"), "https://api.z.ai/api/paas/v4"),
		Type:                "openai-compat",
		DefaultLargeModelID: "glm-4.7",
		Models: []catwalk.Model{
			// GLM-4.7 - Latest flagship
			{
				ID:               "glm-4.7",
				Name:             "GLM-4.7",
				CostPer1MIn:      0.6,
				CostPer1MOut:     2.2,
				ContextWindow:    204800,
				DefaultMaxTokens: 16000,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			// GLM-4.6 - Enhanced reasoning
			{
				ID:               "glm-4.6",
				Name:             "GLM-4.6",
				CostPer1MIn:      0.6,
				CostPer1MOut:     2.2,
				ContextWindow:    200000,
				DefaultMaxTokens: 16000,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			// GLM-4.5 - Stable mid-tier
			{
				ID:               "glm-4.5",
				Name:             "GLM-4.5",
				CostPer1MIn:      0.6,
				CostPer1MOut:     2.2,
				ContextWindow:    128000,
				DefaultMaxTokens: 16000,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			// GLM-4.5-Air - Budget variant
			{
				ID:               "glm-4.5-air",
				Name:             "GLM-4.5-Air",
				CostPer1MIn:      0.2,
				CostPer1MOut:     1.1,
				ContextWindow:    128000,
				DefaultMaxTokens: 8000,
				CanReason:        false,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			// GLM Vision Models
			{
				ID:               "glm-4.6v",
				Name:             "GLM-4.6V (Vision)",
				CostPer1MIn:      0.3,
				CostPer1MOut:     0.9,
				ContextWindow:    128000,
				DefaultMaxTokens: 16000,
				CanReason:        true,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "glm-4.5v",
				Name:             "GLM-4.5V (Vision)",
				CostPer1MIn:      0.6,
				CostPer1MOut:     1.8,
				ContextWindow:    128000,
				DefaultMaxTokens: 16000,
				CanReason:        true,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			// GLM-4.5-Flash - Free model
			{
				ID:               "glm-4.5-flash",
				Name:             "GLM-4.5-Flash (Free)",
				CostPer1MIn:      0.0,
				CostPer1MOut:     0.0,
				ContextWindow:    128000,
				DefaultMaxTokens: 8000,
				CanReason:        false,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			// GLM-4.6v-Flash - Free vision model
			{
				ID:               "glm-4.6v-flash",
				Name:             "GLM-4.6V-Flash (Free Vision)",
				CostPer1MIn:      0.0,
				CostPer1MOut:     0.0,
				ContextWindow:    128000,
				DefaultMaxTokens: 8000,
				CanReason:        false,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
		},
		DefaultHeaders: map[string]string{},
	}

	return zaiProvider
}
