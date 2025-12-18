package agent

import (
	"time"

	"charm.land/fantasy"
)

// Local event tracking stubs - no telemetry sent
func (a *sessionAgent) eventPromptSent(sessionID string) {
	// No telemetry - removed
}

func (a *sessionAgent) eventPromptResponded(sessionID string, duration time.Duration) {
	// No telemetry - removed
}

func (a *sessionAgent) eventTokensUsed(sessionID string, model Model, usage fantasy.Usage, cost float64) {
	// No telemetry - removed
}

func (a *sessionAgent) eventCommon(sessionID string, model Model) []any {
	m := model.ModelCfg

	return []any{
		"session id", sessionID,
		"provider", m.Provider,
		"model", m.Model,
		"reasoning effort", m.ReasoningEffort,
		"thinking mode", m.Think,
		"yolo mode", a.isYolo,
	}
}
