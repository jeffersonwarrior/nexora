package qa

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestAll runs the complete QA suite
func TestAll(t *testing.T) {
	t.Run("go test", testGoTest)
	t.Run("compile", testCompile)
	t.Run("config files", testConfigFiles)
}

func testGoTest(t *testing.T) {
	cmd := exec.Command("go", "test", "./...")
	cmd.Dir = "/home/nexora"
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go test failed:\n%s\nError: %v", output, err)
	}
	t.Log("✅ go test passed")
}

func testCompile(t *testing.T) {
	cmd := exec.Command("go", "build", "-o", "/tmp/nexora-test", ".")
	cmd.Dir = "/home/nexora"
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("compile failed:\n%s\nError: %v", output, err)
	}

	// Cleanup
	os.Remove("/tmp/nexora-test")
	t.Log("✅ compile passed")
}

func testConfigFiles(t *testing.T) {
	configFiles := []string{
		"/home/agent/.local/share/nexora/nexora.json",
		"/home/nexora/nexora.example.json",
		"/home/nexora/.nexora/config.json",
		"/home/nexora/nexora.json",
	}

	for _, configFile := range configFiles {
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			continue // Skip missing files
		}

		t.Run(fmt.Sprintf("validate %s", filepath.Base(configFile)), func(t *testing.T) {
			data, err := os.ReadFile(configFile)
			if err != nil {
				t.Fatalf("failed to read %s: %v", configFile, err)
			}

			var js interface{}
			if err := json.Unmarshal(data, &js); err != nil {
				t.Fatalf("%s is invalid JSON: %v", configFile, err)
			}
			t.Logf("✅ %s valid JSON", configFile)
		})
	}
}
