package shell

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

// TmuxSession represents a TMUX session managed by Nexora
type TmuxSession struct {
	ID          string // Nexora's internal ID (e.g., "nexora-abc123-def456")
	SessionName string // TMUX session name
	PaneID      string // TMUX pane identifier
	WorkingDir  string
	Command     string // Last command executed
	Description string
	StartedAt   time.Time
	Output      bytes.Buffer
	mu          sync.RWMutex
	done        chan struct{}
}

// TmuxManager manages TMUX sessions
type TmuxManager struct {
	sessions        map[string]*TmuxSession
	defaultSessions map[string]string // Maps conversation sessionID -> default TMUX sessionID
	mu              sync.RWMutex
}

var (
	tmuxManager     *TmuxManager
	tmuxManagerOnce sync.Once
)

// GetTmuxManager returns the singleton TMUX manager
func GetTmuxManager() *TmuxManager {
	tmuxManagerOnce.Do(func() {
		tmuxManager = &TmuxManager{
			sessions:        make(map[string]*TmuxSession),
			defaultSessions: make(map[string]string),
		}
	})
	return tmuxManager
}

// IsAvailable checks if TMUX is installed and accessible
func IsTmuxAvailable() bool {
	cmd := exec.Command("tmux", "-V")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	// Extract version string
	versionOutput := strings.TrimSpace(string(output))
	return strings.HasPrefix(versionOutput, "tmux")
}

// NewTmuxSession creates a new TMUX session with a pane
func (m *TmuxManager) NewTmuxSession(sessionID, workingDir, command, description string) (*TmuxSession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if session already exists
	if _, exists := m.sessions[sessionID]; exists {
		return nil, fmt.Errorf("session already exists: %s", sessionID)
	}

	// Create TMUX session
	sessionName := "nexora-" + sessionID

	// Create new TMUX session with a window
	// Format: tmux new-session -d -s sessionName -x 200 -y 50 -c workingDir
	createCmd := exec.Command("tmux", "new-session", "-d", "-s", sessionName, "-x", "200", "-y", "50", "-c", workingDir)
	if err := createCmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to create TMUX session: %w", err)
	}

	// Get the pane ID
	paneCmd := exec.Command("tmux", "list-panes", "-t", sessionName, "-F", "#{pane_id}")
	paneOutput, err := paneCmd.CombinedOutput()
	if err != nil {
		// Cleanup on failure
		exec.Command("tmux", "kill-session", "-t", sessionName).Run()
		return nil, fmt.Errorf("failed to get pane ID: %w", err)
	}
	paneID := strings.TrimSpace(string(paneOutput))
	if paneID == "" {
		paneID = "%0" // Default pane
	}

	session := &TmuxSession{
		ID:          sessionID,
		SessionName: sessionName,
		PaneID:      paneID,
		WorkingDir:  workingDir,
		Command:     command,
		Description: description,
		StartedAt:   time.Now(),
		done:        make(chan struct{}),
	}

	m.sessions[sessionID] = session

	// If a command is provided, execute it using internal method (avoids deadlock)
	if command != "" {
		if err := m.sendCommandToSession(session, command); err != nil {
			m.RemoveSession(sessionID)
			return nil, err
		}
	}

	return session, nil
}

// SendCommand sends a command to a TMUX session
func (m *TmuxManager) SendCommand(sessionID, command string) error {
	m.mu.RLock()
	session, exists := m.sessions[sessionID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	return m.sendCommandToSession(session, command)
}

// sendCommandToSession sends a command directly to a session (internal method, no lock needed)
func (m *TmuxManager) sendCommandToSession(session *TmuxSession, command string) error {
	// Process special key sequences (e.g., <Esc>, <Enter>, <Tab>)
	processed := processSpecialKeys(command)

	// Send command using literal mode (-l) to avoid over-escaping
	// This allows shell metacharacters like ; | & to work correctly
	cmd := exec.Command("tmux", "send-keys", "-l", "-t", session.SessionName, processed)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to send command: %w", err)
	}

	// Send Enter key separately (not literal)
	enterCmd := exec.Command("tmux", "send-keys", "-t", session.SessionName, "Enter")
	if err := enterCmd.Run(); err != nil {
		return fmt.Errorf("failed to send Enter: %w", err)
	}

	session.mu.Lock()
	session.Command = command
	session.mu.Unlock()

	return nil
}

// CaptureOutput captures the current pane output
func (m *TmuxManager) CaptureOutput(sessionID string) (string, error) {
	m.mu.RLock()
	session, exists := m.sessions[sessionID]
	m.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("session not found: %s", sessionID)
	}

	// Capture pane content
	cmd := exec.Command("tmux", "capture-pane", "-t", session.SessionName, "-p")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to capture output: %w", err)
	}

	return string(output), nil
}

// KillSession terminates a TMUX session
func (m *TmuxManager) KillSession(sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	// Kill TMUX session
	cmd := exec.Command("tmux", "kill-session", "-t", session.SessionName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to kill TMUX session: %w", err)
	}

	delete(m.sessions, sessionID)
	return nil
}

// KillAll kills all TMUX sessions for a given prefix
func (m *TmuxManager) KillAll(prefix string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errors []string
	for id, session := range m.sessions {
		if strings.HasPrefix(id, prefix) {
			cmd := exec.Command("tmux", "kill-session", "-t", session.SessionName)
			if err := cmd.Run(); err != nil {
				errors = append(errors, fmt.Sprintf("failed to kill %s: %v", id, err))
			} else {
				delete(m.sessions, id)
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors killing sessions: %s", strings.Join(errors, "; "))
	}

	return nil
}

// RemoveSession removes a session from tracking without killing TMUX
func (m *TmuxManager) RemoveSession(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, sessionID)
}

// GetSession returns a session by ID
func (m *TmuxManager) GetSession(sessionID string) (*TmuxSession, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	session, exists := m.sessions[sessionID]
	return session, exists
}

// ListSessions returns all session IDs
func (m *TmuxManager) ListSessions() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := make([]string, 0, len(m.sessions))
	for id := range m.sessions {
		ids = append(ids, id)
	}
	return ids
}

// SessionCount returns the number of active sessions
func (m *TmuxManager) SessionCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.sessions)
}

// IsSessionRunning checks if a session is still running
func (m *TmuxManager) IsSessionRunning(sessionID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return false
	}

	// Check if TMUX session still exists
	cmd := exec.Command("tmux", "has-session", "-t", session.SessionName)
	err := cmd.Run()
	return err == nil
}

// GetPaneDimensions returns the dimensions of a TMUX pane
func (m *TmuxManager) GetPaneDimensions(sessionID string) (width, height int, err error) {
	m.mu.RLock()
	session, exists := m.sessions[sessionID]
	m.mu.RUnlock()

	if !exists {
		return 0, 0, fmt.Errorf("session not found: %s", sessionID)
	}

	// Format: width x height (e.g., "200 x 50")
	cmd := exec.Command("tmux", "display-message", "-t", session.SessionName, "-p", "#{pane_width} #{pane_height}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get pane dimensions: %w", err)
	}

	parts := strings.Split(string(output), " ")
	if len(parts) >= 2 {
		width, _ = strconv.Atoi(strings.TrimSpace(parts[0]))
		height, _ = strconv.Atoi(strings.TrimSpace(parts[1]))
	}

	return width, height, nil
}

// ResizePane resizes a TMUX pane
func (m *TmuxManager) ResizePane(sessionID string, width, height int) error {
	m.mu.RLock()
	session, exists := m.sessions[sessionID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	cmd := exec.Command("tmux", "resize-pane", "-t", session.SessionName, "-x", strconv.Itoa(width), "-y", strconv.Itoa(height))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to resize pane: %w", err)
	}

	return nil
}

// processSpecialKeys converts special key sequences like <Esc>, <Enter>, <Tab> to tmux format
// When using tmux send-keys -l (literal mode), we DON'T escape shell metacharacters
// They need to be interpreted by bash running inside the TMUX session
func processSpecialKeys(input string) string {
	// For literal mode (-l flag), we pass the command as-is
	// Shell metacharacters (;, |, &, etc.) work correctly
	// Special key sequences like <Esc>, <Tab> should be sent separately, not as literal text

	// For now, just return input unchanged - literal mode handles everything
	// TODO: In future, process <Esc>, <Enter>, <Tab> sequences and send them separately
	return input
}

// GetOrCreateDefaultSession returns the default session for a conversation, creating if needed
func (m *TmuxManager) GetOrCreateDefaultSession(conversationID, workingDir string) (*TmuxSession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if default session exists for this conversation
	if tmuxSessionID, exists := m.defaultSessions[conversationID]; exists {
		if session, ok := m.sessions[tmuxSessionID]; ok {
			// Verify TMUX session still exists
			cmd := exec.Command("tmux", "has-session", "-t", session.SessionName)
			if cmd.Run() == nil {
				return session, nil
			}
			// Session died, clean up
			delete(m.sessions, tmuxSessionID)
			delete(m.defaultSessions, conversationID)
		}
	}

	// Create new default session with readable ID
	// Format: nexora-main-001, nexora-main-002, etc.
	tmuxSessionID := m.generateReadableSessionID()

	// Create TMUX session
	sessionName := "nexora-" + tmuxSessionID
	createCmd := exec.Command("tmux", "new-session", "-d", "-s", sessionName, "-x", "200", "-y", "50", "-c", workingDir)
	if err := createCmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to create default TMUX session: %w", err)
	}

	// Get pane ID
	paneCmd := exec.Command("tmux", "list-panes", "-t", sessionName, "-F", "#{pane_id}")
	paneOutput, err := paneCmd.CombinedOutput()
	if err != nil {
		exec.Command("tmux", "kill-session", "-t", sessionName).Run()
		return nil, fmt.Errorf("failed to get pane ID: %w", err)
	}
	paneID := strings.TrimSpace(string(paneOutput))
	if paneID == "" {
		paneID = "%0"
	}

	session := &TmuxSession{
		ID:          tmuxSessionID,
		SessionName: sessionName,
		PaneID:      paneID,
		WorkingDir:  workingDir,
		Description: "Default persistent shell",
		StartedAt:   time.Now(),
		done:        make(chan struct{}),
	}

	m.sessions[tmuxSessionID] = session
	m.defaultSessions[conversationID] = tmuxSessionID

	return session, nil
}

// generateReadableSessionID creates a human-readable session ID
func (m *TmuxManager) generateReadableSessionID() string {
	words := []string{
		"alpha", "bravo", "charlie", "delta", "echo", "foxtrot", "golf", "hotel",
		"india", "juliet", "kilo", "lima", "mike", "november", "oscar", "papa",
		"quebec", "romeo", "sierra", "tango", "uniform", "victor", "whiskey", "xray",
		"yankee", "zulu", "atom", "byte", "code", "data", "edge", "flux", "gate",
		"hash", "iron", "jade", "kite", "lens", "mesh", "node", "opus", "peak",
		"quad", "ruby", "sync", "tide", "unit", "vibe", "wave", "zero",
	}

	// Count existing sessions to get next number
	sessionCount := len(m.sessions)
	wordIndex := sessionCount % len(words)
	sequenceNum := (sessionCount / len(words)) + 1

	return fmt.Sprintf("%s-%03d", words[wordIndex], sequenceNum)
}

// SessionInfo returns information about a session
func (s *TmuxSession) SessionInfo() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"id":           s.ID,
		"session_name": s.SessionName,
		"pane_id":      s.PaneID,
		"working_dir":  s.WorkingDir,
		"command":      s.Command,
		"description":  s.Description,
		"started_at":   s.StartedAt,
		"running_time": time.Since(s.StartedAt).String(),
	}
}
