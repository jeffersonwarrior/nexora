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
		if result.Model != "zai-glm-4.6" {
			t.Errorf("Expected zai-glm-4.6 model, got: %s", result.Model)
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
		if result.Model != "grok-4-1-fast" {
			t.Errorf("Expected grok-4-1-fast model, got: %s", result.Model)
		}
	})

	t.Run("Z.AI fallback", func(t *testing.T) {
		providers := csync.NewMap[string, config.ProviderConfig]()
		providers.Set("zai", config.ProviderConfig{
			APIKey: "test-key",
		})

		cfg := config.Config{
			Providers: providers,
		}

		result := DetectFastestSummarizer(cfg)
		if result.Provider != "zai" {
			t.Errorf("Expected zai provider, got: %s", result.Provider)
		}
		if result.Model != "glm-4.5-flash" {
			t.Errorf("Expected glm-4.5-flash model, got: %s", result.Model)
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
	t.Run("Cerebras zai-glm-4.6 is fast", func(t *testing.T) {
		if !IsFastSummarizer("cerebras", "zai-glm-4.6") {
			t.Error("Expected cerebras zai-glm-4.6 to be considered fast")
		}
	})

	t.Run("xAI grok-4-1-fast is fast", func(t *testing.T) {
		if !IsFastSummarizer("xai", "grok-4-1-fast") {
			t.Error("Expected xai grok-4-1-fast to be considered fast")
		}
	})

	t.Run("Z.AI glm-4.5-flash is fast", func(t *testing.T) {
		if !IsFastSummarizer("zai", "glm-4.5-flash") {
			t.Error("Expected zai glm-4.5-flash to be considered fast")
		}
	})

	t.Run("Synthetic minimax is fast", func(t *testing.T) {
		if !IsFastSummarizer("synthetic", "minimax/minimax-m2.1") {
			t.Error("Expected synthetic minimax/minimax-m2.1 to be considered fast")
		}
	})

	t.Run("OpenAI gpt-4 is not fast", func(t *testing.T) {
		if IsFastSummarizer("openai", "gpt-4") {
			t.Error("Did not expect openai gpt-4 to be considered fast")
		}
	})
}
