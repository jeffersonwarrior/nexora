package filepicker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/nexora/nexora/internal/tui/components/dialogs"
)

func TestNewFilePickerCmp(t *testing.T) {
	tests := []struct {
		name       string
		workingDir string
	}{
		{
			name:       "with working directory",
			workingDir: "/tmp",
		},
		{
			name:       "empty working directory",
			workingDir: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			picker := NewFilePickerCmp(tt.workingDir)
			if picker == nil {
				t.Fatal("NewFilePickerCmp returned nil")
			}

			if picker.ID() != FilePickerID {
				t.Errorf("expected picker ID %q, got %q", FilePickerID, picker.ID())
			}
		})
	}
}

func TestFilePicker_Init(t *testing.T) {
	picker := NewFilePickerCmp("/tmp")
	cmd := picker.Init()
	// Init returns the filepicker init command
	if cmd == nil {
		t.Error("expected Init to return non-nil command")
	}
}

func TestFilePicker_Update(t *testing.T) {
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
			name: "down key navigates (handled by filepicker)",
			msg: tea.KeyPressMsg(tea.Key{
				Code: tea.KeyDown,
			}),
			expectClose: false,
		},
		{
			name: "up key navigates (handled by filepicker)",
			msg: tea.KeyPressMsg(tea.Key{
				Code: tea.KeyUp,
			}),
			expectClose: false,
		},
		{
			name: "j key navigates (handled by filepicker)",
			msg: tea.KeyPressMsg(tea.Key{
				Code: 'j',
				Text: "j",
			}),
			expectClose: false,
		},
		{
			name: "k key navigates (handled by filepicker)",
			msg: tea.KeyPressMsg(tea.Key{
				Code: 'k',
				Text: "k",
			}),
			expectClose: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			picker := NewFilePickerCmp("/tmp")
			picker.Init()

			_, cmd := picker.Update(tt.msg)

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

func TestFilePicker_View(t *testing.T) {
	picker := NewFilePickerCmp("/tmp")

	view := picker.View()

	if view == "" {
		t.Error("expected non-empty view")
	}

	// Check for key content
	if !strings.Contains(view, "Add Image") {
		t.Error("expected view to contain title")
	}
}

func TestFilePicker_Position(t *testing.T) {
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
			picker := NewFilePickerCmp("/tmp")

			// Update window size
			picker.Update(tea.WindowSizeMsg{
				Width:  tt.width,
				Height: tt.height,
			})

			row, col := picker.Position()

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

func TestFilePicker_ID(t *testing.T) {
	picker := NewFilePickerCmp("/tmp")

	id := picker.ID()

	if id != FilePickerID {
		t.Errorf("expected ID %q, got %q", FilePickerID, id)
	}

	// Verify FilePickerID is the expected value
	const expectedID = "filepicker"
	if FilePickerID != expectedID {
		t.Errorf("expected FilePickerID to be %q, got %q", expectedID, FilePickerID)
	}
}

func TestIsFileTooBig(t *testing.T) {
	tests := []struct {
		name      string
		fileSize  int64
		sizeLimit int64
		expectBig bool
	}{
		{
			name:      "file within limit",
			fileSize:  1024,
			sizeLimit: 2048,
			expectBig: false,
		},
		{
			name:      "file at limit",
			fileSize:  2048,
			sizeLimit: 2048,
			expectBig: false,
		},
		{
			name:      "file exceeds limit",
			fileSize:  3072,
			sizeLimit: 2048,
			expectBig: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary file with the specified size
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test.txt")

			f, err := os.Create(tmpFile)
			if err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}

			// Write data to reach the desired size
			if tt.fileSize > 0 {
				if err := f.Truncate(tt.fileSize); err != nil {
					f.Close()
					t.Fatalf("failed to set file size: %v", err)
				}
			}
			f.Close()

			// Test the function
			isTooBig, err := IsFileTooBig(tmpFile, tt.sizeLimit)
			if err != nil {
				t.Fatalf("IsFileTooBig returned error: %v", err)
			}

			if isTooBig != tt.expectBig {
				t.Errorf("expected isTooBig=%v, got %v", tt.expectBig, isTooBig)
			}
		})
	}
}

func TestIsFileTooBig_NonExistentFile(t *testing.T) {
	_, err := IsFileTooBig("/nonexistent/file.txt", 1024)
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestAllowedTypes(t *testing.T) {
	expected := []string{".jpg", ".jpeg", ".png"}

	if len(AllowedTypes) != len(expected) {
		t.Errorf("expected %d allowed types, got %d", len(expected), len(AllowedTypes))
	}

	for i, ext := range expected {
		if i >= len(AllowedTypes) {
			break
		}
		if AllowedTypes[i] != ext {
			t.Errorf("expected AllowedTypes[%d] to be %q, got %q", i, ext, AllowedTypes[i])
		}
	}
}

func TestMaxAttachmentSize(t *testing.T) {
	expectedSize := int64(5 * 1024 * 1024) // 5MB

	if MaxAttachmentSize != expectedSize {
		t.Errorf("expected MaxAttachmentSize to be %d, got %d", expectedSize, MaxAttachmentSize)
	}
}
