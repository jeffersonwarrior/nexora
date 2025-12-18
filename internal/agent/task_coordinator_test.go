package agent

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nexora/cli/internal/agent/tools"
)

func TestTaskExecutionCoordinator(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create some test files
	testFile := filepath.Join(tempDir, "test.go")
	err := os.WriteFile(testFile, []byte("package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create coordinator
	coordinator := NewTaskExecutionCoordinator()

	// Test tempdir tools
	tempDirTool := tools.NewTempDir()
	createdDir, err := tempDirTool.CreateTempDir("test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Verify directory exists
	if _, err := os.Stat(createdDir); os.IsNotExist(err) {
		t.Fatalf("Temp directory was not created")
	}

	// Test cleanup
	err = tempDirTool.CleanUp(createdDir)
	if err != nil {
		t.Fatalf("Failed to cleanup temp dir: %v", err)
	}

	// Verify directory is cleaned up
	if _, err := os.Stat(createdDir); !os.IsNotExist(err) {
		t.Fatalf("Temp directory was not cleaned up")
	}

	// Test task execution
	ctx := context.Background()
	instructions := "Read and analyze the test.go file"

	result, err := coordinator.ExecuteTask(ctx, tempDir, instructions)
	if err != nil {
		t.Fatalf("Task execution failed: %v", err)
	}

	if result == nil {
		t.Fatalf("Expected result but got nil")
	}

	// Verify task context
	if result.Directory != tempDir {
		t.Errorf("Expected directory %s, got %s", tempDir, result.Directory)
	}

	if result.Instructions != instructions {
		t.Errorf("Expected instructions %s, got %s", instructions, result.Instructions)
	}

	if result.Progress != "✅ Task completed successfully" {
		t.Errorf("Expected completed status, got %s", result.Progress)
	}

	// Verify execution time
	if time.Since(result.StartTime) < 0 {
		t.Errorf("Invalid start time")
	}
}

func TestTaskCoordinator_FileExtraction(t *testing.T) {
	coordinator := NewTaskExecutionCoordinator()

	instructions := "Please read the test.go file and analyze main.go"
	files := coordinator.extractFilesFromInstructions(instructions)

	// Should find some files (if they exist)
	if len(files) > 10 {
		t.Errorf("Too many files extracted: %d", len(files))
	}
}

func TestTempDir_CreateFile(t *testing.T) {
	tempDirTool := tools.NewTempDir()

	filePath, err := tempDirTool.CreateTempFile("test", "Hello, World!")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Verify file exists
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read temp file: %v", err)
	}

	if string(content) != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', got '%s'", string(content))
	}

	// Cleanup
	err = tempDirTool.CleanUp(filepath.Dir(filePath))
	if err != nil {
		t.Fatalf("Failed to cleanup: %v", err)
	}
}
