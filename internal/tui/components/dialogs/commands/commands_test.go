package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/nexora/nexora/internal/config"
	"github.com/nexora/nexora/internal/tui/components/dialogs"
)

// setupTestConfig creates a minimal test config for tests that need it.
func setupTestConfig(t *testing.T) {
	t.Helper()
	tempDir := t.TempDir()
	_, err := config.Init(tempDir, tempDir, false)
	if err != nil {
		t.Skipf("Failed to init test config: %v", err)
	}
}

func TestNewCommandDialog(t *testing.T) {
	dialog := NewCommandDialog("test-session")
	if dialog == nil {
		t.Fatal("NewCommandDialog returned nil")
	}

	if dialog.ID() != CommandsDialogID {
		t.Errorf("expected dialog ID %q, got %q", CommandsDialogID, dialog.ID())
	}
}

func TestCommandDialogInit(t *testing.T) {
	dialog := NewCommandDialog("test-session")
	cmd := dialog.Init()
	// Init loads custom commands and sets command type, returns a cmd
	if cmd == nil {
		t.Error("expected Init to return a command")
	}
}

func TestCommandDialogID(t *testing.T) {
	dialog := NewCommandDialog("test-session")
	id := dialog.ID()

	if id != CommandsDialogID {
		t.Errorf("expected ID %q, got %q", CommandsDialogID, id)
	}

	const expectedID = "commands"
	if CommandsDialogID != expectedID {
		t.Errorf("expected CommandsDialogID to be %q, got %q", expectedID, CommandsDialogID)
	}
}

func TestCommandDialogPosition(t *testing.T) {
	dialog := NewCommandDialog("test-session").(*commandDialogCmp)

	// Set window dimensions directly
	dialog.wWidth = 100
	dialog.wHeight = 50

	row, col := dialog.Position()

	// Position should be valid
	if row < 0 || row >= 50 {
		t.Errorf("row %d out of bounds for height %d", row, 50)
	}
	if col < 0 {
		t.Errorf("col %d should not be negative", col)
	}

	// Verify positioning logic
	expectedRow := 50/4 - 2
	expectedCol := 100/2 - defaultWidth/2

	if row != expectedRow {
		t.Errorf("expected row %d, got %d", expectedRow, row)
	}
	if col != expectedCol {
		t.Errorf("expected col %d, got %d", expectedCol, col)
	}
}

func TestCommandDialogUpdate_CloseKey(t *testing.T) {
	dialog := NewCommandDialog("test-session")

	msg := tea.KeyPressMsg(tea.Key{
		Code: tea.KeyEscape,
	})

	_, cmd := dialog.Update(msg)

	if cmd == nil {
		t.Error("expected escape key to return a command")
	}

	// Execute the command to get the message
	resultMsg := cmd()
	if _, ok := resultMsg.(dialogs.CloseDialogMsg); !ok {
		t.Error("expected CloseDialogMsg")
	}
}

func TestCommandType_String(t *testing.T) {
	tests := []struct {
		name     string
		cmdType  commandType
		expected string
	}{
		{
			name:     "system commands",
			cmdType:  SystemCommands,
			expected: "System",
		},
		{
			name:     "user commands",
			cmdType:  UserCommands,
			expected: "User",
		},
		{
			name:     "mcp prompts",
			cmdType:  MCPPrompts,
			expected: "MCP",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cmdType.String()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestCommandDialog_Next(t *testing.T) {
	// Note: User commands are now combined with System commands,
	// so next() should skip UserCommands and go directly to MCP
	tests := []struct {
		name          string
		current       commandType
		hasUserCmds   bool
		hasMCPPrompts bool
		expectedNext  commandType
	}{
		{
			name:          "system to mcp when mcp exists (user commands combined with system)",
			current:       SystemCommands,
			hasUserCmds:   true,
			hasMCPPrompts: true,
			expectedNext:  MCPPrompts,
		},
		{
			name:          "system to mcp when no user but mcp exists",
			current:       SystemCommands,
			hasUserCmds:   false,
			hasMCPPrompts: true,
			expectedNext:  MCPPrompts,
		},
		{
			name:          "system to system when nothing else",
			current:       SystemCommands,
			hasUserCmds:   false,
			hasMCPPrompts: false,
			expectedNext:  SystemCommands,
		},
		{
			name:          "system stays system when only user commands (no separate tab)",
			current:       SystemCommands,
			hasUserCmds:   true,
			hasMCPPrompts: false,
			expectedNext:  SystemCommands,
		},
		{
			name:          "user to mcp (legacy)",
			current:       UserCommands,
			hasUserCmds:   true,
			hasMCPPrompts: true,
			expectedNext:  MCPPrompts,
		},
		{
			name:          "mcp to system",
			current:       MCPPrompts,
			hasUserCmds:   true,
			hasMCPPrompts: true,
			expectedNext:  SystemCommands,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog := NewCommandDialog("test-session").(*commandDialogCmp)
			dialog.selected = tt.current

			if tt.hasUserCmds {
				dialog.userCommands = []Command{{ID: "test", Title: "Test"}}
			}

			if tt.hasMCPPrompts {
				dialog.mcpPrompts.SetSlice([]Command{{ID: "test-mcp", Title: "Test MCP"}})
			}

			next := dialog.next()

			if next != tt.expectedNext {
				t.Errorf("expected next to be %v, got %v", tt.expectedNext, next)
			}
		})
	}
}

func TestCommand_Handler(t *testing.T) {
	called := false
	cmd := Command{
		ID:    "test",
		Title: "Test Command",
		Handler: func(c Command) tea.Cmd {
			called = true
			return nil
		},
	}

	handler := cmd.Handler(cmd)
	if handler != nil {
		handler()
	}

	if !called {
		t.Error("expected handler to be called")
	}
}

func TestCommandDialogView(t *testing.T) {
	setupTestConfig(t)
	dialog := NewCommandDialog("test-session").(*commandDialogCmp)
	dialog.wWidth = 100
	dialog.wHeight = 50

	view := dialog.View()

	if len(view) == 0 {
		t.Error("expected View to return non-empty string")
	}
}

func TestCommandDialogCursor(t *testing.T) {
	dialog := NewCommandDialog("test-session").(*commandDialogCmp)
	dialog.wWidth = 100
	dialog.wHeight = 50

	// Cursor returns nil or a valid cursor position
	cursor := dialog.Cursor()
	// The result depends on the command list state, but should not panic
	_ = cursor
}

func TestCommandDialogUpdate_WindowSizeMsg(t *testing.T) {
	setupTestConfig(t)
	dialog := NewCommandDialog("test-session").(*commandDialogCmp)

	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}

	_, _ = dialog.Update(msg)

	if dialog.wWidth != 120 {
		t.Errorf("expected wWidth to be 120, got %d", dialog.wWidth)
	}
	if dialog.wHeight != 40 {
		t.Errorf("expected wHeight to be 40, got %d", dialog.wHeight)
	}
	// Command may or may not be returned depending on internal state
}

func TestCommandDialogUpdate_TabKey(t *testing.T) {
	dialog := NewCommandDialog("test-session").(*commandDialogCmp)
	// User commands are now combined with system, so we need MCP prompts to switch
	dialog.mcpPrompts.SetSlice([]Command{{ID: "mcp-test", Title: "MCP Test"}})

	initialSelected := dialog.selected
	msg := tea.KeyPressMsg(tea.Key{
		Code: tea.KeyTab,
	})

	_, _ = dialog.Update(msg)

	// Tab key switches to MCP prompts when they exist
	if dialog.selected == initialSelected {
		t.Error("expected Tab key to switch command type when MCP prompts exist")
	}
}

func TestCommandDialogUpdate_TabKey_NoMCPPrompts(t *testing.T) {
	dialog := NewCommandDialog("test-session").(*commandDialogCmp)
	// User commands exist but no MCP prompts - tab should do nothing
	// (user commands are combined with system now)
	dialog.userCommands = []Command{{ID: "test", Title: "Test"}}

	msg := tea.KeyPressMsg(tea.Key{
		Code: tea.KeyTab,
	})

	_, cmd := dialog.Update(msg)

	// Tab should return nil when no MCP prompts (user commands don't create separate tab)
	if cmd != nil {
		t.Error("expected Tab key to return nil when no MCP prompts exist")
	}
}

func TestCommandDialogUpdate_SelectKey_NoItem(t *testing.T) {
	dialog := NewCommandDialog("test-session").(*commandDialogCmp)

	msg := tea.KeyPressMsg(tea.Key{
		Code: tea.KeyEnter,
	})

	_, cmd := dialog.Update(msg)

	// With no item selected, should return nil
	if cmd != nil {
		t.Error("expected Enter key to return nil when no item selected")
	}
}

func TestCommandDialogUpdate_DefaultKeyHandling(t *testing.T) {
	dialog := NewCommandDialog("test-session").(*commandDialogCmp)

	// Test with a regular character key
	msg := tea.KeyPressMsg(tea.Key{
		Code: 'a',
		Text: "a",
	})

	_, cmd := dialog.Update(msg)

	// Should delegate to command list and return a command
	_ = cmd
}

func TestCommandDialogListWidth(t *testing.T) {
	dialog := NewCommandDialog("test-session").(*commandDialogCmp)

	width := dialog.listWidth()

	expected := defaultWidth - 2
	if width != expected {
		t.Errorf("expected listWidth %d, got %d", expected, width)
	}
}

func TestCommandDialogListHeight(t *testing.T) {
	dialog := NewCommandDialog("test-session").(*commandDialogCmp)
	dialog.wHeight = 50

	height := dialog.listHeight()

	// Should be calculated based on items + padding
	if height <= 0 {
		t.Error("expected listHeight to be positive")
	}

	// Should be capped at wHeight/2
	if height > dialog.wHeight/2 {
		t.Errorf("listHeight %d exceeds max %d", height, dialog.wHeight/2)
	}
}

func TestCommandDialogDefaultWidth(t *testing.T) {
	if defaultWidth != 70 {
		t.Errorf("expected defaultWidth to be 70, got %d", defaultWidth)
	}
}

func TestCommandDialogSetCommandType(t *testing.T) {
	setupTestConfig(t)
	dialog := NewCommandDialog("test-session").(*commandDialogCmp)
	dialog.userCommands = []Command{{ID: "user1", Title: "User 1"}}
	dialog.mcpPrompts.SetSlice([]Command{{ID: "mcp1", Title: "MCP 1"}})

	tests := []struct {
		name     string
		cmdType  commandType
		expected commandType
	}{
		{"system", SystemCommands, SystemCommands},
		{"user", UserCommands, UserCommands},
		{"mcp", MCPPrompts, MCPPrompts},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = dialog.setCommandType(tt.cmdType)
			// setCommandType may return nil if the list has no items
			if dialog.selected != tt.expected {
				t.Errorf("expected selected to be %v, got %v", tt.expected, dialog.selected)
			}
		})
	}
}

func TestCommandDialogCommandTypeRadio(t *testing.T) {
	dialog := NewCommandDialog("test-session").(*commandDialogCmp)

	// Without user commands or MCP prompts - should show System only
	radio := dialog.commandTypeRadio()
	if radio == "" {
		t.Error("expected commandTypeRadio to return non-empty string")
	}

	// With user commands - should still only show System (user combined with system)
	dialog.userCommands = []Command{{ID: "test", Title: "Test"}}
	radio = dialog.commandTypeRadio()
	if radio == "" {
		t.Error("expected commandTypeRadio with user commands to return non-empty string")
	}
	// User tab should not appear in radio
	if strings.Contains(radio, "User") {
		t.Error("expected commandTypeRadio NOT to show User tab (user commands combined with system)")
	}

	// With MCP prompts - should show System and MCP
	dialog.mcpPrompts.SetSlice([]Command{{ID: "mcp", Title: "MCP"}})
	radio = dialog.commandTypeRadio()
	if radio == "" {
		t.Error("expected commandTypeRadio with MCP prompts to return non-empty string")
	}
	if !strings.Contains(radio, "MCP") {
		t.Error("expected commandTypeRadio to show MCP tab when MCP prompts exist")
	}
}

func TestStripCommandPrefix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "strip user prefix",
			input:    "user:rewrite",
			expected: "rewrite",
		},
		{
			name:     "strip project prefix",
			input:    "project:deploy",
			expected: "deploy",
		},
		{
			name:     "no prefix",
			input:    "command",
			expected: "command",
		},
		{
			name:     "nested user prefix",
			input:    "user:sub:cmd",
			expected: "sub:cmd",
		},
		{
			name:     "nested project prefix",
			input:    "project:dir:cmd",
			expected: "dir:cmd",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripCommandPrefix(tt.input)
			if result != tt.expected {
				t.Errorf("stripCommandPrefix(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCommandDialogStyle(t *testing.T) {
	dialog := NewCommandDialog("test-session").(*commandDialogCmp)

	style := dialog.style()

	// Should return a valid style
	rendered := style.Render("test")
	if rendered == "" {
		t.Error("expected style to render non-empty string")
	}
}

func TestCommandDialogMoveCursor(t *testing.T) {
	dialog := NewCommandDialog("test-session").(*commandDialogCmp)
	dialog.wWidth = 100
	dialog.wHeight = 50

	cursor := &tea.Cursor{}
	cursor.X = 5
	cursor.Y = 10
	moved := dialog.moveCursor(cursor)

	if moved.X == 5 && moved.Y == 10 {
		t.Error("expected cursor to be moved")
	}
}

// Tests for loader.go

func TestExtractArgNames(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "single arg",
			content:  "Hello $NAME",
			expected: []string{"NAME"},
		},
		{
			name:     "multiple args",
			content:  "Hello $NAME, your age is $AGE",
			expected: []string{"NAME", "AGE"},
		},
		{
			name:     "duplicate args",
			content:  "$NAME said hello to $NAME",
			expected: []string{"NAME"},
		},
		{
			name:     "no args",
			content:  "Hello world",
			expected: nil,
		},
		{
			name:     "complex arg names",
			content:  "$USER_NAME and $API_KEY_123",
			expected: []string{"USER_NAME", "API_KEY_123"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractArgNames(tt.content)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d args, got %d", len(tt.expected), len(result))
				return
			}
			for i, arg := range result {
				if arg != tt.expected[i] {
					t.Errorf("expected arg %d to be %s, got %s", i, tt.expected[i], arg)
				}
			}
		})
	}
}

func TestIsMarkdownFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected bool
	}{
		{"lowercase md", "file.md", true},
		{"uppercase MD", "file.MD", true},
		{"mixed case", "file.Md", true},
		{"not markdown", "file.txt", false},
		{"no extension", "file", false},
		{"go file", "main.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isMarkdownFile(tt.filename)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestBuildCommandID(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		baseDir  string
		prefix   string
		expected string
	}{
		{
			name:     "simple file",
			path:     "/base/test.md",
			baseDir:  "/base",
			prefix:   "user:",
			expected: "user:test",
		},
		{
			name:     "nested file",
			path:     "/base/dir/subdir/test.md",
			baseDir:  "/base",
			prefix:   "project:",
			expected: "project:dir:subdir:test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildCommandID(tt.path, tt.baseDir, tt.prefix)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestEnsureDir(t *testing.T) {
	tmpDir := t.TempDir()
	newDir := filepath.Join(tmpDir, "new", "nested", "dir")

	err := ensureDir(newDir)
	if err != nil {
		t.Errorf("ensureDir failed: %v", err)
	}

	// Verify directory was created
	info, err := os.Stat(newDir)
	if err != nil {
		t.Errorf("directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected a directory")
	}

	// Calling again should succeed (directory already exists)
	err = ensureDir(newDir)
	if err != nil {
		t.Errorf("ensureDir on existing dir failed: %v", err)
	}
}

func TestLoadCustomCommands_WithConfig(t *testing.T) {
	setupTestConfig(t)
	// With config loaded, should return commands (empty or not)
	commands, err := LoadCustomCommands()
	if err != nil {
		t.Errorf("unexpected error when config loaded: %v", err)
	}
	// Commands may be empty, that's ok - just verify no error
	_ = commands
}

func TestCommandLoader_LoadFromSource(t *testing.T) {
	tmpDir := t.TempDir()
	commandsDir := filepath.Join(tmpDir, "commands")

	// Create commands directory with a test command
	err := os.MkdirAll(commandsDir, 0o755)
	if err != nil {
		t.Fatalf("failed to create commands dir: %v", err)
	}

	// Create a test command file
	cmdContent := "Hello $NAME, welcome!"
	cmdFile := filepath.Join(commandsDir, "greet.md")
	err = os.WriteFile(cmdFile, []byte(cmdContent), 0o644)
	if err != nil {
		t.Fatalf("failed to write command file: %v", err)
	}

	loader := &commandLoader{
		sources: []commandSource{
			{path: commandsDir, prefix: "test:"},
		},
	}

	commands, err := loader.loadFromSource(commandSource{path: commandsDir, prefix: "test:"})
	if err != nil {
		t.Errorf("loadFromSource failed: %v", err)
	}

	if len(commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(commands))
	}

	if commands[0].ID != "test:greet" {
		t.Errorf("expected ID 'test:greet', got '%s'", commands[0].ID)
	}
}

func TestCommandLoader_LoadAll(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple source directories
	source1 := filepath.Join(tmpDir, "source1")
	source2 := filepath.Join(tmpDir, "source2")

	for _, dir := range []string{source1, source2} {
		err := os.MkdirAll(dir, 0o755)
		if err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
		err = os.WriteFile(filepath.Join(dir, "cmd.md"), []byte("test"), 0o644)
		if err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
	}

	loader := &commandLoader{
		sources: []commandSource{
			{path: source1, prefix: "a:"},
			{path: source2, prefix: "b:"},
		},
	}

	commands, err := loader.loadAll()
	if err != nil {
		t.Errorf("loadAll failed: %v", err)
	}

	if len(commands) != 2 {
		t.Errorf("expected 2 commands, got %d", len(commands))
	}
}

func TestCreateCommandHandler_NoArgs(t *testing.T) {
	content := "Simple command without arguments"
	handler := createCommandHandler("test", "Test command", content)

	cmd := Command{ID: "test", Title: "Test"}
	result := handler(cmd)

	if result == nil {
		t.Error("expected handler to return a command")
	}

	// Execute the command
	msg := result()
	if customMsg, ok := msg.(CommandRunCustomMsg); ok {
		if customMsg.Content != content {
			t.Errorf("expected content %q, got %q", content, customMsg.Content)
		}
	} else {
		t.Errorf("expected CommandRunCustomMsg, got %T", msg)
	}
}

func TestCreateCommandHandler_WithArgs(t *testing.T) {
	content := "Hello $NAME!"
	handler := createCommandHandler("test", "Test command", content)

	cmd := Command{ID: "test", Title: "Test"}
	result := handler(cmd)

	if result == nil {
		t.Error("expected handler to return a command")
	}

	// Should return ShowArgumentsDialogMsg
	msg := result()
	if dialogMsg, ok := msg.(ShowArgumentsDialogMsg); ok {
		if dialogMsg.CommandID != "test" {
			t.Errorf("expected CommandID 'test', got '%s'", dialogMsg.CommandID)
		}
		if len(dialogMsg.ArgNames) != 1 || dialogMsg.ArgNames[0] != "NAME" {
			t.Errorf("expected ArgNames [NAME], got %v", dialogMsg.ArgNames)
		}
	} else {
		t.Errorf("expected ShowArgumentsDialogMsg, got %T", msg)
	}
}

func TestExecUserPrompt(t *testing.T) {
	content := "Hello $NAME, your age is $AGE"
	args := map[string]string{
		"NAME": "Alice",
		"AGE":  "30",
	}

	cmd := execUserPrompt(content, args)
	msg := cmd()

	if customMsg, ok := msg.(CommandRunCustomMsg); ok {
		expected := "Hello Alice, your age is 30"
		if customMsg.Content != expected {
			t.Errorf("expected content %q, got %q", expected, customMsg.Content)
		}
	} else {
		t.Errorf("expected CommandRunCustomMsg, got %T", msg)
	}
}

func TestGetXDGCommandsDir(t *testing.T) {
	// Test with XDG_CONFIG_HOME set
	original := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", original)

	os.Setenv("XDG_CONFIG_HOME", "/custom/config")
	result := getXDGCommandsDir()
	expected := "/custom/config/nexora/commands"
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestBuildCommandSources(t *testing.T) {
	cfg := &config.Config{
		Options: &config.Options{
			DataDirectory: "/project/.nexora",
		},
	}

	sources := buildCommandSources(cfg)

	// Should have at least the project source
	found := false
	for _, source := range sources {
		if source.prefix == projectCommandPrefix {
			found = true
			if source.path != "/project/.nexora/commands" {
				t.Errorf("expected project path '/project/.nexora/commands', got '%s'", source.path)
			}
		}
	}

	if !found {
		t.Error("expected to find project command source")
	}
}

func TestCommandsDialogID_Constant(t *testing.T) {
	if CommandsDialogID != "commands" {
		t.Errorf("expected CommandsDialogID to be 'commands', got '%s'", CommandsDialogID)
	}
}

func TestUserCommandPrefix(t *testing.T) {
	if userCommandPrefix != "user:" {
		t.Errorf("expected userCommandPrefix to be 'user:', got '%s'", userCommandPrefix)
	}
}

func TestProjectCommandPrefix(t *testing.T) {
	if projectCommandPrefix != "project:" {
		t.Errorf("expected projectCommandPrefix to be 'project:', got '%s'", projectCommandPrefix)
	}
}

// Tests for arguments.go

func TestNewCommandArgumentsDialog(t *testing.T) {
	args := []Argument{
		{Name: "NAME", Title: "Name", Description: "Enter your name", Required: true},
		{Name: "AGE", Title: "Age", Description: "Enter your age"},
	}

	onSubmit := func(args map[string]string) tea.Cmd { return nil }
	dialog := NewCommandArgumentsDialog("test", "Test Dialog", "test-cmd", "A test command", args, onSubmit)

	if dialog == nil {
		t.Fatal("NewCommandArgumentsDialog returned nil")
	}

	if dialog.ID() != argumentsDialogID {
		t.Errorf("expected ID %q, got %q", argumentsDialogID, dialog.ID())
	}
}

func TestCommandArgumentsDialogInit(t *testing.T) {
	args := []Argument{{Name: "NAME", Title: "Name"}}
	onSubmit := func(args map[string]string) tea.Cmd { return nil }
	dialog := NewCommandArgumentsDialog("test", "Test", "cmd", "desc", args, onSubmit)

	cmd := dialog.Init()
	if cmd != nil {
		t.Error("expected Init to return nil")
	}
}

func TestCommandArgumentsDialogUpdate_WindowSize(t *testing.T) {
	args := []Argument{{Name: "NAME", Title: "Name"}}
	onSubmit := func(args map[string]string) tea.Cmd { return nil }
	dialog := NewCommandArgumentsDialog("test", "Test", "cmd", "desc", args, onSubmit).(*commandArgumentsDialogCmp)

	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	_, _ = dialog.Update(msg)

	if dialog.wWidth != 100 {
		t.Errorf("expected wWidth 100, got %d", dialog.wWidth)
	}
	if dialog.wHeight != 50 {
		t.Errorf("expected wHeight 50, got %d", dialog.wHeight)
	}
}

func TestCommandArgumentsDialogUpdate_CloseKey(t *testing.T) {
	args := []Argument{{Name: "NAME", Title: "Name"}}
	onSubmit := func(args map[string]string) tea.Cmd { return nil }
	dialog := NewCommandArgumentsDialog("test", "Test", "cmd", "desc", args, onSubmit)

	msg := tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape})
	_, cmd := dialog.Update(msg)

	if cmd == nil {
		t.Error("expected Close key to return a command")
	}
}

func TestCommandArgumentsDialogUpdate_NextKey(t *testing.T) {
	args := []Argument{
		{Name: "NAME", Title: "Name"},
		{Name: "AGE", Title: "Age"},
	}
	onSubmit := func(args map[string]string) tea.Cmd { return nil }
	dialog := NewCommandArgumentsDialog("test", "Test", "cmd", "desc", args, onSubmit).(*commandArgumentsDialogCmp)

	// Press tab to move to next
	msg := tea.KeyPressMsg(tea.Key{Code: tea.KeyTab})
	_, _ = dialog.Update(msg)

	if dialog.focused != 1 {
		t.Errorf("expected focused to be 1, got %d", dialog.focused)
	}
}

func TestCommandArgumentsDialogUpdate_PreviousKey(t *testing.T) {
	args := []Argument{
		{Name: "NAME", Title: "Name"},
		{Name: "AGE", Title: "Age"},
	}
	onSubmit := func(args map[string]string) tea.Cmd { return nil }
	dialog := NewCommandArgumentsDialog("test", "Test", "cmd", "desc", args, onSubmit).(*commandArgumentsDialogCmp)

	// Move to second input first
	dialog.focused = 1
	dialog.inputs[0].Blur()
	dialog.inputs[1].Focus()

	// Press shift+tab to move to previous
	msg := tea.KeyPressMsg(tea.Key{
		Code: tea.KeyTab,
		Mod:  tea.ModShift,
	})
	_, _ = dialog.Update(msg)

	if dialog.focused != 0 {
		t.Errorf("expected focused to be 0, got %d", dialog.focused)
	}
}

func TestCommandArgumentsDialogUpdate_ConfirmKey_NotLast(t *testing.T) {
	args := []Argument{
		{Name: "NAME", Title: "Name"},
		{Name: "AGE", Title: "Age"},
	}
	onSubmit := func(args map[string]string) tea.Cmd { return nil }
	dialog := NewCommandArgumentsDialog("test", "Test", "cmd", "desc", args, onSubmit).(*commandArgumentsDialogCmp)

	// Press enter while on first input - should move to next
	msg := tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter})
	_, _ = dialog.Update(msg)

	if dialog.focused != 1 {
		t.Errorf("expected focused to be 1, got %d", dialog.focused)
	}
}

func TestCommandArgumentsDialogUpdate_ConfirmKey_Last(t *testing.T) {
	args := []Argument{
		{Name: "NAME", Title: "Name"},
	}
	onSubmit := func(args map[string]string) tea.Cmd {
		return nil
	}
	dialog := NewCommandArgumentsDialog("test", "Test", "cmd", "desc", args, onSubmit).(*commandArgumentsDialogCmp)

	// Press enter while on last input - should submit
	msg := tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter})
	_, cmd := dialog.Update(msg)

	if cmd == nil {
		t.Error("expected Confirm key on last input to return a command")
	}
}

func TestCommandArgumentsDialogUpdate_TextInput(t *testing.T) {
	args := []Argument{{Name: "NAME", Title: "Name"}}
	onSubmit := func(args map[string]string) tea.Cmd { return nil }
	dialog := NewCommandArgumentsDialog("test", "Test", "cmd", "desc", args, onSubmit).(*commandArgumentsDialogCmp)

	// Type a character
	msg := tea.KeyPressMsg(tea.Key{
		Code: 'a',
		Text: "a",
	})
	_, cmd := dialog.Update(msg)

	// Should be passed to text input
	_ = cmd
}

func TestCommandArgumentsDialogUpdate_PasteMsg(t *testing.T) {
	args := []Argument{{Name: "NAME", Title: "Name"}}
	onSubmit := func(args map[string]string) tea.Cmd { return nil }
	dialog := NewCommandArgumentsDialog("test", "Test", "cmd", "desc", args, onSubmit).(*commandArgumentsDialogCmp)

	msg := tea.PasteMsg{Content: "pasted text"}
	_, cmd := dialog.Update(msg)

	// Should be passed to text input
	_ = cmd
}

func TestCommandArgumentsDialogView(t *testing.T) {
	args := []Argument{
		{Name: "NAME", Title: "Name", Required: true},
		{Name: "AGE", Title: "Age", Description: "Your age"},
	}
	onSubmit := func(args map[string]string) tea.Cmd { return nil }
	dialog := NewCommandArgumentsDialog("test", "Test Dialog", "cmd", "A test command", args, onSubmit).(*commandArgumentsDialogCmp)
	dialog.wWidth = 100
	dialog.wHeight = 50

	view := dialog.View()

	if len(view) == 0 {
		t.Error("expected View to return non-empty string")
	}
}

func TestCommandArgumentsDialogCursor(t *testing.T) {
	args := []Argument{{Name: "NAME", Title: "Name"}}
	onSubmit := func(args map[string]string) tea.Cmd { return nil }
	dialog := NewCommandArgumentsDialog("test", "Test", "cmd", "desc", args, onSubmit).(*commandArgumentsDialogCmp)
	dialog.wWidth = 100
	dialog.wHeight = 50

	cursor := dialog.Cursor()
	// Should return a cursor for the focused input
	_ = cursor
}

func TestCommandArgumentsDialogCursor_NoInputs(t *testing.T) {
	args := []Argument{}
	onSubmit := func(args map[string]string) tea.Cmd { return nil }
	dialog := NewCommandArgumentsDialog("test", "Test", "cmd", "desc", args, onSubmit).(*commandArgumentsDialogCmp)

	cursor := dialog.Cursor()
	if cursor != nil {
		t.Error("expected nil cursor with no inputs")
	}
}

func TestCommandArgumentsDialogPosition(t *testing.T) {
	args := []Argument{{Name: "NAME", Title: "Name"}}
	onSubmit := func(args map[string]string) tea.Cmd { return nil }
	dialog := NewCommandArgumentsDialog("test", "Test", "cmd", "desc", args, onSubmit).(*commandArgumentsDialogCmp)
	dialog.wWidth = 100
	dialog.wHeight = 50
	dialog.width = 60
	dialog.height = 15

	row, col := dialog.Position()

	expectedRow := (50 / 2) - (15 / 2)
	expectedCol := (100 / 2) - (60 / 2)

	if row != expectedRow {
		t.Errorf("expected row %d, got %d", expectedRow, row)
	}
	if col != expectedCol {
		t.Errorf("expected col %d, got %d", expectedCol, col)
	}
}

func TestArgumentsDialogID_Constant(t *testing.T) {
	if argumentsDialogID != "arguments" {
		t.Errorf("expected argumentsDialogID to be 'arguments', got '%s'", argumentsDialogID)
	}
}

func TestArgumentStruct(t *testing.T) {
	arg := Argument{
		Name:        "TEST_ARG",
		Title:       "Test Argument",
		Description: "A test argument",
		Required:    true,
	}

	if arg.Name != "TEST_ARG" {
		t.Error("Name field incorrect")
	}
	if arg.Title != "Test Argument" {
		t.Error("Title field incorrect")
	}
	if arg.Description != "A test argument" {
		t.Error("Description field incorrect")
	}
	if !arg.Required {
		t.Error("Required field incorrect")
	}
}

// Message type tests
func TestMessageTypes(t *testing.T) {
	// Test that all message types can be instantiated
	_ = SwitchSessionsMsg{}
	_ = NewSessionsMsg{}
	_ = SwitchModelMsg{}
	_ = QuitMsg{}
	_ = OpenFilePickerMsg{}
	_ = ToggleHelpMsg{}
	_ = ToggleCompactModeMsg{}
	_ = ToggleThinkingMsg{}
	_ = OpenReasoningDialogMsg{}
	_ = OpenExternalEditorMsg{}
	_ = ToggleYoloModeMsg{}
	_ = AboutNexoraMsg{}
	_ = CompactMsg{SessionID: "test-session"}
	_ = ShowArgumentsDialogMsg{CommandID: "test", ArgNames: []string{"ARG1"}}
	_ = CloseArgumentsDialogMsg{Submit: true, CommandID: "test"}
	_ = CommandRunCustomMsg{Content: "test content"}
	_ = ShowMCPPromptArgumentsDialogMsg{}
}

func TestCompactMsg(t *testing.T) {
	msg := CompactMsg{SessionID: "session-123"}
	if msg.SessionID != "session-123" {
		t.Errorf("expected SessionID 'session-123', got '%s'", msg.SessionID)
	}
}

func TestShowArgumentsDialogMsg(t *testing.T) {
	msg := ShowArgumentsDialogMsg{
		CommandID:   "cmd-1",
		Description: "Test description",
		ArgNames:    []string{"ARG1", "ARG2"},
	}

	if msg.CommandID != "cmd-1" {
		t.Error("CommandID incorrect")
	}
	if msg.Description != "Test description" {
		t.Error("Description incorrect")
	}
	if len(msg.ArgNames) != 2 {
		t.Error("ArgNames count incorrect")
	}
}

func TestCloseArgumentsDialogMsg(t *testing.T) {
	msg := CloseArgumentsDialogMsg{
		Submit:    true,
		CommandID: "cmd-1",
		Content:   "test content",
		Args:      map[string]string{"key": "value"},
	}

	if !msg.Submit {
		t.Error("Submit incorrect")
	}
	if msg.CommandID != "cmd-1" {
		t.Error("CommandID incorrect")
	}
	if msg.Content != "test content" {
		t.Error("Content incorrect")
	}
	if msg.Args["key"] != "value" {
		t.Error("Args incorrect")
	}
}
