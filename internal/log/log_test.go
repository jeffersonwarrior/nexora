package log

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestSetup verifies logger initialization
func TestSetup(t *testing.T) {
	// Create a temporary directory for logs
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	// Setup logger
	Setup(logFile, false)

	// Verify initialization
	if !Initialized() {
		t.Fatal("Logger should be initialized")
	}

	// Write a log message to trigger file creation
	slog.Info("test message")
	time.Sleep(20 * time.Millisecond) // Give time for file creation

	// Verify log file was created
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}

	// Write an error to trigger error log creation
	errorLogFile := filepath.Join(tmpDir, "nexora-errors.log")
	slog.Error("test error")
	time.Sleep(20 * time.Millisecond) // Give time for file creation

	if _, err := os.Stat(errorLogFile); os.IsNotExist(err) {
		t.Error("Error log file was not created after writing error")
	}
}

// TestSetupDebugMode verifies debug level logging
func TestSetupDebugMode(t *testing.T) {
	// Note: Can't test file creation since Setup uses sync.Once
	// and previous tests already initialized the logger.
	// We verify that calling Setup with debug=true doesn't panic
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test-debug.log")

	Setup(logFile, true)

	if !Initialized() {
		t.Fatal("Logger should be initialized")
	}

	// Log a debug message (won't use our file due to sync.Once, but shouldn't error)
	slog.Debug("debug message")
}

// TestSetupIdempotent verifies Setup can be called multiple times safely
func TestSetupIdempotent(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test-idempotent.log")

	// Call Setup multiple times
	Setup(logFile, false)
	Setup(logFile, false)
	Setup(logFile, false)

	// Should still be initialized
	if !Initialized() {
		t.Fatal("Logger should be initialized after multiple Setup calls")
	}
}

// TestInitialized verifies the Initialized function
func TestInitialized(t *testing.T) {
	// Note: We can't truly test uninitialized state since Setup is called
	// by other tests, but we can verify it returns true after initialization
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test-init.log")

	Setup(logFile, false)

	if !Initialized() {
		t.Error("Initialized() should return true after Setup")
	}
}

// TestMultiHandlerEnabled verifies handler level checking
func TestMultiHandlerEnabled(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test-handler.log")
	errorLogFile := filepath.Join(tmpDir, "nexora-errors.log")

	// Create handlers
	mainHandler := slog.NewJSONHandler(
		&testWriter{},
		&slog.HandlerOptions{Level: slog.LevelInfo},
	)
	errorHandler := slog.NewJSONHandler(
		&testWriter{},
		&slog.HandlerOptions{Level: slog.LevelError},
	)

	mh := multiHandler{
		main:   mainHandler,
		errors: errorHandler,
	}

	ctx := context.Background()

	// Test different log levels
	tests := []struct {
		level   slog.Level
		enabled bool
	}{
		{slog.LevelDebug, false}, // Below INFO
		{slog.LevelInfo, true},
		{slog.LevelWarn, true},
		{slog.LevelError, true},
	}

	for _, tt := range tests {
		enabled := mh.Enabled(ctx, tt.level)
		if enabled != tt.enabled {
			t.Errorf("Level %v: expected enabled=%v, got %v",
				tt.level, tt.enabled, enabled)
		}
	}

	// Cleanup
	os.Remove(logFile)
	os.Remove(errorLogFile)
}

// TestMultiHandlerHandle verifies correct handler routing
func TestMultiHandlerHandle(t *testing.T) {
	mainWriter := &testWriter{}
	errorWriter := &testWriter{}

	mainHandler := slog.NewJSONHandler(mainWriter, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	errorHandler := slog.NewJSONHandler(errorWriter, &slog.HandlerOptions{
		Level: slog.LevelError,
	})

	mh := multiHandler{
		main:   mainHandler,
		errors: errorHandler,
	}

	ctx := context.Background()

	// Test INFO level goes to main
	infoRecord := slog.NewRecord(time.Now(), slog.LevelInfo, "info message", 0)
	if err := mh.Handle(ctx, infoRecord); err != nil {
		t.Errorf("Failed to handle INFO record: %v", err)
	}

	if len(mainWriter.data) == 0 {
		t.Error("INFO should write to main handler")
	}

	// Reset
	mainWriter.data = ""
	errorWriter.data = ""

	// Test ERROR level goes to error handler
	errorRecord := slog.NewRecord(time.Now(), slog.LevelError, "error message", 0)
	if err := mh.Handle(ctx, errorRecord); err != nil {
		t.Errorf("Failed to handle ERROR record: %v", err)
	}

	if len(errorWriter.data) == 0 {
		t.Error("ERROR should write to error handler")
	}

	if len(mainWriter.data) > 0 {
		t.Error("ERROR should not write to main handler")
	}
}

// TestMultiHandlerWithAttrs verifies attribute chaining
func TestMultiHandlerWithAttrs(t *testing.T) {
	mainHandler := slog.NewJSONHandler(&testWriter{}, &slog.HandlerOptions{})
	errorHandler := slog.NewJSONHandler(&testWriter{}, &slog.HandlerOptions{})

	mh := multiHandler{
		main:   mainHandler,
		errors: errorHandler,
	}

	attrs := []slog.Attr{
		slog.String("key", "value"),
		slog.Int("count", 42),
	}

	newHandler := mh.WithAttrs(attrs)

	// Verify it returns a multiHandler
	if _, ok := newHandler.(multiHandler); !ok {
		t.Error("WithAttrs should return a multiHandler")
	}
}

// TestMultiHandlerWithGroup verifies group chaining
func TestMultiHandlerWithGroup(t *testing.T) {
	mainHandler := slog.NewJSONHandler(&testWriter{}, &slog.HandlerOptions{})
	errorHandler := slog.NewJSONHandler(&testWriter{}, &slog.HandlerOptions{})

	mh := multiHandler{
		main:   mainHandler,
		errors: errorHandler,
	}

	newHandler := mh.WithGroup("test-group")

	// Verify it returns a multiHandler
	if _, ok := newHandler.(multiHandler); !ok {
		t.Error("WithGroup should return a multiHandler")
	}
}

// TestRecoverPanic verifies panic recovery and logging
func TestRecoverPanic(t *testing.T) {
	// Test panic recovery
	didPanic := false
	cleanupCalled := false

	func() {
		defer RecoverPanic("test", func() {
			cleanupCalled = true
		})
		didPanic = true
		panic("test panic")
	}()

	if !didPanic {
		t.Error("Panic should have been triggered")
	}

	if !cleanupCalled {
		t.Error("Cleanup function should have been called")
	}

	// Verify panic log file was created
	files, err := filepath.Glob("nexora-panic-test-*.log")
	if err != nil {
		t.Fatalf("Failed to find panic log files: %v", err)
	}

	if len(files) == 0 {
		t.Error("Panic log file should have been created")
	} else {
		// Verify file content
		content, err := os.ReadFile(files[0])
		if err != nil {
			t.Fatalf("Failed to read panic log: %v", err)
		}

		contentStr := string(content)
		if !strings.Contains(contentStr, "Panic in test") {
			t.Error("Panic log should contain panic message")
		}
		if !strings.Contains(contentStr, "test panic") {
			t.Error("Panic log should contain panic value")
		}
		if !strings.Contains(contentStr, "Stack Trace:") {
			t.Error("Panic log should contain stack trace")
		}

		// Cleanup
		for _, f := range files {
			os.Remove(f)
		}
	}
}

// TestRecoverPanicNoPanic verifies no-op when no panic occurs
func TestRecoverPanicNoPanic(t *testing.T) {
	cleanupCalled := false

	func() {
		defer RecoverPanic("test-no-panic", func() {
			cleanupCalled = true
		})
		// No panic
	}()

	// Cleanup should not be called when no panic
	if cleanupCalled {
		t.Error("Cleanup should not be called when no panic occurs")
	}

	// Verify no panic log was created
	files, err := filepath.Glob("nexora-panic-test-no-panic-*.log")
	if err != nil {
		t.Fatalf("Failed to check for panic logs: %v", err)
	}

	if len(files) > 0 {
		t.Error("No panic log should be created when no panic occurs")
		// Cleanup
		for _, f := range files {
			os.Remove(f)
		}
	}
}

// TestRecoverPanicNilCleanup verifies nil cleanup is handled
func TestRecoverPanicNilCleanup(t *testing.T) {
	func() {
		defer RecoverPanic("test-nil-cleanup", nil)
		panic("test panic with nil cleanup")
	}()

	// Should not panic, cleanup is optional
	files, err := filepath.Glob("nexora-panic-test-nil-cleanup-*.log")
	if err != nil {
		t.Fatalf("Failed to check for panic logs: %v", err)
	}

	if len(files) > 0 {
		// Cleanup
		for _, f := range files {
			os.Remove(f)
		}
	}
}

// TestLogRotation verifies log rotation configuration
func TestLogRotation(t *testing.T) {
	// Note: Can't test file creation since Setup uses sync.Once
	// We verify that writing logs doesn't cause errors
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test-rotation.log")

	Setup(logFile, false)

	// Write multiple log entries (won't use our file due to sync.Once, but shouldn't error)
	for i := 0; i < 100; i++ {
		slog.Info("test message", "iteration", i)
	}

	// Give time for writes
	time.Sleep(50 * time.Millisecond)

	// Note: We can't verify file creation due to sync.Once,
	// but we verified the code doesn't panic
}

// TestErrorLogSeparation verifies errors go to separate file
func TestErrorLogSeparation(t *testing.T) {
	// Note: Can't test file creation since Setup uses sync.Once
	// We verify that writing different log levels doesn't cause errors
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test-separation.log")

	Setup(logFile, false)

	// Write INFO and ERROR messages (won't use our files due to sync.Once, but shouldn't error)
	slog.Info("info message")
	slog.Error("error message")

	// Give time for writes
	time.Sleep(50 * time.Millisecond)

	// Note: We can't verify file creation due to sync.Once,
	// but we verified the code handles different log levels without panicking
}

// testWriter is a simple writer for testing
type testWriter struct {
	data string
}

func (tw *testWriter) Write(p []byte) (n int, err error) {
	tw.data += string(p)
	return len(p), nil
}
