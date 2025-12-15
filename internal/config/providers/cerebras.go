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
		DefaultSmallModelID: "llama3.1-8b",
		Models: []catwalk.Model{
			// GPT Open Source 120B - Fastest inference
			{
				ID:               "gpt-oss-120b",
				Name:             "GPT OSS 120B",
				CostPer1MIn:      0.35,
				CostPer1MOut:     0.75,
				ContextWindow:    4096,
				DefaultMaxTokens: 2048,
				CanReason:        false,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			// Llama 3.3 70B - Latest Llama
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
			// Llama 3.1 8B - Small Llama
			{
				ID:               "llama3.1-8b",
				Name:             "Llama 3.1 8B",
				CostPer1MIn:      0.3,
				CostPer1MOut:     0.6,
				ContextWindow:    128000,
				DefaultMaxTokens: 4096,
				CanReason:        false,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			// Qwen 3 32B
			{
				ID:               "qwen-3-32b",
				Name:             "Qwen 3 32B",
				CostPer1MIn:      0.35,
				CostPer1MOut:     0.75,
				ContextWindow:    262144,
				DefaultMaxTokens: 8192,
				CanReason:        false,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			// Qwen 3 235B - Large Qwen
			{
				ID:               "qwen-3-235b",
				Name:             "Qwen 3 235B",
				CostPer1MIn:      0.95,
				CostPer1MOut:     1.3,
				ContextWindow:    262144,
				DefaultMaxTokens: 16000,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			// Z.AI GLM 4.6 via Cerebras
			{
				ID:               "zai-glm-4.6",
				Name:             "GLM 4.6 (via Cerebras)",
				CostPer1MIn:      0.6,
				CostPer1MOut:     2.2,
				ContextWindow:    200000,
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
