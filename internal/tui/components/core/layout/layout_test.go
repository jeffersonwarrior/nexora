package layout

import (
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
)

// testHelp implements the Help interface for testing
type testHelp struct {
	bindings []key.Binding
}

func (t *testHelp) Bindings() []key.Binding {
	return t.bindings
}

// testPositional implements the Positional interface for testing
type testPositional struct {
	x, y int
}

func (t *testPositional) SetPosition(x, y int) tea.Cmd {
	t.x = x
	t.y = y
	return nil
}

// TestHelpInterface verifies the Help interface is properly defined
func TestHelpInterface(t *testing.T) {
	t.Parallel()

	var _ Help = (*testHelp)(nil)
}

// TestPositionalInterface verifies the Positional interface is properly defined
func TestPositionalInterface(t *testing.T) {
	t.Parallel()

	var _ Positional = (*testPositional)(nil)
}

// TestHelpImplementation tests the Help interface implementation
func TestHelpImplementation(t *testing.T) {
	t.Parallel()

	bindings := []key.Binding{
		key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
		key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "exit")),
	}

	help := &testHelp{bindings: bindings}
	assert.Equal(t, bindings, help.Bindings(), "should return correct bindings")
	assert.Len(t, help.Bindings(), 2, "should have 2 bindings")
}

// TestPositionalImplementation tests the Positional interface implementation
func TestPositionalImplementation(t *testing.T) {
	t.Parallel()

	pos := &testPositional{}

	// Test initial position
	assert.Equal(t, 0, pos.x, "initial x should be 0")
	assert.Equal(t, 0, pos.y, "initial y should be 0")

	// Test SetPosition
	cmd := pos.SetPosition(10, 20)
	assert.Nil(t, cmd, "SetPosition should return nil cmd")
	assert.Equal(t, 10, pos.x, "x should be set to 10")
	assert.Equal(t, 20, pos.y, "y should be set to 20")

	// Test another SetPosition
	pos.SetPosition(50, 100)
	assert.Equal(t, 50, pos.x, "x should be updated to 50")
	assert.Equal(t, 100, pos.y, "y should be updated to 100")
}

// TestHelpWithEmptyBindings tests Help with empty bindings
func TestHelpWithEmptyBindings(t *testing.T) {
	t.Parallel()

	help := &testHelp{bindings: []key.Binding{}}
	assert.Empty(t, help.Bindings(), "should have no bindings")
	assert.Len(t, help.Bindings(), 0, "length should be 0")
}

// TestHelpWithMultipleBindings tests Help with multiple bindings
func TestHelpWithMultipleBindings(t *testing.T) {
	t.Parallel()

	bindings := make([]key.Binding, 5)
	for i := 0; i < 5; i++ {
		bindings[i] = key.NewBinding()
	}

	help := &testHelp{bindings: bindings}
	assert.Len(t, help.Bindings(), 5, "should have 5 bindings")
}

// TestPositionalCoordinates tests various coordinate values
func TestPositionalCoordinates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		x, y int
	}{
		{name: "origin", x: 0, y: 0},
		{name: "positive coords", x: 10, y: 20},
		{name: "large coords", x: 1000, y: 2000},
		{name: "negative coords", x: -10, y: -20},
		{name: "large negative coords", x: -1000, y: -2000},
		{name: "mixed coords", x: 100, y: -50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pos := &testPositional{}
			pos.SetPosition(tt.x, tt.y)

			assert.Equal(t, tt.x, pos.x, "x coordinate should match")
			assert.Equal(t, tt.y, pos.y, "y coordinate should match")
		})
	}
}

// BenchmarkPositionalSetPosition benchmarks the SetPosition method
func BenchmarkPositionalSetPosition(b *testing.B) {
	pos := &testPositional{}

	for i := 0; i < b.N; i++ {
		pos.SetPosition(i, i*2)
	}
}

// BenchmarkHelpBindings benchmarks the Bindings method
func BenchmarkHelpBindings(b *testing.B) {
	bindings := make([]key.Binding, 10)
	for i := 0; i < 10; i++ {
		bindings[i] = key.NewBinding()
	}

	help := &testHelp{bindings: bindings}

	for i := 0; i < b.N; i++ {
		_ = help.Bindings()
	}
}
