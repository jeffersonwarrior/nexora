package status

import (
	"testing"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/nexora/nexora/internal/tui/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// dummyKeyMap implements help.KeyMap for testing
type dummyKeyMap struct{}

func (d *dummyKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{}
}

func (d *dummyKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{}
}

// TestNewStatusCmp tests the NewStatusCmp constructor
func TestNewStatusCmp(t *testing.T) {
	t.Parallel()

	cmp := NewStatusCmp()
	require.NotNil(t, cmp, "NewStatusCmp should return non-nil component")
	cmd := cmp.Init()
	assert.Nil(t, cmd, "Init should return nil cmd")
}

// TestStatusCmpInit tests the Init method
func TestStatusCmpInit(t *testing.T) {
	t.Parallel()

	cmp := NewStatusCmp()
	cmd := cmp.Init()
	assert.Nil(t, cmd, "Init should return nil cmd")
}

// TestStatusCmpViewInitial tests the View method with initial state
func TestStatusCmpViewInitial(t *testing.T) {
	t.Parallel()

	cmp := NewStatusCmp()
	cmp.SetKeyMap(&dummyKeyMap{})
	view := cmp.View()
	assert.NotEmpty(t, view, "View should return non-empty string")
}

// TestStatusCmpWindowSizeMsg tests Update with WindowSizeMsg
func TestStatusCmpWindowSizeMsg(t *testing.T) {
	t.Parallel()

	cmp := NewStatusCmp()
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}

	model, cmd := cmp.Update(msg)
	assert.NotNil(t, model, "Update should return non-nil model")
	assert.Nil(t, cmd, "Update should return nil cmd for WindowSizeMsg")
}

// TestStatusCmpInfoMsg tests Update with InfoMsg
func TestStatusCmpInfoMsg(t *testing.T) {
	t.Parallel()

	cmp := NewStatusCmp()
	infoMsg := util.InfoMsg{
		Type: util.InfoTypeSuccess,
		Msg:  "Operation successful",
	}

	model, cmd := cmp.Update(infoMsg)
	assert.NotNil(t, model, "Update should return non-nil model")
	assert.NotNil(t, cmd, "Update should return non-nil cmd for InfoMsg")

	// Execute the command to ensure it works
	if cmd != nil {
		msg := cmd()
		assert.NotNil(t, msg, "Clear message command should produce a message")
	}
}

// TestStatusCmpClearStatusMsg tests Update with ClearStatusMsg
func TestStatusCmpClearStatusMsg(t *testing.T) {
	t.Parallel()

	cmp := NewStatusCmp()

	// First set an info message
	infoMsg := util.InfoMsg{Type: util.InfoTypeInfo, Msg: "Test"}
	cmp.Update(infoMsg)

	// Then clear it
	clearMsg := util.ClearStatusMsg{}
	model, cmd := cmp.Update(clearMsg)

	assert.NotNil(t, model, "Update should return non-nil model")
	assert.Nil(t, cmd, "Update should return nil cmd for ClearStatusMsg")
}

// TestStatusCmpToggleFullHelp tests the ToggleFullHelp method
func TestStatusCmpToggleFullHelp(t *testing.T) {
	t.Parallel()

	cmp := NewStatusCmp()
	cmp.ToggleFullHelp()
	// Method should not panic
	cmp.ToggleFullHelp()
	// Method should not panic
}

// TestStatusCmpSetKeyMap tests the SetKeyMap method
func TestStatusCmpSetKeyMap(t *testing.T) {
	t.Parallel()

	cmp := NewStatusCmp()
	cmp.SetKeyMap(&dummyKeyMap{})
	// Method should not panic
}

// TestStatusCmpInfoTypeError tests error info message rendering
func TestStatusCmpInfoTypeError(t *testing.T) {
	t.Parallel()

	cmp := NewStatusCmp()
	cmp.SetKeyMap(&dummyKeyMap{})
	cmp.Update(tea.WindowSizeMsg{Width: 80, Height: 40})

	infoMsg := util.InfoMsg{
		Type: util.InfoTypeError,
		Msg:  "An error occurred",
	}
	cmp.Update(infoMsg)

	view := cmp.View()
	assert.NotEmpty(t, view, "View should show error message")
}

// TestStatusCmpInfoTypeWarn tests warning info message rendering
func TestStatusCmpInfoTypeWarn(t *testing.T) {
	t.Parallel()

	cmp := NewStatusCmp()
	cmp.SetKeyMap(&dummyKeyMap{})
	cmp.Update(tea.WindowSizeMsg{Width: 80, Height: 40})

	infoMsg := util.InfoMsg{
		Type: util.InfoTypeWarn,
		Msg:  "Warning message",
	}
	cmp.Update(infoMsg)

	view := cmp.View()
	assert.NotEmpty(t, view, "View should show warning message")
}

// TestStatusCmpInfoTypeUpdate tests update info message rendering
func TestStatusCmpInfoTypeUpdate(t *testing.T) {
	t.Parallel()

	cmp := NewStatusCmp()
	cmp.SetKeyMap(&dummyKeyMap{})
	cmp.Update(tea.WindowSizeMsg{Width: 80, Height: 40})

	infoMsg := util.InfoMsg{
		Type: util.InfoTypeUpdate,
		Msg:  "Update available",
	}
	cmp.Update(infoMsg)

	view := cmp.View()
	assert.NotEmpty(t, view, "View should show update message")
}

// TestStatusCmpInfoTypeSuccess tests success info message rendering
func TestStatusCmpInfoTypeSuccess(t *testing.T) {
	t.Parallel()

	cmp := NewStatusCmp()
	cmp.SetKeyMap(&dummyKeyMap{})
	cmp.Update(tea.WindowSizeMsg{Width: 80, Height: 40})

	infoMsg := util.InfoMsg{
		Type: util.InfoTypeSuccess,
		Msg:  "Success!",
	}
	cmp.Update(infoMsg)

	view := cmp.View()
	assert.NotEmpty(t, view, "View should show success message")
}

// TestStatusCmpCustomTTL tests custom TTL in InfoMsg
func TestStatusCmpCustomTTL(t *testing.T) {
	t.Parallel()

	cmp := NewStatusCmp()

	infoMsg := util.InfoMsg{
		Type: util.InfoTypeInfo,
		Msg:  "Test",
		TTL:  2 * time.Second,
	}

	model, cmd := cmp.Update(infoMsg)
	assert.NotNil(t, model, "Update should return non-nil model")
	assert.NotNil(t, cmd, "Update should return non-nil cmd")
}

// TestStatusCmpNarrowWidth tests rendering with very narrow width
func TestStatusCmpNarrowWidth(t *testing.T) {
	t.Parallel()

	cmp := NewStatusCmp()
	cmp.SetKeyMap(&dummyKeyMap{})
	cmp.Update(tea.WindowSizeMsg{Width: 30, Height: 10})

	infoMsg := util.InfoMsg{
		Type: util.InfoTypeInfo,
		Msg:  "This is a very long message that should be truncated for narrow widths",
	}
	cmp.Update(infoMsg)

	view := cmp.View()
	assert.NotEmpty(t, view, "View should handle narrow widths")
}

// TestStatusCmpWideWidth tests rendering with wide width
func TestStatusCmpWideWidth(t *testing.T) {
	t.Parallel()

	cmp := NewStatusCmp()
	cmp.SetKeyMap(&dummyKeyMap{})
	cmp.Update(tea.WindowSizeMsg{Width: 200, Height: 40})

	infoMsg := util.InfoMsg{
		Type: util.InfoTypeSuccess,
		Msg:  "This message should display nicely in a wide terminal",
	}
	cmp.Update(infoMsg)

	view := cmp.View()
	assert.NotEmpty(t, view, "View should handle wide widths")
}

// TestStatusCmpMultipleUpdates tests multiple sequential updates
func TestStatusCmpMultipleUpdates(t *testing.T) {
	t.Parallel()

	cmp := NewStatusCmp()
	cmp.SetKeyMap(&dummyKeyMap{})
	cmp.Update(tea.WindowSizeMsg{Width: 80, Height: 40})

	// First info message
	info1 := util.InfoMsg{Type: util.InfoTypeInfo, Msg: "First"}
	cmp.Update(info1)

	// Second info message (overwrites)
	info2 := util.InfoMsg{Type: util.InfoTypeSuccess, Msg: "Second"}
	cmp.Update(info2)

	view := cmp.View()
	assert.NotEmpty(t, view, "View should display latest message")
}

// TestStatusCmpEmptyMessage tests empty message handling
func TestStatusCmpEmptyMessage(t *testing.T) {
	t.Parallel()

	cmp := NewStatusCmp()
	cmp.SetKeyMap(&dummyKeyMap{})
	cmp.Update(tea.WindowSizeMsg{Width: 80, Height: 40})

	infoMsg := util.InfoMsg{
		Type: util.InfoTypeInfo,
		Msg:  "",
	}
	cmp.Update(infoMsg)

	view := cmp.View()
	assert.NotEmpty(t, view, "View should handle empty message")
}

// TestStatusCmpInterface verifies StatusCmp implements StatusCmp interface
func TestStatusCmpInterface(t *testing.T) {
	t.Parallel()

	cmp := NewStatusCmp()
	var _ StatusCmp = cmp
}
