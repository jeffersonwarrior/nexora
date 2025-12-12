package config

import (
	"cmp"
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
	"github.com/nexora/cli/internal/home"
)

type ProviderClient interface {
	GetProviders() ([]catwalk.Provider, error)
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
	slog.Info("Saving provider data to disk", "path", path)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("failed to create directory for provider cache: %w", err)
	}

	data, err := json.MarshalIndent(providers, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal provider data: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write provider data to cache: %w", err)
	}
	return nil
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

	// Inject mistral provider if it doesn't exist
	providers = injectMistralProviders(providers)
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
		providers, err = catwalk.NewWithURL(pathOrUrl).GetProviders()
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

		// Inject mistral provider if it doesn't exist
		if providerErr == nil {
			providerList = injectMistralProviders(providerList)
		}
	})
	return providerList, providerErr
}

func loadProviders(autoUpdateDisabled bool, client ProviderClient, path string) ([]catwalk.Provider, error) {
	catwalkGetAndSave := func() ([]catwalk.Provider, error) {
		providers, err := client.GetProviders()
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

	switch {
	case autoUpdateDisabled:
		slog.Warn("Providers auto-update is disabled")

		if _, err := os.Stat(path); err == nil {
			slog.Warn("Using locally cached providers")
			return loadProvidersFromCache(path)
		}

		slog.Warn("Saving embedded providers to cache")
		providers := embedded.GetAll()
		// Inject mistral provider
		providers = injectMistralProviders(providers)
		if err := saveProvidersInCache(path, providers); err != nil {
			return nil, err
		}
		return providers, nil

	default:
		slog.Info("Fetching providers from Catwalk.", "path", path)

		providers, err := catwalkGetAndSave()
		if err != nil {
			catwalkUrl := fmt.Sprintf("%s/v2/providers", cmp.Or(os.Getenv("CATWALK_URL"), defaultCatwalkURL))
			return nil, fmt.Errorf("Nexora was unable to fetch an updated list of providers from %s. Consider setting NEXORA_DISABLE_PROVIDER_AUTO_UPDATE=1 to use the embedded providers bundled at the time of this Nexora release. You can also update providers manually. For more info see nexora update-providers --help. %w", catwalkUrl, err) //nolint:staticcheck
		}
		// Inject mistral provider
		providers = injectMistralProviders(providers)
		return providers, nil
	}
}

// injectMistralProviders adds Mistral provider to the providers list if it doesn't exist
func injectMistralProviders(providers []catwalk.Provider) []catwalk.Provider {
	slog.Info("Injecting Mistral provider", "existing_count", len(providers))

	// Check if mistral already exists
	for _, provider := range providers {
		if provider.ID == "mistral" {
			slog.Info("Mistral provider already exists")
			return providers
		}
	}

	// Create Mistral provider with the specified models
	mistralProvider := catwalk.Provider{
		Name:                "Mistral",
		ID:                  "mistral",
		APIKey:              "$MISTRAL_API_KEY",
		APIEndpoint:         cmp.Or(os.Getenv("MISTRAL_API_ENDPOINT"), "https://api.mistral.ai/v1"),
		Type:                "openai-compat",
		DefaultLargeModelID: "mistral-large-3-25-12",
		DefaultSmallModelID: "ministral-3-8b-25-12",
		Models: []catwalk.Model{
			{
				ID:               "mistral-large-3-25-12",
				Name:             "Mistral Large 3",
				CostPer1MIn:      2.0,
				CostPer1MOut:     6.0,
				ContextWindow:    131072,
				DefaultMaxTokens: 32000,
				CanReason:        true,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "mistral-medium-3-1-25-08",
				Name:             "Mistral Medium 3.1",
				CostPer1MIn:      1.5,
				CostPer1MOut:     4.5,
				ContextWindow:    131072,
				DefaultMaxTokens: 32000,
				CanReason:        true,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "mistral-small-3-2-25-06",
				Name:             "Mistral Small 3.2",
				CostPer1MIn:      0.2,
				CostPer1MOut:     0.6,
				ContextWindow:    128000,
				DefaultMaxTokens: 16000,
				CanReason:        true,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "ministral-3-14b-25-12",
				Name:             "Ministral 3 14B",
				CostPer1MIn:      0.15,
				CostPer1MOut:     0.45,
				ContextWindow:    128000,
				DefaultMaxTokens: 16000,
				CanReason:        true,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "ministral-3-8b-25-12",
				Name:             "Ministral 3 8B",
				CostPer1MIn:      0.1,
				CostPer1MOut:     0.3,
				ContextWindow:    128000,
				DefaultMaxTokens: 8000,
				CanReason:        false,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "ministral-3-3b-25-12",
				Name:             "Ministral 3 3B",
				CostPer1MIn:      0.05,
				CostPer1MOut:     0.15,
				ContextWindow:    128000,
				DefaultMaxTokens: 4000,
				CanReason:        false,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "devstral-2512",
				Name:             "Devstral 2",
				CostPer1MIn:      0.3,
				CostPer1MOut:     0.9,
				ContextWindow:    262144,
				DefaultMaxTokens: 16000,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "codestral-25-08",
				Name:             "Codestral",
				CostPer1MIn:      0.3,
				CostPer1MOut:     0.9,
				ContextWindow:    131072,
				DefaultMaxTokens: 32000,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "codestral-embed-25-05",
				Name:             "Codestral Embed",
				CostPer1MIn:      0.1,
				CostPer1MOut:     0.1,
				ContextWindow:    8192,
				DefaultMaxTokens: 8192,
				CanReason:        false,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "mistral-embed",
				Name:             "Mistral Embed",
				CostPer1MIn:      0.1,
				CostPer1MOut:     0.1,
				ContextWindow:    8192,
				DefaultMaxTokens: 8192,
				CanReason:        false,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
		},
		DefaultHeaders: map[string]string{},
	}

	slog.Info("Added Mistral provider with models", "model_count", len(mistralProvider.Models))
	return append(providers, mistralProvider)
}

// injectMiniMaxProviders adds MiniMax provider to the providers list if it doesn't exist
func injectMiniMaxProviders(providers []catwalk.Provider) []catwalk.Provider {
	slog.Info("Injecting MiniMax provider", "existing_count", len(providers))

	// Check if minimax already exists
	for _, provider := range providers {
		if provider.ID == "minimax" {
			slog.Info("MiniMax provider already exists")
			return providers
		}
	}

	// Create MiniMax provider with Anthropic-compatible models
	minimaxProvider := catwalk.Provider{
		Name:                "MiniMax",
		ID:                  "minimax",
		APIKey:              "$MINIMAX_API_KEY",
		APIEndpoint:         cmp.Or(os.Getenv("MINIMAX_API_ENDPOINT"), "https://api.minimax.io/anthropic"),
		Type:                "anthropic",
		DefaultLargeModelID: "MiniMax-M2",
		DefaultSmallModelID: "MiniMax-M2-Stable",
		Models: []catwalk.Model{
			{
				ID:               "MiniMax-M2",
				Name:             "MiniMax M2",
				CostPer1MIn:      0.5,
				CostPer1MOut:     1.5,
				ContextWindow:    200000,
				DefaultMaxTokens: 8000,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
			{
				ID:               "MiniMax-M2-Stable",
				Name:             "MiniMax M2 Stable",
				CostPer1MIn:      0.4,
				CostPer1MOut:     1.2,
				ContextWindow:    200000,
				DefaultMaxTokens: 8000,
				CanReason:        true,
				SupportsImages:   false,
				Options:          catwalk.ModelOptions{},
			},
		},
		DefaultHeaders: map[string]string{},
	}

	slog.Info("Added MiniMax provider with models", "model_count", len(minimaxProvider.Models))
	return append(providers, minimaxProvider)
}
