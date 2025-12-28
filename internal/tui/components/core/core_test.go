package core

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFocusableInterface verifies the Focusable interface is properly defined
func TestFocusableInterface(t *testing.T) {
	// Test that interface exists and methods are properly typed
	var _ Focusable = (*testFocusable)(nil)
}

// TestSizeableInterface verifies the Sizeable interface is properly defined
func TestSizeableInterface(t *testing.T) {
	// Test that interface exists and methods are properly typed
	var _ Sizeable = (*testSizeable)(nil)
}

// testFocusable implements the Focusable interface for testing
type testFocusable struct {
	focused bool
}

func (t *testFocusable) Focus() tea.Cmd {
	t.focused = true
	return nil
}

func (t *testFocusable) Blur() tea.Cmd {
	t.focused = false
	return nil
}

func (t *testFocusable) IsFocused() bool {
	return t.focused
}

// testSizeable implements the Sizeable interface for testing
type testSizeable struct {
	width, height int
}

func (t *testSizeable) SetSize(width, height int) tea.Cmd {
	t.width, t.height = width, height
	return nil
}

func (t *testSizeable) GetSize() (int, int) {
	return t.width, t.height
}

// TestFocusableImplementation tests the focus/blur behavior
func TestFocusableImplementation(t *testing.T) {
	t.Parallel()

	tf := &testFocusable{focused: false}

	// Test initial state
	assert.False(t, tf.IsFocused(), "should not be focused initially")

	// Test Focus
	cmd := tf.Focus()
	assert.Nil(t, cmd, "Focus should return nil cmd")
	assert.True(t, tf.IsFocused(), "should be focused after Focus()")

	// Test Blur
	cmd = tf.Blur()
	assert.Nil(t, cmd, "Blur should return nil cmd")
	assert.False(t, tf.IsFocused(), "should not be focused after Blur()")
}

// TestSizeableImplementation tests the sizing behavior
func TestSizeableImplementation(t *testing.T) {
	t.Parallel()

	ts := &testSizeable{}

	// Test initial size
	width, height := ts.GetSize()
	assert.Equal(t, 0, width, "initial width should be 0")
	assert.Equal(t, 0, height, "initial height should be 0")

	// Test SetSize
	cmd := ts.SetSize(100, 50)
	assert.Nil(t, cmd, "SetSize should return nil cmd")

	width, height = ts.GetSize()
	assert.Equal(t, 100, width, "width should be set to 100")
	assert.Equal(t, 50, height, "height should be set to 50")

	// Test another SetSize
	ts.SetSize(200, 80)
	width, height = ts.GetSize()
	assert.Equal(t, 200, width, "width should be updated to 200")
	assert.Equal(t, 80, height, "height should be updated to 80")
}

// TestSection tests the Section helper function
func TestSection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		text  string
		width int
	}{
		{
			name:  "simple section",
			text:  "Status",
			width: 80,
		},
		{
			name:  "wide width",
			text:  "Title",
			width: 120,
		},
		{
			name:  "narrow width",
			text:  "X",
			width: 20,
		},
		{
			name:  "very narrow width",
			text:  "Test",
			width: 10,
		},
		{
			name:  "exact width",
			text:  "Exact",
			width: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := Section(tt.text, tt.width)
			assert.NotEmpty(t, result, "should return non-empty string")
			assert.Contains(t, result, tt.text, "should contain the text")
		})
	}
}

// TestSectionWithInfo tests the SectionWithInfo helper function
func TestSectionWithInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		text  string
		info  string
		width int
	}{
		{
			name:  "with info",
			text:  "Status",
			info:  "[active]",
			width: 80,
		},
		{
			name:  "empty info",
			text:  "Title",
			info:  "",
			width: 80,
		},
		{
			name:  "long info",
			text:  "Status",
			info:  "[very long info here]",
			width: 80,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := SectionWithInfo(tt.text, tt.width, tt.info)
			assert.NotEmpty(t, result, "should return non-empty string")
			assert.Contains(t, result, tt.text, "should contain the text")
			if tt.info != "" {
				assert.Contains(t, result, tt.info, "should contain the info when provided")
			}
		})
	}
}

// TestTitle tests the Title helper function
func TestTitle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		title string
		width int
	}{
		{
			name:  "simple title",
			title: "Main",
			width: 80,
		},
		{
			name:  "wide width",
			title: "Content",
			width: 120,
		},
		{
			name:  "narrow width",
			title: "X",
			width: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := Title(tt.title, tt.width)
			assert.NotEmpty(t, result, "should return non-empty string")
			assert.Contains(t, result, tt.title, "should contain the title")
		})
	}
}

// TestSelectableButton tests the SelectableButton function
func TestSelectableButton(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts ButtonOpts
	}{
		{
			name: "unselected button",
			opts: ButtonOpts{
				Text:           "Cancel",
				UnderlineIndex: 0,
				Selected:       false,
			},
		},
		{
			name: "selected button",
			opts: ButtonOpts{
				Text:           "OK",
				UnderlineIndex: 0,
				Selected:       true,
			},
		},
		{
			name: "button with middle underline",
			opts: ButtonOpts{
				Text:           "Delete",
				UnderlineIndex: 2,
				Selected:       false,
			},
		},
		{
			name: "button with negative underline index",
			opts: ButtonOpts{
				Text:           "Yes",
				UnderlineIndex: -1,
				Selected:       false,
			},
		},
		{
			name: "button with out of bounds underline index",
			opts: ButtonOpts{
				Text:           "No",
				UnderlineIndex: 10,
				Selected:       false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := SelectableButton(tt.opts)
			assert.NotEmpty(t, result, "should return non-empty string")
		})
	}
}

// TestSelectableButtons tests the SelectableButtons function
func TestSelectableButtons(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		buttons []ButtonOpts
		spacing string
	}{
		{
			name: "two buttons with default spacing",
			buttons: []ButtonOpts{
				{Text: "OK", UnderlineIndex: 0, Selected: true},
				{Text: "Cancel", UnderlineIndex: 0, Selected: false},
			},
			spacing: "",
		},
		{
			name: "two buttons with custom spacing",
			buttons: []ButtonOpts{
				{Text: "Yes", UnderlineIndex: 0, Selected: true},
				{Text: "No", UnderlineIndex: 0, Selected: false},
			},
			spacing: " | ",
		},
		{
			name: "three buttons",
			buttons: []ButtonOpts{
				{Text: "Save", UnderlineIndex: 0, Selected: false},
				{Text: "Discard", UnderlineIndex: 0, Selected: false},
				{Text: "Cancel", UnderlineIndex: 0, Selected: true},
			},
			spacing: "  ",
		},
		{
			name:    "single button",
			buttons: []ButtonOpts{{Text: "OK", UnderlineIndex: 0, Selected: true}},
			spacing: "  ",
		},
		{
			name:    "empty buttons",
			buttons: []ButtonOpts{},
			spacing: "  ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := SelectableButtons(tt.buttons, tt.spacing)
			// Result can be empty for empty buttons
			if len(tt.buttons) > 0 {
				assert.NotEmpty(t, result, "should return non-empty string for non-empty buttons")
			}
		})
	}
}

// TestSelectableButtonsVertical tests the SelectableButtonsVertical function
func TestSelectableButtonsVertical(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		buttons []ButtonOpts
		spacing int
	}{
		{
			name: "two buttons no spacing",
			buttons: []ButtonOpts{
				{Text: "OK", UnderlineIndex: 0, Selected: true},
				{Text: "Cancel", UnderlineIndex: 0, Selected: false},
			},
			spacing: 0,
		},
		{
			name: "two buttons with spacing",
			buttons: []ButtonOpts{
				{Text: "Yes", UnderlineIndex: 0, Selected: true},
				{Text: "No", UnderlineIndex: 0, Selected: false},
			},
			spacing: 1,
		},
		{
			name: "three buttons with double spacing",
			buttons: []ButtonOpts{
				{Text: "Save", UnderlineIndex: 0, Selected: false},
				{Text: "Discard", UnderlineIndex: 0, Selected: false},
				{Text: "Cancel", UnderlineIndex: 0, Selected: true},
			},
			spacing: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := SelectableButtonsVertical(tt.buttons, tt.spacing)
			if len(tt.buttons) > 0 {
				assert.NotEmpty(t, result, "should return non-empty string")
			}
		})
	}
}

// TestNewSimpleHelp tests the NewSimpleHelp function
func TestNewSimpleHelp(t *testing.T) {
	t.Parallel()

	shortList := []key.Binding{
		key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
	}
	fullList := [][]key.Binding{
		{key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "exit"))},
	}

	help := NewSimpleHelp(shortList, fullList)
	require.NotNil(t, help, "NewSimpleHelp should return non-nil help")

	shortHelp := help.ShortHelp()
	assert.Equal(t, len(shortList), len(shortHelp), "ShortHelp should return correct number of bindings")

	fullHelp := help.FullHelp()
	assert.Equal(t, len(fullList), len(fullHelp), "FullHelp should return correct number of binding groups")
}

// TestDiffFormatter tests the DiffFormatter function
func TestDiffFormatter(t *testing.T) {
	t.Parallel()

	formatter := DiffFormatter()
	require.NotNil(t, formatter, "DiffFormatter should return non-nil formatter")
}

// BenchmarkSelectableButton benchmarks the SelectableButton function
func BenchmarkSelectableButton(b *testing.B) {
	opts := ButtonOpts{
		Text:           "Button Text",
		UnderlineIndex: 0,
		Selected:       true,
	}

	for i := 0; i < b.N; i++ {
		SelectableButton(opts)
	}
}

// BenchmarkSection benchmarks the Section function
func BenchmarkSection(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Section("Test Section", 80)
	}
}
