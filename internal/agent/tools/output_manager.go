package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Output management constants
const (
	// Token limits for tool output (conservative estimates)
	SmallOutputLimit   = 4000   // Safe for all models
	MediumOutputLimit  = 12000  // Good for most models  
	LargeOutputLimit   = 28000  // For models with 32K+ context
	
	// Fallback to byte limit if token counting fails
	MaxOutputBytes = 100 * 1024 // 100KB
	
	// When to write to tmp file instead of truncating
	TooLargeForContext = 50000 // Tokens - definitely too big
)

// OutputResult represents the managed output
type OutputResult struct {
	Content       string // The actual content (possibly truncated)
	WasTruncated  bool   // Whether output was truncated
	WasWrittenToFile bool // Whether output was written to a tmp file
	FilePath      string // Path to tmp file if written
	TokenCount    int    // Estimated token count
	OriginalSize  int    // Original size in bytes
	ActionTaken   string // What we did: "returned", "truncated", "written_to_file"
}

// ManageOutput is the main smart output wrapper
// It takes raw tool output and manages it for model context safety
func ManageOutput(output string, toolName string, workingDir string, sessionID string) OutputResult {
	result := OutputResult{
		Content:      output,
		OriginalSize: len(output),
	}
	
	if output == "" {
		result.ActionTaken = "returned_empty"
		return result
	}
	
	// Count tokens
	tokenCount := countTokens(output)
	result.TokenCount = tokenCount
	
	// Determine appropriate limit based on tool type
	limit := getOutputLimitForTool(toolName)
	
	// If within limit, return as-is
	if tokenCount <= limit {
		result.ActionTaken = "returned"
		return result
	}
	
	// If extremely large, write to tmp file
	if tokenCount > TooLargeForContext {
		filePath, err := writeToTmpFile(output, toolName, workingDir, sessionID)
		if err == nil {
			result.WasWrittenToFile = true
			result.FilePath = filePath
			result.Content = fmt.Sprintf("Output too large (%d tokens, %d bytes).\n\nWritten to: %s\n\nUse view tool to read this file.\n\nThis file will be deleted when the session ends.", 
				tokenCount, len(output), filePath)
			result.ActionTaken = "written_to_file"
			return result
		}
		// If write fails, fall through to truncate
	}
	
	// Truncate to fit token limit
	result.WasTruncated = true
	result.Content = truncateToTokenLimit(output, limit)
	result.ActionTaken = "truncated"
	
	return result
}

// getOutputLimitForTool returns appropriate output limit based on tool type
func getOutputLimitForTool(toolName string) int {
	switch toolName {
	case "bash", "shell", "exec":
		// Bash output can be very large
		return MediumOutputLimit
	case "grep", "search", "find", "rg":
		// Search results - usually line-based, can be large
		return SmallOutputLimit
	case "view", "read", "cat":
		// File reading - depends on file size
		return SmallOutputLimit
	case "ls", "list", "dir":
		// Directory listing - usually small
		return SmallOutputLimit
	case "fetch", "curl", "wget":
		// Web fetch - can be large
		return MediumOutputLimit
	case "glob":
		// File finding - can return many paths
		return SmallOutputLimit
	default:
		// Default conservative limit
		return SmallOutputLimit
	}
}

// truncateToTokenLimit truncates content to fit within token limit
func truncateToTokenLimit(content string, maxTokens int) string {
	// First try byte-based truncation (faster)
	maxBytes := maxTokens * 4 // Rough estimate: 4 chars per token
	if len(content) <= maxBytes {
		return content
	}
	
	// Truncate to maxBytes first
	truncated := content[:maxBytes]
	
	// Then refine by counting tokens and adjusting
	currentTokens := countTokens(truncated)
	if currentTokens <= maxTokens {
		return truncated
	}
	
	// Binary search for exact token limit
	low, high := 0, len(truncated)
	for low < high {
		mid := (low + high + 1) / 2
		testContent := truncated[:mid]
		if countTokens(testContent) <= maxTokens {
			low = mid
		} else {
			high = mid - 1
		}
	}
	
	return truncated[:low]
}

// writeToTmpFile writes content to a session-scoped tmp file
func writeToTmpFile(content string, toolName string, workingDir string, sessionID string) (string, error) {
	// Create session-scoped tmp directory
	tmpDir := filepath.Join(workingDir, "nexora-output-"+sessionID)
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create tmp directory: %w", err)
	}
	
	// Generate filename
	timestamp := time.Now().Format("20060102150405.000000000")
	safeToolName := strings.ReplaceAll(toolName, "/", "_")
	filename := fmt.Sprintf("%s-%s.txt", timestamp, safeToolName)
	tmpPath := filepath.Join(tmpDir, filename)
	
	// Write content
	if err := os.WriteFile(tmpPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write tmp file: %w", err)
	}
	
	return tmpPath, nil
}

// CleanupSessionOutputFiles removes all output files for a session
func CleanupSessionOutputFiles(workingDir string, sessionID string) error {
	tmpDir := filepath.Join(workingDir, "nexora-output-"+sessionID)
	return os.RemoveAll(tmpDir)
}

// FormatOutputForModel creates a user-friendly message about what happened to the output
func FormatOutputForModel(result OutputResult, toolName string) string {
	switch result.ActionTaken {
	case "returned":
		return ""
	case "truncated":
		return fmt.Sprintf("[Output truncated from %d bytes (%d tokens) to fit context limit. Use view/grep tools for complete results.]",
			result.OriginalSize, result.TokenCount)
	case "written_to_file":
		return fmt.Sprintf("[Output written to file due to size (%d tokens, %d bytes). File: %s]",
			result.TokenCount, result.OriginalSize, result.FilePath)
	case "returned_empty":
		return "[No output]"
	default:
		return ""
	}
}
