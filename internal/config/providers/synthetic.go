package providers

import (
	"cmp"
	"os"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
)

// SyntheticProvider creates Synthetic.new provider if it doesn't exist.
// Synthetic.new is an OpenRouter-compatible aggregator providing access to multiple models.
func SyntheticProvider(providers []catwalk.Provider) catwalk.Provider {
	// Check if synthetic already exists
	for _, provider := range providers {
		if provider.ID == "synthetic" {
			return catwalk.Provider{}
		}
	}

	syntheticProvider := catwalk.Provider{
		Name:                "Synthetic",
		ID:                  "synthetic",
		APIKey:              "$SYNTHETIC_API_KEY",
		APIEndpoint:         cmp.Or(os.Getenv("SYNTHETIC_API_ENDPOINT"), "https://api.synthetic.new/v1"),
		Type:                "openai-compat",
		DefaultLargeModelID: "minimax/minimax-m2.1",
		DefaultSmallModelID: "minimax/minimax-m2.1",
		Models: []catwalk.Model{
			{
				ID:               "minimax/minimax-m2.1",
				Name:             "MiniMax M2.1 (Synthetic)",
				CostPer1MIn:      0.3,
				CostPer1MOut:     1.2,
				ContextWindow:    204800,
				DefaultMaxTokens: 8000,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "hf:zai-org/GLM-4.6",
				Name:             "Z.AI GLM 4.6 (Synthetic)",
				CostPer1MIn:      0.6,
				CostPer1MOut:     2.2,
				ContextWindow:    198000,
				DefaultMaxTokens: 8192,
				CanReason:        true,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
		},
		DefaultHeaders: map[string]string{},
	}

	return syntheticProvider
}
