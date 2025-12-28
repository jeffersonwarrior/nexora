package about

import (
	"runtime"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/nexora/nexora/internal/tui/components/dialogs"
	"github.com/nexora/nexora/internal/version"
)

func TestNewAboutDialog(t *testing.T) {
	dialog := NewAboutDialog()
	if dialog == nil {
		t.Fatal("NewAboutDialog returned nil")
	}

	if dialog.ID() != AboutDialogID {
		t.Errorf("expected dialog ID %q, got %q", AboutDialogID, dialog.ID())
	}
}

func TestAboutDialog_Init(t *testing.T) {
	dialog := NewAboutDialog()
	cmd := dialog.Init()
	if cmd != nil {
		t.Error("expected Init to return nil")
	}
}

func TestAboutDialog_Update(t *testing.T) {
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
			name: "enter key closes dialog",
			msg: tea.KeyPressMsg(tea.Key{
				Code: tea.KeyEnter,
			}),
			expectClose: true,
		},
		{
			name: "q key closes dialog",
			msg: tea.KeyPressMsg(tea.Key{
				Code: 'q',
				Text: "q",
			}),
			expectClose: true,
		},
		{
			name: "other key does nothing",
			msg: tea.KeyPressMsg(tea.Key{
				Code: 'x',
				Text: "x",
			}),
			expectClose: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog := NewAboutDialog()

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

func TestAboutDialog_View(t *testing.T) {
	dialog := NewAboutDialog()

	view := dialog.View()

	if view == "" {
		t.Error("expected non-empty view")
	}

	// Check for key content
	expectedContent := []string{
		"NEXORA",
		version.Display(),
		"AI-native terminal application",
		runtime.GOOS,
		runtime.GOARCH,
		"Go",
		"Press any key to close",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(view, expected) {
			t.Errorf("expected view to contain %q", expected)
		}
	}
}

func TestAboutDialog_Position(t *testing.T) {
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
			dialog := NewAboutDialog()

			// Update window size
			dialog.Update(tea.WindowSizeMsg{
				Width:  tt.width,
				Height: tt.height,
			})

			row, col := dialog.Position()

			// Position should be centered
			if row < 0 || row >= tt.height {
				t.Errorf("row %d out of bounds for height %d", row, tt.height)
			}
			if col < 0 || col >= tt.width {
				t.Errorf("col %d out of bounds for width %d", col, tt.width)
			}

			// Check that it's approximately centered
			contentWidth := 40
			contentHeight := 9
			expectedRow := (tt.height - contentHeight) / 2
			expectedCol := (tt.width - contentWidth) / 2

			if row != expectedRow {
				t.Errorf("expected row %d, got %d", expectedRow, row)
			}
			if col != expectedCol {
				t.Errorf("expected col %d, got %d", expectedCol, col)
			}
		})
	}
}

func TestAboutDialog_ID(t *testing.T) {
	dialog := NewAboutDialog()

	id := dialog.ID()

	if id != AboutDialogID {
		t.Errorf("expected ID %q, got %q", AboutDialogID, id)
	}

	// Verify DialogID is the expected value
	const expectedID = "about"
	if AboutDialogID != expectedID {
		t.Errorf("expected AboutDialogID to be %q, got %q", expectedID, AboutDialogID)
	}
}
