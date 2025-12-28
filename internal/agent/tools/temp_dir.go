package tools

import (
	"fmt"
	"os"
	"path/filepath"

	"log/slog"
)

// TempDir provides temporary directory management capabilities
type TempDir struct {
	basePath string
}

// NewTempDir creates a new temporary directory manager
func NewTempDir() *TempDir {
	// Use .nexora to avoid conflict with the nexora executable
	basePath := filepath.Join(os.TempDir(), ".nexora", "nexora")
	return &TempDir{basePath: basePath}
}

// CreateTempDir creates a new temporary directory with a given prefix
func (t *TempDir) CreateTempDir(prefix string) (string, error) {
	if err := os.MkdirAll(t.basePath, 0o700); err != nil {
		return "", fmt.Errorf("failed to create base temp directory: %w", err)
	}

	fullPath, err := os.MkdirTemp(t.basePath, prefix)
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	slog.Debug("Created temporary directory", "path", fullPath)
	return fullPath, nil
}

// CreateTempFile creates a new temporary file with a given prefix and content
func (t *TempDir) CreateTempFile(prefix, content string) (string, error) {
	tempDir, err := t.CreateTempDir(prefix)
	if err != nil {
		return "", err
	}

	filePath := filepath.Join(tempDir, prefix+".txt")
	if err := os.WriteFile(filePath, []byte(content), 0o600); err != nil {
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}

	slog.Debug("Created temporary file", "path", filePath)
	return filePath, nil
}

// CleanUp removes the temporary directory and all its contents
func (t *TempDir) CleanUp(path string) error {
	if path == "" {
		return nil
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}

	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("failed to cleanup temp directory: %w", err)
	}

	slog.Debug("Cleaned up temporary directory", "path", path)
	return nil
}

// GetBasePath returns the base temporary directory path
func (t *TempDir) GetBasePath() string {
	return t.basePath
}
