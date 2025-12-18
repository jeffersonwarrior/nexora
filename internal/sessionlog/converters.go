package sessionlog

import (
	"github.com/nexora/nexora/internal/agent/tools"
)

// ConvertEditDiagnosticsToLog converts edit diagnostics to a log entry
func ConvertEditDiagnosticsToLog(sessionID, instanceID string, diag tools.EditDiagnosticsInfo, analysis tools.WhitespaceAnalysis) EditOperationLog {
	return EditOperationLog{
		SessionID:       sessionID,
		InstanceID:      instanceID,
		FilePath:        diag.FilePath,
		Status:          "failure",
		FailureReason:   diag.FailureReason,
		OldStringLength: len(diag.OldString),
		NewStringLength: len(diag.NewString),
		DurationMS:      0, // Will be set by caller
		HasTabs:         analysis.ContainsTab,
		HasMixedIndent:  analysis.HasMixedIndent,
		FileLineEndings: analysis.LineEndings,
		Metadata: map[string]interface{}{
			"old_string_preview": truncateStringPreview(diag.OldString, 80),
			"file_size":          len(diag.FileContent),
			"old_string_lines":   countLines(diag.OldString),
			"context":            diag.Context,
		},
	}
}

// ConvertEditSuccessToLog converts a successful edit to a log entry
func ConvertEditSuccessToLog(sessionID, instanceID, filePath string, oldLen, newLen, replacementCount, attemptCount int, durationMS float64) EditOperationLog {
	return EditOperationLog{
		SessionID:        sessionID,
		InstanceID:       instanceID,
		FilePath:         filePath,
		Status:           "success",
		OldStringLength:  oldLen,
		NewStringLength:  newLen,
		ReplacementCount: replacementCount,
		AttemptCount:     attemptCount,
		DurationMS:       durationMS,
		Metadata: map[string]interface{}{
			"replacement_count": replacementCount,
			"attempt_count":     attemptCount,
		},
	}
}

// ConvertViewOperationToLog converts view diagnostics to a log entry
func ConvertViewOperationToLog(sessionID, instanceID string, diag tools.ViewDiagnosticsInfo, durationMS float64) ViewOperationLog {
	return ViewOperationLog{
		SessionID:     sessionID,
		InstanceID:    instanceID,
		FilePath:      diag.FilePath,
		OffsetLine:    diag.Offset,
		LimitLines:    diag.Limit,
		FileSizeBytes: diag.FileSize,
		TotalLines:    diag.LineCount,
		Status:        "success",
		DurationMS:    durationMS,
		Metadata:      diag.Context,
	}
}

// ConvertViewErrorToLog converts a view error to a log entry
func ConvertViewErrorToLog(sessionID, instanceID string, diag tools.ViewDiagnosticsInfo, errorReason string) ViewOperationLog {
	return ViewOperationLog{
		SessionID:   sessionID,
		InstanceID:  instanceID,
		FilePath:    diag.FilePath,
		OffsetLine:  diag.Offset,
		LimitLines:  diag.Limit,
		Status:      "error",
		ErrorReason: errorReason,
	}
}

// Helper function
func countLines(s string) int {
	if len(s) == 0 {
		return 0
	}
	count := 1
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			count++
		}
	}
	return count
}

// TruncatePreview provides a preview of a string for logging
// (This is a wrapper - the actual function is in tools package)
var TruncatePreview = func(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s[:]
	}
	return s[:maxLen] + "..."
}

// truncateStringPreview truncates a string for logging
func truncateStringPreview(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
