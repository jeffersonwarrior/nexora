package quit

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/nexora/nexora/internal/tui/components/dialogs"
)

func TestNewQuitDialog(t *testing.T) {
	dialog := NewQuitDialog()
	if dialog == nil {
		t.Fatal("NewQuitDialog returned nil")
	}

	if dialog.ID() != QuitDialogID {
		t.Errorf("expected dialog ID %q, got %q", QuitDialogID, dialog.ID())
	}
}

func TestQuitDialog_Init(t *testing.T) {
	dialog := NewQuitDialog()
	cmd := dialog.Init()
	if cmd != nil {
		t.Error("expected Init to return nil")
	}
}

func TestQuitDialog_Update(t *testing.T) {
	tests := []struct {
		name        string
		msg         tea.Msg
		expectClose bool
		expectQuit  bool
	}{
		{
			name: "window size message updates dimensions",
			msg: tea.WindowSizeMsg{
				Width:  100,
				Height: 50,
			},
			expectClose: false,
			expectQuit:  false,
		},
		{
			name: "tab key toggles selection",
			msg: tea.KeyPressMsg(tea.Key{
				Code: tea.KeyTab,
			}),
			expectClose: false,
			expectQuit:  false,
		},
		{
			name: "left key toggles selection",
			msg: tea.KeyPressMsg(tea.Key{
				Code: tea.KeyLeft,
			}),
			expectClose: false,
			expectQuit:  false,
		},
		{
			name: "right key toggles selection",
			msg: tea.KeyPressMsg(tea.Key{
				Code: tea.KeyRight,
			}),
			expectClose: false,
			expectQuit:  false,
		},
		{
			name: "y key quits immediately",
			msg: tea.KeyPressMsg(tea.Key{
				Code: 'y',
				Text: "y",
			}),
			expectClose: false,
			expectQuit:  true,
		},
		{
			name: "n key closes dialog",
			msg: tea.KeyPressMsg(tea.Key{
				Code: 'n',
				Text: "n",
			}),
			expectClose: true,
			expectQuit:  false,
		},
		{
			name: "esc key closes dialog",
			msg: tea.KeyPressMsg(tea.Key{
				Code: tea.KeyEscape,
			}),
			expectClose: true,
			expectQuit:  false,
		},
		{
			name: "other key does nothing",
			msg: tea.KeyPressMsg(tea.Key{
				Code: 'x',
				Text: "x",
			}),
			expectClose: false,
			expectQuit:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog := NewQuitDialog()

			_, cmd := dialog.Update(tt.msg)

			gotClose := false
			gotQuit := false
			if cmd != nil {
				msg := cmd()
				if _, ok := msg.(dialogs.CloseDialogMsg); ok {
					gotClose = true
				}
				if _, ok := msg.(tea.QuitMsg); ok {
					gotQuit = true
				}
			}

			if gotClose != tt.expectClose {
				t.Errorf("expected close=%v, got close=%v", tt.expectClose, gotClose)
			}
			if gotQuit != tt.expectQuit {
				t.Errorf("expected quit=%v, got quit=%v", tt.expectQuit, gotQuit)
			}
		})
	}
}

func TestQuitDialog_Update_EnterOnYes(t *testing.T) {
	dialog := NewQuitDialog().(*quitDialogCmp)

	// Default is "No" selected, toggle to "Yes"
	dialog.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyTab}))

	// Now press enter - should quit
	_, cmd := dialog.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))

	if cmd == nil {
		t.Fatal("expected command when pressing enter on Yes")
	}

	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Error("expected QuitMsg when pressing enter on Yes")
	}
}

func TestQuitDialog_Update_EnterOnNo(t *testing.T) {
	dialog := NewQuitDialog()

	// Default is "No" selected
	_, cmd := dialog.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))

	if cmd == nil {
		t.Fatal("expected command when pressing enter on No")
	}

	msg := cmd()
	if _, ok := msg.(dialogs.CloseDialogMsg); !ok {
		t.Error("expected CloseDialogMsg when pressing enter on No")
	}
}

func TestQuitDialog_View(t *testing.T) {
	dialog := NewQuitDialog()

	view := dialog.View()

	if view == "" {
		t.Error("expected non-empty view")
	}

	// Check for key content
	if !strings.Contains(view, "Are you sure you want to quit?") {
		t.Error("expected view to contain quit question")
	}
}

func TestQuitDialog_Position(t *testing.T) {
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
			dialog := NewQuitDialog()

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

func TestQuitDialog_ID(t *testing.T) {
	dialog := NewQuitDialog()

	id := dialog.ID()

	if id != QuitDialogID {
		t.Errorf("expected ID %q, got %q", QuitDialogID, id)
	}

	// Verify DialogID is the expected value
	const expectedID = "quit"
	if QuitDialogID != expectedID {
		t.Errorf("expected QuitDialogID to be %q, got %q", expectedID, QuitDialogID)
	}
}

func TestQuitDialog_DefaultSelection(t *testing.T) {
	dialog := NewQuitDialog().(*quitDialogCmp)

	// Default should be "No" for safety
	if !dialog.selectedNo {
		t.Error("expected default selection to be No for safety")
	}
}
