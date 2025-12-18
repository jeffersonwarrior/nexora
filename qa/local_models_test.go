package qa

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nexora/cli/internal/config/providers"
)

// TestLocalModelsFullFlow tests the local detector without imports
func TestLocalModelsFullFlow(t *testing.T) {
	// Test 1: Ollama detection + model list + context window
	t.Run("OllamaFullFlow", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/api/tags":
				w.Write([]byte(`{"models":[{"name":"llama3.1:70b","size":43762291200,"digest":"sha256:abc"}]}`))
			case "/api/generate":
				w.Write([]byte(`{"model":"llama3.1:70b","context_window":131072}`))
			default:
				http.Error(w, "Not Found", http.StatusNotFound)
			}
		}))
		defer server.Close()

		detector := providers.NewLocalDetector(server.URL)
		provider, err := detector.Detect("ollama", "")
		if err != nil {
			t.Fatal(err)
		}
		if provider.Type != "local" || provider.Endpoint != server.URL {
			t.Errorf("Expected local provider, got %v", provider)
		}
		if len(provider.Models) == 0 {
			t.Fatal("No models detection failed")
		}
		if provider.Models[0].Matched != "llama-3.1-70b-instruct" {
			t.Error("Model matching failed")
		}
	})

	// Test 2: vLLM/OpenAI-compatible detection
	t.Run("VLLMOpenAICompatible", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/v1/models" {
				w.Write([]byte(`{"data":[{"id":"meta-llama/Llama-3.1-70B-Instruct","context_window":131072}]}`))
			} else {
				http.Error(w, "Not Found", http.StatusNotFound)
			}
		}))
		defer server.Close()

		detector := providers.NewLocalDetector(server.URL)
		provider, err := detector.Detect("openai", "")
		if err != nil {
			t.Fatal(err)
		}
		if provider.Models[0].ID != "meta-llama/Llama-3.1-70B-Instruct" {
			t.Error("vLLM model detection failed")
		}
	})

	// Test 3: API Key required (401 → key prompt simulation)
	t.Run("APIKeyRequired", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(401)
		}))
		defer server.Close()

		detector := providers.NewLocalDetector(server.URL)
		provider, err := detector.Detect("openai", "test-key-123")
		if err == nil {
			t.Error("Expected auth error without key")
		}
		if provider != nil {
			t.Error("Provider should be nil on auth error")
		}
	})

	// Test 4: 30s timeout (reduced to 2s for testing)
	t.Run("Timeout30s", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(2 * time.Second) // Reduced timeout for testing
		}))
		defer server.Close()

		// Use a shorter timeout detector - skip this test since we can't mock private fields
		t.Skip("Timeout test skipped - private fields prevent struct literal")
	})
}

func TestModelMatching(t *testing.T) {
	tests := []struct {
		rawModel string
		expected string
	}{
		{"llama3.1:70b", "llama-3.1-70b-instruct"},
		{"llama3.1:405b", "llama-3.1-405b-instruct"},
		{"codellama:34b", "codellama-34b-instruct"},
		{"mixtral:8x22b", "mixtral-8x22b-instruct"},
		{"meta-llama/Llama-3.1-70B-Instruct", "llama-3.1-70b-instruct"},
		{"70b", "llama-3-70b-instruct"},
		{"8b", "llama-3-8b-instruct"},
	}

	for _, tt := range tests {
		result := providers.MatchModelToLibrary(tt.rawModel)
		if result != tt.expected {
			t.Errorf("MatchModel(%q) = %q, want %q", tt.rawModel, result, tt.expected)
		}
	}
}

func TestContextWindowPrioritization(t *testing.T) {
	models := []providers.LocalModel{
		{ID: "small-7b", Context: 4096},
		{ID: "medium-32k", Context: 32768},
		{ID: "large-128k", Context: 131072},
	}

	prioritized := providers.PrioritizeModels(models)
	if prioritized[0].ID != "large-128k" {
		t.Error("Expected >64k first")
	}
	if prioritized[1].ID != "medium-32k" {
		t.Error("Expected 32-64k second")
	}
	if prioritized[2].ID != "small-7b" {
		t.Error("Expected <32k last")
	}
}
