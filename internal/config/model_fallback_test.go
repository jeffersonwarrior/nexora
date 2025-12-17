package config

import (
	"testing"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
	"github.com/nexora/cli/internal/csync"
)

func TestModelsNeedSetup(t *testing.T) {
	cfg := &Config{
		modelsNeedSetup: true,
	}

	if !cfg.ModelsNeedSetup() {
		t.Error("Expected ModelsNeedSetup to return true")
	}

	cfg.modelsNeedSetup = false
	if cfg.ModelsNeedSetup() {
		t.Error("Expected ModelsNeedSetup to return false")
	}
}

func TestValidateModel(t *testing.T) {
	// Setup config with valid providers
	cfg := &Config{
		Providers: csync.NewMap[string, ProviderConfig](),
	}

	cfg.Providers.Set("openai", ProviderConfig{
		ID:     "openai",
		APIKey: "test-key",
		Models: []catwalk.Model{
			{ID: "gpt-4o", Name: "GPT-4"},
			{ID: "gpt-3.5-turbo", Name: "GPT-3.5"},
		},
	})

	// Test valid model
	cfg.Models = map[SelectedModelType]SelectedModel{
		SelectedModelTypeLarge: {Provider: "openai", Model: "gpt-4o"},
	}

	valid, err := cfg.validateModel(SelectedModelTypeLarge)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !valid {
		t.Error("Expected model to be valid")
	}

	// Test invalid provider
	cfg.Models[SelectedModelTypeLarge] = SelectedModel{Provider: "invalid", Model: "gpt-4o"}
	valid, err = cfg.validateModel(SelectedModelTypeLarge)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if valid {
		t.Error("Expected model with invalid provider to be invalid")
	}

	// Test invalid model
	cfg.Models[SelectedModelTypeLarge] = SelectedModel{Provider: "openai", Model: "invalid-model"}
	valid, err = cfg.validateModel(SelectedModelTypeLarge)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if valid {
		t.Error("Expected invalid model to be invalid")
	}

	// Test no model selected
	delete(cfg.Models, SelectedModelTypeLarge)
	valid, err = cfg.validateModel(SelectedModelTypeLarge)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if valid {
		t.Error("Expected no model to be invalid")
	}
}

func TestIsValidRecentModel(t *testing.T) {
	// Setup config with valid providers
	cfg := &Config{
		Providers: csync.NewMap[string, ProviderConfig](),
	}

	cfg.Providers.Set("openai", ProviderConfig{
		ID:     "openai",
		APIKey: "test-key",
		Models: []catwalk.Model{
			{ID: "gpt-4o", Name: "GPT-4"},
		},
	})

	// Test valid recent model
	validModel := SelectedModel{Provider: "openai", Model: "gpt-4o"}
	if !cfg.isValidRecentModel(SelectedModelTypeLarge, validModel) {
		t.Error("Expected valid recent model to be valid")
	}

	// Test invalid provider
	invalidProvider := SelectedModel{Provider: "invalid", Model: "gpt-4o"}
	if cfg.isValidRecentModel(SelectedModelTypeLarge, invalidProvider) {
		t.Error("Expected model with invalid provider to be invalid")
	}

	// Test invalid model
	invalidModel := SelectedModel{Provider: "openai", Model: "invalid-model"}
	if cfg.isValidRecentModel(SelectedModelTypeLarge, invalidModel) {
		t.Error("Expected invalid model to be invalid")
	}

	// Test empty values
	emptyProvider := SelectedModel{Provider: "", Model: "gpt-4o"}
	if cfg.isValidRecentModel(SelectedModelTypeLarge, emptyProvider) {
		t.Error("Expected model with empty provider to be invalid")
	}

	emptyModel := SelectedModel{Provider: "openai", Model: ""}
	if cfg.isValidRecentModel(SelectedModelTypeLarge, emptyModel) {
		t.Error("Expected model with empty model to be invalid")
	}
}
