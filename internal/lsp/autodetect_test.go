package lsp

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"Go file", "main.go", "go"},
		{"Go mod file", "go.mod", "go"},
		{"TypeScript file", "app.ts", "typescript"},
		{"TypeScript JSX", "Component.tsx", "typescript"},
		{"JavaScript file", "script.js", "javascript"},
		{"Python file", "script.py", "python"},
		{"Rust file", "main.rs", "rust"},
		{"Unknown extension", "readme.txt", ""},
		{"No extension", "Makefile", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectLanguage(tt.path)
			if result != tt.expected {
				t.Errorf("DetectLanguage(%q) = %q, want %q", tt.path, result, tt.expected)
			}
		})
	}
}

func TestDetectLanguageFromDirectory(t *testing.T) {
	tests := []struct {
		name         string
		files        []string
		expectedLang string
	}{
		{
			name:         "Go project",
			files:        []string{"go.mod", "main.go"},
			expectedLang: "go",
		},
		{
			name:         "TypeScript project",
			files:        []string{"package.json", "tsconfig.json"},
			expectedLang: "typescript",
		},
		{
			name:         "Python project",
			files:        []string{"requirements.txt", "setup.py"},
			expectedLang: "python",
		},
		{
			name:         "Rust project",
			files:        []string{"Cargo.toml", "Cargo.lock"},
			expectedLang: "rust",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory with marker files
			tmpDir := t.TempDir()
			for _, file := range tt.files {
				path := filepath.Join(tmpDir, file)
				if err := os.WriteFile(path, []byte{}, 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
			}

			result := DetectLanguageFromDirectory(tmpDir)
			if result != tt.expectedLang {
				t.Errorf("DetectLanguageFromDirectory(%q) = %q, want %q", tmpDir, result, tt.expectedLang)
			}
		})
	}
}

func TestFindLSPServer(t *testing.T) {
	tests := []struct {
		name     string
		language string
		wantName string
	}{
		{"Go LSP", "go", "gopls"},
		{"TypeScript LSP", "typescript", "typescript-language-server"},
		{"Python LSP", "python", "pyright"},
		{"Rust LSP", "rust", "rust-analyzer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, found := FindLSPServer(tt.language)

			// If found, verify the binary name matches expected
			if found {
				binary := filepath.Base(path)
				if binary != tt.wantName {
					t.Errorf("FindLSPServer(%q) returned binary %q, want %q", tt.language, binary, tt.wantName)
				}

				// Verify the path is executable
				if _, err := exec.LookPath(path); err != nil {
					t.Errorf("FindLSPServer(%q) returned non-executable path: %v", tt.language, err)
				}
			}
			// Note: We don't fail if not found, as LSP servers may not be installed in test environment
		})
	}
}

func TestGetLSPServerCommand(t *testing.T) {
	tests := []struct {
		name     string
		language string
		wantCmd  string
	}{
		{"Go", "go", "gopls"},
		{"TypeScript", "typescript", "typescript-language-server"},
		{"Python", "python", "pyright-langserver"},
		{"Rust", "rust", "rust-analyzer"},
		{"Unknown", "unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := GetLSPServerCommand(tt.language)
			if cmd != tt.wantCmd {
				t.Errorf("GetLSPServerCommand(%q) = %q, want %q", tt.language, cmd, tt.wantCmd)
			}
		})
	}
}

func TestGetInstallCommand(t *testing.T) {
	tests := []struct {
		name     string
		language string
		wantCmds []string // Multiple possible install commands
	}{
		{
			name:     "Go",
			language: "go",
			wantCmds: []string{"go install golang.org/x/tools/gopls@latest"},
		},
		{
			name:     "TypeScript",
			language: "typescript",
			wantCmds: []string{"npm install -g typescript-language-server typescript"},
		},
		{
			name:     "Python",
			language: "python",
			wantCmds: []string{"pip install pyright", "pip3 install pyright"},
		},
		{
			name:     "Rust",
			language: "rust",
			wantCmds: []string{"rustup component add rust-analyzer"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := GetInstallCommand(tt.language)
			if cmd == "" {
				t.Errorf("GetInstallCommand(%q) returned empty string", tt.language)
				return
			}

			// Check if command matches any expected variant
			found := false
			for _, expected := range tt.wantCmds {
				if cmd == expected {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("GetInstallCommand(%q) = %q, want one of %v", tt.language, cmd, tt.wantCmds)
			}
		})
	}
}

func TestNeedsAutoInstall(t *testing.T) {
	tests := []struct {
		name     string
		language string
		// We can't predict if LSP is installed, so just verify function doesn't panic
	}{
		{"Go", "go"},
		{"TypeScript", "typescript"},
		{"Python", "python"},
		{"Rust", "rust"},
		{"Unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just call the function to ensure it doesn't panic
			_ = NeedsAutoInstall(tt.language)
		})
	}
}

func TestInstallLSPServer(t *testing.T) {
	// Skip in short mode as this would attempt actual installation
	if testing.Short() {
		t.Skip("Skipping installation test in short mode")
	}

	tests := []struct {
		name       string
		language   string
		shouldSkip bool
	}{
		{"Go", "go", true},         // Skip actual install in tests
		{"Unknown", "unknown", true}, // Unknown language should return error
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldSkip {
				t.Skip("Skipping actual installation test")
			}

			ctx := context.Background()
			confirmFunc := func(prompt string) bool {
				return false // Always decline in tests
			}

			err := InstallLSPServer(ctx, tt.language, confirmFunc)
			// For unknown languages, expect an error
			if tt.language == "unknown" && err == nil {
				t.Error("InstallLSPServer(unknown) should return error")
			}
		})
	}
}
