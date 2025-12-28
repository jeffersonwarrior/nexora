package reasoning

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/nexora/nexora/internal/tui/components/dialogs"
)

func TestNewReasoningDialog(t *testing.T) {
	dialog := NewReasoningDialog()
	if dialog == nil {
		t.Fatal("NewReasoningDialog returned nil")
	}

	if dialog.ID() != ReasoningDialogID {
		t.Errorf("expected dialog ID %q, got %q", ReasoningDialogID, dialog.ID())
	}
}

func TestReasoningDialog_Init(t *testing.T) {
	// Skip this test since it requires config to be initialized
	// which is not trivial in a unit test context
	t.Skip("Requires config initialization")
}

func TestReasoningDialog_Update(t *testing.T) {
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
			dialog := NewReasoningDialog()
			// Don't call Init() as it requires config initialization

			_, cmd := dialog.Update(tt.msg)

			gotClose := false
			if cmd != nil {
				// Execute the command to get the message
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

func TestReasoningDialog_View(t *testing.T) {
	dialog := NewReasoningDialog()

	view := dialog.View()

	if view == "" {
		t.Error("expected non-empty view")
	}

	// Check for key content
	if !strings.Contains(view, "Select Reasoning Effort") {
		t.Error("expected view to contain title")
	}
}

func TestReasoningDialog_Position(t *testing.T) {
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
			dialog := NewReasoningDialog()

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

func TestReasoningDialog_ID(t *testing.T) {
	dialog := NewReasoningDialog()

	id := dialog.ID()

	if id != ReasoningDialogID {
		t.Errorf("expected ID %q, got %q", ReasoningDialogID, id)
	}

	// Verify DialogID is the expected value
	const expectedID = "reasoning"
	if ReasoningDialogID != expectedID {
		t.Errorf("expected ReasoningDialogID to be %q, got %q", expectedID, ReasoningDialogID)
	}
}

func TestReasoningDialog_Cursor(t *testing.T) {
	dialog := NewReasoningDialog().(*reasoningDialogCmp)

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

func TestReasoningDialogKeyMap_ShortHelp(t *testing.T) {
	keyMap := DefaultReasoningDialogKeyMap()

	bindings := keyMap.ShortHelp()

	if len(bindings) != 2 {
		t.Errorf("expected 2 short help bindings, got %d", len(bindings))
	}
}

func TestReasoningDialogKeyMap_FullHelp(t *testing.T) {
	keyMap := DefaultReasoningDialogKeyMap()

	bindings := keyMap.FullHelp()

	if len(bindings) != 2 {
		t.Errorf("expected 2 full help rows, got %d", len(bindings))
	}
}
