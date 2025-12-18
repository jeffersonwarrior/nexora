package testutil

import (
	"os"
	"testing"
)

// SimpleTestProvider provides a minimal mock for testing
type SimpleTestProvider struct {
	responses []string
	callIndex int
}

// NewSimpleTestProvider creates a basic test provider
func NewSimpleTestProvider(responses ...string) *SimpleTestProvider {
	return &SimpleTestProvider{
		responses: responses,
	}
}

// MockResponse returns a simple response for testing
func (p *SimpleTestProvider) MockResponse(content string) string {
	if p.callIndex >= len(p.responses) {
		return content
	}
	resp := p.responses[p.callIndex]
	p.callIndex++
	return resp
}

// Reset call index
func (p *SimpleTestProvider) Reset() {
	p.callIndex = 0
}

// SetupTempTestDir creates a temporary directory for testing
func SetupTempTestDir(t *testing.T, files map[string]string) string {
	t.Helper()

	dir := t.TempDir()

	for path, content := range files {
		fullPath := dir + "/" + path
		err := os.WriteFile(fullPath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", fullPath, err)
		}
	}

	return dir
}

// CleanupTempTestDir is kept for compatibility, but t.TempDir() is preferred
func CleanupTempTestDir(dir string) {
	os.RemoveAll(dir)
}
