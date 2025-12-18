// Package state provides state machine management for Nexora agents.
//
// The state machine tracks agent execution state, progress, and detects
// stuck loops while allowing productive long-running tasks.
package state

import (
	"fmt"
	"time"
)

// AgentState represents the current state of the agent execution.
type AgentState int

const (
	// StateIdle indicates the agent is waiting for work
	StateIdle AgentState = iota

	// StateProcessingPrompt indicates the agent is processing a user prompt
	StateProcessingPrompt

	// StateStreamingResponse indicates the agent is streaming LLM response
	StateStreamingResponse

	// StateExecutingTool indicates the agent is executing a tool call
	StateExecutingTool

	// StateAwaitingPermission indicates the agent is waiting for user approval
	StateAwaitingPermission

	// StateErrorRecovery indicates the agent is recovering from an error
	StateErrorRecovery

	// StatePhaseTransition indicates moving between refactor phases
	StatePhaseTransition

	// StateProgressCheck indicates validating forward progress
	StateProgressCheck

	// StateResourcePaused indicates agent paused due to resource limits
	StateResourcePaused

	// StateHalted indicates the agent has stopped (loop detected or user request)
	StateHalted
)

// String returns a human-readable state name.
func (s AgentState) String() string {
	switch s {
	case StateIdle:
		return "Idle"
	case StateProcessingPrompt:
		return "ProcessingPrompt"
	case StateStreamingResponse:
		return "StreamingResponse"
	case StateExecutingTool:
		return "ExecutingTool"
	case StateAwaitingPermission:
		return "AwaitingPermission"
	case StateErrorRecovery:
		return "ErrorRecovery"
	case StatePhaseTransition:
		return "PhaseTransition"
	case StateProgressCheck:
		return "ProgressCheck"
	case StateResourcePaused:
		return "ResourcePaused"
	case StateHalted:
		return "Halted"
	default:
		return fmt.Sprintf("Unknown(%d)", s)
	}
}

// IsTerminal returns true if this state is terminal (execution should stop).
func (s AgentState) IsTerminal() bool {
	return s == StateHalted
}

// CanTransitionTo returns true if transitioning from this state to target is valid.
func (s AgentState) CanTransitionTo(target AgentState) bool {
	transitions, ok := validTransitions[s]
	if !ok {
		return false
	}

	for _, valid := range transitions {
		if valid == target {
			return true
		}
	}
	return false
}

// validTransitions defines the allowed state transitions.
var validTransitions = map[AgentState][]AgentState{
	StateIdle: {
		StateProcessingPrompt,
		StatePhaseTransition,
		StateHalted,
	},
	StateProcessingPrompt: {
		StateStreamingResponse,
		StateErrorRecovery,
		StateHalted,
		StateProcessingPrompt, // Allow re-prompting for better recovery
	},
	StateStreamingResponse: {
		StateExecutingTool,
		StateAwaitingPermission,
		StateProgressCheck,
		StateIdle,
		StateErrorRecovery,
		StateHalted,
	},
	StateExecutingTool: {
		StateStreamingResponse,
		StateProgressCheck,
		StateErrorRecovery,
		StateHalted,
	},
	StateAwaitingPermission: {
		StateStreamingResponse,
		StateExecutingTool,
		StateIdle,
		StateHalted,
	},
	StateErrorRecovery: {
		StateStreamingResponse,
		StateExecutingTool,
		StateProgressCheck,
		StateIdle,
		StateHalted,
	},
	StatePhaseTransition: {
		StateProcessingPrompt,
		StateProgressCheck,
		StatePhaseTransition,
		StateIdle,
		StateHalted,
	},
	StateProgressCheck: {
		StateStreamingResponse,
		StateExecutingTool,
		StatePhaseTransition,
		StateErrorRecovery,
		StateIdle,
		StateHalted,
	},
	StateHalted: {}, // Terminal state
}

// TransitionError represents an invalid state transition.
type TransitionError struct {
	From    AgentState
	To      AgentState
	Reason  string
	Context map[string]interface{}
}

func (e *TransitionError) Error() string {
	return fmt.Sprintf("invalid transition from %s to %s: %s", e.From, e.To, e.Reason)
}

// StateMetrics tracks metrics for a particular state.
type StateMetrics struct {
	State        AgentState
	EnterTime    time.Time
	ExitTime     time.Time
	Duration     time.Duration
	TransitionTo AgentState
	ErrorCount   int
	LastError    error
}
