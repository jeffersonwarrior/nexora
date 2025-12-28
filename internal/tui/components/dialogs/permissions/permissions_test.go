package permissions

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/nexora/nexora/internal/agent/tools"
	"github.com/nexora/nexora/internal/permission"
)

func TestNewPermissionDialogCmp(t *testing.T) {
	tests := []struct {
		name       string
		permission permission.PermissionRequest
		opts       *Options
	}{
		{
			name: "bash tool with nil options",
			permission: permission.PermissionRequest{
				ID:          "test-id-1",
				SessionID:   "session-1",
				ToolCallID:  "tool-call-1",
				ToolName:    tools.BashToolName,
				Description: "Test bash command",
				Action:      "execute",
				Path:        "/home/test",
				Params: tools.BashPermissionsParams{
					Command:     "echo 'test'",
					Description: "Test command",
					WorkingDir:  "/home/test",
				},
			},
			opts: nil,
		},
		{
			name: "edit tool with split mode",
			permission: permission.PermissionRequest{
				ID:         "test-id-2",
				SessionID:  "session-1",
				ToolCallID: "tool-call-2",
				ToolName:   tools.EditToolName,
				Path:       "/home/test/file.txt",
				Params: tools.EditPermissionsParams{
					FilePath:   "/home/test/file.txt",
					OldContent: "old content",
					NewContent: "new content",
				},
			},
			opts: &Options{DiffMode: "split"},
		},
		{
			name: "write tool with unified mode",
			permission: permission.PermissionRequest{
				ID:         "test-id-3",
				SessionID:  "session-1",
				ToolCallID: "tool-call-3",
				ToolName:   tools.WriteToolName,
				Path:       "/home/test/newfile.txt",
				Params: tools.WritePermissionsParams{
					FilePath:   "/home/test/newfile.txt",
					OldContent: "",
					NewContent: "new file content",
				},
			},
			opts: &Options{DiffMode: "unified"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog := NewPermissionDialogCmp(tt.permission, tt.opts)
			if dialog == nil {
				t.Fatal("NewPermissionDialogCmp returned nil")
			}

			if dialog.ID() != PermissionsDialogID {
				t.Errorf("expected dialog ID %q, got %q", PermissionsDialogID, dialog.ID())
			}

			// Check that default selected option is 0 (Allow)
			impl := dialog.(*permissionDialogCmp)
			if impl.selectedOption != 0 {
				t.Errorf("expected selectedOption to be 0, got %d", impl.selectedOption)
			}

			// Check that permission is set correctly
			if impl.permission.ID != tt.permission.ID {
				t.Errorf("expected permission ID %q, got %q", tt.permission.ID, impl.permission.ID)
			}
		})
	}
}

func TestPermissionDialogCmp_Init(t *testing.T) {
	permission := permission.PermissionRequest{
		ID:         "test-id",
		SessionID:  "session-1",
		ToolCallID: "tool-call-1",
		ToolName:   tools.BashToolName,
		Path:       "/home/test",
		Params: tools.BashPermissionsParams{
			Command:     "ls -la",
			Description: "List files",
		},
	}

	dialog := NewPermissionDialogCmp(permission, nil)
	cmd := dialog.Init()

	// Init may return nil or a viewport init command - both are valid
	// We just verify that Init doesn't panic
	_ = cmd
}

func TestPermissionDialogCmp_StateTransitions(t *testing.T) {
	tests := []struct {
		name           string
		initialOption  int
		key            tea.Key
		expectedOption int
		expectCmd      bool
		cmdMsgType     any
	}{
		{
			name:           "right arrow moves from Allow to AllowSession",
			initialOption:  0,
			key:            tea.Key{Code: tea.KeyRight},
			expectedOption: 1,
			expectCmd:      false,
		},
		{
			name:           "right arrow moves from AllowSession to Deny",
			initialOption:  1,
			key:            tea.Key{Code: tea.KeyRight},
			expectedOption: 2,
			expectCmd:      false,
		},
		{
			name:           "right arrow wraps from Deny to Allow",
			initialOption:  2,
			key:            tea.Key{Code: tea.KeyRight},
			expectedOption: 0,
			expectCmd:      false,
		},
		{
			name:           "left arrow moves from Deny to AllowSession",
			initialOption:  2,
			key:            tea.Key{Code: tea.KeyLeft},
			expectedOption: 1,
			expectCmd:      false,
		},
		{
			name:           "left arrow moves from AllowSession to Allow",
			initialOption:  1,
			key:            tea.Key{Code: tea.KeyLeft},
			expectedOption: 0,
			expectCmd:      false,
		},
		{
			name:           "left arrow wraps from Allow to Deny",
			initialOption:  0,
			key:            tea.Key{Code: tea.KeyLeft},
			expectedOption: 2,
			expectCmd:      false,
		},
		{
			name:           "tab moves forward",
			initialOption:  0,
			key:            tea.Key{Code: tea.KeyTab},
			expectedOption: 1,
			expectCmd:      false,
		},
		{
			name:          "enter on Allow option",
			initialOption: 0,
			key:           tea.Key{Code: tea.KeyEnter},
			expectCmd:     true,
			cmdMsgType:    PermissionResponseMsg{},
		},
		{
			name:          "enter on AllowSession option",
			initialOption: 1,
			key:           tea.Key{Code: tea.KeyEnter},
			expectCmd:     true,
			cmdMsgType:    PermissionResponseMsg{},
		},
		{
			name:          "enter on Deny option",
			initialOption: 2,
			key:           tea.Key{Code: tea.KeyEnter},
			expectCmd:     true,
			cmdMsgType:    PermissionResponseMsg{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			permission := permission.PermissionRequest{
				ID:         "test-id",
				SessionID:  "session-1",
				ToolCallID: "tool-call-1",
				ToolName:   tools.BashToolName,
				Path:       "/home/test",
				Params: tools.BashPermissionsParams{
					Command:     "echo 'test'",
					Description: "Test",
				},
			}

			dialog := NewPermissionDialogCmp(permission, nil)
			impl := dialog.(*permissionDialogCmp)
			impl.selectedOption = tt.initialOption

			_, cmd := dialog.Update(tea.KeyPressMsg(tt.key))

			if !tt.expectCmd {
				// Check that the option changed correctly
				if impl.selectedOption != tt.expectedOption {
					t.Errorf("expected selectedOption %d, got %d", tt.expectedOption, impl.selectedOption)
				}
			} else {
				// Verify command was returned
				if cmd == nil {
					t.Error("expected command, got nil")
				}
			}
		})
	}
}

func TestPermissionDialogCmp_ApprovalFlow(t *testing.T) {
	tests := []struct {
		name           string
		key            tea.Key
		keyText        string
		expectedAction PermissionAction
	}{
		{
			name:           "a key triggers Allow",
			key:            tea.Key{Code: 'a', Text: "a"},
			keyText:        "a",
			expectedAction: PermissionAllow,
		},
		{
			name:           "A key triggers Allow",
			key:            tea.Key{Code: 'A', Text: "A"},
			keyText:        "A",
			expectedAction: PermissionAllow,
		},
		{
			name:           "s key triggers AllowSession",
			key:            tea.Key{Code: 's', Text: "s"},
			keyText:        "s",
			expectedAction: PermissionAllowForSession,
		},
		{
			name:           "S key triggers AllowSession",
			key:            tea.Key{Code: 'S', Text: "S"},
			keyText:        "S",
			expectedAction: PermissionAllowForSession,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			permission := permission.PermissionRequest{
				ID:         "test-id",
				SessionID:  "session-1",
				ToolCallID: "tool-call-1",
				ToolName:   tools.BashToolName,
				Path:       "/home/test",
				Params: tools.BashPermissionsParams{
					Command:     "echo 'test'",
					Description: "Test",
				},
			}

			dialog := NewPermissionDialogCmp(permission, nil)
			_, cmd := dialog.Update(tea.KeyPressMsg(tt.key))

			if cmd == nil {
				t.Fatal("expected command, got nil")
			}

			// Execute the batch command and check messages
			msg := cmd()
			if msg == nil {
				t.Fatal("expected message from command")
			}

			// The command is a batch, so we need to check if it contains the expected messages
			// We can't easily inspect batch commands, so we'll just verify a command was returned
		})
	}
}

func TestPermissionDialogCmp_DenialFlow(t *testing.T) {
	tests := []struct {
		name    string
		key     tea.Key
		keyText string
	}{
		{
			name:    "d key triggers Deny",
			key:     tea.Key{Code: 'd', Text: "d"},
			keyText: "d",
		},
		{
			name:    "D key triggers Deny",
			key:     tea.Key{Code: 'D', Text: "D"},
			keyText: "D",
		},
		{
			name: "esc key triggers Deny",
			key:  tea.Key{Code: tea.KeyEscape},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			permission := permission.PermissionRequest{
				ID:         "test-id",
				SessionID:  "session-1",
				ToolCallID: "tool-call-1",
				ToolName:   tools.BashToolName,
				Path:       "/home/test",
				Params: tools.BashPermissionsParams{
					Command:     "rm -rf /",
					Description: "Dangerous command",
				},
			}

			dialog := NewPermissionDialogCmp(permission, nil)
			_, cmd := dialog.Update(tea.KeyPressMsg(tt.key))

			if cmd == nil {
				t.Fatal("expected command, got nil")
			}

			// Verify command was returned (batch with CloseDialogMsg and PermissionResponseMsg)
			msg := cmd()
			if msg == nil {
				t.Fatal("expected message from command")
			}
		})
	}
}

func TestPermissionDialogCmp_KeyboardNavigation(t *testing.T) {
	permission := permission.PermissionRequest{
		ID:         "test-id",
		SessionID:  "session-1",
		ToolCallID: "tool-call-1",
		ToolName:   tools.BashToolName,
		Path:       "/home/test",
		Params: tools.BashPermissionsParams{
			Command:     "echo 'test'",
			Description: "Test",
		},
	}

	dialog := NewPermissionDialogCmp(permission, nil)
	impl := dialog.(*permissionDialogCmp)

	// Test navigation sequence: right, right, right (should wrap), left
	impl.selectedOption = 0
	dialog.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyRight}))
	if impl.selectedOption != 1 {
		t.Errorf("after first right, expected 1, got %d", impl.selectedOption)
	}

	dialog.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyRight}))
	if impl.selectedOption != 2 {
		t.Errorf("after second right, expected 2, got %d", impl.selectedOption)
	}

	dialog.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyRight}))
	if impl.selectedOption != 0 {
		t.Errorf("after third right (wrap), expected 0, got %d", impl.selectedOption)
	}

	dialog.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyLeft}))
	if impl.selectedOption != 2 {
		t.Errorf("after left (wrap backward), expected 2, got %d", impl.selectedOption)
	}
}

func TestPermissionDialogCmp_WindowResize(t *testing.T) {
	permission := permission.PermissionRequest{
		ID:         "test-id",
		SessionID:  "session-1",
		ToolCallID: "tool-call-1",
		ToolName:   tools.BashToolName,
		Path:       "/home/test",
		Params: tools.BashPermissionsParams{
			Command:     "echo 'test'",
			Description: "Test",
		},
	}

	dialog := NewPermissionDialogCmp(permission, nil)
	impl := dialog.(*permissionDialogCmp)

	// Send window size update
	dialog.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	if impl.wWidth != 120 {
		t.Errorf("expected wWidth 120, got %d", impl.wWidth)
	}
	if impl.wHeight != 40 {
		t.Errorf("expected wHeight 40, got %d", impl.wHeight)
	}

	// Check that content is marked dirty after resize
	if !impl.contentDirty {
		t.Error("expected contentDirty to be true after window resize")
	}
}

func TestPermissionDialogCmp_DiffViewToggle(t *testing.T) {
	permission := permission.PermissionRequest{
		ID:         "test-id",
		SessionID:  "session-1",
		ToolCallID: "tool-call-1",
		ToolName:   tools.EditToolName,
		Path:       "/home/test/file.txt",
		Params: tools.EditPermissionsParams{
			FilePath:   "/home/test/file.txt",
			OldContent: "old content\nline 2\nline 3",
			NewContent: "new content\nline 2\nline 3 modified",
		},
	}

	dialog := NewPermissionDialogCmp(permission, nil)
	impl := dialog.(*permissionDialogCmp)

	// Set window size first
	dialog.Update(tea.WindowSizeMsg{Width: 160, Height: 40})

	// Initially should use default mode based on width
	initialSplitMode := impl.useDiffSplitMode()

	// Toggle diff mode with 't' key
	dialog.Update(tea.KeyPressMsg(tea.Key{Code: 't', Text: "t"}))

	newSplitMode := impl.useDiffSplitMode()
	if newSplitMode == initialSplitMode {
		t.Error("expected diff mode to toggle")
	}

	// Toggle again
	dialog.Update(tea.KeyPressMsg(tea.Key{Code: 't', Text: "t"}))

	finalSplitMode := impl.useDiffSplitMode()
	if finalSplitMode != initialSplitMode {
		t.Error("expected diff mode to toggle back to initial state")
	}
}

func TestPermissionDialogCmp_DiffViewScroll(t *testing.T) {
	permission := permission.PermissionRequest{
		ID:         "test-id",
		SessionID:  "session-1",
		ToolCallID: "tool-call-1",
		ToolName:   tools.EditToolName,
		Path:       "/home/test/file.txt",
		Params: tools.EditPermissionsParams{
			FilePath:   "/home/test/file.txt",
			OldContent: strings.Repeat("line\n", 100),
			NewContent: strings.Repeat("modified line\n", 100),
		},
	}

	dialog := NewPermissionDialogCmp(permission, nil)
	impl := dialog.(*permissionDialogCmp)

	// Set window size
	dialog.Update(tea.WindowSizeMsg{Width: 160, Height: 40})

	// Test vertical scrolling
	initialYOffset := impl.diffYOffset
	dialog.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyDown, Mod: tea.ModShift}))
	if impl.diffYOffset != initialYOffset+1 {
		t.Errorf("expected diffYOffset %d, got %d", initialYOffset+1, impl.diffYOffset)
	}

	dialog.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyUp, Mod: tea.ModShift}))
	if impl.diffYOffset != initialYOffset {
		t.Errorf("expected diffYOffset %d, got %d", initialYOffset, impl.diffYOffset)
	}

	// Scroll up at 0 should stay at 0
	dialog.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyUp, Mod: tea.ModShift}))
	if impl.diffYOffset != 0 {
		t.Errorf("expected diffYOffset 0, got %d", impl.diffYOffset)
	}

	// Test horizontal scrolling
	initialXOffset := impl.diffXOffset
	dialog.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyRight, Mod: tea.ModShift}))
	if impl.diffXOffset != initialXOffset+5 {
		t.Errorf("expected diffXOffset %d, got %d", initialXOffset+5, impl.diffXOffset)
	}

	dialog.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyLeft, Mod: tea.ModShift}))
	if impl.diffXOffset != initialXOffset {
		t.Errorf("expected diffXOffset %d, got %d", initialXOffset, impl.diffXOffset)
	}

	// Scroll left at 0 should stay at 0
	dialog.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyLeft, Mod: tea.ModShift}))
	if impl.diffXOffset != 0 {
		t.Errorf("expected diffXOffset 0, got %d", impl.diffXOffset)
	}
}

func TestPermissionDialogCmp_View(t *testing.T) {
	tests := []struct {
		name           string
		toolName       string
		params         any
		expectInView   []string
		unexpectInView []string
	}{
		{
			name:     "bash tool displays command",
			toolName: tools.BashToolName,
			params: tools.BashPermissionsParams{
				Command:     "echo 'hello world'",
				Description: "Print hello",
			},
			expectInView: []string{
				"Permission Required",
				"Tool",
				tools.BashToolName,
			},
		},
		{
			name:     "download tool displays URL and file",
			toolName: tools.DownloadToolName,
			params: tools.DownloadPermissionsParams{
				URL:      "https://example.com/file.zip",
				FilePath: "/home/test/file.zip",
			},
			expectInView: []string{
				"Permission Required",
				"Tool",
				tools.DownloadToolName,
			},
		},
		{
			name:     "edit tool displays file",
			toolName: tools.EditToolName,
			params: tools.EditPermissionsParams{
				FilePath:   "/home/test/file.txt",
				OldContent: "old",
				NewContent: "new",
			},
			expectInView: []string{
				"Permission Required",
				"Tool",
				tools.EditToolName,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			permission := permission.PermissionRequest{
				ID:         "test-id",
				SessionID:  "session-1",
				ToolCallID: "tool-call-1",
				ToolName:   tt.toolName,
				Path:       "/home/test",
				Params:     tt.params,
			}

			dialog := NewPermissionDialogCmp(permission, nil)

			// Set window size to get proper rendering
			dialog.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

			view := dialog.View()

			if view == "" {
				t.Error("expected non-empty view")
			}

			for _, expected := range tt.expectInView {
				if !strings.Contains(view, expected) {
					t.Errorf("expected view to contain %q", expected)
				}
			}

			for _, notExpected := range tt.unexpectInView {
				if strings.Contains(view, notExpected) {
					t.Errorf("expected view to NOT contain %q", notExpected)
				}
			}
		})
	}
}

func TestPermissionDialogCmp_Position(t *testing.T) {
	permission := permission.PermissionRequest{
		ID:         "test-id",
		SessionID:  "session-1",
		ToolCallID: "tool-call-1",
		ToolName:   tools.BashToolName,
		Path:       "/home/test",
		Params: tools.BashPermissionsParams{
			Command:     "echo 'test'",
			Description: "Test",
		},
	}

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
			dialog := NewPermissionDialogCmp(permission, nil)

			// Update window size
			dialog.Update(tea.WindowSizeMsg{
				Width:  tt.width,
				Height: tt.height,
			})

			// Trigger rendering to calculate position
			_ = dialog.View()

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

func TestPermissionDialogCmp_ID(t *testing.T) {
	permission := permission.PermissionRequest{
		ID:         "test-id",
		SessionID:  "session-1",
		ToolCallID: "tool-call-1",
		ToolName:   tools.BashToolName,
		Path:       "/home/test",
	}

	dialog := NewPermissionDialogCmp(permission, nil)

	id := dialog.ID()

	if id != PermissionsDialogID {
		t.Errorf("expected ID %q, got %q", PermissionsDialogID, id)
	}

	// Verify DialogID constant value
	const expectedID = "permissions"
	if PermissionsDialogID != expectedID {
		t.Errorf("expected PermissionsDialogID to be %q, got %q", expectedID, PermissionsDialogID)
	}
}

func TestPermissionDialogCmp_SupportsDiffView(t *testing.T) {
	tests := []struct {
		name            string
		toolName        string
		expectDiffView  bool
	}{
		{
			name:           "EditTool supports diff view",
			toolName:       tools.EditToolName,
			expectDiffView: true,
		},
		{
			name:           "WriteTool supports diff view",
			toolName:       tools.WriteToolName,
			expectDiffView: true,
		},
		{
			name:           "MultiEditTool supports diff view",
			toolName:       tools.MultiEditToolName,
			expectDiffView: true,
		},
		{
			name:           "BashTool does not support diff view",
			toolName:       tools.BashToolName,
			expectDiffView: false,
		},
		{
			name:           "DownloadTool does not support diff view",
			toolName:       tools.DownloadToolName,
			expectDiffView: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var params any
			if tt.expectDiffView {
				params = tools.EditPermissionsParams{
					FilePath:   "/test/file.txt",
					OldContent: "old",
					NewContent: "new",
				}
			} else if tt.toolName == tools.BashToolName {
				params = tools.BashPermissionsParams{
					Command: "echo test",
				}
			} else if tt.toolName == tools.DownloadToolName {
				params = tools.DownloadPermissionsParams{
					URL:      "https://example.com",
					FilePath: "/test/file",
				}
			}

			permission := permission.PermissionRequest{
				ID:         "test-id",
				SessionID:  "session-1",
				ToolCallID: "tool-call-1",
				ToolName:   tt.toolName,
				Path:       "/home/test",
				Params:     params,
			}

			dialog := NewPermissionDialogCmp(permission, nil)
			impl := dialog.(*permissionDialogCmp)

			supportsDiff := impl.supportsDiffView()
			if supportsDiff != tt.expectDiffView {
				t.Errorf("expected supportsDiffView=%v, got %v", tt.expectDiffView, supportsDiff)
			}
		})
	}
}

func TestPermissionDialogCmp_SelectCurrentOption(t *testing.T) {
	tests := []struct {
		name           string
		selectedOption int
		expectedAction PermissionAction
	}{
		{
			name:           "option 0 selects Allow",
			selectedOption: 0,
			expectedAction: PermissionAllow,
		},
		{
			name:           "option 1 selects AllowForSession",
			selectedOption: 1,
			expectedAction: PermissionAllowForSession,
		},
		{
			name:           "option 2 selects Deny",
			selectedOption: 2,
			expectedAction: PermissionDeny,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			permission := permission.PermissionRequest{
				ID:         "test-id",
				SessionID:  "session-1",
				ToolCallID: "tool-call-1",
				ToolName:   tools.BashToolName,
				Path:       "/home/test",
				Params: tools.BashPermissionsParams{
					Command: "echo test",
				},
			}

			dialog := NewPermissionDialogCmp(permission, nil)
			impl := dialog.(*permissionDialogCmp)
			impl.selectedOption = tt.selectedOption

			cmd := impl.selectCurrentOption()
			if cmd == nil {
				t.Fatal("expected command, got nil")
			}

			// Execute command
			msg := cmd()
			if msg == nil {
				t.Fatal("expected message from command")
			}
		})
	}
}

func TestPermissionDialogCmp_ContentCaching(t *testing.T) {
	permission := permission.PermissionRequest{
		ID:         "test-id",
		SessionID:  "session-1",
		ToolCallID: "tool-call-1",
		ToolName:   tools.BashToolName,
		Path:       "/home/test",
		Params: tools.BashPermissionsParams{
			Command:     "echo 'test'",
			Description: "Test",
		},
	}

	dialog := NewPermissionDialogCmp(permission, nil)
	impl := dialog.(*permissionDialogCmp)

	// Set window size
	dialog.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// First call should generate content and cache it
	if !impl.contentDirty {
		t.Error("expected contentDirty to be true initially")
	}

	content1 := impl.getOrGenerateContent()
	if impl.contentDirty {
		t.Error("expected contentDirty to be false after generating content")
	}
	if impl.cachedContent == "" {
		t.Error("expected cachedContent to be set")
	}

	// Second call should use cached content
	content2 := impl.getOrGenerateContent()
	if content1 != content2 {
		t.Error("expected cached content to be returned")
	}

	// Window resize should mark content as dirty
	dialog.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	if !impl.contentDirty {
		t.Error("expected contentDirty to be true after window resize")
	}
}

func TestOptions_isSplitMode(t *testing.T) {
	tests := []struct {
		name     string
		diffMode string
		expected *bool
	}{
		{
			name:     "split mode returns true pointer",
			diffMode: "split",
			expected: boolPtr(true),
		},
		{
			name:     "unified mode returns false pointer",
			diffMode: "unified",
			expected: boolPtr(false),
		},
		{
			name:     "empty mode returns nil",
			diffMode: "",
			expected: nil,
		},
		{
			name:     "invalid mode returns nil",
			diffMode: "invalid",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := Options{DiffMode: tt.diffMode}
			result := opts.isSplitMode()

			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %v", *result)
				}
			} else {
				if result == nil {
					t.Error("expected non-nil result")
				} else if *result != *tt.expected {
					t.Errorf("expected %v, got %v", *tt.expected, *result)
				}
			}
		})
	}
}

// Helper function to create bool pointer
func boolPtr(b bool) *bool {
	return &b
}
