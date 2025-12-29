package shell

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestShell_GetWorkingDir(t *testing.T) {
	tmpDir := t.TempDir()
	shell := NewShell(&Options{
		WorkingDir: tmpDir,
	})

	wd := shell.GetWorkingDir()
	if wd != tmpDir {
		t.Errorf("expected working dir %q, got %q", tmpDir, wd)
	}
}

func TestShell_SetWorkingDir(t *testing.T) {
	tmpDir := t.TempDir()
	shell := NewShell(nil)

	// Set to valid directory
	err := shell.SetWorkingDir(tmpDir)
	if err != nil {
		t.Errorf("SetWorkingDir failed: %v", err)
	}

	wd := shell.GetWorkingDir()
	if wd != tmpDir {
		t.Errorf("expected working dir %q after set, got %q", tmpDir, wd)
	}

	// Set to invalid directory
	err = shell.SetWorkingDir("/nonexistent/directory")
	if err == nil {
		t.Error("expected error when setting invalid directory")
	}
}

func TestShell_GetEnv(t *testing.T) {
	testEnv := []string{
		"TEST_VAR=test_value",
		"ANOTHER_VAR=another_value",
	}

	shell := NewShell(&Options{
		Env: testEnv,
	})

	env := shell.GetEnv()

	if len(env) != len(testEnv) {
		t.Errorf("expected %d env vars, got %d", len(testEnv), len(env))
	}

	// Verify it's a copy (modifying returned slice doesn't affect shell)
	env[0] = "MODIFIED=value"
	env2 := shell.GetEnv()
	if env2[0] == "MODIFIED=value" {
		t.Error("GetEnv should return a copy, not the original slice")
	}
}

func TestShell_SetEnv(t *testing.T) {
	shell := NewShell(nil)

	// Set new variable
	shell.SetEnv("TEST_VAR", "test_value")

	env := shell.GetEnv()
	found := false
	for _, e := range env {
		if strings.HasPrefix(e, "TEST_VAR=") {
			found = true
			if e != "TEST_VAR=test_value" {
				t.Errorf("expected TEST_VAR=test_value, got %s", e)
			}
			break
		}
	}
	if !found {
		t.Error("TEST_VAR not found in environment")
	}

	// Update existing variable
	shell.SetEnv("TEST_VAR", "new_value")
	env = shell.GetEnv()
	count := 0
	for _, e := range env {
		if strings.HasPrefix(e, "TEST_VAR=") {
			count++
			if e != "TEST_VAR=new_value" {
				t.Errorf("expected TEST_VAR=new_value, got %s", e)
			}
		}
	}
	if count != 1 {
		t.Errorf("expected exactly 1 TEST_VAR entry, got %d", count)
	}
}

func TestShell_SetBlockFuncs(t *testing.T) {
	shell := NewShell(nil)

	called := false
	blockFunc := func(args []string) bool {
		called = true
		return true
	}

	shell.SetBlockFuncs([]BlockFunc{blockFunc})

	// Verify block funcs were set by checking they're used
	if shell.blockFuncs == nil {
		t.Error("blockFuncs not set")
	}

	if len(shell.blockFuncs) != 1 {
		t.Errorf("expected 1 block func, got %d", len(shell.blockFuncs))
	}

	// Call the block func to verify it works
	shell.blockFuncs[0]([]string{"test"})
	if !called {
		t.Error("block func was not called")
	}
}

func TestNoopLogger_InfoPersist(t *testing.T) {
	logger := noopLogger{}
	// Should not panic
	logger.InfoPersist("test message", "key", "value")
}

func TestShell_NewShell_WithLogger(t *testing.T) {
	mockLogger := &mockLogger{}
	shell := NewShell(&Options{
		Logger: mockLogger,
	})

	if shell.logger != mockLogger {
		t.Error("logger not set correctly")
	}
}

func TestShell_NewShell_NilOptions(t *testing.T) {
	shell := NewShell(nil)

	if shell == nil {
		t.Fatal("NewShell with nil options returned nil")
	}

	// Should use defaults
	if shell.cwd == "" {
		t.Error("expected default working directory")
	}

	if len(shell.env) == 0 {
		t.Error("expected environment variables from os.Environ()")
	}
}

func TestShell_NewShell_CustomOptions(t *testing.T) {
	tmpDir := t.TempDir()
	customEnv := []string{"CUSTOM=value"}
	mockLogger := &mockLogger{}
	blockFunc := func(args []string) bool { return false }

	shell := NewShell(&Options{
		WorkingDir: tmpDir,
		Env:        customEnv,
		Logger:     mockLogger,
		BlockFuncs: []BlockFunc{blockFunc},
	})

	if shell.cwd != tmpDir {
		t.Errorf("expected cwd %q, got %q", tmpDir, shell.cwd)
	}

	if len(shell.env) != 1 || shell.env[0] != "CUSTOM=value" {
		t.Errorf("expected custom env, got %v", shell.env)
	}

	if shell.logger != mockLogger {
		t.Error("logger not set")
	}

	if len(shell.blockFuncs) != 1 {
		t.Errorf("expected 1 block func, got %d", len(shell.blockFuncs))
	}
}

// Mock logger for testing
type mockLogger struct {
	messages []string
}

func (m *mockLogger) InfoPersist(msg string, keysAndValues ...any) {
	m.messages = append(m.messages, msg)
}

func TestShell_Concurrency(t *testing.T) {
	shell := NewShell(nil)

	// Test that concurrent reads are safe
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_ = shell.GetWorkingDir()
			_ = shell.GetEnv()
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestShell_SetWorkingDir_Concurrency(t *testing.T) {
	tmpDir1 := t.TempDir()
	tmpDir2 := t.TempDir()

	shell := NewShell(&Options{WorkingDir: tmpDir1})

	// Test concurrent SetWorkingDir calls
	done := make(chan bool, 2)

	go func() {
		_ = shell.SetWorkingDir(tmpDir1)
		done <- true
	}()

	go func() {
		_ = shell.SetWorkingDir(tmpDir2)
		done <- true
	}()

	<-done
	<-done

	// Final working dir should be one of the two
	wd := shell.GetWorkingDir()
	if wd != tmpDir1 && wd != tmpDir2 {
		t.Errorf("unexpected working dir: %s", wd)
	}
}

func TestShell_SetEnv_Concurrency(t *testing.T) {
	shell := NewShell(nil)

	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		i := i
		go func() {
			shell.SetEnv("TEST_VAR", filepath.Join("value", string(rune(i))))
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify environment is consistent
	env := shell.GetEnv()
	count := 0
	for _, e := range env {
		if strings.HasPrefix(e, "TEST_VAR=") {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected exactly 1 TEST_VAR entry after concurrent sets, got %d", count)
	}
}
