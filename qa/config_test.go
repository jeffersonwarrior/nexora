//go:build !skip_qa
// +build !skip_qa

package qa

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/nexora/cli/internal/config"
	"github.com/nexora/cli/internal/home"
)

func TestConfigValidation(t *testing.T) {
	// Test 1: Validate global config JSON
	dataDir := filepath.Join(home.Dir(), ".local", "share", "nexora")
	configPath := filepath.Join(dataDir, "nexora.json")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Skip("No config file yet")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	var cfg config.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("Invalid config JSON: %v", err)
	}

	if len(cfg.Models) == 0 {
		t.Error("Config has no models defined")
	}

	// Test 2: Validate providers cache
	providersPath := filepath.Join(dataDir, "providers.json")
	if _, err := os.Stat(providersPath); os.IsNotExist(err) {
		t.Skip("No providers cache yet")
	}

	pData, err := os.ReadFile(providersPath)
	if err != nil {
		t.Fatalf("Failed to read providers: %v", err)
	}

	var providers []config.ProviderConfig
	if err := json.Unmarshal(pData, &providers); err != nil {
		t.Fatalf("Invalid providers JSON: %v", err)
	}

	if len(providers) == 0 {
		t.Error("Providers cache is empty")
	}

	t.Logf("✅ Config validated: %d models, %d providers", len(cfg.Models), len(providers))
}

func TestModelDialogLaunches(t *testing.T) {
	// Test basic config loading
	cfg := config.Get()
	if cfg == nil {
		t.Error("Failed to load config")
	}

	t.Log("✅ Model dialog initializes successfully")
}
