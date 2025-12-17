package providers

import (
	"cmp"
	"os"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
)

// MistralNativeProvider creates native Mistral provider using OpenAI-compatible API with Mistral-specific parameters.
func MistralNativeProvider(providers []catwalk.Provider) catwalk.Provider {
	// Check if mistral already exists
	for _, provider := range providers {
		if provider.ID == "mistral" {
			return catwalk.Provider{}
		}
	}

	return catwalk.Provider{
		Name:                "Mistral AI (Native)",
		ID:                  "mistral",
		APIKey:              "$MISTRAL_API_KEY",
		APIEndpoint:         cmp.Or(os.Getenv("MISTRAL_API_ENDPOINT"), "https://api.mistral.ai/v1"),
		Type:                "openaicompat", // Uses openaicompat but with Mistral-specific ExtraBody
		DefaultLargeModelID: "devstral-2512",
		DefaultSmallModelID: "devstral-small-2512",
		ExtraBody: map[string]any{
			"tool_choice":          "auto",
			"prompt_mode":          "reasoning",
			"parallel_tool_calls":  true,
			"safe_prompt":          false,
		},
		Models: []catwalk.Model{
			// Devstral 2 - Agentic coding (72.2% SWE-bench)
			{
				ID:               "devstral-2512",
				Name:             "Devstral 2 (123B)",
				CostPer1MIn:      0.4,
				CostPer1MOut:     2.0,
				ContextWindow:    262144,
				DefaultMaxTokens: 16000,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			// Devstral Small 2 - Lightweight agentic tasks
			{
				ID:               "devstral-small-2512",
				Name:             "Devstral Small 2 (24B)",
				CostPer1MIn:      0.2,
				CostPer1MOut:     0.6,
				ContextWindow:    262144,
				DefaultMaxTokens: 8000,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
		},
		DefaultHeaders: map[string]string{},
	}
}