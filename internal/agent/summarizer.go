package agent

import (
	"github.com/nexora/nexora/internal/config"
)

// ModelProvider represents a model/provider pair for summarization
type ModelProvider struct {
	Provider string
	Model    string
}

// DetectFastestSummarizer determines the fastest available summarization model
// based on the provider configuration and known performance characteristics.
// Returns empty ModelProvider if no fast summarizer is available.
func DetectFastestSummarizer(cfg config.Config) ModelProvider {
	// Known fast summarizers in order of preference (fastest first):
	// 1. Cerebras llama-3.1-8b: ~2000 tokens/second
	// 2. xAI grok-3-mini: ~1200 tokens/second

	// Check for Cerebras provider first (highest preference)
	if cfg.Providers != nil {
		if provider, ok := cfg.Providers.Get("cerebras"); ok && !provider.Disable {
			return ModelProvider{
				Provider: "cerebras",
				Model:    "llama3.1-8b",
			}
		}

		// Check for xAI provider as second choice
		if provider, ok := cfg.Providers.Get("xai"); ok && !provider.Disable {
			return ModelProvider{
				Provider: "xai",
				Model:    "grok-3-mini",
			}
		}
	}

	// No fast summarizer available
	return ModelProvider{}
}

// IsFastSummarizer checks if the given provider/model combination is considered
// a "fast" summarizer (can handle >1000 tokens/second).
func IsFastSummarizer(provider, model string) bool {
	switch provider {
	case "cerebras":
		return model == "llama3.1-8b" || model == "llama-3.1-8b"
	case "xai":
		return model == "grok-3-mini" || model == "grok-3-mini-beta"
	default:
		return false
	}
}
