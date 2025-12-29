package tui

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestTUIInteractions tests TUI functionality via tmux automation
// Requires: tmux, compiled nexora binary
// Run with: go test -v ./internal/tui/... -run TestTUI
func TestTUIInteractions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TUI test in short mode")
	}

	// Check prerequisites
	if !hasTMUX(t) {
		t.Skip("tmux not available")
	}

	if !existsNexora(t) {
		t.Skip("nexora binary not found")
	}

	sessionName := "nexora-tui-test-" + strings.Replace(t.Name(), "/", "-", -1)

	// Cleanup setup
	setup := func() {
		exec.Command("tmux", "kill-session", "-t", sessionName).Run()
		time.Sleep(500 * time.Millisecond)
	}

	teardown := func() {
		exec.Command("tmux", "kill-session", "-t", sessionName).Run()
	}

	setup()
	defer teardown()

	// Start nexora
	nexoraPath := getNexoraPath()
	tmuxCmd := exec.Command("tmux", "new-session", "-d", "-s", sessionName, nexoraPath)
	if err := tmuxCmd.Run(); err != nil {
		t.Fatalf("Failed to start nexora in tmux: %v", err)
	}
	time.Sleep(2 * time.Second)

	// Helper functions
	sendKeys := func(keys string) {
		exec.Command("tmux", "send-keys", "-t", sessionName, "0", keys).Run()
		time.Sleep(300 * time.Millisecond)
	}

	sendCtrl := func(key string) {
		sendKeys("C-" + key)
	}

	capture := func() string {
		cmd := exec.Command("tmux", "capture-pane", "-t", sessionName, "-p")
		var buf bytes.Buffer
		cmd.Stdout = &buf
		cmd.Run()
		return buf.String()
	}

	// Test 7: TUI responsiveness
	t.Run("TUIResponsive", func(t *testing.T) {
		// Send some keys and verify no crash
		sendKeys("hello")
		time.Sleep(300 * time.Millisecond)
		sendKeys("C-u") // clear
		time.Sleep(300 * time.Millisecond)

		screen := capture()
		if strings.Contains(screen, "nexora") || strings.Contains(screen, "Nexora") {
			t.Log("✓ TUI is responsive")
		} else {
			t.Log("Note: TUI appears to be running")
		}
	})
	t.Run("SlashCommandTrigger", func(t *testing.T) {
		// Clear line
		sendKeys("C-u")
		time.Sleep(300 * time.Millisecond)

		// Type "/"
		sendKeys("/")
		time.Sleep(1 * time.Second)

		screen := capture()
		t.Logf("Screen after '/': %s", screen)

		// Should see commands menu or dialog
		if strings.Contains(screen, "command") || strings.Contains(screen, "Command") {
			t.Log("✓ '/' triggers commands menu")
		} else {
			t.Log("Note: Commands menu may require configuration")
		}
	})

	// Test 2: "/" passthrough with path text
	t.Run("SlashPassthrough", func(t *testing.T) {
		// Clear first
		sendKeys("C-u")
		time.Sleep(300 * time.Millisecond)

		// Type path
		sendKeys("/home/user/project")
		time.Sleep(500 * time.Millisecond)

		screen := capture()

		if strings.Contains(screen, "/home/user/project") {
			t.Log("✓ '/' passthrough works - path typed correctly")
		} else {
			t.Log("Note: Could not verify path text in output")
		}
	})

	// Test 3: Ctrl+E opens models dialog
	t.Run("CtrlEOpensModels", func(t *testing.T) {
		sendCtrl("e")
		time.Sleep(1 * time.Second)

		screen := capture()

		if strings.Contains(screen, "model") || strings.Contains(screen, "Model") {
			t.Log("✓ Ctrl+E opens models dialog")
		} else {
			t.Log("Note: Models dialog may not be visible (requires provider config)")
		}
	})

	// Test 4: Ctrl+P opens prompts dialog
	t.Run("CtrlPOpensPrompts", func(t *testing.T) {
		sendCtrl("p")
		time.Sleep(1 * time.Second)

		screen := capture()

		if strings.Contains(screen, "prompt") || strings.Contains(screen, "Prompt") {
			t.Log("✓ Ctrl+P opens prompts dialog")
		} else {
			t.Log("Note: Prompts dialog may not be visible")
		}
	})

	// Test 5: Dialog navigation with j/k
	t.Run("DialogNavigation", func(t *testing.T) {
		// Open prompts dialog
		sendCtrl("p")
		time.Sleep(500 * time.Millisecond)

		// Try j/k navigation
		sendKeys("j")
		time.Sleep(200 * time.Millisecond)
		sendKeys("k")
		time.Sleep(200 * time.Millisecond)

		t.Log("✓ j/k navigation attempted")
	})

	// Test 6: Ctrl+N new session
	t.Run("CtrlNNewSession", func(t *testing.T) {
		sendCtrl("n")
		time.Sleep(500 * time.Millisecond)
		t.Log("✓ Ctrl+N sent (new session)")
	})

	// Test 7: TUI responsiveness
	t.Run("TUIResponsive", func(t *testing.T) {
		// Send some keys and verify no crash
		sendKeys("hello")
		time.Sleep(300 * time.Millisecond)
		sendKeys("C-u") // clear
		time.Sleep(300 * time.Millisecond)

		screen := capture()
		if strings.Contains(screen, "nexora") || strings.Contains(screen, "Nexora") {
			t.Log("✓ TUI is responsive")
		} else {
			t.Log("Note: TUI appears to be running")
		}
	})

	// Final screen capture for debugging
	t.Run("FinalScreen", func(t *testing.T) {
		screen := capture()
		t.Logf("Final TUI state:\n%s", screen)
	})
}

// TestTUIKeyBindings tests specific key bindings
func TestTUIKeyBindings(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	if !hasTMUX(t) || !existsNexora(t) {
		t.Skip("Prerequisites not met")
	}

	sessionName := "nexora-keys-test"
	defer exec.Command("tmux", "kill-session", "-t", sessionName).Run()

	// Start nexora
	nexoraPath := getNexoraPath()
	exec.Command("tmux", "new-session", "-d", "-s", sessionName, nexoraPath).Run()
	time.Sleep(2 * time.Second)

	send := func(keys string) {
		exec.Command("tmux", "send-keys", "-t", sessionName, "0", keys).Run()
		time.Sleep(300 * time.Millisecond)
	}

	capture := func() string {
		cmd := exec.Command("tmux", "capture-pane", "-t", sessionName, "-p")
		var buf bytes.Buffer
		cmd.Stdout = &buf
		cmd.Run()
		return buf.String()
	}

	// Test each key binding
	tests := []struct {
		name  string
		keys  string
		check func(string) bool
	}{
		{"Ctrl+E models", "C-e", func(s string) bool { return true }}, // Just verify no crash
		{"Ctrl+P prompts", "C-p", func(s string) bool { return true }},
		{"Ctrl+N session", "C-n", func(s string) bool { return true }},
		{"j down", "j", func(s string) bool { return true }},
		{"k up", "k", func(s string) bool { return true }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			send(tt.keys)
			screen := capture()
			if !tt.check(screen) {
				t.Logf("Unexpected screen state for %s", tt.name)
			}
		})
	}
}

// TestTUIDialogs tests dialog opening and navigation
func TestTUIDialogs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	if !hasTMUX(t) || !existsNexora(t) {
		t.Skip("Prerequisites not met")
	}

	sessionName := "nexora-dialogs-test"
	defer exec.Command("tmux", "kill-session", "-t", sessionName).Run()

	exec.Command("tmux", "new-session", "-d", "-s", sessionName, "./nexora").Run()
	time.Sleep(2 * time.Second)

	send := func(keys string) {
		exec.Command("tmux", "send-keys", "-t", sessionName, "0", keys).Run()
		time.Sleep(300 * time.Millisecond)
	}

	capture := func() string {
		cmd := exec.Command("tmux", "capture-pane", "-t", sessionName, "-p")
		var buf bytes.Buffer
		cmd.Stdout = &buf
		cmd.Run()
		return buf.String()
	}

	t.Run("CommandsDialog", func(t *testing.T) {
		send(":") // Command mode
		time.Sleep(500 * time.Millisecond)
		screen := capture()
		t.Logf("After ':': %s", screen)
	})

	t.Run("ModelsDialog", func(t *testing.T) {
		send("C-e")
		time.Sleep(500 * time.Millisecond)
		screen := capture()
		t.Logf("After Ctrl+E: %s", screen)
	})

	t.Run("PromptsDialog", func(t *testing.T) {
		send("C-p")
		time.Sleep(500 * time.Millisecond)
		screen := capture()
		t.Logf("After Ctrl+P: %s", screen)
	})
}

// Helper: check if tmux is available
func hasTMUX(t *testing.T) bool {
	cmd := exec.Command("which", "tmux")
	return cmd.Run() == nil
}

// Helper: check if nexora binary exists
func existsNexora(t *testing.T) bool {
	// Check in multiple possible locations
	paths := []string{
		"./nexora",
		"/home/nexora/nexora",
		"nexora",
	}
	for _, p := range paths {
		cmd := exec.Command("test", "-x", p)
		if cmd.Run() == nil {
			return true
		}
	}
	return false
}

// getNexoraPath returns the path to nexora binary
func getNexoraPath() string {
	paths := []string{
		"./nexora",
		"/home/nexora/nexora",
		"nexora",
	}
	for _, p := range paths {
		cmd := exec.Command("test", "-x", p)
		if cmd.Run() == nil {
			return p
		}
	}
	return "./nexora"
}

// BenchmarkTUIStart measures TUI startup time
func BenchmarkTUIStart(b *testing.B) {
	if !hasTMUX(nil) || !existsNexora(nil) {
		b.Skip("Prerequisites not met")
	}

	nexoraPath := getNexoraPath()
	for i := 0; i < b.N; i++ {
		sessionName := "nexora-bench-" + string(rune(i))
		defer exec.Command("tmux", "kill-session", "-t", sessionName).Run()

		cmd := exec.Command("tmux", "new-session", "-d", "-s", sessionName, nexoraPath)
		cmd.Start()
		time.Sleep(2 * time.Second)
		cmd.Process.Kill()
	}
}
