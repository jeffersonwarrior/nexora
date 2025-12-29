package lsp

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DetectLanguage detects the programming language from a file path based on extension.
// Returns the language identifier (e.g., "go", "typescript", "python", "rust") or empty string if unknown.
func DetectLanguage(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	base := strings.ToLower(filepath.Base(path))

	switch ext {
	case ".go":
		return "go"
	case ".mod":
		if strings.HasSuffix(base, "go.mod") {
			return "go"
		}
	case ".ts", ".tsx":
		return "typescript"
	case ".js", ".jsx":
		return "javascript"
	case ".py":
		return "python"
	case ".rs":
		return "rust"
	}

	return ""
}

// DetectLanguageFromDirectory detects the language of a project by examining marker files.
// Returns the language identifier or empty string if detection fails.
func DetectLanguageFromDirectory(dir string) string {
	// Check for common project marker files
	markers := map[string]string{
		"go.mod":           "go",
		"go.sum":           "go",
		"Cargo.toml":       "rust",
		"Cargo.lock":       "rust",
		"tsconfig.json":    "typescript",
		"package.json":     "typescript", // Could be JS too, but assume TS
		"requirements.txt": "python",
		"setup.py":         "python",
		"pyproject.toml":   "python",
	}

	for marker, lang := range markers {
		path := filepath.Join(dir, marker)
		if _, err := os.Stat(path); err == nil {
			return lang
		}
	}

	return ""
}

// LSPServerInfo contains information about an LSP server for a language.
type LSPServerInfo struct {
	Language    string
	Command     string
	Args        []string
	InstallCmd  string
	RootMarkers []string
	FileTypes   []string
}

// GetLSPServerInfo returns configuration information for a given language's LSP server.
func GetLSPServerInfo(language string) *LSPServerInfo {
	switch language {
	case "go":
		return &LSPServerInfo{
			Language:    "go",
			Command:     "gopls",
			Args:        []string{},
			InstallCmd:  "go install golang.org/x/tools/gopls@latest",
			RootMarkers: []string{"go.mod", "go.sum"},
			FileTypes:   []string{".go", ".mod"},
		}
	case "typescript", "javascript":
		return &LSPServerInfo{
			Language:    "typescript",
			Command:     "typescript-language-server",
			Args:        []string{"--stdio"},
			InstallCmd:  "npm install -g typescript-language-server typescript",
			RootMarkers: []string{"package.json", "tsconfig.json"},
			FileTypes:   []string{".ts", ".tsx", ".js", ".jsx"},
		}
	case "python":
		return &LSPServerInfo{
			Language:    "python",
			Command:     "pyright-langserver",
			Args:        []string{"--stdio"},
			InstallCmd:  "pip install pyright",
			RootMarkers: []string{"requirements.txt", "setup.py", "pyproject.toml"},
			FileTypes:   []string{".py"},
		}
	case "rust":
		return &LSPServerInfo{
			Language:    "rust",
			Command:     "rust-analyzer",
			Args:        []string{},
			InstallCmd:  "rustup component add rust-analyzer",
			RootMarkers: []string{"Cargo.toml", "Cargo.lock"},
			FileTypes:   []string{".rs"},
		}
	default:
		return nil
	}
}

// GetLSPServerCommand returns the command name for a language's LSP server.
func GetLSPServerCommand(language string) string {
	info := GetLSPServerInfo(language)
	if info == nil {
		return ""
	}
	return info.Command
}

// GetInstallCommand returns the installation command for a language's LSP server.
func GetInstallCommand(language string) string {
	info := GetLSPServerInfo(language)
	if info == nil {
		return ""
	}
	return info.InstallCmd
}

// FindLSPServer searches for an LSP server binary for the given language.
// Returns the path to the binary and a boolean indicating if it was found.
func FindLSPServer(language string) (string, bool) {
	info := GetLSPServerInfo(language)
	if info == nil {
		return "", false
	}

	// Try to find the command in PATH
	path, err := exec.LookPath(info.Command)
	if err != nil {
		return "", false
	}

	return path, true
}

// NeedsAutoInstall checks if a language needs LSP server auto-installation.
// Returns true if the LSP server is not found and can be auto-installed.
func NeedsAutoInstall(language string) bool {
	info := GetLSPServerInfo(language)
	if info == nil {
		return false
	}

	_, found := FindLSPServer(language)
	return !found
}

// ConfirmFunc is a function type for requesting user confirmation.
// It takes a prompt string and returns true if the user confirms, false otherwise.
type ConfirmFunc func(prompt string) bool

// InstallLSPServer attempts to install the LSP server for the given language.
// It requests user confirmation before proceeding with installation.
// Returns an error if installation fails or is declined.
func InstallLSPServer(ctx context.Context, language string, confirm ConfirmFunc) error {
	info := GetLSPServerInfo(language)
	if info == nil {
		return fmt.Errorf("unsupported language: %s", language)
	}

	// Check if already installed
	if _, found := FindLSPServer(language); found {
		return nil // Already installed
	}

	// Request confirmation
	prompt := fmt.Sprintf("LSP server for %s not found. Install using: %s?", language, info.InstallCmd)
	if !confirm(prompt) {
		return fmt.Errorf("installation declined by user")
	}

	// Parse and execute install command
	parts := strings.Fields(info.InstallCmd)
	if len(parts) == 0 {
		return fmt.Errorf("empty install command")
	}

	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("installation failed: %w", err)
	}

	// Verify installation succeeded
	if _, found := FindLSPServer(language); !found {
		return fmt.Errorf("installation appeared to succeed but %s not found in PATH", info.Command)
	}

	return nil
}

// AutoDetectAndInstall detects the language from a file or directory and optionally installs
// the LSP server if missing. Returns the detected language and any error.
func AutoDetectAndInstall(ctx context.Context, path string, autoInstall bool, confirm ConfirmFunc) (string, error) {
	// Try file-based detection first
	lang := DetectLanguage(path)

	// If that fails and path is a directory, try directory-based detection
	if lang == "" {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			lang = DetectLanguageFromDirectory(path)
		}
	}

	if lang == "" {
		return "", fmt.Errorf("could not detect language for: %s", path)
	}

	// Check if LSP server needs installation
	if autoInstall && NeedsAutoInstall(lang) {
		if err := InstallLSPServer(ctx, lang, confirm); err != nil {
			return lang, fmt.Errorf("failed to install LSP server: %w", err)
		}
	}

	return lang, nil
}
