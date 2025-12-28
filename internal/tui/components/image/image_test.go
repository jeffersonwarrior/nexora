package image

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestNewModel(t *testing.T) {
	tests := []struct {
		name   string
		width  uint
		height uint
		url    string
		check  func(*Model) bool
	}{
		{
			name:   "with dimensions and url",
			width:  100,
			height: 50,
			url:    "http://example.com/image.jpg",
			check: func(m *Model) bool {
				return m.width == 100 && m.height == 50 && m.url == "http://example.com/image.jpg"
			},
		},
		{
			name:   "zero dimensions",
			width:  0,
			height: 0,
			url:    "file.png",
			check: func(m *Model) bool {
				return m.width == 0 && m.height == 0
			},
		},
		{
			name:   "empty url",
			width:  50,
			height: 50,
			url:    "",
			check: func(m *Model) bool {
				return m.url == ""
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := New(tt.width, tt.height, tt.url)
			if !tt.check(&model) {
				t.Errorf("New() failed validation")
			}
		})
	}
}

func TestModelInit(t *testing.T) {
	model := New(100, 50, "http://example.com/image.jpg")
	cmd := model.Init()

	// Init should return nil (no loading command by default)
	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestModelView(t *testing.T) {
	tests := []struct {
		name  string
		model *Model
		check func(string) bool
	}{
		{
			name: "with error",
			model: func() *Model {
				m := New(100, 50, "test.jpg")
				m.err = errMsg{error: &mockError{"test error"}}
				return &m
			}(),
			check: func(s string) bool {
				return len(s) > 0 && s == "couldn't load image(s): test error"
			},
		},
		{
			name: "with image content",
			model: func() *Model {
				m := New(100, 50, "test.jpg")
				m.image = "test image content"
				return &m
			}(),
			check: func(s string) bool {
				return s == "test image content"
			},
		},
		{
			name: "empty state",
			model: func() *Model {
				m := New(100, 50, "test.jpg")
				return &m
			}(),
			check: func(s string) bool {
				return s == ""
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view := tt.model.View()
			if !tt.check(view) {
				t.Errorf("View() returned %q, failed validation", view)
			}
		})
	}
}

func TestModelUpdateErrMsg(t *testing.T) {
	model := New(100, 50, "test.jpg")
	testErr := mockError{"test error"}
	msg := errMsg{error: &testErr}

	resultModel, cmd := model.Update(msg)

	// Should store the error
	if resultModel.err == nil {
		t.Error("error not stored in model")
	}

	// Should return nil command
	if cmd != nil {
		t.Error("Update(errMsg) should return nil command")
	}
}

func TestModelUpdateRedrawMsg(t *testing.T) {
	model := New(100, 50, "old.jpg")
	msg := redrawMsg{
		width:  200,
		height: 100,
		url:    "new.jpg",
	}

	resultModel, cmd := model.Update(msg)

	// Should update dimensions and url
	if resultModel.width != 200 || resultModel.height != 100 || resultModel.url != "new.jpg" {
		t.Errorf("redraw message not processed correctly")
	}

	// Should return a command (loadURL)
	if cmd == nil {
		t.Error("Update(redrawMsg) should return a command")
	}
}

func TestModelUpdateLoadMsg(t *testing.T) {
	// We can't fully test loadMsg without actual image data,
	// but we can verify that the message path exists in Update
	// This would require mocking io.ReadCloser which is complex
}

func TestModelUpdateOtherMsg(t *testing.T) {
	model := New(100, 50, "test.jpg")
	var msg tea.Msg = tea.BatchMsg{}

	resultModel, cmd := model.Update(msg)

	// Should return unchanged model
	if resultModel != model {
		t.Error("Update() returned wrong model")
	}

	// Should return nil command
	if cmd != nil {
		t.Error("Update(BatchMsg) should return nil command")
	}
}

func TestModelRedraw(t *testing.T) {
	model := New(100, 50, "old.jpg")
	cmd := model.Redraw(200, 100, "new.jpg")

	// Should return a command
	if cmd == nil {
		t.Error("Redraw() should return a command")
	}

	// Execute the command and check the message
	msg := cmd()
	redrawMsg, ok := msg.(redrawMsg)
	if !ok {
		t.Errorf("Redraw() command returned wrong message type")
	}

	if redrawMsg.width != 200 || redrawMsg.height != 100 || redrawMsg.url != "new.jpg" {
		t.Errorf("Redraw() command message has wrong values")
	}
}

func TestModelUpdateURL(t *testing.T) {
	model := New(100, 50, "old.jpg")
	cmd := model.UpdateURL("new.jpg")

	// Should return a command
	if cmd == nil {
		t.Error("UpdateURL() should return a command")
	}

	// Execute the command and check the message
	msg := cmd()
	redrawMsg, ok := msg.(redrawMsg)
	if !ok {
		t.Errorf("UpdateURL() command returned wrong message type")
	}

	// Should keep old dimensions but update url
	if redrawMsg.width != 100 || redrawMsg.height != 50 || redrawMsg.url != "new.jpg" {
		t.Errorf("UpdateURL() command message has wrong values")
	}
}

func TestModelIsLoading(t *testing.T) {
	tests := []struct {
		name      string
		imageStr  string
		expected  bool
	}{
		{
			name:      "empty image string means loading",
			imageStr:  "",
			expected:  true,
		},
		{
			name:      "non-empty image string means not loading",
			imageStr:  "some image data",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := New(100, 50, "test.jpg")
			model.image = tt.imageStr

			result := model.IsLoading()
			if result != tt.expected {
				t.Errorf("IsLoading() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Mock error for testing
type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}

func TestModelFieldAccess(t *testing.T) {
	// Test that we can set and read model fields
	model := New(100, 50, "test.jpg")

	// These are private fields, but we test them indirectly through Update
	model.image = "test image"
	model.err = nil

	if model.image != "test image" {
		t.Error("unable to set/get image field")
	}

	if model.err != nil {
		t.Error("error field should be nil")
	}
}

func TestModelDimensions(t *testing.T) {
	tests := []struct {
		name   string
		width  uint
		height uint
	}{
		{
			name:   "standard dimensions",
			width:  1920,
			height: 1080,
		},
		{
			name:   "square",
			width:  512,
			height: 512,
		},
		{
			name:   "portrait",
			width:  480,
			height: 640,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := New(tt.width, tt.height, "test.jpg")
			if model.width != tt.width || model.height != tt.height {
				t.Errorf("dimensions not set correctly")
			}
		})
	}
}
