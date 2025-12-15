package providers

import (
	"github.com/charmbracelet/catwalk/pkg/catwalk"
)

// NexoraProvider creates Nexora provider if it doesn't exist.
func NexoraProvider(providers []catwalk.Provider) catwalk.Provider {
	// Check if nexora already exists
	for _, provider := range providers {
		if provider.ID == "nexora" {
			return catwalk.Provider{}
		}
	}

	nexoraProvider := catwalk.Provider{
		Name:                "Nexora",
		ID:                  "nexora",
		APIKey:              "",
		APIEndpoint:         "http://localhost:9000/v1",
		Type:                "openai-compat",
		DefaultLargeModelID: "devstral-small-2",
		DefaultSmallModelID: "devstral-small-2",
		Models: []catwalk.Model{
			{
				ID:               "devstral-small-2",
				Name:             "Devstral Small 2",
				CostPer1MIn:      0.3,
				CostPer1MOut:     0.9,
				ContextWindow:    262144,
				DefaultMaxTokens: 16000,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
		},
		DefaultHeaders: map[string]string{},
	}

	return nexoraProvider
}
