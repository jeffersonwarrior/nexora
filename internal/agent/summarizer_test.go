package agent

import (
	"testing"

	"github.com/nexora/nexora/internal/config"
	"github.com/nexora/nexora/internal/csync"
)

func TestDetectFastestSummarizer(t *testing.T) {
	t.Run("Cerebras prioritized", func(t *testing.T) {
		providers := csync.NewMap[string, config.ProviderConfig]()
		providers.Set("cerebras", config.ProviderConfig{
			APIKey: "test-key",
		})
		providers.Set("xai", config.ProviderConfig{
			APIKey: "test-key",
		})

		cfg := config.Config{
			Providers: providers,
		}

		result := DetectFastestSummarizer(cfg)
		if result.Provider != "cerebras" {
			t.Errorf("Expected cerebras provider, got: %s", result.Provider)
		}
		if result.Model != "llama3.1-8b" {
			t.Errorf("Expected llama3.1-8b model, got: %s", result.Model)
		}
	})

	t.Run("xAI fallback", func(t *testing.T) {
		providers := csync.NewMap[string, config.ProviderConfig]()
		providers.Set("xai", config.ProviderConfig{
			APIKey: "test-key",
		})

		cfg := config.Config{
			Providers: providers,
		}

		result := DetectFastestSummarizer(cfg)
		if result.Provider != "xai" {
			t.Errorf("Expected xai provider, got: %s", result.Provider)
		}
		if result.Model != "grok-3-mini" {
			t.Errorf("Expected grok-3-mini model, got: %s", result.Model)
		}
	})

	t.Run("No fast providers", func(t *testing.T) {
		providers := csync.NewMap[string, config.ProviderConfig]()
		providers.Set("openai", config.ProviderConfig{})

		cfg := config.Config{
			Providers: providers,
		}

		result := DetectFastestSummarizer(cfg)
		if result.Provider != "" || result.Model != "" {
			t.Errorf("Expected empty result for non-fast providers, got: %v", result)
		}
	})
}

func TestIsFastSummarizer(t *testing.T) {
	t.Run("Cerebras llama3.1-8b is fast", func(t *testing.T) {
		if !IsFastSummarizer("cerebras", "llama3.1-8b") {
			t.Error("Expected cerebras llama3.1-8b to be considered fast")
		}
	})

	t.Run("xAI grok-3-mini is fast", func(t *testing.T) {
		if !IsFastSummarizer("xai", "grok-3-mini") {
			t.Error("Expected xai grok-3-mini to be considered fast")
		}
	})

	t.Run("OpenAI gpt-4 is not fast", func(t *testing.T) {
		if IsFastSummarizer("openai", "gpt-4") {
			t.Error("Did not expect openai gpt-4 to be considered fast")
		}
	})
}
