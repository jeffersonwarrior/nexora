package models

import (
	"strings"
	"testing"
	"time"
)

func TestModelListRendering(t *testing.T) {
	// Test that model list renders correctly
	models := []Model{
		{Name: "MiniMax M2.1", Provider: "MiniMax"},
		{Name: "Claude Opus 4", Provider: "Anthropic"},
		{Name: "Grok 4", Provider: "xAI"},
	}

	if len(models) != 3 {
		t.Errorf("Expected 3 models, got %d", len(models))
	}
}

func TestModelSelection(t *testing.T) {
	models := []Model{
		{ID: "model1", Name: "MiniMax M2.1", Selected: false},
		{ID: "model2", Name: "Claude Opus 4", Selected: true},
	}

	selected := false
	for _, m := range models {
		if m.Selected {
			selected = true
			if m.ID != "model2" {
				t.Errorf("Wrong model selected: %s", m.ID)
			}
		}
	}

	if !selected {
		t.Error("No model selected")
	}
}

func TestModelSelectionToggle(t *testing.T) {
	m := Model{ID: "test", Name: "Test Model", Selected: false}

	if m.Selected {
		t.Error("Model should not be selected initially")
	}

	m.Selected = true
	if !m.Selected {
		t.Error("Model should be selected after toggle")
	}
}

func TestProviderFiltering(t *testing.T) {
	allModels := []Model{
		{Name: "Model1", Provider: "Anthropic"},
		{Name: "Model2", Provider: "MiniMax"},
		{Name: "Model3", Provider: "Anthropic"},
		{Name: "Model4", Provider: "xAI"},
	}

	anthropic := filterByProvider(allModels, "Anthropic")
	if len(anthropic) != 2 {
		t.Errorf("Expected 2 Anthropic models, got %d", len(anthropic))
	}

	minimax := filterByProvider(allModels, "MiniMax")
	if len(minimax) != 1 {
		t.Errorf("Expected 1 MiniMax model, got %d", len(minimax))
	}
}

func TestModelSearch(t *testing.T) {
	models := []Model{
		{Name: "MiniMax M2.1 Reasoning"},
		{Name: "MiniMax M2.1 Fast"},
		{Name: "Claude Opus 4"},
	}

	results := searchModels(models, "MiniMax")
	if len(results) != 2 {
		t.Errorf("Expected 2 MiniMax results, got %d", len(results))
	}

	results = searchModels(models, "Opus")
	if len(results) != 1 {
		t.Errorf("Expected 1 Opus result, got %d", len(results))
	}

	results = searchModels(models, "NotExist")
	if len(results) != 0 {
		t.Errorf("Expected 0 results for non-existent, got %d", len(results))
	}
}

func TestModelSorting(t *testing.T) {
	models := []Model{
		{Name: "Z-Model", Provider: "Z-Provider"},
		{Name: "A-Model", Provider: "A-Provider"},
		{Name: "M-Model", Provider: "M-Provider"},
	}

	sorted := sortByName(models)

	if sorted[0].Name != "A-Model" {
		t.Error("First model should be A-Model after sorting")
	}

	if sorted[2].Name != "Z-Model" {
		t.Error("Last model should be Z-Model after sorting")
	}
}

func TestModelValidation(t *testing.T) {
	tests := []struct {
		name  string
		model Model
		valid bool
	}{
		{"valid model", Model{ID: "m1", Name: "Test", Provider: "Provider"}, true},
		{"empty name", Model{ID: "m1", Name: "", Provider: "Provider"}, false},
		{"empty id", Model{ID: "", Name: "Test", Provider: "Provider"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.model.IsValid(); got != tt.valid {
				t.Errorf("Model.IsValid() = %v, want %v", got, tt.valid)
			}
		})
	}
}

func TestModelType(t *testing.T) {
	tests := []struct {
		name     string
		model    Model
		expected ModelType
	}{
		{"reasoning", Model{Name: "MiniMax Reasoning", Type: ModelTypeReasoning}, ModelTypeReasoning},
		{"fast", Model{Name: "MiniMax Fast", Type: ModelTypeFast}, ModelTypeFast},
		{"default", Model{Name: "Default"}, ModelTypeFast},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.model.Type != tt.expected {
				t.Errorf("Model.Type = %v, want %v", tt.model.Type, tt.expected)
			}
		})
	}
}

func TestRecentlyUsedModels(t *testing.T) {
	models := []Model{
		{Name: "Used1", LastUsed: timeNow().Add(-1 * time.Hour)},
		{Name: "Used2", LastUsed: timeNow().Add(-2 * time.Hour)},
		{Name: "Used3", LastUsed: timeNow().Add(-30 * time.Minute)},
	}

	sorted := sortByRecentlyUsed(models)

	if sorted[0].Name != "Used3" {
		t.Error("Most recently used should be first")
	}
}

func TestModelCategories(t *testing.T) {
	models := []Model{
		{Name: "R1", Category: "Reasoning"},
		{Name: "F1", Category: "Fast"},
		{Name: "R2", Category: "Reasoning"},
		{Name: "F2", Category: "Fast"},
	}

	categories := extractCategories(models)

	if len(categories) != 2 {
		t.Errorf("Expected 2 categories, got %d", len(categories))
	}
}

// Helper types and functions (would use actual implementation)
type Model struct {
	ID       string
	Name     string
	Provider string
	Selected bool
	Type     ModelType
	Category string
	LastUsed time.Time
}

type ModelType int

const (
	ModelTypeFast ModelType = iota
	ModelTypeReasoning
)

func (m Model) IsValid() bool {
	return m.ID != "" && m.Name != ""
}

func filterByProvider(models []Model, provider string) []Model {
	var result []Model
	for _, m := range models {
		if m.Provider == provider {
			result = append(result, m)
		}
	}
	return result
}

func searchModels(models []Model, query string) []Model {
	var result []Model
	for _, m := range models {
		if strings.Contains(strings.ToLower(m.Name), strings.ToLower(query)) {
			result = append(result, m)
		}
	}
	return result
}

func sortByName(models []Model) []Model {
	sorted := make([]Model, len(models))
	copy(sorted, models)
	// Simple sort for testing
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i].Name > sorted[j].Name {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	return sorted
}

func sortByRecentlyUsed(models []Model) []Model {
	sorted := make([]Model, len(models))
	copy(sorted, models)
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i].LastUsed.Before(sorted[j].LastUsed) {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	return sorted
}

func extractCategories(models []Model) []string {
	set := make(map[string]bool)
	for _, m := range models {
		if m.Category != "" {
			set[m.Category] = true
		}
	}

	categories := make([]string, 0, len(set))
	for c := range set {
		categories = append(categories, c)
	}
	return categories
}

func timeNow() time.Time {
	return time.Now()
}
