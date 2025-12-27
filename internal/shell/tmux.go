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

// SessionStatus represents the current state of a pooled session
type SessionStatus int

const (
	SessionAvailable SessionStatus = iota // Ready for reuse
	SessionBusy                           // Currently in use
	SessionDraining                       // Marked for cleanup
)

func (s SessionStatus) String() string {
	switch s {
	case SessionAvailable:
		return "available"
	case SessionBusy:
		return "busy"
	case SessionDraining:
		return "draining"
	default:
		return "unknown"
	}
}

// PoolConfig configures session pool behavior
type PoolConfig struct {
	MaxSize     int           // Maximum pool size (default: 10)
	IdleTimeout time.Duration // Cleanup idle sessions after this (default: 5min)
}

// DefaultPoolConfig returns sensible defaults
func DefaultPoolConfig() PoolConfig {
	return PoolConfig{
		MaxSize:     10,
		IdleTimeout: 5 * time.Minute,
	}
}

// PoolMetrics tracks session pool statistics
type PoolMetrics struct {
	Created  int64 // Total sessions created
	Reused   int64 // Sessions reused from pool
	Released int64 // Sessions returned to pool
	Cleaned  int64 // Sessions cleaned up due to idle
}

// PooledSession extends TmuxSession with pool metadata
type PooledSession struct {
	*TmuxSession
	Status     SessionStatus
	LastUsedAt time.Time
}

// TmuxSession represents a TMUX session managed by Nexora
type TmuxSession struct {
	ID          string    // Nexora's internal ID (e.g., "nexora-abc123-def456")
	SessionName string    // TMUX session name
	PaneID      string    // TMUX pane identifier
	WorkingDir  string
	Command     string    // Last command executed
	Description string
	StartedAt   time.Time
	Output      bytes.Buffer
	mu          sync.RWMutex
	done        chan struct{}
}

// TmuxManager manages TMUX sessions with pooling support
type TmuxManager struct {
	sessions        map[string]*TmuxSession
	defaultSessions map[string]string // Maps conversation sessionID -> default TMUX sessionID
	mu              sync.RWMutex

	// Session pool
	pool       map[string]*PooledSession // Pooled sessions by ID
	poolConfig PoolConfig
	poolMu     sync.RWMutex

	// Metrics
	metrics   PoolMetrics
	metricsMu sync.RWMutex

	// Cleanup
	cleanupStop chan struct{}
	cleanupOnce sync.Once
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
			pool:            make(map[string]*PooledSession),
			poolConfig:      DefaultPoolConfig(),
			cleanupStop:     make(chan struct{}),
		}
		// Start background cleanup
		tmuxManager.startCleanup()
	})
	return tmuxManager
}

// GetTmuxManagerWithConfig returns manager with custom config (for testing)
func GetTmuxManagerWithConfig(config PoolConfig) *TmuxManager {
	m := GetTmuxManager()
	m.poolMu.Lock()
	m.poolConfig = config
	m.poolMu.Unlock()
	return m
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
// Now uses session pooling for efficient resource management
func (m *TmuxManager) GetOrCreateDefaultSession(conversationID, workingDir string) (*TmuxSession, error) {
	m.mu.Lock()

	// Check if default session exists for this conversation
	if tmuxSessionID, exists := m.defaultSessions[conversationID]; exists {
		if session, ok := m.sessions[tmuxSessionID]; ok {
			// Verify TMUX session still exists
			cmd := exec.Command("tmux", "has-session", "-t", session.SessionName)
			if cmd.Run() == nil {
				m.mu.Unlock()
				return session, nil
			}
			// Session died, clean up
			delete(m.sessions, tmuxSessionID)
			delete(m.defaultSessions, conversationID)
		}
	}
	m.mu.Unlock()

	// Use pooled session instead of creating new
	session, err := m.GetPooledSession(workingDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get pooled session: %w", err)
	}

	// Map this conversation to the pooled session
	m.mu.Lock()
	m.defaultSessions[conversationID] = session.ID
	m.mu.Unlock()

	return session, nil
}

// ReleaseDefaultSession releases a conversation's session back to pool
func (m *TmuxManager) ReleaseDefaultSession(conversationID string) {
	m.mu.Lock()
	tmuxSessionID, exists := m.defaultSessions[conversationID]
	if exists {
		delete(m.defaultSessions, conversationID)
	}
	m.mu.Unlock()

	if exists {
		m.ReleaseSession(tmuxSessionID)
	}
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

// ============================================================================
// Session Pool Methods
// ============================================================================

// startCleanup starts background goroutine for idle session cleanup
func (m *TmuxManager) startCleanup() {
	m.cleanupOnce.Do(func() {
		go func() {
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					m.cleanupIdleSessions()
				case <-m.cleanupStop:
					return
				}
			}
		}()
	})
}

// StopCleanup stops the background cleanup goroutine
func (m *TmuxManager) StopCleanup() {
	select {
	case <-m.cleanupStop:
		// Already closed
	default:
		close(m.cleanupStop)
	}
}

// cleanupIdleSessions removes sessions that have been idle too long
func (m *TmuxManager) cleanupIdleSessions() {
	m.poolMu.Lock()
	defer m.poolMu.Unlock()

	now := time.Now()
	var toRemove []string

	for id, ps := range m.pool {
		if ps.Status == SessionAvailable && now.Sub(ps.LastUsedAt) > m.poolConfig.IdleTimeout {
			toRemove = append(toRemove, id)
		}
	}

	for _, id := range toRemove {
		ps := m.pool[id]
		// Kill the actual tmux session
		exec.Command("tmux", "kill-session", "-t", ps.SessionName).Run()
		delete(m.pool, id)

		m.metricsMu.Lock()
		m.metrics.Cleaned++
		m.metricsMu.Unlock()
	}
}

// GetPooledSession returns an available session from pool or creates new one
// This is the main entry point for session pooling
func (m *TmuxManager) GetPooledSession(workingDir string) (*TmuxSession, error) {
	// First try to find an available session with matching working dir
	m.poolMu.Lock()

	for _, ps := range m.pool {
		if ps.Status == SessionAvailable && ps.WorkingDir == workingDir {
			// Verify session still exists
			cmd := exec.Command("tmux", "has-session", "-t", ps.SessionName)
			if cmd.Run() == nil {
				ps.Status = SessionBusy
				ps.LastUsedAt = time.Now()
				m.poolMu.Unlock()

				m.metricsMu.Lock()
				m.metrics.Reused++
				m.metricsMu.Unlock()

				return ps.TmuxSession, nil
			}
			// Session died, remove from pool
			delete(m.pool, ps.ID)
		}
	}

	// Check pool size limit
	poolSize := len(m.pool)
	maxSize := m.poolConfig.MaxSize
	m.poolMu.Unlock()

	if poolSize >= maxSize {
		// Try to find ANY available session and change its working dir
		m.poolMu.Lock()
		for _, ps := range m.pool {
			if ps.Status == SessionAvailable {
				cmd := exec.Command("tmux", "has-session", "-t", ps.SessionName)
				if cmd.Run() == nil {
					// Change working directory
					cdCmd := exec.Command("tmux", "send-keys", "-t", ps.SessionName, "cd "+workingDir, "Enter")
					cdCmd.Run()

					ps.WorkingDir = workingDir
					ps.Status = SessionBusy
					ps.LastUsedAt = time.Now()
					m.poolMu.Unlock()

					m.metricsMu.Lock()
					m.metrics.Reused++
					m.metricsMu.Unlock()

					return ps.TmuxSession, nil
				}
				delete(m.pool, ps.ID)
			}
		}
		m.poolMu.Unlock()
		return nil, fmt.Errorf("session pool exhausted (max %d)", maxSize)
	}

	// Create new session
	sessionID := m.generateReadableSessionID()
	sessionName := "nexora-" + sessionID

	createCmd := exec.Command("tmux", "new-session", "-d", "-s", sessionName, "-x", "200", "-y", "50", "-c", workingDir)
	if err := createCmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to create pooled session: %w", err)
	}

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
		ID:          sessionID,
		SessionName: sessionName,
		PaneID:      paneID,
		WorkingDir:  workingDir,
		Description: "Pooled session",
		StartedAt:   time.Now(),
		done:        make(chan struct{}),
	}

	pooled := &PooledSession{
		TmuxSession: session,
		Status:      SessionBusy,
		LastUsedAt:  time.Now(),
	}

	m.poolMu.Lock()
	m.pool[sessionID] = pooled
	m.poolMu.Unlock()

	m.metricsMu.Lock()
	m.metrics.Created++
	m.metricsMu.Unlock()

	// Also add to main sessions map for compatibility
	m.mu.Lock()
	m.sessions[sessionID] = session
	m.mu.Unlock()

	return session, nil
}

// ReleaseSession returns a session to the pool for reuse
func (m *TmuxManager) ReleaseSession(sessionID string) {
	m.poolMu.Lock()
	defer m.poolMu.Unlock()

	if ps, exists := m.pool[sessionID]; exists {
		// Clear the terminal for next use
		exec.Command("tmux", "send-keys", "-t", ps.SessionName, "clear", "Enter").Run()

		ps.Status = SessionAvailable
		ps.LastUsedAt = time.Now()

		m.metricsMu.Lock()
		m.metrics.Released++
		m.metricsMu.Unlock()
	}
}

// GetMetrics returns current pool metrics
func (m *TmuxManager) GetMetrics() PoolMetrics {
	m.metricsMu.RLock()
	defer m.metricsMu.RUnlock()
	return m.metrics
}

// PoolStatus returns pool status info
func (m *TmuxManager) PoolStatus() map[string]interface{} {
	m.poolMu.RLock()
	defer m.poolMu.RUnlock()

	available := 0
	busy := 0
	draining := 0

	for _, ps := range m.pool {
		switch ps.Status {
		case SessionAvailable:
			available++
		case SessionBusy:
			busy++
		case SessionDraining:
			draining++
		}
	}

	metrics := m.GetMetrics()

	return map[string]interface{}{
		"pool_size":  len(m.pool),
		"max_size":   m.poolConfig.MaxSize,
		"available":  available,
		"busy":       busy,
		"draining":   draining,
		"created":    metrics.Created,
		"reused":     metrics.Reused,
		"released":   metrics.Released,
		"cleaned":    metrics.Cleaned,
		"reuse_rate": calculateReuseRate(metrics),
	}
}

// calculateReuseRate computes session reuse percentage
func calculateReuseRate(m PoolMetrics) float64 {
	total := m.Created + m.Reused
	if total == 0 {
		return 0
	}
	return float64(m.Reused) / float64(total) * 100
}

// DrainPool marks all sessions for cleanup
func (m *TmuxManager) DrainPool() {
	m.poolMu.Lock()
	defer m.poolMu.Unlock()

	for _, ps := range m.pool {
		if ps.Status == SessionAvailable {
			ps.Status = SessionDraining
		}
	}
}

// KillAllPooled kills all pooled sessions immediately
func (m *TmuxManager) KillAllPooled() int {
	m.poolMu.Lock()
	defer m.poolMu.Unlock()

	killed := 0
	for id, ps := range m.pool {
		exec.Command("tmux", "kill-session", "-t", ps.SessionName).Run()
		delete(m.pool, id)
		killed++
	}

	return killed
}
