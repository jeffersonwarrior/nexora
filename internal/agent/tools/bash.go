package tools

import (
	"bytes"
	"cmp"
	"context"
	_ "embed"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"charm.land/fantasy"
	"github.com/nexora/nexora/internal/config"
	"github.com/nexora/nexora/internal/permission"
	"github.com/nexora/nexora/internal/shell"
)

type BashParams struct {
	Description     string `json:"description" description:"A brief description of what the command does, try to keep it under 30 characters or so"`
	Command         string `json:"command" description:"The command to execute"`
	WorkingDir      string `json:"working_dir,omitempty" description:"The working directory to execute the command in (defaults to current directory)"`
	RunInBackground bool   `json:"run_in_background,omitempty" description:"Set to true (boolean) to run this command in the background. Use job_output to read the output later."`
	ShellID         string `json:"shell_id,omitempty" description:"Shell ID to continue (for TMUX sessions or background jobs)"`
}

type BashPermissionsParams struct {
	Description     string `json:"description"`
	Command         string `json:"command"`
	WorkingDir      string `json:"working_dir"`
	RunInBackground bool   `json:"run_in_background"`
	ShellID         string `json:"shell_id"`
}

type BashResponseMetadata struct {
	StartTime        int64  `json:"start_time"`
	EndTime          int64  `json:"end_time"`
	Output           string `json:"output"`
	Description      string `json:"description"`
	WorkingDirectory string `json:"working_directory"`
	Background       bool   `json:"background,omitempty"`
	ShellID          string `json:"shell_id,omitempty"`
	
	// TMUX fields
	TmuxSessionID string `json:"tmux_session_id,omitempty"`
	TmuxPaneID    string `json:"tmux_pane_id,omitempty"`
	TmuxAvailable bool   `json:"tmux_available,omitempty"`
}

const (
	BashToolName = "bash"

	AutoBackgroundThreshold = 1 * time.Minute // Commands taking longer automatically become background jobs
	MaxOutputLength         = 30000
	BashNoOutput            = "no output"
)

//go:embed bash.tpl
var bashDescriptionTmpl []byte

var bashDescriptionTpl = template.Must(
	template.New("bashDescription").
		Parse(string(bashDescriptionTmpl)),
)

type bashDescriptionData struct {
	BannedCommands  string
	MaxOutputLength int
	Attribution     config.Attribution
	ModelName       string
}

var bannedCommands = []string{}

func bashDescription(attribution *config.Attribution, modelName string) string {
	bannedCommandsStr := strings.Join(bannedCommands, ", ")
	var out bytes.Buffer
	if err := bashDescriptionTpl.Execute(&out, bashDescriptionData{
		BannedCommands:  bannedCommandsStr,
		MaxOutputLength: MaxOutputLength,
		Attribution:     *attribution,
		ModelName:       modelName,
	}); err != nil {
		// this should never happen.
		panic("failed to execute bash description template: " + err.Error())
	}
	return out.String()
}

func blockFuncs() []shell.BlockFunc {
	return []shell.BlockFunc{
		// ========================================
		// Phase 1: Critical Safety Blockers
		// ========================================

		// Block recursive force removal (rm -rf)
		shell.ArgumentsBlocker("rm", []string{}, []string{"-rf"}),
		shell.ArgumentsBlocker("rm", []string{}, []string{"-fr"}),
		shell.ArgumentsBlocker("rm", []string{}, []string{"--recursive", "--force"}),
		shell.ArgumentsBlocker("rm", []string{}, []string{"-r", "-f"}),

		// Block killing Nexora or TMUX processes
		func(args []string) bool {
			if len(args) == 0 {
				return false
			}
			cmd := args[0]
			if cmd == "pkill" || cmd == "killall" {
				for _, arg := range args[1:] {
					lower := strings.ToLower(arg)
					if strings.Contains(lower, "nexora") ||
						strings.Contains(lower, "tmux") {
						return true
					}
				}
			}
			return false
		},

		// Block killing init/systemd (PID 1)
		func(args []string) bool {
			if len(args) >= 2 && args[0] == "kill" {
				for _, arg := range args[1:] {
					if arg == "1" || arg == "-1" {
						return true
					}
				}
			}
			return false
		},

		// Block disk format and wipe commands
		shell.CommandsBlocker([]string{
			"mkfs",
			"mkfs.ext4",
			"mkfs.ext3",
			"mkfs.xfs",
			"mkfs.btrfs",
			"fdisk",
			"dd",
			"shred",
		}),

		// Block fork bombs and infinite loops
		func(args []string) bool {
			cmdStr := strings.Join(args, " ")
			forkBombPatterns := []string{
				":()",      // Classic bash fork bomb
				"while true", // Infinite while loop
				":|:",      // Fork bomb variant
			}
			for _, pattern := range forkBombPatterns {
				if strings.Contains(cmdStr, pattern) {
					return true
				}
			}
			return false
		},

		// Block dangerous git operations
		shell.ArgumentsBlocker("git", []string{"push"}, []string{"--force"}),
		shell.ArgumentsBlocker("git", []string{"push"}, []string{"-f"}),

		// Block dangerous chmod operations
		func(args []string) bool {
			if len(args) >= 2 && args[0] == "chmod" {
				for _, arg := range args[1:] {
					if arg == "777" || arg == "000" {
						return true
					}
				}
			}
			return false
		},

		// Block operations on critical system directories
		func(args []string) bool {
			if len(args) < 2 {
				return false
			}

			dangerousPaths := []string{"/bin", "/sbin", "/usr/bin", "/usr/sbin", "/etc", "/sys", "/proc"}
			destructiveCmds := []string{"rm", "mv", "chmod", "chown"}

			cmd := args[0]
			if !sliceContains(destructiveCmds, cmd) {
				return false
			}

			for _, arg := range args[1:] {
				for _, dangerous := range dangerousPaths {
					if strings.HasPrefix(arg, dangerous) {
						return true
					}
				}
			}
			return false
		},
	}
}

func sliceContains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func NewBashTool(permissions permission.Service, workingDir string, attribution *config.Attribution, modelName string) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		BashToolName,
		string(bashDescription(attribution, modelName)),
		func(ctx context.Context, params BashParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if params.Command == "" {
				return fantasy.NewTextErrorResponse("missing command"), nil
			}

			// Determine working directory
			execWorkingDir := cmp.Or(params.WorkingDir, workingDir)

			isSafeReadOnly := false
			cmdLower := strings.ToLower(params.Command)

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
				return fantasy.ToolResponse{}, fmt.Errorf("session ID is required for executing shell command")
			}
			if !isSafeReadOnly {
				p := permissions.Request(
					permission.CreatePermissionRequest{
						SessionID:   sessionID,
						Path:        execWorkingDir,
						ToolCallID:  call.ID,
						ToolName:    BashToolName,
						Action:      "execute",
						Description: fmt.Sprintf("Execute command: %s", params.Command),
						Params:      BashPermissionsParams(params),
					},
				)
				if !p {
					return fantasy.ToolResponse{}, permission.ErrorPermissionDenied
				}
			}

			// === TMUX Integration ===
			// If TMUX is available, use it for command execution
			if shell.IsTmuxAvailable() {
				return executeTmuxCommand(ctx, params, call, execWorkingDir, sessionID)
			}

			// === Legacy Execution (no TMUX) ===

			// If explicitly requested as background, start immediately with detached context
			if params.RunInBackground {
				startTime := time.Now()
				bgManager := shell.GetBackgroundShellManager()
				bgManager.Cleanup()
				// Use background context so it continues after tool returns
				bgShell, err := bgManager.Start(context.Background(), execWorkingDir, blockFuncs(), params.Command, params.Description)
				if err != nil {
					return fantasy.ToolResponse{}, fmt.Errorf("error starting background shell: %w", err)
				}

				// Wait a short time to detect fast failures (blocked commands, syntax errors, etc.)
				time.Sleep(1 * time.Second)
				stdout, stderr, done, execErr := bgShell.GetOutput()

				if done {
					// Command failed or completed very quickly
					bgManager.Remove(bgShell.ID)

					interrupted := shell.IsInterrupt(execErr)
					exitCode := shell.ExitCode(execErr)
					if exitCode == 0 && !interrupted && execErr != nil {
						return fantasy.ToolResponse{}, fmt.Errorf("[Job %s] error executing command: %w", bgShell.ID, execErr)
					}

					stdout = formatOutput(stdout, stderr, execErr)

					metadata := BashResponseMetadata{
						StartTime:        startTime.UnixMilli(),
						EndTime:          time.Now().UnixMilli(),
						Output:           stdout,
						Description:      params.Description,
						Background:       params.RunInBackground,
						WorkingDirectory: bgShell.WorkingDir,
					}
					if stdout == "" {
						return fantasy.WithResponseMetadata(fantasy.NewTextResponse(BashNoOutput), metadata), nil
					}
					stdout += fmt.Sprintf("\n\n<cwd>%s</cwd>", normalizeWorkingDir(bgShell.WorkingDir))
					return fantasy.WithResponseMetadata(fantasy.NewTextResponse(stdout), metadata), nil
				}

				// Still running after fast-failure check - return as background job
				metadata := BashResponseMetadata{
					StartTime:        startTime.UnixMilli(),
					EndTime:          time.Now().UnixMilli(),
					Description:      params.Description,
					WorkingDirectory: bgShell.WorkingDir,
					Background:       true,
					ShellID:          bgShell.ID,
				}
				response := fmt.Sprintf("Background shell started with ID: %s\n\nUse job_output tool to view output or job_kill to terminate.", bgShell.ID)
				return fantasy.WithResponseMetadata(fantasy.NewTextResponse(response), metadata), nil
			}

			// Start synchronous execution with auto-background support
			startTime := time.Now()

			// Start with detached context so it can survive if moved to background
			bgManager := shell.GetBackgroundShellManager()
			bgManager.Cleanup()
			bgShell, err := bgManager.Start(context.Background(), execWorkingDir, blockFuncs(), params.Command, params.Description)
			if err != nil {
				return fantasy.ToolResponse{}, fmt.Errorf("error starting shell: %w", err)
			}

			// Wait for either completion, auto-background threshold, or context cancellation
			ticker := time.NewTicker(100 * time.Millisecond)
			defer ticker.Stop()
			timeout := time.After(AutoBackgroundThreshold)

			var stdout, stderr string
			var done bool
			var execErr error

		waitLoop:
			for {
				select {
				case <-ticker.C:
					stdout, stderr, done, execErr = bgShell.GetOutput()
					if done {
						break waitLoop
					}
				case <-timeout:
					stdout, stderr, done, execErr = bgShell.GetOutput()
					break waitLoop
				case <-ctx.Done():
					// Incoming context was cancelled before we moved to background
					// Kill the shell and return error
					bgManager.Kill(bgShell.ID)
					return fantasy.ToolResponse{}, ctx.Err()
				}
			}

			if done {
				// Command completed within threshold - return synchronously
				// Remove from background manager since we're returning directly
				// Don't call Kill() as it cancels the context and corrupts the exit code
				bgManager.Remove(bgShell.ID)

				interrupted := shell.IsInterrupt(execErr)
				exitCode := shell.ExitCode(execErr)
				if exitCode == 0 && !interrupted && execErr != nil {
					return fantasy.ToolResponse{}, fmt.Errorf("[Job %s] error executing command: %w", bgShell.ID, execErr)
				}

				stdout = formatOutput(stdout, stderr, execErr)

				metadata := BashResponseMetadata{
					StartTime:        startTime.UnixMilli(),
					EndTime:          time.Now().UnixMilli(),
					Output:           stdout,
					Description:      params.Description,
					Background:       params.RunInBackground,
					WorkingDirectory: bgShell.WorkingDir,
				}
				if stdout == "" {
					return fantasy.WithResponseMetadata(fantasy.NewTextResponse(BashNoOutput), metadata), nil
				}
				stdout += fmt.Sprintf("\n\n<cwd>%s</cwd>", normalizeWorkingDir(bgShell.WorkingDir))
				return fantasy.WithResponseMetadata(fantasy.NewTextResponse(stdout), metadata), nil
			}

			// Still running - keep as background job
			metadata := BashResponseMetadata{
				StartTime:        startTime.UnixMilli(),
				EndTime:          time.Now().UnixMilli(),
				Description:      params.Description,
				WorkingDirectory: bgShell.WorkingDir,
				Background:       true,
				ShellID:          bgShell.ID,
			}
			response := fmt.Sprintf("Command is taking longer than expected and has been moved to background.\n\nBackground shell ID: %s\n\nUse job_output tool to view output or job_kill to terminate.", bgShell.ID)
			return fantasy.WithResponseMetadata(fantasy.NewTextResponse(response), metadata), nil
		})
}

// formatOutput formats the output of a completed command with error handling
func formatOutput(stdout, stderr string, execErr error) string {
	interrupted := shell.IsInterrupt(execErr)
	exitCode := shell.ExitCode(execErr)

	stdout = truncateOutput(stdout)
	stderr = truncateOutput(stderr)

	errorMessage := stderr
	if errorMessage == "" && execErr != nil {
		errorMessage = execErr.Error()
	}

	if interrupted {
		if errorMessage != "" {
			errorMessage += "\n"
		}
		errorMessage += "Command was aborted before completion"
	} else if exitCode != 0 {
		if errorMessage != "" {
			errorMessage += "\n"
		}
		errorMessage += fmt.Sprintf("Exit code %d", exitCode)
	}

	hasBothOutputs := stdout != "" && stderr != ""

	if hasBothOutputs {
		stdout += "\n"
	}

	if errorMessage != "" {
		stdout += "\n" + errorMessage
	}

	return stdout
}

func truncateOutput(content string) string {
	if len(content) <= MaxOutputLength {
		return content
	}

	halfLength := MaxOutputLength / 2
	start := content[:halfLength]
	end := content[len(content)-halfLength:]

	truncatedLinesCount := countLines(content[halfLength : len(content)-halfLength])
	return fmt.Sprintf("%s\n\n... [%d lines truncated] ...\n\n%s", start, truncatedLinesCount, end)
}

func countLines(s string) int {
	if s == "" {
		return 0
	}
	return len(strings.Split(s, "\n"))
}

func normalizeWorkingDir(path string) string {
	if runtime.GOOS == "windows" {
		cwd, err := os.Getwd()
		if err != nil {
			cwd = "C:"
		}
		path = strings.ReplaceAll(path, filepath.VolumeName(cwd), "")
	}

	return filepath.ToSlash(path)
}

// === TMUX Integration ===

// executeTmuxCommand handles bash execution using TMUX if available
func executeTmuxCommand(ctx context.Context, params BashParams, call fantasy.ToolCall, execWorkingDir string, sessionID string) (fantasy.ToolResponse, error) {
	tmuxManager := shell.GetTmuxManager()
	tmuxAvailable := shell.IsTmuxAvailable()
	
	startTime := time.Now()
	
	var session *shell.TmuxSession
	var err error
	var ok bool
	
	// Determine if we should create a new session or use existing
	if params.ShellID != "" {
		// Continue existing session
		session, ok = tmuxManager.GetSession(params.ShellID)
		if !ok {
			// Session not found, create new
			session, err = tmuxManager.NewTmuxSession(params.ShellID, execWorkingDir, params.Command, params.Description)
		}
	} else {
		// Create new session
		newSessionID := fmt.Sprintf("%s-%d", sessionID, time.Now().UnixNano())
		session, err = tmuxManager.NewTmuxSession(newSessionID, execWorkingDir, params.Command, params.Description)
	}
	
	if err != nil {
		return fantasy.ToolResponse{}, fmt.Errorf("failed to manage TMUX session: %w", err)
	}
	
	// Small delay to allow TMUX to execute the command
	time.Sleep(100 * time.Millisecond)
	
	// Get output
	output, err := tmuxManager.CaptureOutput(session.ID)
	if err != nil {
		return fantasy.ToolResponse{}, fmt.Errorf("failed to capture TMUX output: %w", err)
	}
	
	// Apply output management (truncation, tmp file, etc.)
	managedOutput := ManageOutput(output, "bash", execWorkingDir, sessionID)
	
	// Format output with working directory
	if managedOutput.Content != "" {
		managedOutput.Content += fmt.Sprintf("\n\n<cwd>%s</cwd>", normalizeWorkingDir(execWorkingDir))
	}
	
	// Build metadata
	metadata := BashResponseMetadata{
		StartTime:        startTime.UnixMilli(),
		EndTime:          time.Now().UnixMilli(),
		Output:           managedOutput.Content,
		Description:      params.Description,
		WorkingDirectory: execWorkingDir,
		ShellID:          session.ID,
		TmuxSessionID:    session.SessionName,
		TmuxPaneID:       session.PaneID,
		TmuxAvailable:    tmuxAvailable,
	}
	
	// Add truncation notice if applicable
	if truncationMsg := FormatOutputForModel(managedOutput, "bash"); truncationMsg != "" {
		managedOutput.Content += "\n\n" + truncationMsg
	}
	
	return fantasy.WithResponseMetadata(fantasy.NewTextResponse(managedOutput.Content), metadata), nil
}

// === Alias Integration for job_kill and job_output ===

// executeJobManagement handles job_kill and job_output functionality via bash tool
func executeJobManagement(params BashParams, startTime time.Time, execWorkingDir string) (fantasy.ToolResponse, error) {
	bgManager := shell.GetBackgroundShellManager()
	
	// Get the background shell
	bgShell, ok := bgManager.Get(params.ShellID)
	if !ok {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("background shell not found: %s", params.ShellID)), nil
	}
	
	// Determine if this is a kill or output request
	if params.Command == "" || strings.HasPrefix(strings.ToLower(params.Command), "output") {
		// job_output functionality
		stdout, stderr, done, execErr := bgShell.GetOutput()
		
		var outputParts []string
		if stdout != "" {
			outputParts = append(outputParts, stdout)
		}
		if stderr != "" {
			outputParts = append(outputParts, stderr)
		}
		
		status := "running"
		if done {
			status = "completed"
			if execErr != nil {
				exitCode := shell.ExitCode(execErr)
				if exitCode != 0 {
					outputParts = append(outputParts, fmt.Sprintf("Exit code %d", exitCode))
				}
			}
		}
		
		output := strings.Join(outputParts, "\n")
		if output == "" {
			output = BashNoOutput
		}
		
		metadata := BashResponseMetadata{
			StartTime:        startTime.UnixMilli(),
			EndTime:          time.Now().UnixMilli(),
			Output:           output,
			Description:      bgShell.Description,
			WorkingDirectory: bgShell.WorkingDir,
			ShellID:          params.ShellID,
			Background:       true,
		}
		
		result := fmt.Sprintf("Status: %s\n\n%s", status, output)
		return fantasy.WithResponseMetadata(fantasy.NewTextResponse(result), metadata), nil
	} else {
		// job_kill functionality - command should be "exit" or kill indicator
		err := bgManager.Kill(params.ShellID)
		if err != nil {
			return fantasy.NewTextErrorResponse(err.Error()), nil
		}
		
		result := fmt.Sprintf("Background shell %s terminated successfully", params.ShellID)
		metadata := BashResponseMetadata{
			StartTime:        startTime.UnixMilli(),
			EndTime:          time.Now().UnixMilli(),
			Description:      bgShell.Description,
			WorkingDirectory: execWorkingDir,
			ShellID:          params.ShellID,
		}
		
		return fantasy.WithResponseMetadata(fantasy.NewTextResponse(result), metadata), nil
	}
}
