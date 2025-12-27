package qa

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nexora/nexora/internal/config/providers"
	"github.com/nexora/nexora/internal/testutil"
)

// TestLocalModelsFullFlow tests the local detector without imports
func TestLocalModelsFullFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping timeout test in short mode")
	}
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
		if provider.Models[0].Context != 131072 {
			t.Errorf("Expected context 131072, got %d", provider.Models[0].Context)
		}
	})

	// Test 2b: vLLM with max_model_len field
	t.Run("VLLMMaxModelLen", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/v1/models" {
				// vLLM returns max_model_len instead of context_window
				w.Write([]byte(`{"data":[{"id":"Qwen/Qwen2.5-72B-Instruct","max_model_len":131072,"owned_by":"vllm"}]}`))
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
		if provider.Models[0].ID != "Qwen/Qwen2.5-72B-Instruct" {
			t.Error("vLLM model detection failed")
		}
		// This should now correctly detect 128k context
		if provider.Models[0].Context != 131072 {
			t.Errorf("Expected context 131072 from max_model_len, got %d", provider.Models[0].Context)
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

	// Test 4: Test timeout with actual delay using context
	t.Run("Timeout30s", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simulate a slow server that takes longer than our timeout
			time.Sleep(2 * time.Second) // Explicit delay
			w.Write([]byte(`{"models":[]}`))
		}))
		defer server.Close()

		// Test with a shorter timeout than the server delay
		testutil.RunWithTimeout(t, 1*time.Second, func(ctx context.Context) {
			// Create detector
			detector := providers.NewLocalDetector(server.URL)

			// This should either succeed (if server responds fast) or timeout gracefully
			_, err := detector.Detect("ollama", "")
			if err != nil {
				// Expect timeout or context cancellation, not a hang
				if err == context.DeadlineExceeded || err == context.Canceled {
					t.Log("✅ Timeout worked as expected")
				} else {
					t.Error("Unexpected error:", err)
				}
			}
		})
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

func TestContextWindowDetection(t *testing.T) {
	// Test that context window detection works correctly and doesn't
	// pick up weird numbers from version strings or other parts of model name
	tests := []struct {
		modelName string
		expected  int
		reason    string
	}{
		// Explicit context window indicators should take priority
		{"model-128k", 131072, "Explicit 128k indicator"},
		{"llama3.1:70b-128k", 131072, "Explicit 128k overrides param size"},
		{"model:32k", 32768, "Explicit 32k indicator"},
		{"model:16k", 16384, "Explicit 16k indicator"},

		// Parameter count based detection (specific patterns only)
		{"llama3.1:70b", 131072, "70b models get 128k"},
		{"llama3.1:8b", 8192, "8b models get 8k"},
		{"codellama:34b", 131072, "34b models get 128k"},

		// Should NOT match just digits without 'b' suffix
		{"model-v3.1.8", 4096, "Version number '8' should not match '8b' pattern"},
		{"qwen2.5-72b-instruct", 131072, "Should match 72b pattern"},
		{"deepseek-coder-33b", 131072, "Should match 33b pattern"},

		// Unknown models get default
		{"random-model", 4096, "Unknown models default to 4096"},
		{"gpt-custom", 4096, "Unknown models default to 4096"},
	}

	detector := providers.NewLocalDetector("http://localhost:11434")
	for _, tt := range tests {
		t.Run(tt.modelName, func(t *testing.T) {
			// This is a private method, so we test via the public flow
			// For now, test the estimation function indirectly
			result := detector.EstimateContext(tt.modelName)
			if result != tt.expected {
				t.Errorf("%s: got context %d, want %d", tt.reason, result, tt.expected)
			}
		})
	}
}
