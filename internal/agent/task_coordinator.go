package agent

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nexora/nexora/internal/agent/prompt"
	"github.com/nexora/nexora/internal/agent/tools"
	"log/slog"
)

// TaskExecutionContext holds context for task execution
type TaskExecutionContext struct {
	ID             string
	Directory      string
	Instructions   string
	Progress       string
	StartTime      time.Time
	_TEMP_DIR_PATH string // For temporary files
}

// TaskExecutionCoordinator manages task execution with simplified approach
type TaskExecutionCoordinator struct {
	tempDir *tools.TempDir
	tasks   map[string]*TaskExecutionContext
}

// NewTaskExecutionCoordinator creates a new task coordinator
func NewTaskExecutionCoordinator() *TaskExecutionCoordinator {
	return &TaskExecutionCoordinator{
		tempDir: tools.NewTempDir(),
		tasks:   make(map[string]*TaskExecutionContext),
	}
}

// ExecuteTask executes a task with the given instructions
func (tec *TaskExecutionCoordinator) ExecuteTask(ctx context.Context, directory, instructions string) (*TaskExecutionContext, error) {
	slog.Info("🎯 Executing task", "directory", directory, "instructions", instructions[:min(len(instructions), 100)])

	// Create execution context
	taskID := fmt.Sprintf("task_%d", time.Now().UnixNano())
	taskCtx := &TaskExecutionContext{
		ID:           taskID,
		Directory:    directory,
		Instructions: instructions,
		StartTime:    time.Now(),
		Progress:     "📍 Starting task execution...",
	}

	// Create temporary directory for this task
	tempDir, err := tec.tempDir.CreateTempDir(taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	taskCtx._TEMP_DIR_PATH = tempDir

	// Store task
	tec.tasks[taskID] = taskCtx

	// Execute the task
	defer func() {
		if err := tec.tempDir.CleanUp(tempDir); err != nil {
			slog.Error("Failed to cleanup temp directory", "error", err)
		}
		delete(tec.tasks, taskID)
	}()

	// Create a simple execution plan
	taskCtx.Progress = "📋 Creating execution plan..."
	plan, err := tec.createSimpleExecutionPlan(instructions)
	if err != nil {
		return nil, fmt.Errorf("failed to create execution plan: %w", err)
	}

	taskCtx.Progress = "🚀 Executing plan steps..."
	result, err := tec.executeSimplePlan(ctx, taskCtx, plan)
	if err != nil {
		return nil, fmt.Errorf("task execution failed: %w", err)
	}

	taskCtx.Progress = "✅ Task completed successfully"
	slog.Info("🎉 Task execution completed", "duration", time.Since(taskCtx.StartTime))

	return taskCtx, result
}

// createSimpleExecutionPlan creates a simple execution plan
func (tec *TaskExecutionCoordinator) createSimpleExecutionPlan(instructions string) (*prompt.Prompt, error) {
	return prompt.NewPrompt(
		"task-execution",
		"system: You are executing a specific task. Follow the instructions precisely and report the results.\n\nuser: "+instructions,
	)
}

// executeSimplePlan executes a simple plan
func (tec *TaskExecutionCoordinator) executeSimplePlan(ctx context.Context, taskCtx *TaskExecutionContext, plan *prompt.Prompt) (error, error) {
	// Extract files mentioned in instructions
	files := tec.extractFilesFromInstructions(taskCtx.Instructions)

	slog.Info("📄 Found relevant files", "count", len(files))

	// For each file, create a simple task
	for _, file := range files {
		select {
		case <-ctx.Done():
			return fmt.Errorf("task execution cancelled"), nil
		default:
			taskCtx.Progress = fmt.Sprintf("🔧 Processing %s...", filepath.Base(file))

			// Simple file processing - just read it
			content, err := tec.readSimpleFile(file)
			if err != nil {
				slog.Warn("Failed to read file", "file", file, "error", err)
				continue
			}

			slog.Info("📖 Read file", "file", file, "size", len(content))
		}
	}

	return nil, nil
}

// extractFilesFromInstructions extracts file paths from instructions
func (tec *TaskExecutionCoordinator) extractFilesFromInstructions(instructions string) []string {
	var files []string

	// Simple extraction - look for common patterns in working directory
	scanner := bufio.NewScanner(strings.NewReader(instructions))
	for scanner.Scan() {
		line := scanner.Text()
		for _, ext := range []string{".go", ".md", ".txt", ".json", ".yaml", ".yml"} {
			if strings.Contains(line, ext) {
				// Try to find files in current directory
				matches, _ := filepath.Glob("*" + ext)
				files = append(files, matches...)
			}
		}
	}

	// Remove duplicates and limit
	seen := make(map[string]bool)
	unique := []string{}
	for _, file := range files {
		if !seen[file] {
			seen[file] = true
			unique = append(unique, file)
			if len(unique) >= 10 { // Limit to 10 files
				break
			}
		}
	}

	return unique
}

// readSimpleFile reads a file simply
func (tec *TaskExecutionCoordinator) readSimpleFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
