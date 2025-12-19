package agent

import (
	"log/slog"
	"strings"
	"time"

	"github.com/nexora/nexora/internal/message"
)

// ConversationState represents the current state of a conversation
type ConversationState int

const (
	StateActive ConversationState = iota
	StateWaitingForUser
	StateAgentProcessing
	StateCompleted
	StateClosed
)

func (cs ConversationState) String() string {
	switch cs {
	case StateActive:
		return "active"
	case StateWaitingForUser:
		return "waiting_for_user"
	case StateAgentProcessing:
		return "agent_processing"
	case StateCompleted:
		return "completed"
	case StateClosed:
		return "closed"
	default:
		return "unknown"
	}
}

// CanTransition checks if a state transition is valid
func (cs ConversationState) CanTransition(newState ConversationState) bool {
	transitions := map[ConversationState][]ConversationState{
		StateActive:         {StateWaitingForUser, StateAgentProcessing, StateClosed},
		StateWaitingForUser: {StateAgentProcessing, StateClosed},
		StateAgentProcessing:{StateWaitingForUser, StateCompleted},
		StateCompleted:      {StateClosed},
		StateClosed:         {},
	}
	
	for _, valid := range transitions[cs] {
		if valid == newState {
			return true
		}
	}
	return false
}

// ThreadedConversation represents a properly managed conversation thread
type ThreadedConversation struct {
	SessionID      string
	State          ConversationState
	CreatedAt      time.Time
	LastActivity   time.Time
	MessageCount   int
	RequiresInput  bool
	CompletionScore float64 // 0.0 to 1.0
	Token          string   // For continuation requests
}

// ConversationManager manages conversation threads with proper state
type ConversationManager struct {
	sessions map[string]*ThreadedConversation
}

func NewConversationManager() *ConversationManager {
	return &ConversationManager{
		sessions: make(map[string]*ThreadedConversation),
	}
}

// GetOrCreateSession retrieves or creates a conversation session
func (cm *ConversationManager) GetOrCreateSession(sessionID string) *ThreadedConversation {
	if session, exists := cm.sessions[sessionID]; exists {
		return session
	}
	
	session := &ThreadedConversation{
		SessionID:      sessionID,
		State:          StateActive,
		CreatedAt:      time.Now(),
		LastActivity:   time.Now(),
		MessageCount:   0,
		RequiresInput:  true,
		CompletionScore: 0.0,
		Token:          generateToken(),
	}
	
	cm.sessions[sessionID] = session
	return session
}

// TransitionState attempts to transition to a new state
func (cm *ConversationManager) TransitionState(sessionID string, newState ConversationState) bool {
	session := cm.GetOrCreateSession(sessionID)
	
	if !session.State.CanTransition(newState) {
		slog.Warn("invalid state transition",
			"session_id", sessionID,
			"from", session.State,
			"to", newState)
		return false
	}
	
	oldState := session.State
	session.State = newState
	session.LastActivity = time.Now()
	
	slog.Debug("conversation state transitioned",
		"session_id", sessionID,
		"from", oldState,
		"to", newState)
	
	return true
}

// RecordMessage records a message and updates conversation state
func (cm *ConversationManager) RecordMessage(sessionID string, msg message.Message) {
	session := cm.GetOrCreateSession(sessionID)
	session.MessageCount++
	session.LastActivity = time.Now()
	
	// Analyze message for completion indicators
	if msg.Role == message.Assistant {
		session.CompletionScore = analyzeCompletion(msg.Content().Text)
		session.RequiresInput = session.CompletionScore < 0.8
		
		if session.CompletionScore > 0.8 && session.State == StateAgentProcessing {
			cm.TransitionState(sessionID, StateCompleted)
			// Schedule auto-close in 30 minutes
			go cm.scheduleAutoClose(sessionID)
		} else if session.State == StateAgentProcessing {
			cm.TransitionState(sessionID, StateWaitingForUser)
		}
	} else if msg.Role == message.User {
		session.RequiresInput = false
		if session.State == StateWaitingForUser {
			cm.TransitionState(sessionID, StateAgentProcessing)
		}
	}
}

// ShouldContinue determines if conversation should continue based on state
func (cm *ConversationManager) ShouldContinue(sessionID string) bool {
	session := cm.GetOrCreateSession(sessionID)
	
	// Only continue if we're in processing state
	return session.State == StateAgentProcessing
}

// IsConversationCompleted checks if conversation is marked as completed
func (cm *ConversationManager) IsConversationCompleted(sessionID string) bool {
	session := cm.GetOrCreateSession(sessionID)
	return session.State == StateCompleted
}

// GetCompletionScore returns the completion confidence score
func (cm *ConversationManager) GetCompletionScore(sessionID string) float64 {
	session := cm.GetOrCreateSession(sessionID)
	return session.CompletionScore
}

// scheduleAutoClose automatically closes completed conversations after inactivity
func (cm *ConversationManager) scheduleAutoClose(sessionID string) {
	<-time.After(30 * time.Minute)
	
	if session, exists := cm.sessions[sessionID]; exists && 
		session.State == StateCompleted {
		cm.TransitionState(sessionID, StateClosed)
		slog.Info("conversation auto-closed", "session_id", sessionID)
	}
}

// analyzeCompletion performs semantic analysis to determine completion confidence
func analyzeCompletion(content string) float64 {
	content = strings.ToLower(content)
	
	// Strong completion indicators (0.9 confidence)
	completionPhrases := []string{
		"is there anything else",
		"what else can i help you with",
		"let me know if you need anything else",
		"task completed",
		"finished helping",
		"all set",
		"done",
		"complete",
	}
	
	for _, phrase := range completionPhrases {
		if strings.Contains(content, phrase) {
			return 0.9
		}
	}
	
	// Moderate completion indicators (0.7 confidence)
	moderatePhrases := []string{
		"should now",
		"has been",
		"you can now",
		"successfully",
		"as requested",
	}
	
	moderateScore := 0.0
	for _, phrase := range moderatePhrases {
		if strings.Contains(content, phrase) {
			moderateScore += 0.2
		}
	}
return minFloat(0.7, moderateScore)
	// Work continuators reduce completion score
	workContinuers := []string{
		"now let me",
		"next, i'll",
		"i'll now",
		"let me also",
		"let me create",
		"let me implement",
		"let me update",
		"next step",
	}
	
	for _, phrase := range workContinuers {
		if strings.Contains(content, phrase) {
			return 0.1 // Very low completion score
		}
	}
	
	// Default moderate confidence
	return 0.5
}

func generateToken() string {
	return strings.ReplaceAll(time.Now().Format("20060102150405.000"), ".", "")
}

// minFloat returns the minimum of two float64 values
func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

