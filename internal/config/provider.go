package config

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
	"github.com/charmbracelet/catwalk/pkg/embedded"
	"github.com/nexora/cli/internal/config/providers"
	"github.com/nexora/cli/internal/home"
)

type ProviderClient interface {
	GetProviders(context.Context, string) ([]catwalk.Provider, error)
}

var (
	providerOnce sync.Once
	providerList []catwalk.Provider
	providerErr  error
)

// file to cache provider data
func providerCacheFileData() string {
	xdgDataHome := os.Getenv("XDG_DATA_HOME")
	if xdgDataHome != "" {
		return filepath.Join(xdgDataHome, appName, "providers.json")
	}

	// return the path to the main data directory
	// for windows, it should be in `%LOCALAPPDATA%/nexora/`
	// for linux and macOS, it should be in `$HOME/.local/share/nexora/`
	if runtime.GOOS == "windows" {
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			localAppData = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local")
		}
		return filepath.Join(localAppData, appName, "providers.json")
	}

	return filepath.Join(home.Dir(), ".local", "share", appName, "providers.json")
}

func saveProvidersInCache(path string, providers []catwalk.Provider) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("failed to create provider cache directory: %w", err)
	}

	data, err := json.MarshalIndent(providers, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal providers: %w", err)
	}

	return os.WriteFile(path, data, 0o644)
}

func loadProvidersFromCache(path string) ([]catwalk.Provider, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read provider cache file: %w", err)
	}

	var providers []catwalk.Provider
	if err := json.Unmarshal(data, &providers); err != nil {
		return nil, fmt.Errorf("failed to unmarshal provider data from cache: %w", err)
	}

	// Inject custom providers if they don't exist
	providers = injectCustomProviders(providers)
	return providers, nil
}

func UpdateProviders(pathOrUrl string) error {
	var providers []catwalk.Provider
	pathOrUrl = cmp.Or(pathOrUrl, os.Getenv("CATWALK_URL"), defaultCatwalkURL)

	switch {
	case pathOrUrl == "embedded":
		providers = embedded.GetAll()
	case strings.HasPrefix(pathOrUrl, "http://") || strings.HasPrefix(pathOrUrl, "https://"):
		var err error
		providers, err = catwalk.NewWithURL(pathOrUrl).GetProviders(context.Background(), "")
		if err != nil {
			return fmt.Errorf("failed to fetch providers from Catwalk: %w", err)
		}
	default:
		content, err := os.ReadFile(pathOrUrl)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
		if err := json.Unmarshal(content, &providers); err != nil {
			return fmt.Errorf("failed to unmarshal provider data: %w", err)
		}
		if len(providers) == 0 {
			return fmt.Errorf("no providers found in the provided source")
		}
	}

	cachePath := providerCacheFileData()
	if err := saveProvidersInCache(cachePath, providers); err != nil {
		return fmt.Errorf("failed to save providers to cache: %w", err)
	}

	slog.Info("Providers updated successfully", "count", len(providers), "from", pathOrUrl, "to", cachePath)
	return nil
}

func Providers(cfg *Config) ([]catwalk.Provider, error) {
	providerOnce.Do(func() {
		catwalkURL := cmp.Or(os.Getenv("CATWALK_URL"), defaultCatwalkURL)
		client := catwalk.NewWithURL(catwalkURL)
		path := providerCacheFileData()

		autoUpdateDisabled := cfg.Options.DisableProviderAutoUpdate
		providerList, providerErr = loadProviders(autoUpdateDisabled, client, path)
	})
	return providerList, providerErr
}

func loadProviders(autoUpdateDisabled bool, client ProviderClient, path string) ([]catwalk.Provider, error) {
	catwalkGetAndSave := func() ([]catwalk.Provider, error) {
		providers, err := client.GetProviders(context.Background(), "")
		if err != nil {
			return nil, fmt.Errorf("failed to fetch providers from catwalk: %w", err)
		}
		if len(providers) == 0 {
			return nil, fmt.Errorf("empty providers list from catwalk")
		}
		if err := saveProvidersInCache(path, providers); err != nil {
			return nil, err
		}
		return providers, nil
	}

	// Try cache first if auto-update is disabled
	if autoUpdateDisabled {
		slog.Debug("Auto-update disabled, loading from cache")
		providers, err := loadProvidersFromCache(path)
		if err == nil {
			return providers, nil
		}
		slog.Warn("Failed to load from cache, falling back to embedded providers", "error", err)
		providers = embedded.GetAll()
		providers = injectCustomProviders(providers)
		if err := saveProvidersInCache(path, providers); err != nil {
			return nil, err
		}
		return providers, nil
	}

	// Try to fetch from Catwalk with cache fallback
	if providers, err := catwalkGetAndSave(); err == nil {
		// Inject custom providers when we get providers from Catwalk
		providers = injectCustomProviders(providers)
		return providers, nil
	}

	// Fallback to embedded providers
	slog.Info("Using embedded providers as fallback")
	providers := embedded.GetAll()
	// Don't cache embedded providers as they should be updated via releases
	return providers, nil
}

// injectCustomProviders consolidates all custom provider injections.
// Configured providers are prepended to the list (sorted to top).
func injectCustomProviders(providerList []catwalk.Provider) []catwalk.Provider {
	injectors := []func([]catwalk.Provider) catwalk.Provider{
		// Mistral variants (General, Devstral, Codestral)
		providers.MistralGeneralProvider,
		providers.MistralDevstralProvider,
		providers.MistralCodestralProvider,
		// New major providers
		providers.OpenAIProvider,
		providers.AnthropicProvider,
		providers.GeminiProvider,
		providers.ZAIProvider,
		providers.CerebrasProvider,
		providers.MistralNativeProvider,
		providers.RentalH200Provider,
		// Existing providers
		providers.NexoraProvider,
		providers.XAIProvider,
		providers.MiniMaxProvider,
	}

	// Collect injected providers
	injectedProviders := []catwalk.Provider{}
	injectedCount := 0
	for _, injector := range injectors {
		if p := injector(providerList); p.ID != "" {
			injectedProviders = append(injectedProviders, p)
			injectedCount++
		}
	}

	// Prepend injected providers to the list (configured providers on top)
	if injectedCount > 0 {
		injectedProviders = append(injectedProviders, providerList...)
		providerList = injectedProviders
		slog.Info("Injected custom providers", "count", injectedCount, "total", len(providerList))
	}

	return providerList
}
