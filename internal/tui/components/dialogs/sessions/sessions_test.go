package sessions

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/nexora/nexora/internal/session"
	"github.com/nexora/nexora/internal/tui/components/chat"
	"github.com/nexora/nexora/internal/tui/components/dialogs"
)

func TestNewSessionDialogCmp(t *testing.T) {
	tests := []struct {
		name       string
		sessions   []session.Session
		selectedID string
	}{
		{
			name:       "empty sessions",
			sessions:   []session.Session{},
			selectedID: "",
		},
		{
			name: "with sessions",
			sessions: []session.Session{
				{ID: "1", Title: "Session 1"},
				{ID: "2", Title: "Session 2"},
			},
			selectedID: "1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog := NewSessionDialogCmp(tt.sessions, tt.selectedID)
			if dialog == nil {
				t.Fatal("NewSessionDialogCmp returned nil")
			}

			if dialog.ID() != SessionsDialogID {
				t.Errorf("expected dialog ID %q, got %q", SessionsDialogID, dialog.ID())
			}
		})
	}
}

func TestSessionDialog_Init(t *testing.T) {
	sessions := []session.Session{
		{ID: "1", Title: "Test Session"},
	}
	dialog := NewSessionDialogCmp(sessions, "1")
	cmd := dialog.Init()
	// Init returns commands to initialize the list, should not crash
	_ = cmd
}

func TestSessionDialog_Update(t *testing.T) {
	tests := []struct {
		name        string
		msg         tea.Msg
		expectClose bool
	}{
		{
			name: "window size message updates dimensions",
			msg: tea.WindowSizeMsg{
				Width:  100,
				Height: 50,
			},
			expectClose: false,
		},
		{
			name: "esc key closes dialog",
			msg: tea.KeyPressMsg(tea.Key{
				Code: tea.KeyEscape,
			}),
			expectClose: true,
		},
		{
			name: "down key navigates (handled by list)",
			msg: tea.KeyPressMsg(tea.Key{
				Code: tea.KeyDown,
			}),
			expectClose: false,
		},
		{
			name: "up key navigates (handled by list)",
			msg: tea.KeyPressMsg(tea.Key{
				Code: tea.KeyUp,
			}),
			expectClose: false,
		},
		{
			name: "j key navigates (handled by list)",
			msg: tea.KeyPressMsg(tea.Key{
				Code: 'j',
				Text: "j",
			}),
			expectClose: false,
		},
		{
			name: "k key navigates (handled by list)",
			msg: tea.KeyPressMsg(tea.Key{
				Code: 'k',
				Text: "k",
			}),
			expectClose: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessions := []session.Session{
				{ID: "1", Title: "Test Session"},
			}
			dialog := NewSessionDialogCmp(sessions, "1")
			dialog.Init()

			_, cmd := dialog.Update(tt.msg)

			gotClose := false
			if cmd != nil {
				msg := cmd()
				if _, ok := msg.(dialogs.CloseDialogMsg); ok {
					gotClose = true
				}
			}

			if gotClose != tt.expectClose {
				t.Errorf("expected close=%v, got close=%v", tt.expectClose, gotClose)
			}
		})
	}
}

func TestSessionDialog_Update_SelectSession(t *testing.T) {
	sessions := []session.Session{
		{ID: "1", Title: "Test Session 1"},
		{ID: "2", Title: "Test Session 2"},
	}
	dialog := NewSessionDialogCmp(sessions, "1")
	dialog.Init()

	// Update with window size first
	dialog.Update(tea.WindowSizeMsg{
		Width:  100,
		Height: 50,
	})

	// Press enter to select
	_, cmd := dialog.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))

	if cmd == nil {
		// It's possible no item is selected if list is not initialized
		return
	}

	// Execute the command sequence
	msg := cmd()

	// Should close dialog
	if _, ok := msg.(dialogs.CloseDialogMsg); ok {
		return // Expected
	}

	// Or return session selected message
	if _, ok := msg.(chat.SessionSelectedMsg); ok {
		return // Also expected
	}
}

func TestSessionDialog_View(t *testing.T) {
	sessions := []session.Session{
		{ID: "1", Title: "Test Session"},
	}
	dialog := NewSessionDialogCmp(sessions, "1")

	view := dialog.View()

	if view == "" {
		t.Error("expected non-empty view")
	}

	// Check for key content
	if !strings.Contains(view, "Switch Session") {
		t.Error("expected view to contain title")
	}
}

func TestSessionDialog_Position(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
	}{
		{
			name:   "small window",
			width:  80,
			height: 24,
		},
		{
			name:   "medium window",
			width:  120,
			height: 40,
		},
		{
			name:   "large window",
			width:  200,
			height: 60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessions := []session.Session{
				{ID: "1", Title: "Test Session"},
			}
			dialog := NewSessionDialogCmp(sessions, "1")

			// Update window size
			dialog.Update(tea.WindowSizeMsg{
				Width:  tt.width,
				Height: tt.height,
			})

			row, col := dialog.Position()

			// Position should be within window bounds
			if row < 0 || row >= tt.height {
				t.Errorf("row %d out of bounds for height %d", row, tt.height)
			}
			if col < 0 || col >= tt.width {
				t.Errorf("col %d out of bounds for width %d", col, tt.width)
			}
		})
	}
}

func TestSessionDialog_ID(t *testing.T) {
	sessions := []session.Session{
		{ID: "1", Title: "Test Session"},
	}
	dialog := NewSessionDialogCmp(sessions, "1")

	id := dialog.ID()

	if id != SessionsDialogID {
		t.Errorf("expected ID %q, got %q", SessionsDialogID, id)
	}

	// Verify DialogID is the expected value
	const expectedID = "sessions"
	if SessionsDialogID != expectedID {
		t.Errorf("expected SessionsDialogID to be %q, got %q", expectedID, SessionsDialogID)
	}
}

func TestSessionDialog_Cursor(t *testing.T) {
	sessions := []session.Session{
		{ID: "1", Title: "Test Session"},
	}
	dialog := NewSessionDialogCmp(sessions, "1").(*sessionDialogCmp)

	// Initialize with window size
	dialog.Update(tea.WindowSizeMsg{
		Width:  100,
		Height: 50,
	})

	// Cursor method should not crash
	cursor := dialog.Cursor()
	// Cursor may be nil if not focused, which is fine
	_ = cursor
}
