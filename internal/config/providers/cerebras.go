package providers

import (
	"cmp"
	"os"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
)

// CerebrasProvider creates Cerebras provider if it doesn't exist.
func CerebrasProvider(providers []catwalk.Provider) catwalk.Provider {
	// Check if cerebras already exists
	for _, provider := range providers {
		if provider.ID == "cerebras" {
			return catwalk.Provider{}
		}
	}

	cerebrasProvider := catwalk.Provider{
		Name:                "Cerebras",
		ID:                  "cerebras",
		APIKey:              "$CEREBRAS_API_KEY",
		APIEndpoint:         cmp.Or(os.Getenv("CEREBRAS_API_ENDPOINT"), "https://api.cerebras.ai/v1"),
		Type:                "openai-compat",
		DefaultLargeModelID: "llama-3.3-70b",
		Models: []catwalk.Model{
			// Production Models
			{
				ID:               "llama-3.3-70b",
				Name:             "Llama 3.3 70B",
				CostPer1MIn:      0.85,
				CostPer1MOut:     1.2,
				ContextWindow:    128000,
				DefaultMaxTokens: 8192,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "llama-4-scout-17b-16e-instruct",
				Name:             "Llama 4 Scout (17B-16E MoE)",
				CostPer1MIn:      0.65,
				CostPer1MOut:     0.85,
				ContextWindow:    128000,
				DefaultMaxTokens: 8192,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "llama3.1-8b",
				Name:             "Llama 3.1 8B",
				CostPer1MIn:      0.1,
				CostPer1MOut:     0.1,
				ContextWindow:    128000,
				DefaultMaxTokens: 4096,
				CanReason:        false,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "qwen-3-32b",
				Name:             "Qwen 3 32B",
				CostPer1MIn:      0.4,
				CostPer1MOut:     0.8,
				ContextWindow:    131000,
				DefaultMaxTokens: 8192,
				CanReason:        false,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "gpt-oss-120b",
				Name:             "GPT OSS 120B",
				CostPer1MIn:      0.35,
				CostPer1MOut:     0.75,
				ContextWindow:    131000,
				DefaultMaxTokens: 2048,
				CanReason:        false,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			// Preview Models
			{
				ID:               "qwen-3-235b-a22b-instruct-2507",
				Name:             "Qwen 3 235B Instruct (Preview)",
				CostPer1MIn:      0.6,
				CostPer1MOut:     1.2,
				ContextWindow:    131000,
				DefaultMaxTokens: 16000,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "zai-glm-4.6",
				Name:             "Z.ai GLM 4.6 (Preview)",
				CostPer1MIn:      2.25,
				CostPer1MOut:     2.75,
				ContextWindow:    131000,
				DefaultMaxTokens: 8192,
				CanReason:        true,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
		},
		DefaultHeaders: map[string]string{},
	}

	return cerebrasProvider
}
