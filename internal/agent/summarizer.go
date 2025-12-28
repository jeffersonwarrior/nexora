package agent

import (
	"strings"

	"github.com/nexora/nexora/internal/config"
)

// ModelProvider represents a model/provider pair for summarization
type ModelProvider struct {
	Provider string
	Model    string
}

// FastSummarizerModels defines the explicit fast models for summarization/compaction.
// These are high-throughput models optimized for fast inference.
// Order: fastest inference first.
var FastSummarizerModels = []ModelProvider{
	// Cerebras GLM 4.6 - ~2000+ tokens/sec via Cerebras' custom silicon
	{Provider: "cerebras", Model: "zai-glm-4.6"},
	// Grok 4.1 Fast - non-thinking mode, optimized for speed
	{Provider: "xai", Model: "grok-4-1-fast"},
	// Z.AI GLM 4.5 Flash - free tier, fast inference
	{Provider: "zai", Model: "glm-4.5-flash"},
	// Z.AI GLM 4.5 Air - budget tier fallback
	{Provider: "zai", Model: "glm-4.5-air"},
	// Synthetic MiniMax - fast via aggregator
	{Provider: "synthetic", Model: "minimax/minimax-m2.1"},
}

// DetectFastestSummarizer finds the first available fast summarizer model.
// Iterates through FastSummarizerModels in priority order and returns
// the first one whose provider is configured and enabled.
func DetectFastestSummarizer(cfg config.Config) ModelProvider {
	if cfg.Providers == nil {
		return ModelProvider{}
	}

	for _, mp := range FastSummarizerModels {
		cfgProvider, ok := cfg.Providers.Get(mp.Provider)
		if !ok || cfgProvider.Disable {
			continue
		}

		// Provider is configured and enabled, use this fast model
		return mp
	}

	return ModelProvider{}
}

// IsFastSummarizer checks if the given provider/model is a known fast summarizer.
func IsFastSummarizer(provider, model string) bool {
	provider = strings.ToLower(provider)
	model = strings.ToLower(model)

	for _, mp := range FastSummarizerModels {
		if strings.ToLower(mp.Provider) == provider &&
			strings.Contains(model, strings.ToLower(mp.Model)) {
			return true
		}
	}
	return false
}
