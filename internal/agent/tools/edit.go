package tools

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"charm.land/fantasy"
	"github.com/nexora/cli/internal/aiops"
	"github.com/nexora/cli/internal/csync"
	"github.com/nexora/cli/internal/diff"
	"github.com/nexora/cli/internal/filepathext"
	"github.com/nexora/cli/internal/fsext"
	"github.com/nexora/cli/internal/history"

	"github.com/nexora/cli/internal/lsp"
	"github.com/nexora/cli/internal/permission"
)

type EditParams struct {
	FilePath   string `json:"file_path" description:"The absolute path to the file to modify"`
	OldString  string `json:"old_string" description:"The text to replace"`
	NewString  string `json:"new_string" description:"The text to replace it with"`
	ReplaceAll bool   `json:"replace_all,omitempty" description:"Replace all occurrences of old_string (default false)"`
}

type EditPermissionsParams struct {
	FilePath   string `json:"file_path"`
	OldContent string `json:"old_content,omitempty"`
	NewContent string `json:"new_content,omitempty"`
}

type EditResponseMetadata struct {
	Additions  int    `json:"additions"`
	Removals   int    `json:"removals"`
	OldContent string `json:"old_content,omitempty"`
	NewContent string `json:"new_content,omitempty"`
}

const EditToolName = "edit"

//go:embed edit.md
var editDescription []byte

type editContext struct {
	ctx         context.Context
	permissions permission.Service
	files       history.Service
	workingDir  string
	aiops       aiops.Ops
}

func NewEditTool(lspClients *csync.Map[string, *lsp.Client], permissions permission.Service, files history.Service, workingDir string, aiops aiops.Ops) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		EditToolName,
		string(editDescription),
		func(ctx context.Context, params EditParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if params.FilePath == "" {
				return fantasy.NewTextErrorResponse("file_path is required"), nil
			}

			params.FilePath = filepathext.SmartJoin(workingDir, params.FilePath)

			var response fantasy.ToolResponse
			var err error

			editCtx := editContext{ctx, permissions, files, workingDir, aiops}

			if params.OldString == "" {
				response, err = createNewFile(editCtx, params.FilePath, params.NewString, call)
				if err != nil {
					return response, err
				}
			}

			if params.NewString == "" {
				response, err = deleteContent(editCtx, params.FilePath, params.OldString, params.ReplaceAll, call)
				if err != nil {
					return response, err
				}
			}

			response, err = replaceContent(editCtx, params.FilePath, params.OldString, params.NewString, params.ReplaceAll, call)
			if err != nil {
				return response, err
			}
			if response.IsError {
				// Return early if there was an error during content replacement
				// This prevents unnecessary LSP diagnostics processing
				return response, nil
			}

			notifyLSPs(ctx, lspClients, params.FilePath)

			text := fmt.Sprintf("<result>\n%s\n</result>\n", response.Content)
			text += getDiagnostics(params.FilePath, lspClients)
			response.Content = text
			return response, nil
		})
}

func createNewFile(edit editContext, filePath, content string, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
	fileInfo, err := os.Stat(filePath)
	if err == nil {
		if fileInfo.IsDir() {
			return fantasy.NewTextErrorResponse(fmt.Sprintf("path is a directory, not a file: %s", filePath)), nil
		}
		return fantasy.NewTextErrorResponse(fmt.Sprintf("file already exists: %s", filePath)), nil
	} else if !os.IsNotExist(err) {
		return fantasy.ToolResponse{}, fmt.Errorf("failed to access file: %w", err)
	}

	dir := filepath.Dir(filePath)
	if err = os.MkdirAll(dir, 0o755); err != nil {
		return fantasy.ToolResponse{}, fmt.Errorf("failed to create parent directories: %w", err)
	}

	sessionID := GetSessionFromContext(edit.ctx)
	if sessionID == "" {
		return fantasy.ToolResponse{}, fmt.Errorf("session ID is required for creating a new file")
	}

	_, additions, removals := diff.GenerateDiff(
		"",
		content,
		strings.TrimPrefix(filePath, edit.workingDir),
	)
	p := edit.permissions.Request(
		permission.CreatePermissionRequest{
			SessionID:   sessionID,
			Path:        fsext.PathOrPrefix(filePath, edit.workingDir),
			ToolCallID:  call.ID,
			ToolName:    EditToolName,
			Action:      "write",
			Description: fmt.Sprintf("Create file %s", filePath),
			Params: EditPermissionsParams{
				FilePath:   filePath,
				OldContent: "",
				NewContent: content,
			},
		},
	)
	if !p {
		return fantasy.ToolResponse{}, permission.ErrorPermissionDenied
	}

	err = os.WriteFile(filePath, []byte(content), 0o644)
	if err != nil {
		return fantasy.ToolResponse{}, fmt.Errorf("failed to write file: %w", err)
	}

	// File can't be in the history so we create a new file history
	_, err = edit.files.Create(edit.ctx, sessionID, filePath, "")
	if err != nil {
		// Log error but don't fail the operation
		return fantasy.ToolResponse{}, fmt.Errorf("error creating file history: %w", err)
	}

	// Add the new content to the file history
	_, err = edit.files.CreateVersion(edit.ctx, sessionID, filePath, content)
	if err != nil {
		// Log error but don't fail the operation
		slog.Error("Error creating file history version", "error", err)
	}

	recordFileWrite(filePath)
	recordFileRead(filePath)

	return fantasy.WithResponseMetadata(
		fantasy.NewTextResponse("File created: "+filePath),
		EditResponseMetadata{
			OldContent: "",
			NewContent: content,
			Additions:  additions,
			Removals:   removals,
		},
	), nil
}

func deleteContent(edit editContext, filePath, oldString string, replaceAll bool, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fantasy.NewTextErrorResponse(fmt.Sprintf("file not found: %s", filePath)), nil
		}
		return fantasy.ToolResponse{}, fmt.Errorf("failed to access file: %w", err)
	}

	if fileInfo.IsDir() {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("path is a directory, not a file: %s", filePath)), nil
	}

	if getLastReadTime(filePath).IsZero() {
		return fantasy.NewTextErrorResponse("you must read the file before editing it. Use the View tool first"), nil
	}

	modTime := fileInfo.ModTime()
	lastRead := getLastReadTime(filePath)
	if modTime.After(lastRead) {
		return fantasy.NewTextErrorResponse(
			fmt.Sprintf("file %s has been modified since it was last read (mod time: %s, last read: %s)",
				filePath, modTime.Format(time.RFC3339), lastRead.Format(time.RFC3339),
			)), nil
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return fantasy.ToolResponse{}, fmt.Errorf("failed to read file: %w", err)
	}

	oldContent, isCrlf := fsext.ToUnixLineEndings(string(content))

	var newContent string
	var deletionCount int

	if replaceAll {
		newContent = strings.ReplaceAll(oldContent, oldString, "")
		deletionCount = strings.Count(oldContent, oldString)
		if deletionCount == 0 {
			return fantasy.NewTextErrorResponse("old_string not found in file. Make sure it matches exactly, including whitespace and line breaks"), nil
		}
	} else {
		index := strings.Index(oldContent, oldString)
		if index == -1 {
			return fantasy.NewTextErrorResponse("old_string not found in file. Make sure it matches exactly, including whitespace and line breaks"), nil
		}

		lastIndex := strings.LastIndex(oldContent, oldString)
		if index != lastIndex {
			return fantasy.NewTextErrorResponse("old_string appears multiple times in the file. Please provide more context to ensure a unique match, or set replace_all to true"), nil
		}

		newContent = oldContent[:index] + oldContent[index+len(oldString):]
		deletionCount = 1
	}

	sessionID := GetSessionFromContext(edit.ctx)

	if sessionID == "" {
		return fantasy.ToolResponse{}, fmt.Errorf("session ID is required for creating a new file")
	}

	_, additions, removals := diff.GenerateDiff(
		oldContent,
		newContent,
		strings.TrimPrefix(filePath, edit.workingDir),
	)

	p := edit.permissions.Request(
		permission.CreatePermissionRequest{
			SessionID:   sessionID,
			Path:        fsext.PathOrPrefix(filePath, edit.workingDir),
			ToolCallID:  call.ID,
			ToolName:    EditToolName,
			Action:      "write",
			Description: fmt.Sprintf("Delete content from file %s", filePath),
			Params: EditPermissionsParams{
				FilePath:   filePath,
				OldContent: oldContent,
				NewContent: newContent,
			},
		},
	)
	if !p {
		return fantasy.ToolResponse{}, permission.ErrorPermissionDenied
	}

	if isCrlf {
		newContent, _ = fsext.ToWindowsLineEndings(newContent)
	}

	err = os.WriteFile(filePath, []byte(newContent), 0o644)
	if err != nil {
		return fantasy.ToolResponse{}, fmt.Errorf("failed to write file: %w", err)
	}

	// Check if file exists in history
	file, err := edit.files.GetByPathAndSession(edit.ctx, filePath, sessionID)
	if err != nil {
		_, err = edit.files.Create(edit.ctx, sessionID, filePath, oldContent)
		if err != nil {
			// Log error but don't fail the operation
			return fantasy.ToolResponse{}, fmt.Errorf("error creating file history: %w", err)
		}
	}
	if file.Content != oldContent {
		// User Manually changed the content store an intermediate version
		_, err = edit.files.CreateVersion(edit.ctx, sessionID, filePath, oldContent)
		if err != nil {
			slog.Error("Error creating file history version", "error", err)
		}
	}
	// Store the new version
	_, err = edit.files.CreateVersion(edit.ctx, sessionID, filePath, "")
	if err != nil {
		slog.Error("Error creating file history version", "error", err)
	}

	recordFileWrite(filePath)
	recordFileRead(filePath)

	return fantasy.WithResponseMetadata(
		fantasy.NewTextResponse("Content deleted from file: "+filePath),
		EditResponseMetadata{
			OldContent: oldContent,
			NewContent: newContent,
			Additions:  additions,
			Removals:   removals,
		},
	), nil
}

func replaceContent(edit editContext, filePath, oldString, newString string, replaceAll bool, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
	attemptCount := 0
	// Auto-view file before every edit to ensure we have latest context
	if err := autoViewFileBeforeEdit(edit.ctx, filePath); err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("failed to view file before edit: %v", err)), nil
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fantasy.NewTextErrorResponse(fmt.Sprintf("file not found: %s", filePath)), nil
		}
		return fantasy.ToolResponse{}, fmt.Errorf("failed to access file: %w", err)
	}

	if fileInfo.IsDir() {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("path is a directory, not a file: %s", filePath)), nil
	}

	if getLastReadTime(filePath).IsZero() {
		return fantasy.NewTextErrorResponse("you must read the file before editing it. Use the View tool first"), nil
	}

	modTime := fileInfo.ModTime()
	lastRead := getLastReadTime(filePath)
	if modTime.After(lastRead) {
		return fantasy.NewTextErrorResponse(
			fmt.Sprintf("file %s has been modified since it was last read (mod time: %s, last read: %s)",
				filePath, modTime.Format(time.RFC3339), lastRead.Format(time.RFC3339),
			)), nil
	}

	// Validate old_string exists before proceeding with edit
	if err := ValidateEditString(filePath, oldString, replaceAll); err != nil {
		return fantasy.NewTextErrorResponse(err.Error()), nil
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return fantasy.ToolResponse{}, fmt.Errorf("failed to read file: %w", err)
	}

	oldContent, isCrlf := fsext.ToUnixLineEndings(string(content))

	var newContent string
	var replacementCount int

	if replaceAll {
		newContent = strings.ReplaceAll(oldContent, oldString, newString)
		replacementCount = strings.Count(oldContent, oldString)
		if replacementCount == 0 {
		if replacementCount == 0 {
			// Try whitespace normalization (handles tabs vs spaces from View output)
			if normalized, found := tryNormalizedMatch(oldContent, oldString); found {
				oldString = normalized
				newContent = strings.ReplaceAll(oldContent, oldString, newString)
				replacementCount = strings.Count(oldContent, oldString)
				if replacementCount > 0 {
					// Successfully resolved with whitespace normalization
					goto foundReplaceAll
				}
			}
		}
			// Try AIOPS edit resolution first if available
			if edit.aiops != nil {
				resolution, err := edit.aiops.ResolveEdit(edit.ctx, oldContent, oldString, newString)
				if err == nil && resolution.Confidence > 0.8 {
					oldString = resolution.ExactOldString
					newContent = strings.ReplaceAll(oldContent, oldString, newString)
					replacementCount = strings.Count(oldContent, oldString)
					if replacementCount > 0 {
						// Successfully resolved with AIOPS
						goto foundReplaceAll
					}
				}
			}
			// Attempt self-healing retry with better context
			attemptCount++
			retryParams, err := attemptSelfHealingRetry(edit.ctx, filePath, oldString, newString)
			if err != nil {
				LogEditFailure(EditDiagnosticsInfo{
					FilePath:      filePath,
					OldString:     oldString,
					NewString:     newString,
					FileContent:   oldContent,
					FileSizeBytes: len(oldContent),
					LineCount:     strings.Count(oldContent, "\n") + 1,
					FailureReason: "old_string not found (replaceAll)",
					Context:       PatternMatchAnalysis(oldContent, oldString),
				})
				return fantasy.NewTextErrorResponse("old_string not found in file. Make sure it matches exactly, including whitespace and line breaks"), nil
			}
			// Use the improved parameters from retry
			oldString = retryParams.OldString
			newContent = strings.ReplaceAll(oldContent, oldString, newString)
			replacementCount = strings.Count(oldContent, oldString)
			if replacementCount == 0 {
				return fantasy.NewTextErrorResponse("old_string not found in file. Make sure it matches exactly, including whitespace and line breaks"), nil
			}
		}
	foundReplaceAll:
	} else {
		index := strings.Index(oldContent, oldString)
		if index == -1 {
			// Try AIOPS edit resolution first if available
			if edit.aiops != nil {
				resolution, err := edit.aiops.ResolveEdit(edit.ctx, oldContent, oldString, newString)
				if err == nil && resolution.Confidence > 0.8 {
					oldString = resolution.ExactOldString
					index = strings.Index(oldContent, oldString)
					if index != -1 {
						// Successfully resolved with AIOPS
						goto found
					}
				}
			}
			// Attempt self-healing retry with better context
			attemptCount++
			retryParams, err := attemptSelfHealingRetry(edit.ctx, filePath, oldString, newString)
			if err != nil {
				LogEditFailure(EditDiagnosticsInfo{
					FilePath:      filePath,
					OldString:     oldString,
					NewString:     newString,
					FileContent:   oldContent,
					FileSizeBytes: len(oldContent),
					LineCount:     strings.Count(oldContent, "\n") + 1,
					FailureReason: "old_string not found (replaceAll)",
					Context:       PatternMatchAnalysis(oldContent, oldString),
				})
				return fantasy.NewTextErrorResponse("old_string not found in file. Make sure it matches exactly, including whitespace and line breaks"), nil
			}
			// Use the improved parameters from retry
			oldString = retryParams.OldString
			index = strings.Index(oldContent, oldString)
			if index == -1 {
				return fantasy.NewTextErrorResponse("old_string not found in file. Make sure it matches exactly, including whitespace and line breaks"), nil
			}
		}
	found:

		lastIndex := strings.LastIndex(oldContent, oldString)
		if index != lastIndex {
			// Attempt self-healing retry with better context
			attemptCount++
			retryParams, err := attemptSelfHealingRetry(edit.ctx, filePath, oldString, newString)
			if err != nil {
				return fantasy.NewTextErrorResponse("old_string appears multiple times in the file. Please provide more context to ensure a unique match, or set replace_all to true"), nil
			}
			// Use the improved parameters from retry
			oldString = retryParams.OldString
			index = strings.Index(oldContent, oldString)
			lastIndex = strings.LastIndex(oldContent, oldString)
			if index != lastIndex {
				return fantasy.NewTextErrorResponse("old_string appears multiple times in the file. Please provide more context to ensure a unique match, or set replace_all to true"), nil
			}
		}

		newContent = oldContent[:index] + newString + oldContent[index+len(oldString):]
		replacementCount = 1
	}

	if oldContent == newContent {
		return fantasy.NewTextErrorResponse("new content is the same as old content. No changes made."), nil
	}
	sessionID := GetSessionFromContext(edit.ctx)

	if sessionID == "" {
		return fantasy.ToolResponse{}, fmt.Errorf("session ID is required for creating a new file")
	}
	_, additions, removals := diff.GenerateDiff(
		oldContent,
		newContent,
		strings.TrimPrefix(filePath, edit.workingDir),
	)

	p := edit.permissions.Request(
		permission.CreatePermissionRequest{
			SessionID:   sessionID,
			Path:        fsext.PathOrPrefix(filePath, edit.workingDir),
			ToolCallID:  call.ID,
			ToolName:    EditToolName,
			Action:      "write",
			Description: fmt.Sprintf("Replace content in file %s", filePath),
			Params: EditPermissionsParams{
				FilePath:   filePath,
				OldContent: oldContent,
				NewContent: newContent,
			},
		},
	)
	if !p {
		return fantasy.ToolResponse{}, permission.ErrorPermissionDenied
	}

	if isCrlf {
		newContent, _ = fsext.ToWindowsLineEndings(newContent)
	}

	err = os.WriteFile(filePath, []byte(newContent), 0o644)
	if err != nil {
		return fantasy.ToolResponse{}, fmt.Errorf("failed to write file: %w", err)
	}

	// Check if file exists in history
	file, err := edit.files.GetByPathAndSession(edit.ctx, filePath, sessionID)
	if err != nil {
		_, err = edit.files.Create(edit.ctx, sessionID, filePath, oldContent)
		if err != nil {
			// Log error but don't fail the operation
			return fantasy.ToolResponse{}, fmt.Errorf("error creating file history: %w", err)
		}
	}
	if file.Content != oldContent {
		// User Manually changed the content store an intermediate version
		_, err = edit.files.CreateVersion(edit.ctx, sessionID, filePath, oldContent)
		if err != nil {
			slog.Debug("Error creating file history version", "error", err)
		}
	}
	// Store the new version
	_, err = edit.files.CreateVersion(edit.ctx, sessionID, filePath, newContent)
	if err != nil {
		slog.Error("Error creating file history version", "error", err)
	}

	recordFileWrite(filePath)
	recordFileRead(filePath)
	LogEditSuccess(filePath, len(oldString), len(newString), replacementCount, attemptCount)

	return fantasy.WithResponseMetadata(
		fantasy.NewTextResponse("Content replaced in file: "+filePath),
		EditResponseMetadata{
			OldContent: oldContent,
			NewContent: newContent,
			Additions:  additions,
			Removals:   removals,
		}), nil
}

// attemptSelfHealingRetry uses the self-healing strategy to improve the old_string
// by extracting better context from the file when the initial match fails
func attemptSelfHealingRetry(ctx context.Context, filePath string, oldString string, newString string) (EditParams, error) {
	strategy := NewEditRetryStrategy(ctx)
	retryParams, err := strategy.RetryWithContext(filePath, oldString, newString, "old_string not found")
	if err != nil {
		return EditParams{}, err
	}
	return retryParams, nil
}

// autoViewFileBeforeEdit automatically views a file before editing to ensure we have the latest context
func autoViewFileBeforeEdit(ctx context.Context, filePath string) error {
	_, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file for auto-view: %w", err)
	}

	// Update the last read time to reflect this automatic view
	recordFileRead(filePath)

	// Log that we performed an auto-view
	slog.Debug("Auto-viewed file before edit", "file", filePath)

	return nil
}

// normalizeWhitespace converts mixed whitespace to consistent tabs for matching
// This helps match text copied from View output (which may have spaces from display padding)
