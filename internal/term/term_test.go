package term

import (
	"os"
	"testing"
)

func TestSupportsProgressBar(t *testing.T) {
	tests := []struct {
		name            string
		termProgram     string
		wtSession       string
		setWtSession    bool
		expectedSupport bool
	}{
		{
			name:            "Windows Terminal with WT_SESSION",
			termProgram:     "",
			wtSession:       "some-value",
			setWtSession:    true,
			expectedSupport: true,
		},
		{
			name:            "Windows Terminal with empty WT_SESSION",
			termProgram:     "",
			wtSession:       "",
			setWtSession:    true,
			expectedSupport: true,
		},
		{
			name:            "Ghostty terminal (lowercase)",
			termProgram:     "ghostty",
			wtSession:       "",
			setWtSession:    false,
			expectedSupport: true,
		},
		{
			name:            "Ghostty terminal (uppercase)",
			termProgram:     "GHOSTTY",
			wtSession:       "",
			setWtSession:    false,
			expectedSupport: true,
		},
		{
			name:            "Ghostty terminal (mixed case)",
			termProgram:     "GhOsTtY",
			wtSession:       "",
			setWtSession:    false,
			expectedSupport: true,
		},
		{
			name:            "Ghostty in program name",
			termProgram:     "my-ghostty-fork",
			wtSession:       "",
			setWtSession:    false,
			expectedSupport: true,
		},
		{
			name:            "Regular terminal (no support)",
			termProgram:     "xterm",
			wtSession:       "",
			setWtSession:    false,
			expectedSupport: false,
		},
		{
			name:            "iTerm2 (no support)",
			termProgram:     "iTerm.app",
			wtSession:       "",
			setWtSession:    false,
			expectedSupport: false,
		},
		{
			name:            "VS Code terminal (no support)",
			termProgram:     "vscode",
			wtSession:       "",
			setWtSession:    false,
			expectedSupport: false,
		},
		{
			name:            "Empty TERM_PROGRAM (no support)",
			termProgram:     "",
			wtSession:       "",
			setWtSession:    false,
			expectedSupport: false,
		},
		{
			name:            "Both Ghostty and WT_SESSION (should support)",
			termProgram:     "ghostty",
			wtSession:       "session-123",
			setWtSession:    true,
			expectedSupport: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original env vars
			origTermProgram := os.Getenv("TERM_PROGRAM")
			origWtSession, origWtSessionSet := os.LookupEnv("WT_SESSION")

			// Restore after test
			defer func() {
				if origTermProgram != "" {
					os.Setenv("TERM_PROGRAM", origTermProgram)
				} else {
					os.Unsetenv("TERM_PROGRAM")
				}

				if origWtSessionSet {
					os.Setenv("WT_SESSION", origWtSession)
				} else {
					os.Unsetenv("WT_SESSION")
				}
			}()

			// Set test environment
			if tt.termProgram != "" {
				os.Setenv("TERM_PROGRAM", tt.termProgram)
			} else {
				os.Unsetenv("TERM_PROGRAM")
			}

			if tt.setWtSession {
				os.Setenv("WT_SESSION", tt.wtSession)
			} else {
				os.Unsetenv("WT_SESSION")
			}

			// Test the function
			result := SupportsProgressBar()
			if result != tt.expectedSupport {
				t.Errorf("SupportsProgressBar() = %v, want %v", result, tt.expectedSupport)
			}
		})
	}
}

// TestSupportsProgressBarRealEnvironment tests with the actual environment
func TestSupportsProgressBarRealEnvironment(t *testing.T) {
	// This test just ensures the function doesn't panic
	result := SupportsProgressBar()

	termProgram := os.Getenv("TERM_PROGRAM")
	_, wtSession := os.LookupEnv("WT_SESSION")

	t.Logf("Real environment: TERM_PROGRAM=%q, WT_SESSION=%v, supports=%v",
		termProgram, wtSession, result)

	// Just ensure it returns a boolean (which it always will)
	if result != true && result != false {
		t.Error("SupportsProgressBar() must return a boolean")
	}
}

// TestSupportsProgressBarEdgeCases tests edge cases
func TestSupportsProgressBarEdgeCases(t *testing.T) {
	t.Run("TERM_PROGRAM with whitespace", func(t *testing.T) {
		origTermProgram := os.Getenv("TERM_PROGRAM")
		defer func() {
			if origTermProgram != "" {
				os.Setenv("TERM_PROGRAM", origTermProgram)
			} else {
				os.Unsetenv("TERM_PROGRAM")
			}
		}()

		os.Setenv("TERM_PROGRAM", "  ghostty  ")
		result := SupportsProgressBar()
		if !result {
			t.Error("Expected support for TERM_PROGRAM with whitespace and 'ghostty'")
		}
	})

	t.Run("TERM_PROGRAM with ghost (partial match)", func(t *testing.T) {
		origTermProgram := os.Getenv("TERM_PROGRAM")
		defer func() {
			if origTermProgram != "" {
				os.Setenv("TERM_PROGRAM", origTermProgram)
			} else {
				os.Unsetenv("TERM_PROGRAM")
			}
		}()

		os.Setenv("TERM_PROGRAM", "ghost")
		result := SupportsProgressBar()
		if result {
			t.Error("Should not support 'ghost' without 'ty'")
		}
	})
}
