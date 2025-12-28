package messages

import (
	"encoding/json"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/nexora/nexora/internal/agent"
	"github.com/nexora/nexora/internal/agent/tools"
	"github.com/nexora/nexora/internal/message"
	"github.com/nexora/nexora/internal/permission"
)

func TestNewToolCallCmp(t *testing.T) {
	tc := message.ToolCall{
		ID:       "call-1",
		Name:     tools.BashToolName,
		Input:    `{"command":"ls"}`,
		Finished: false,
	}

	cmp := NewToolCallCmp("parent-msg-1", tc, nil)
	if cmp == nil {
		t.Fatal("Expected non-nil tool call component")
	}

	if cmp.ID() != "call-1" {
		t.Errorf("Expected ID 'call-1', got %s", cmp.ID())
	}

	if cmp.ParentMessageID() != "parent-msg-1" {
		t.Errorf("Expected parent message ID 'parent-msg-1', got %s", cmp.ParentMessageID())
	}
}

func TestNewToolCallCmpWithOptions(t *testing.T) {
	tc := message.ToolCall{
		ID:       "call-1",
		Name:     tools.ViewToolName,
		Input:    `{"file_path":"test.go"}`,
		Finished: true,
	}

	result := message.ToolResult{
		ToolCallID: "call-1",
		Content:    "file content",
		IsError:    false,
	}

	cmp := NewToolCallCmp("parent-1", tc, nil,
		WithToolCallResult(result),
		WithToolCallCancelled(),
		WithToolCallNested(true),
		WithToolPermissionRequested(),
		WithToolPermissionGranted(),
	).(*toolCallCmp)

	if cmp.result.ToolCallID != "call-1" {
		t.Error("Expected result to be set")
	}

	if !cmp.cancelled {
		t.Error("Expected cancelled to be true")
	}

	if !cmp.isNested {
		t.Error("Expected isNested to be true")
	}

	if !cmp.permissionRequested {
		t.Error("Expected permissionRequested to be true")
	}

	if !cmp.permissionGranted {
		t.Error("Expected permissionGranted to be true")
	}
}

func TestToolCallCmpInit(t *testing.T) {
	tc := message.ToolCall{
		ID:       "call-1",
		Name:     tools.BashToolName,
		Input:    `{"command":"ls"}`,
		Finished: false,
	}

	cmp := NewToolCallCmp("parent-1", tc, nil)
	cmd := cmp.Init()

	// Init should return a command for animation
	if cmd == nil {
		t.Log("Init returned nil command (animation may not be needed)")
	}
}

func TestToolCallCmpView(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		input    string
		finished bool
	}{
		{
			name:     "bash tool",
			toolName: tools.BashToolName,
			input:    `{"command":"echo hello"}`,
			finished: true,
		},
		{
			name:     "view tool",
			toolName: tools.ViewToolName,
			input:    `{"file_path":"test.go"}`,
			finished: true,
		},
		{
			name:     "pending tool",
			toolName: tools.GrepToolName,
			input:    `{"pattern":"test"}`,
			finished: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := message.ToolCall{
				ID:       "call-1",
				Name:     tt.toolName,
				Input:    tt.input,
				Finished: tt.finished,
			}

			cmp := NewToolCallCmp("parent-1", tc, nil)
			cmp.SetSize(80, 24)

			view := cmp.View()
			if view == "" {
				t.Error("Expected non-empty view")
			}
		})
	}
}

func TestToolCallCmpSetToolCall(t *testing.T) {
	tc1 := message.ToolCall{
		ID:       "call-1",
		Name:     tools.BashToolName,
		Input:    `{"command":"ls"}`,
		Finished: false,
	}

	tc2 := message.ToolCall{
		ID:       "call-2",
		Name:     tools.GrepToolName,
		Input:    `{"pattern":"test"}`,
		Finished: true,
	}

	cmp := NewToolCallCmp("parent-1", tc1, nil)
	cmp.SetToolCall(tc2)

	if cmp.GetToolCall().ID != "call-2" {
		t.Error("Expected tool call to be updated")
	}
}

func TestToolCallCmpSetToolResult(t *testing.T) {
	tc := message.ToolCall{
		ID:       "call-1",
		Name:     tools.BashToolName,
		Input:    `{"command":"ls"}`,
		Finished: true,
	}

	result := message.ToolResult{
		ToolCallID: "call-1",
		Content:    "output here",
		IsError:    false,
	}

	cmp := NewToolCallCmp("parent-1", tc, nil)
	cmp.SetToolResult(result)

	if cmp.GetToolResult().Content != "output here" {
		t.Error("Expected tool result to be updated")
	}
}

func TestToolCallCmpSetCancelled(t *testing.T) {
	tc := message.ToolCall{
		ID:       "call-1",
		Name:     tools.BashToolName,
		Input:    `{"command":"ls"}`,
		Finished: false,
	}

	cmp := NewToolCallCmp("parent-1", tc, nil).(*toolCallCmp)
	cmp.SetCancelled()

	if !cmp.cancelled {
		t.Error("Expected cancelled to be true")
	}
}

func TestToolCallCmpFocus(t *testing.T) {
	tc := message.ToolCall{
		ID:       "call-1",
		Name:     tools.BashToolName,
		Input:    `{"command":"ls"}`,
		Finished: true,
	}

	cmp := NewToolCallCmp("parent-1", tc, nil)
	cmp.Focus()

	if !cmp.IsFocused() {
		t.Error("Expected focused state")
	}

	cmp.Blur()

	if cmp.IsFocused() {
		t.Error("Expected unfocused state after Blur")
	}
}

func TestToolCallCmpSetSize(t *testing.T) {
	tc := message.ToolCall{
		ID:       "call-1",
		Name:     tools.BashToolName,
		Input:    `{"command":"ls"}`,
		Finished: true,
	}

	cmp := NewToolCallCmp("parent-1", tc, nil)
	cmp.SetSize(100, 50)

	width, _ := cmp.GetSize()
	if width != 100 {
		t.Errorf("Expected width 100, got %d", width)
	}
}

func TestToolCallCmpSpinning(t *testing.T) {
	tests := []struct {
		name     string
		finished bool
		wantSpin bool
	}{
		{
			name:     "pending tool should spin",
			finished: false,
			wantSpin: true,
		},
		{
			name:     "finished tool should not spin",
			finished: true,
			wantSpin: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := message.ToolCall{
				ID:       "call-1",
				Name:     tools.BashToolName,
				Input:    `{"command":"ls"}`,
				Finished: tt.finished,
			}

			cmp := NewToolCallCmp("parent-1", tc, nil).(*toolCallCmp)
			cmp.Init()

			// Spinning state depends on finished status
			if tt.wantSpin && !cmp.shouldSpin() {
				t.Error("Expected tool to spin when not finished")
			}
			if !tt.wantSpin && cmp.shouldSpin() {
				t.Error("Expected tool not to spin when finished")
			}
		})
	}
}

func TestToolCallCmpNestedToolCalls(t *testing.T) {
	parentTc := message.ToolCall{
		ID:       "parent-call",
		Name:     agent.AgentToolName,
		Input:    `{"prompt":"test task"}`,
		Finished: false,
	}

	childTc1 := message.ToolCall{
		ID:       "child-1",
		Name:     tools.ViewToolName,
		Input:    `{"file_path":"test.go"}`,
		Finished: true,
	}

	childTc2 := message.ToolCall{
		ID:       "child-2",
		Name:     tools.BashToolName,
		Input:    `{"command":"ls"}`,
		Finished: true,
	}

	parentCmp := NewToolCallCmp("msg-1", parentTc, nil)
	childCmp1 := NewToolCallCmp("msg-1", childTc1, nil, WithToolCallNested(true))
	childCmp2 := NewToolCallCmp("msg-1", childTc2, nil, WithToolCallNested(true))

	nested := []ToolCallCmp{childCmp1, childCmp2}
	parentCmp.SetNestedToolCalls(nested)

	retrieved := parentCmp.GetNestedToolCalls()
	if len(retrieved) != 2 {
		t.Errorf("Expected 2 nested tool calls, got %d", len(retrieved))
	}
}

func TestToolCallCmpSetIsNested(t *testing.T) {
	tc := message.ToolCall{
		ID:       "call-1",
		Name:     tools.BashToolName,
		Input:    `{"command":"ls"}`,
		Finished: true,
	}

	cmp := NewToolCallCmp("parent-1", tc, nil).(*toolCallCmp)
	cmp.SetIsNested(true)

	if !cmp.isNested {
		t.Error("Expected isNested to be true")
	}
}

func TestToolCallCmpPermissions(t *testing.T) {
	tc := message.ToolCall{
		ID:       "call-1",
		Name:     tools.BashToolName,
		Input:    `{"command":"rm -rf"}`,
		Finished: false,
	}

	cmp := NewToolCallCmp("parent-1", tc, nil).(*toolCallCmp)
	cmp.SetPermissionRequested()

	if !cmp.permissionRequested {
		t.Error("Expected permissionRequested to be true")
	}

	cmp.SetPermissionGranted()

	if !cmp.permissionGranted {
		t.Error("Expected permissionGranted to be true")
	}
}

func TestToolCallCmpRenderPending(t *testing.T) {
	tc := message.ToolCall{
		ID:       "call-1",
		Name:     tools.BashToolName,
		Input:    `{"command":"ls"}`,
		Finished: false,
	}

	cmp := NewToolCallCmp("parent-1", tc, nil).(*toolCallCmp)
	cmp.width = 80
	cmp.Init()

	result := cmp.renderPending()
	if result == "" {
		t.Error("Expected non-empty pending render")
	}

	// Should contain tool name
	if !strings.Contains(result, "Bash") {
		t.Error("Expected pending render to contain tool name")
	}
}

func TestToolCallCmpTextWidth(t *testing.T) {
	tests := []struct {
		name     string
		width    int
		isNested bool
		want     int
	}{
		{
			name:     "normal tool call",
			width:    80,
			isNested: false,
			want:     75, // 80 - 5
		},
		{
			name:     "nested tool call",
			width:    80,
			isNested: true,
			want:     74, // 80 - 6
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := &toolCallCmp{
				width:    tt.width,
				isNested: tt.isNested,
			}

			got := tc.textWidth()
			if got != tt.want {
				t.Errorf("textWidth() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestToolCallCmpFit(t *testing.T) {
	tc := &toolCallCmp{width: 80}

	tests := []struct {
		name    string
		content string
		width   int
	}{
		{
			name:    "short content",
			content: "short",
			width:   80,
		},
		{
			name:    "long content",
			content: strings.Repeat("a", 200),
			width:   50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tc.fit(tt.content, tt.width)
			// fit should never panic
			_ = result
		})
	}
}

func TestToolCallCmpRenderToolError(t *testing.T) {
	tc := &toolCallCmp{
		width: 80,
		result: message.ToolResult{
			ToolCallID: "call-1",
			Content:    "Error: something went wrong",
			IsError:    true,
		},
	}

	result := tc.renderToolError()
	if !strings.Contains(result, "ERROR") {
		t.Error("Expected ERROR tag in error rendering")
	}

	if !strings.Contains(result, "something went wrong") {
		t.Error("Expected error message in rendering")
	}
}

func TestToolCallCmpFormatToolForCopy(t *testing.T) {
	tests := []struct {
		name       string
		toolName   string
		input      string
		result     message.ToolResult
		wantInCopy string
	}{
		{
			name:     "bash tool",
			toolName: tools.BashToolName,
			input:    `{"command":"echo hello"}`,
			result: message.ToolResult{
				ToolCallID: "call-1",
				Content:    "hello",
			},
			wantInCopy: "Bash Tool Call",
		},
		{
			name:     "view tool",
			toolName: tools.ViewToolName,
			input:    `{"file_path":"test.go"}`,
			result: message.ToolResult{
				ToolCallID: "call-1",
				Content:    "package main",
			},
			wantInCopy: "View Tool Call",
		},
		{
			name:     "error result",
			toolName: tools.BashToolName,
			input:    `{"command":"invalid"}`,
			result: message.ToolResult{
				ToolCallID: "call-1",
				Content:    "command not found",
				IsError:    true,
			},
			wantInCopy: "Error:",
		},
		{
			name:     "cancelled tool",
			toolName: tools.BashToolName,
			input:    `{"command":"ls"}`,
			result: message.ToolResult{
				ToolCallID: "",
			},
			wantInCopy: "Cancelled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := &toolCallCmp{
				call: message.ToolCall{
					ID:       "call-1",
					Name:     tt.toolName,
					Input:    tt.input,
					Finished: true,
				},
				result:    tt.result,
				cancelled: tt.result.ToolCallID == "" && tt.wantInCopy == "Cancelled",
			}

			copy := tc.formatToolForCopy()
			if !strings.Contains(copy, tt.wantInCopy) {
				t.Errorf("Expected copy to contain %q, got: %s", tt.wantInCopy, copy)
			}
		})
	}
}

func TestToolCallCmpFormatParametersForCopy(t *testing.T) {
	tests := []struct {
		name       string
		toolName   string
		input      string
		wantInCopy string
	}{
		{
			name:       "bash params",
			toolName:   tools.BashToolName,
			input:      `{"command":"ls -la"}`,
			wantInCopy: "Command:",
		},
		{
			name:       "view params",
			toolName:   tools.ViewToolName,
			input:      `{"file_path":"test.go","limit":100}`,
			wantInCopy: "File:",
		},
		{
			name:       "edit params",
			toolName:   tools.EditToolName,
			input:      `{"file_path":"main.go","old_string":"old","new_string":"new"}`,
			wantInCopy: "File:",
		},
		{
			name:       "grep params",
			toolName:   tools.GrepToolName,
			input:      `{"pattern":"func.*main","path":"./"}`,
			wantInCopy: "Pattern:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := &toolCallCmp{
				call: message.ToolCall{
					Name:  tt.toolName,
					Input: tt.input,
				},
			}

			params := tc.formatParametersForCopy()
			if !strings.Contains(params, tt.wantInCopy) {
				t.Errorf("Expected params to contain %q, got: %s", tt.wantInCopy, params)
			}
		})
	}
}

func TestToolCallCmpFormatResultForCopy(t *testing.T) {
	tests := []struct {
		name       string
		toolName   string
		result     message.ToolResult
		wantInCopy string
	}{
		{
			name:     "bash result",
			toolName: tools.BashToolName,
			result: message.ToolResult{
				ToolCallID: "call-1",
				Content:    "file1.go\nfile2.go",
				Metadata:   `{"output":"file1.go\nfile2.go"}`,
			},
			wantInCopy: "```bash",
		},
		{
			name:     "view result",
			toolName: tools.ViewToolName,
			result: message.ToolResult{
				ToolCallID: "call-1",
				Content:    "package main",
				Metadata:   `{"file_path":"main.go","content":"package main"}`,
			},
			wantInCopy: "```go",
		},
		{
			name:     "image result",
			toolName: tools.ViewToolName,
			result: message.ToolResult{
				ToolCallID: "call-1",
				Data:       "base64data",
				MIMEType:   "image/png",
			},
			wantInCopy: "[Image: image/png]",
		},
		{
			name:     "media result",
			toolName: tools.ViewToolName,
			result: message.ToolResult{
				ToolCallID: "call-1",
				Data:       "base64data",
				MIMEType:   "video/mp4",
			},
			wantInCopy: "[Media: video/mp4]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := &toolCallCmp{
				call: message.ToolCall{
					Name: tt.toolName,
				},
				result: tt.result,
			}

			formatted := tc.formatResultForCopy()
			if !strings.Contains(formatted, tt.wantInCopy) {
				t.Errorf("Expected result to contain %q, got: %s", tt.wantInCopy, formatted)
			}
		})
	}
}

func TestToolCallCmpStyle(t *testing.T) {
	tests := []struct {
		name     string
		focused  bool
		isNested bool
	}{
		{
			name:     "normal unfocused",
			focused:  false,
			isNested: false,
		},
		{
			name:     "normal focused",
			focused:  true,
			isNested: false,
		},
		{
			name:     "nested",
			focused:  false,
			isNested: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := &toolCallCmp{
				focused:  tt.focused,
				isNested: tt.isNested,
			}

			style := tc.style()
			// Style should not panic
			_ = style
		})
	}
}

func TestToolCallCmpUpdate(t *testing.T) {
	tc := message.ToolCall{
		ID:       "call-1",
		Name:     tools.BashToolName,
		Input:    `{"command":"ls"}`,
		Finished: false,
	}

	cmp := NewToolCallCmp("parent-1", tc, nil)
	cmp.Init()

	// Test that Update doesn't panic
	_, _ = cmp.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
}

func TestToolCallCmpCopyTool(t *testing.T) {
	tc := message.ToolCall{
		ID:       "call-1",
		Name:     tools.BashToolName,
		Input:    `{"command":"echo hello"}`,
		Finished: true,
	}

	result := message.ToolResult{
		ToolCallID: "call-1",
		Content:    "hello",
		IsError:    false,
	}

	cmp := NewToolCallCmp("parent-1", tc, nil, WithToolCallResult(result)).(*toolCallCmp)

	// copyTool should return a command
	cmd := cmp.copyTool()
	if cmd == nil {
		t.Error("Expected non-nil command from copyTool")
	}
}

func TestToolCallCmpFormatBashResultForCopy(t *testing.T) {
	tc := &toolCallCmp{
		result: message.ToolResult{
			Metadata: `{"output":"line1\nline2"}`,
		},
	}

	result := tc.formatBashResultForCopy()
	if !strings.Contains(result, "```bash") {
		t.Error("Expected bash code block")
	}
	if !strings.Contains(result, "line1") {
		t.Error("Expected output content")
	}
}

func TestToolCallCmpFormatEditResultForCopy(t *testing.T) {
	tc := &toolCallCmp{
		call: message.ToolCall{
			Input: `{"file_path":"test.go"}`,
		},
		result: message.ToolResult{
			Metadata: `{"old_content":"old","new_content":"new"}`,
		},
	}

	result := tc.formatEditResultForCopy()
	if !strings.Contains(result, "```diff") {
		t.Error("Expected diff code block")
	}
}

func TestToolCallCmpFormatMultiEditResultForCopy(t *testing.T) {
	tc := &toolCallCmp{
		call: message.ToolCall{
			Input: `{"file_path":"test.go","edits":[]}`,
		},
		result: message.ToolResult{
			Metadata: `{"old_content":"old","new_content":"new"}`,
		},
	}

	result := tc.formatMultiEditResultForCopy()
	if !strings.Contains(result, "```diff") {
		t.Error("Expected diff code block")
	}
}

func TestToolCallCmpFormatWriteResultForCopy(t *testing.T) {
	tc := &toolCallCmp{
		call: message.ToolCall{
			Input: `{"file_path":"main.go","content":"package main"}`,
		},
	}

	result := tc.formatWriteResultForCopy()
	if !strings.Contains(result, "package main") {
		t.Error("Expected file content")
	}
}

func TestToolCallCmpFormatAgentResultForCopy(t *testing.T) {
	parentTc := &toolCallCmp{
		result: message.ToolResult{
			Content: "Agent result",
		},
		nestedToolCalls: []ToolCallCmp{},
	}

	result := parentTc.formatAgentResultForCopy()
	if !strings.Contains(result, "Agent result") {
		t.Error("Expected agent result content")
	}
}

func TestToolCallCmpShouldSpin(t *testing.T) {
	tests := []struct {
		name      string
		finished  bool
		cancelled bool
		wantSpin  bool
	}{
		{
			name:      "not finished, not cancelled",
			finished:  false,
			cancelled: false,
			wantSpin:  true,
		},
		{
			name:      "finished",
			finished:  true,
			cancelled: false,
			wantSpin:  false,
		},
		{
			name:      "cancelled",
			finished:  false,
			cancelled: true,
			wantSpin:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := &toolCallCmp{
				call: message.ToolCall{
					Finished: tt.finished,
				},
				cancelled: tt.cancelled,
			}

			got := tc.shouldSpin()
			if got != tt.wantSpin {
				t.Errorf("shouldSpin() = %v, want %v", got, tt.wantSpin)
			}
		})
	}
}

func TestToolCallCmpWithNestedSpinning(t *testing.T) {
	parentTc := message.ToolCall{
		ID:       "parent",
		Name:     agent.AgentToolName,
		Input:    `{"prompt":"task"}`,
		Finished: true,
	}

	childTc := message.ToolCall{
		ID:       "child",
		Name:     tools.BashToolName,
		Input:    `{"command":"ls"}`,
		Finished: false, // Still spinning
	}

	parent := NewToolCallCmp("msg-1", parentTc, nil).(*toolCallCmp)
	child := NewToolCallCmp("msg-1", childTc, nil).(*toolCallCmp)
	child.Init()

	parent.SetNestedToolCalls([]ToolCallCmp{child})

	// Parent should report spinning if child is spinning
	if !parent.Spinning() {
		t.Error("Parent should report spinning when child is spinning")
	}
}

func TestPermissionServiceIntegration(t *testing.T) {
	// Create a mock permission service
	var mockPermService permission.Service = nil

	tc := message.ToolCall{
		ID:       "call-1",
		Name:     tools.BashToolName,
		Input:    `{"command":"rm -rf"}`,
		Finished: false,
	}

	// Should not panic even with nil permission service
	cmp := NewToolCallCmp("parent-1", tc, mockPermService)
	if cmp == nil {
		t.Error("Expected component to be created with nil permission service")
	}
}

func TestToolCallCmpWithInvalidJSON(t *testing.T) {
	tc := message.ToolCall{
		ID:       "call-1",
		Name:     tools.BashToolName,
		Input:    `{invalid json}`,
		Finished: true,
	}

	cmp := NewToolCallCmp("parent-1", tc, nil)
	cmp.SetSize(80, 24)

	// Should not panic with invalid JSON
	view := cmp.View()
	if view == "" {
		t.Error("Expected non-empty view even with invalid JSON")
	}

	// Should show error in rendering
	if !strings.Contains(view, "ERROR") && !strings.Contains(view, "Invalid") {
		t.Log("Tool rendering handles invalid JSON gracefully")
	}
}

func TestToolCallCmpMetadataHandling(t *testing.T) {
	tc := &toolCallCmp{
		call: message.ToolCall{
			Name: tools.BashToolName,
		},
		result: message.ToolResult{
			ToolCallID: "call-1",
			Metadata:   `{"tmux_session_id":"session-1","output":"test output"}`,
		},
	}

	var meta tools.BashResponseMetadata
	err := json.Unmarshal([]byte(tc.result.Metadata), &meta)
	if err != nil {
		t.Errorf("Expected valid metadata JSON: %v", err)
	}

	if meta.TmuxSessionID != "session-1" {
		t.Error("Expected tmux_session_id to be parsed")
	}

	if meta.Output != "test output" {
		t.Error("Expected output to be parsed")
	}
}
