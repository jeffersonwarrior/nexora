package core

import (
	"testing"

	tea "charm.land/bubbletea/v2"
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
