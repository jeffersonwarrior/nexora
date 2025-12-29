package providers

import (
	"cmp"
	"os"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
)

// DeepSeekProvider creates DeepSeek provider if it doesn't exist.
func DeepSeekProvider(providers []catwalk.Provider) catwalk.Provider {
	// Check if deepseek already exists
	for _, provider := range providers {
		if provider.ID == "deepseek" {
			return catwalk.Provider{}
		}
	}

	deepseekProvider := catwalk.Provider{
		Name:                "DeepSeek",
		ID:                  "deepseek",
		APIKey:              "$DEEPSEEK_API_KEY",
		APIEndpoint:         cmp.Or(os.Getenv("DEEPSEEK_API_ENDPOINT"), "https://api.deepseek.com"),
		Type:                "openai-compat",
		DefaultLargeModelID: "deepseek-chat",
		Models: []catwalk.Model{
			// DeepSeek V3.2 Chat - Latest flagship (non-thinking)
			{
				ID:               "deepseek-chat",
				Name:             "DeepSeek V3.2 Chat",
				CostPer1MIn:      0.28,
				CostPer1MOut:     0.42,
				ContextWindow:    131072,
				DefaultMaxTokens: 8192,
				CanReason:        false,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			// DeepSeek V3.2 Reasoner - Advanced reasoning (thinking mode)
			{
				ID:               "deepseek-reasoner",
				Name:             "DeepSeek V3.2 Reasoner",
				CostPer1MIn:      0.28,
				CostPer1MOut:     0.42,
				ContextWindow:    131072,
				DefaultMaxTokens: 32768,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
		},
		DefaultHeaders: map[string]string{},
	}

	return deepseekProvider
}
