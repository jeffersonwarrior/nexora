package providers

import (
	"cmp"
	"os"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
)

// CerebrasCodeProvider creates Cerebras Code subscription provider with GLM-4.6 only.
// Cerebras Code subscriptions have limited access to only the GLM-4.6 model.
func CerebrasCodeProvider(providers []catwalk.Provider) catwalk.Provider {
	// Check if cerebras-code already exists
	for _, provider := range providers {
		if provider.ID == "cerebras-code" {
			return catwalk.Provider{}
		}
	}

	cerebrasCodeProvider := catwalk.Provider{
		Name:                "Cerebras Code",
		ID:                  "cerebras-code",
		APIKey:              "$CEREBRAS_API_KEY",
		APIEndpoint:         cmp.Or(os.Getenv("CEREBRAS_API_ENDPOINT"), "https://api.cerebras.ai/v1"),
		Type:                "openai-compat",
		DefaultLargeModelID: "zai-glm-4.6",
		Models: []catwalk.Model{
			// GLM-4.6 - Only model available on Cerebras Code subscription
			{
				ID:               "zai-glm-4.6",
				Name:             "Z.ai GLM 4.6 (Cerebras Code)",
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

	return cerebrasCodeProvider
}
