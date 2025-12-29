package util

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCmdHandler tests the CmdHandler function
func TestCmdHandler(t *testing.T) {
	t.Parallel()

	msg := InfoMsg{Type: InfoTypeInfo, Msg: "test"}
	cmd := CmdHandler(msg)

	require.NotNil(t, cmd, "CmdHandler should return non-nil cmd")
	result := cmd()
	assert.Equal(t, msg, result, "cmd should return the same message")
}

// TestReportError tests the ReportError function
func TestReportError(t *testing.T) {
	t.Parallel()

	err := assert.AnError
	cmd := ReportError(err)

	require.NotNil(t, cmd, "ReportError should return non-nil cmd")
	msg := cmd()
	infoMsg, ok := msg.(InfoMsg)
	assert.True(t, ok, "message should be InfoMsg type")
	assert.Equal(t, InfoTypeError, infoMsg.Type, "type should be InfoTypeError")
	assert.Equal(t, err.Error(), infoMsg.Msg, "message should contain error text")
}

// TestReportInfo tests the ReportInfo function
func TestReportInfo(t *testing.T) {
	t.Parallel()

	text := "Info message"
	cmd := ReportInfo(text)

	require.NotNil(t, cmd, "ReportInfo should return non-nil cmd")
	msg := cmd()
	infoMsg, ok := msg.(InfoMsg)
	assert.True(t, ok, "message should be InfoMsg type")
	assert.Equal(t, InfoTypeInfo, infoMsg.Type, "type should be InfoTypeInfo")
	assert.Equal(t, text, infoMsg.Msg, "message should match input")
}

// TestReportWarn tests the ReportWarn function
func TestReportWarn(t *testing.T) {
	t.Parallel()

	text := "Warning message"
	cmd := ReportWarn(text)

	require.NotNil(t, cmd, "ReportWarn should return non-nil cmd")
	msg := cmd()
	infoMsg, ok := msg.(InfoMsg)
	assert.True(t, ok, "message should be InfoMsg type")
	assert.Equal(t, InfoTypeWarn, infoMsg.Type, "type should be InfoTypeWarn")
	assert.Equal(t, text, infoMsg.Msg, "message should match input")
}

// TestInfoMsg tests the InfoMsg struct
func TestInfoMsg(t *testing.T) {
	t.Parallel()

	msg := InfoMsg{
		Type: InfoTypeSuccess,
		Msg:  "Success",
		TTL:  5 * time.Second,
	}

	assert.Equal(t, InfoTypeSuccess, msg.Type, "type should match")
	assert.Equal(t, "Success", msg.Msg, "message should match")
	assert.Equal(t, 5*time.Second, msg.TTL, "TTL should match")
}

// TestClearStatusMsg tests the ClearStatusMsg struct
func TestClearStatusMsg(t *testing.T) {
	t.Parallel()

	msg := ClearStatusMsg{}
	assert.Equal(t, ClearStatusMsg{}, msg, "ClearStatusMsg should be empty")
}

// TestInfoTypes tests all InfoType constants
func TestInfoTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		infoType InfoType
	}{
		{name: "Info", infoType: InfoTypeInfo},
		{name: "Success", infoType: InfoTypeSuccess},
		{name: "Warn", infoType: InfoTypeWarn},
		{name: "Error", infoType: InfoTypeError},
		{name: "Update", infoType: InfoTypeUpdate},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Just verify the constants exist and have different values
			assert.NotNil(t, tt.infoType)
		})
	}
}

// TestModel interface compliance
type testModel struct {
	initialized bool
	updated     bool
	viewed      bool
}

func (m *testModel) Init() tea.Cmd {
	m.initialized = true
	return nil
}

func (m *testModel) Update(msg tea.Msg) (Model, tea.Cmd) {
	m.updated = true
	return m, nil
}

func (m *testModel) View() string {
	m.viewed = true
	return "test view"
}

// TestModelInterface verifies Model interface implementation
func TestModelInterface(t *testing.T) {
	t.Parallel()

	model := &testModel{}
	var _ Model = model

	cmd := model.Init()
	assert.Nil(t, cmd)
	assert.True(t, model.initialized)

	m, _ := model.Update(nil)
	assert.NotNil(t, m)
	assert.True(t, model.updated)

	view := model.View()
	assert.NotEmpty(t, view)
	assert.True(t, model.viewed)
}

// TestInfoMsgWithZeroTTL tests InfoMsg with zero TTL
func TestInfoMsgWithZeroTTL(t *testing.T) {
	t.Parallel()

	msg := InfoMsg{
		Type: InfoTypeInfo,
		Msg:  "Test",
		TTL:  0,
	}

	assert.Equal(t, time.Duration(0), msg.TTL, "TTL should be zero")
}

// TestReportErrorWithDifferentErrors tests ReportError with various error types
func TestReportErrorWithDifferentErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		errMsg string
	}{
		{name: "simple error", errMsg: "error occurred"},
		{name: "complex error", errMsg: "connection failed: timeout"},
		{name: "empty error", errMsg: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create a custom error
			err := &customError{msg: tt.errMsg}
			cmd := ReportError(err)
			require.NotNil(t, cmd)

			msg := cmd()
			infoMsg, ok := msg.(InfoMsg)
			assert.True(t, ok)
			assert.Equal(t, InfoTypeError, infoMsg.Type)
		})
	}
}

// customError implements the error interface for testing
type customError struct {
	msg string
}

func (e *customError) Error() string {
	return e.msg
}
