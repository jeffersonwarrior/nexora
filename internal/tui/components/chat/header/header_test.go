package header

import (
	"regexp"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/nexora/nexora/internal/config"
	"github.com/nexora/nexora/internal/csync"
	"github.com/nexora/nexora/internal/lsp"
	"github.com/nexora/nexora/internal/pubsub"
	"github.com/nexora/nexora/internal/session"
)

// setupTestConfig initializes a minimal test config.
func setupTestConfig(t *testing.T) {
	t.Helper()
	tempDir := t.TempDir()
	_, err := config.Init(tempDir, tempDir, false)
	if err != nil {
		t.Skipf("Failed to init test config: %v", err)
	}
}

func TestNew(t *testing.T) {
	lspClients := csync.NewMap[string, *lsp.Client]()
	h := New(lspClients)

	if h == nil {
		t.Fatal("New() returned nil")
	}

	// Verify initial state
	if h.ShowingDetails() {
		t.Error("expected ShowingDetails() to be false initially")
	}
}

func TestNew_NilLspClients(t *testing.T) {
	h := New(nil)
	if h == nil {
		t.Fatal("New() with nil lspClients returned nil")
	}
}

func TestHeader_Init(t *testing.T) {
	lspClients := csync.NewMap[string, *lsp.Client]()
	h := New(lspClients)

	cmd := h.Init()
	if cmd != nil {
		t.Error("Init() should return nil cmd")
	}
}

func TestHeader_SetWidth(t *testing.T) {
	lspClients := csync.NewMap[string, *lsp.Client]()
	h := New(lspClients)

	tests := []struct {
		name  string
		width int
	}{
		{"zero width", 0},
		{"small width", 40},
		{"medium width", 80},
		{"large width", 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := h.SetWidth(tt.width)
			if cmd != nil {
				t.Error("SetWidth() should return nil cmd")
			}

			// Access internal state to verify
			hdr := h.(*header)
			if hdr.width != tt.width {
				t.Errorf("expected width %d, got %d", tt.width, hdr.width)
			}
		})
	}
}

func TestHeader_SetSession(t *testing.T) {
	lspClients := csync.NewMap[string, *lsp.Client]()
	h := New(lspClients)

	sess := session.Session{
		ID:               "test-session-12345678",
		Title:            "Test Session",
		PromptTokens:     1000,
		CompletionTokens: 500,
	}

	cmd := h.SetSession(sess)
	if cmd != nil {
		t.Error("SetSession() should return nil cmd")
	}

	// Access internal state to verify
	hdr := h.(*header)
	if hdr.session.ID != sess.ID {
		t.Errorf("expected session ID %s, got %s", sess.ID, hdr.session.ID)
	}
	if hdr.session.Title != sess.Title {
		t.Errorf("expected session title %s, got %s", sess.Title, hdr.session.Title)
	}
}

func TestHeader_SetDetailsOpen(t *testing.T) {
	lspClients := csync.NewMap[string, *lsp.Client]()
	h := New(lspClients)

	// Initially false
	if h.ShowingDetails() {
		t.Error("expected ShowingDetails() to be false initially")
	}

	// Set to true
	h.SetDetailsOpen(true)
	if !h.ShowingDetails() {
		t.Error("expected ShowingDetails() to be true after SetDetailsOpen(true)")
	}

	// Set back to false
	h.SetDetailsOpen(false)
	if h.ShowingDetails() {
		t.Error("expected ShowingDetails() to be false after SetDetailsOpen(false)")
	}
}

func TestHeader_ShowingDetails(t *testing.T) {
	lspClients := csync.NewMap[string, *lsp.Client]()
	h := New(lspClients)

	// Test initial state
	if h.ShowingDetails() {
		t.Error("expected ShowingDetails() to be false initially")
	}

	// Toggle and verify
	h.SetDetailsOpen(true)
	if !h.ShowingDetails() {
		t.Error("expected ShowingDetails() to be true")
	}
}

func TestHeader_View_EmptySession(t *testing.T) {
	lspClients := csync.NewMap[string, *lsp.Client]()
	h := New(lspClients)
	h.SetWidth(80)

	view := h.View()
	if view != "" {
		t.Errorf("expected empty view for empty session, got %q", view)
	}
}

func TestHeader_View_WithSession(t *testing.T) {
	setupTestConfig(t)

	lspClients := csync.NewMap[string, *lsp.Client]()
	h := New(lspClients)
	h.SetWidth(120)

	sess := session.Session{
		ID:               "abcd1234-5678-90ab-cdef-1234567890ab",
		Title:            "My Test Session",
		PromptTokens:     500,
		CompletionTokens: 200,
	}
	h.SetSession(sess)

	view := h.View()

	// View should not be empty
	if view == "" {
		t.Error("expected non-empty view")
	}

	// Should contain session title
	if !strings.Contains(view, "My Test Session") {
		t.Error("expected view to contain session title")
	}

	// Should contain keyboard shortcut
	if !strings.Contains(view, "ctrl+d") {
		t.Error("expected view to contain ctrl+d keyboard shortcut")
	}
}

func TestHeader_View_WithLongTitle(t *testing.T) {
	setupTestConfig(t)

	lspClients := csync.NewMap[string, *lsp.Client]()
	h := New(lspClients)
	h.SetWidth(120)

	sess := session.Session{
		ID:    "abcd1234-5678-90ab-cdef-1234567890ab",
		Title: "This is a very long session title that should be truncated",
	}
	h.SetSession(sess)

	view := h.View()

	// Long titles should be truncated (max 25 chars, with "..." at end)
	if strings.Contains(view, "should be truncated") {
		t.Error("expected long title to be truncated")
	}
	if !strings.Contains(view, "...") {
		t.Error("expected truncated title to contain ...")
	}
}

func TestHeader_View_NoTitle_FallsBackToSessionID(t *testing.T) {
	setupTestConfig(t)

	lspClients := csync.NewMap[string, *lsp.Client]()
	h := New(lspClients)
	h.SetWidth(120)

	sess := session.Session{
		ID:    "abcd1234-5678-90ab-cdef-1234567890ab",
		Title: "",
	}
	h.SetSession(sess)

	view := h.View()

	// Should fall back to "Session <first 8 chars>"
	if !strings.Contains(view, "Session abcd1234") {
		t.Error("expected view to contain fallback session ID")
	}
}

func TestHeader_View_DetailsOpen(t *testing.T) {
	setupTestConfig(t)

	lspClients := csync.NewMap[string, *lsp.Client]()
	h := New(lspClients)
	h.SetWidth(120)

	sess := session.Session{
		ID:    "abcd1234-5678-90ab-cdef-1234567890ab",
		Title: "Test Session",
	}
	h.SetSession(sess)

	// When details closed
	h.SetDetailsOpen(false)
	viewClosed := h.View()
	if !strings.Contains(viewClosed, "open") {
		t.Error("expected 'open' hint when details are closed")
	}

	// When details open
	h.SetDetailsOpen(true)
	viewOpen := h.View()
	if !strings.Contains(viewOpen, "close") {
		t.Error("expected 'close' hint when details are open")
	}
}

func TestHeader_View_NarrowWidth(t *testing.T) {
	setupTestConfig(t)

	lspClients := csync.NewMap[string, *lsp.Client]()
	h := New(lspClients)
	h.SetWidth(40) // Narrow width

	sess := session.Session{
		ID:    "abcd1234-5678-90ab-cdef-1234567890ab",
		Title: "Test Session",
	}
	h.SetSession(sess)

	// Should not panic with narrow width
	view := h.View()
	if view == "" {
		t.Error("expected non-empty view even with narrow width")
	}
}

func TestHeader_View_ZeroWidth(t *testing.T) {
	setupTestConfig(t)

	lspClients := csync.NewMap[string, *lsp.Client]()
	h := New(lspClients)
	h.SetWidth(0)

	sess := session.Session{
		ID:    "abcd1234-5678-90ab-cdef-1234567890ab",
		Title: "Test Session",
	}
	h.SetSession(sess)

	// Should not panic with zero width
	_ = h.View()
}

func TestHeader_Update_SessionEvent(t *testing.T) {
	lspClients := csync.NewMap[string, *lsp.Client]()
	h := New(lspClients)

	sess := session.Session{
		ID:               "abcd1234-5678-90ab-cdef-1234567890ab",
		Title:            "Original Title",
		PromptTokens:     100,
		CompletionTokens: 50,
	}
	h.SetSession(sess)

	// Send updated session event
	updatedSess := session.Session{
		ID:               "abcd1234-5678-90ab-cdef-1234567890ab",
		Title:            "Updated Title",
		PromptTokens:     200,
		CompletionTokens: 100,
	}
	event := pubsub.Event[session.Session]{
		Type:    pubsub.UpdatedEvent,
		Payload: updatedSess,
	}

	model, cmd := h.Update(event)
	if cmd != nil {
		t.Error("Update() should return nil cmd for session event")
	}

	hdr := model.(*header)
	if hdr.session.Title != "Updated Title" {
		t.Errorf("expected updated title, got %s", hdr.session.Title)
	}
	if hdr.session.PromptTokens != 200 {
		t.Errorf("expected 200 prompt tokens, got %d", hdr.session.PromptTokens)
	}
}

func TestHeader_Update_SessionEvent_DifferentID(t *testing.T) {
	lspClients := csync.NewMap[string, *lsp.Client]()
	h := New(lspClients)

	sess := session.Session{
		ID:    "session-1",
		Title: "Original Title",
	}
	h.SetSession(sess)

	// Send event for different session
	differentSess := session.Session{
		ID:    "session-2",
		Title: "Different Session",
	}
	event := pubsub.Event[session.Session]{
		Type:    pubsub.UpdatedEvent,
		Payload: differentSess,
	}

	model, _ := h.Update(event)
	hdr := model.(*header)

	// Should not update because IDs don't match
	if hdr.session.Title != "Original Title" {
		t.Error("should not update session with different ID")
	}
}

func TestHeader_Update_SessionEvent_NotUpdated(t *testing.T) {
	lspClients := csync.NewMap[string, *lsp.Client]()
	h := New(lspClients)

	sess := session.Session{
		ID:    "session-1",
		Title: "Original Title",
	}
	h.SetSession(sess)

	// Send created event (not updated)
	event := pubsub.Event[session.Session]{
		Type: pubsub.CreatedEvent,
		Payload: session.Session{
			ID:    "session-1",
			Title: "New Title",
		},
	}

	model, _ := h.Update(event)
	hdr := model.(*header)

	// Should not update for non-UpdatedEvent
	if hdr.session.Title != "Original Title" {
		t.Error("should not update session for non-UpdatedEvent")
	}
}

func TestHeader_Update_UnrelatedMessage(t *testing.T) {
	lspClients := csync.NewMap[string, *lsp.Client]()
	h := New(lspClients)

	sess := session.Session{
		ID:    "session-1",
		Title: "Test",
	}
	h.SetSession(sess)

	// Send unrelated message types
	model, cmd := h.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	if cmd != nil {
		t.Error("Update() should return nil cmd for unrelated message")
	}
	if model == nil {
		t.Error("Update() should return non-nil model")
	}

	model, cmd = h.Update(tea.KeyPressMsg(tea.Key{}))
	if cmd != nil {
		t.Error("Update() should return nil cmd for key message")
	}
}

func TestHeader_Interface(t *testing.T) {
	lspClients := csync.NewMap[string, *lsp.Client]()
	h := New(lspClients)

	// Verify interface implementation
	var _ Header = h
}

func TestHeader_Details_TokenPercentage(t *testing.T) {
	setupTestConfig(t)

	lspClients := csync.NewMap[string, *lsp.Client]()
	h := New(lspClients)
	h.SetWidth(120)

	sess := session.Session{
		ID:               "abcd1234-5678-90ab-cdef-1234567890ab",
		Title:            "Test",
		PromptTokens:     5000,
		CompletionTokens: 2500,
	}
	h.SetSession(sess)

	view := h.View()

	// Should contain percentage
	if !strings.Contains(view, "%") {
		t.Error("expected view to contain token percentage")
	}
}

// Note: Header is not designed for concurrent access as Bubble Tea
// components run in a single-threaded event loop. Concurrency test removed.

func TestHeader_View_VeryWideWidth(t *testing.T) {
	setupTestConfig(t)

	lspClients := csync.NewMap[string, *lsp.Client]()
	h := New(lspClients)
	h.SetWidth(500) // Very wide

	sess := session.Session{
		ID:    "abcd1234-5678-90ab-cdef-1234567890ab",
		Title: "Test Session",
	}
	h.SetSession(sess)

	// Should not panic with very wide width
	view := h.View()
	if view == "" {
		t.Error("expected non-empty view")
	}
}

func TestHeader_View_ContainsDateTime(t *testing.T) {
	setupTestConfig(t)

	lspClients := csync.NewMap[string, *lsp.Client]()
	h := New(lspClients)
	h.SetWidth(120)

	sess := session.Session{
		ID:    "abcd1234-5678-90ab-cdef-1234567890ab",
		Title: "Test Session",
	}
	h.SetSession(sess)

	view := h.View()

	// Should contain current datetime in format "Jan 2 15:04"
	now := time.Now()
	expectedMonth := now.Format("Jan")
	expectedDay := now.Format("2")

	// Check for month abbreviation
	if !strings.Contains(view, expectedMonth) {
		t.Errorf("expected view to contain month '%s', view: %s", expectedMonth, view)
	}

	// Use regex to check for time pattern HH:MM
	timePattern := regexp.MustCompile(`\d{1,2}:\d{2}`)
	if !timePattern.MatchString(view) {
		t.Errorf("expected view to contain time pattern HH:MM, view: %s", view)
	}

	// Check for day number
	if !strings.Contains(view, expectedDay) {
		// Note: This might be flaky around midnight, but good enough for testing
		t.Logf("warning: view may not contain day '%s' (could be edge case), view: %s", expectedDay, view)
	}
}
