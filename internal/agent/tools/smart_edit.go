package tools

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"charm.land/fantasy"
	"github.com/nexora/nexora/internal/history"
	"github.com/nexora/nexora/internal/permission"
)

// SmartEditParams uses line numbers instead of string matching for 100% reliability
type SmartEditParams struct {
	FilePath  string `json:"file_path" description:"The absolute path to the file to modify"`
	StartLine int    `json:"start_line" description:"The starting line number (1-indexed) to replace"`
	EndLine   int    `json:"end_line" description:"The ending line number (1-indexed) to replace (inclusive)"`
	NewString string `json:"new_string" description:"The replacement text"`
}

const SmartEditToolName = "smart_edit"

//go:embed smart_edit.md
var smartEditDescription []byte

type smartEditContext struct {
	ctx         context.Context
	permissions permission.Service
	files       history.Service
	workingDir  string
}

// NewSmartEditTool creates a line-number based edit tool that never fails due to whitespace
func NewSmartEditTool(permissions permission.Service, files history.Service, workingDir string) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		SmartEditToolName,
		string(smartEditDescription),
		func(ctx context.Context, params SmartEditParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			slog.Info("SMART_EDIT TOOL INVOKED",
				"file", params.FilePath,
				"start_line", params.StartLine,
				"end_line", params.EndLine,
				"new_string_length", len(params.NewString))

			editCtx := smartEditContext{
				ctx:         ctx,
				permissions: permissions,
				files:       files,
				workingDir:  workingDir,
			}

			return executeSmartEdit(editCtx, params)
		},
	)
}

func executeSmartEdit(edit smartEditContext, params SmartEditParams) (fantasy.ToolResponse, error) {
	// Validate parameters
	if params.StartLine < 1 {
		return fantasy.NewTextErrorResponse("start_line must be >= 1"), nil
	}
	if params.EndLine < params.StartLine {
		return fantasy.NewTextErrorResponse("end_line must be >= start_line"), nil
	}

	// Read file
	content, err := os.ReadFile(params.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fantasy.NewTextErrorResponse(fmt.Sprintf("file not found: %s", params.FilePath)), nil
		}
		return fantasy.ToolResponse{}, fmt.Errorf("failed to read file: %w", err)
	}

	oldContent := string(content)
	lines := strings.Split(oldContent, "\n")

	// Validate line numbers
	if params.StartLine > len(lines) {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("start_line %d exceeds file length (%d lines)", params.StartLine, len(lines))), nil
	}
	if params.EndLine > len(lines) {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("end_line %d exceeds file length (%d lines)", params.EndLine, len(lines))), nil
	}

	// Extract the lines to be replaced
	oldString := strings.Join(lines[params.StartLine-1:params.EndLine], "\n")

	// Build new content
	var newLines []string
	newLines = append(newLines, lines[:params.StartLine-1]...)

	// Add new content (split into lines if multi-line)
	if params.NewString != "" {
		newContentLines := strings.Split(params.NewString, "\n")
		newLines = append(newLines, newContentLines...)
	}

	// Add remaining lines after the replaced section
	if params.EndLine < len(lines) {
		newLines = append(newLines, lines[params.EndLine:]...)
	}

	newContent := strings.Join(newLines, "\n")

	// Write back
	err = os.WriteFile(params.FilePath, []byte(newContent), 0o600)
	if err != nil {
		return fantasy.ToolResponse{}, fmt.Errorf("failed to write file: %w", err)
	}

	// Calculate stats
	linesReplaced := params.EndLine - params.StartLine + 1
	newLinesAdded := len(strings.Split(params.NewString, "\n"))

	slog.Info("smart_edit completed",
		"file", params.FilePath,
		"lines_replaced", linesReplaced,
		"new_lines", newLinesAdded)

	return fantasy.WithResponseMetadata(
		fantasy.NewTextResponse(fmt.Sprintf("✓ Edited %s (lines %d-%d replaced with %d new lines)",
			params.FilePath, params.StartLine, params.EndLine, newLinesAdded)),
		map[string]any{
			"lines_replaced": linesReplaced,
			"new_lines":      newLinesAdded,
			"old_content":    oldString,
			"new_content":    params.NewString,
		},
	), nil
}
