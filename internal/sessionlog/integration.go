package sessionlog

// This file contains integration helpers for Nexora v0.25
// The session logger can be integrated by:
// 1. Creating a *Manager during coordinator initialization
// 2. Calling LogEditOperation() in tools/edit.go when failures occur
// 3. Calling LogViewOperation() in tools/view.go
// 4. Calling StartSession/EndSession in coordinator.Run()

import (
	"context"
	"log/slog"
)

// global instance can be set by the application
var globalManager *Manager

// SetGlobalManager sets the global session log manager
func SetGlobalManager(m *Manager) {
	globalManager = m
}

// GetGlobalManager returns the global session log manager
func GetGlobalManager() *Manager {
	return globalManager
}

// LogEditFailureSafe safely logs an edit failure to the global manager
func LogEditFailureSafe(ctx context.Context, sessionID, instanceID string, edit EditOperationLog) {
	if globalManager == nil {
		return
	}
	globalManager.LogEditOperation(ctx, edit)
}

// LogEditSuccessSafe safely logs a successful edit to the global manager
func LogEditSuccessSafe(ctx context.Context, sessionID, instanceID, filePath string, oldLen, newLen, replacementCount, attemptCount int, durationMS float64) {
	if globalManager == nil {
		return
	}

	log := EditOperationLog{
		SessionID:        sessionID,
		InstanceID:       instanceID,
		FilePath:         filePath,
		Status:           "success",
		OldStringLength:  oldLen,
		NewStringLength:  newLen,
		ReplacementCount: replacementCount,
		AttemptCount:     attemptCount,
		DurationMS:       durationMS,
	}
	globalManager.LogEditOperation(ctx, log)
}

// LogViewFailureSafe safely logs a view failure
func LogViewFailureSafe(ctx context.Context, sessionID, instanceID string, view ViewOperationLog) {
	if globalManager == nil {
		return
	}
	globalManager.LogViewOperation(ctx, view)
}

// LogViewSuccessSafe safely logs a successful view
func LogViewSuccessSafe(ctx context.Context, sessionID, instanceID string, view ViewOperationLog) {
	if globalManager == nil {
		return
	}
	globalManager.LogViewOperation(ctx, view)
}

// InitializeFromEnv initializes the global session logger from environment variables
// Set these env vars to enable session logging:
// NEXORA_SESSION_LOG_ENABLED=true
// NEXORA_SESSION_LOG_PG_CONN=user=postgres dbname=nexora_sessions host=localhost sslmode=disable
// NEXORA_SESSION_LOG_INSTANCE_ID=nexora-1 (defaults to hostname)
func InitializeFromEnv() error {
	// TODO: Read from environment and initialize
	slog.Info("Session logging not yet configured from environment")
	return nil
}
