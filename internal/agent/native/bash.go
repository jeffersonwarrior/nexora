package native

import (
	"cmp"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nexora/nexora/internal/permission"
)

// BashParams are the parameters for the bash tool
type BashParams struct {
	Description     string `json:"description" description:"A brief description of what the command does, try to keep it under 30 characters or so"`
	Command         string `json:"command" description:"The command to execute"`
	WorkingDir      string `json:"working_dir,omitempty" description:"The working directory to execute the command in (defaults to current directory)"`
	RunInBackground bool   `json:"run_in_background,omitempty" description:"Set to true (boolean) to run this command in the background. Use job_output to read the output later."`
}

// BashResponseMetadata contains metadata for bash tool responses
type BashResponseMetadata struct {
	StartTime        int64  `json:"start_time"`
	EndTime          int64  `json:"end_time"`
	Output           string `json:"output"`
	Description      string `json:"description"`
	WorkingDirectory string `json:"working_directory"`
	Background       bool   `json:"background,omitempty"`
	ShellID          string `json:"shell_id,omitempty"`
}

// GetSessionFromContext extracts session ID from context
// This function needs to be implemented based on how session IDs are stored in context
func GetSessionFromContext(ctx context.Context) string {
	// TODO: Implement based on actual context structure
	if sessionID, ok := ctx.Value("session_id").(string); ok {
		return sessionID
	}
	return ""
}

// Safe commands that don't require permission prompting
var safeCommands = []string{
	"ls", "pwd", "echo", "cat", "head", "tail", "grep", "find", "which",
	"whereis", "date", "whoami", "id", "uptime", "df", "du", "ps", "top",
}

// NewBashTool creates a new bash tool using the native framework
func NewBashTool(permissions permission.Service, workingDir string, description string) AgentTool {
	info := ToolInfo{
		Name:        "bash",
		Description: description,
		Parameters:  BashParams{},
	}

	handler := func(ctx context.Context, params any, call ToolCall) (ToolResponse, error) {
		bashParams, ok := params.(BashParams)
		if !ok {
			return NewTextErrorResponse("invalid parameters for bash tool"), nil
		}

		if bashParams.Command == "" {
			return NewTextErrorResponse("missing command"), nil
		}

		// Determine working directory
		execWorkingDir := cmp.Or(bashParams.WorkingDir, workingDir)

		isSafeReadOnly := false
		cmdLower := strings.ToLower(bashParams.Command)

		for _, safe := range safeCommands {
			if strings.HasPrefix(cmdLower, safe) {
				if len(cmdLower) == len(safe) || cmdLower[len(safe)] == ' ' || cmdLower[len(safe)] == '-' {
					isSafeReadOnly = true
					break
				}
			}
		}

		sessionID := GetSessionFromContext(ctx)
		if sessionID == "" {
			return ToolResponse{}, fmt.Errorf("session ID is required for executing shell command")
		}

		// Check permissions for unsafe commands
		if !isSafeReadOnly {
			p := permissions.Request(
				permission.CreatePermissionRequest{
					SessionID:   sessionID,
					Path:        execWorkingDir,
					ToolCallID:  call.ID,
					ToolName:    "bash",
					Action:      "execute",
					Description: fmt.Sprintf("Execute command: %s", bashParams.Command),
					Params:      bashParams,
				},
			)
			if !p {
				return ToolResponse{}, fmt.Errorf("permission denied")
			}
		}

		// Execute the command (simplified implementation)
		startTime := time.Now()

		// TODO: Replace with actual shell execution using mvdan/sh
		output := fmt.Sprintf("Command executed: %s\nWorking dir: %s\nBackground: %v",
			bashParams.Command, execWorkingDir, bashParams.RunInBackground)

		endTime := time.Now()

		metadata := BashResponseMetadata{
			StartTime:        startTime.Unix(),
			EndTime:          endTime.Unix(),
			Output:           output,
			Description:      bashParams.Description,
			WorkingDirectory: execWorkingDir,
			Background:       bashParams.RunInBackground,
		}

		return WithResponseMetadata(NewTextResponse(output), map[string]any{
			"bash_metadata": metadata,
		}), nil
	}

	return &BasicTool{
		info:    info,
		handler: handler,
	}
}
