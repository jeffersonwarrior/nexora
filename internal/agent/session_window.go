package agent

import (
	"sync"
	
	"github.com/nexora/cli/internal/message"
)

type SessionWindow struct {
	ID         string                 `json:"id"`
	Title      string                 `json:"title"`
	Agent      SessionAgent           `json:"-"`
	History    []message.Message      `json:"history,omitempty"`
	State      *ConversationState     `json:"state,omitempty"`
	ForkParent string                 `json:"fork_parent,omitempty"`
	IsActive   bool                   `json:"is_active"`
	
	mu sync.RWMutex
}

func (w *SessionWindow) Messages() []message.Message {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.History
}

func (w *SessionWindow) AddMessage(msg message.Message) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.History = append(w.History, msg)
}

func NewSessionWindow(id, title string, agent SessionAgent) *SessionWindow {
	return &SessionWindow{
		ID:      id,
		Title:   title,
		Agent:   agent,
		History: []message.Message{},
	}
}
