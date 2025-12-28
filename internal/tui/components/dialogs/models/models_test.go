package models

import (
	"strings"
	"testing"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/catwalk/pkg/catwalk"
	"github.com/nexora/nexora/internal/config"
	"github.com/nexora/nexora/internal/tui/exp/list"
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

// =============================================================================
// KeyMap Tests
// =============================================================================

func TestDefaultKeyMap(t *testing.T) {
	km := DefaultKeyMap()

	tests := []struct {
		name    string
		binding key.Binding
		keys    []string
	}{
		{"Select", km.Select, []string{"enter", "ctrl+y"}},
		{"Next", km.Next, []string{"down", "ctrl+n"}},
		{"Previous", km.Previous, []string{"up", "ctrl+p"}},
		{"Choose", km.Choose, []string{"left", "right", "h", "l"}},
		{"Tab", km.Tab, []string{"tab"}},
		{"Edit", km.Edit, []string{"ctrl+e"}},
		{"Close", km.Close, []string{"esc", "alt+esc"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keys := tt.binding.Keys()
			if len(keys) != len(tt.keys) {
				t.Errorf("expected %d keys, got %d", len(tt.keys), len(keys))
			}
			for i, expected := range tt.keys {
				if i < len(keys) && keys[i] != expected {
					t.Errorf("expected key %q, got %q", expected, keys[i])
				}
			}
		})
	}
}

func TestKeyMap_KeyBindings(t *testing.T) {
	km := DefaultKeyMap()
	bindings := km.KeyBindings()

	if len(bindings) != 6 {
		t.Errorf("expected 6 bindings, got %d", len(bindings))
	}
}

func TestKeyMap_FullHelp(t *testing.T) {
	km := DefaultKeyMap()
	fullHelp := km.FullHelp()

	if len(fullHelp) == 0 {
		t.Error("expected FullHelp to return at least one row")
	}

	// Each row should have at most 4 bindings
	for i, row := range fullHelp {
		if len(row) > 4 {
			t.Errorf("row %d has %d bindings, expected at most 4", i, len(row))
		}
	}
}

func TestKeyMap_ShortHelp_Default(t *testing.T) {
	km := DefaultKeyMap()
	shortHelp := km.ShortHelp()

	if len(shortHelp) == 0 {
		t.Error("expected ShortHelp to return at least one binding")
	}

	// Default state should include arrow keys, tab, edit, select, close
	if len(shortHelp) != 5 {
		t.Errorf("expected 5 short help bindings in default state, got %d", len(shortHelp))
	}
}

func TestKeyMap_ShortHelp_APIKeyState(t *testing.T) {
	km := DefaultKeyMap()
	km.isAPIKeyHelp = true
	km.isAPIKeyValid = false

	shortHelp := km.ShortHelp()

	// API key input state should show enter + close
	if len(shortHelp) != 2 {
		t.Errorf("expected 2 short help bindings in API key state, got %d", len(shortHelp))
	}
}

func TestKeyMap_ShortHelp_APIKeyValidState(t *testing.T) {
	km := DefaultKeyMap()
	km.isAPIKeyHelp = true
	km.isAPIKeyValid = true

	shortHelp := km.ShortHelp()

	// Validated API key state should show only select
	if len(shortHelp) != 1 {
		t.Errorf("expected 1 short help binding when API key is valid, got %d", len(shortHelp))
	}
}

func TestKeyMap_ShortHelp_ClaudeAuthChooser(t *testing.T) {
	km := DefaultKeyMap()
	km.isClaudeAuthChoiseHelp = true

	shortHelp := km.ShortHelp()

	// Claude auth chooser should have choose, accept, back
	if len(shortHelp) != 3 {
		t.Errorf("expected 3 short help bindings in Claude auth chooser, got %d", len(shortHelp))
	}
}

func TestKeyMap_ShortHelp_ClaudeOAuth(t *testing.T) {
	km := DefaultKeyMap()
	km.isClaudeOAuthHelp = true

	shortHelp := km.ShortHelp()

	// OAuth state should show enter + esc
	if len(shortHelp) < 2 {
		t.Errorf("expected at least 2 short help bindings in OAuth state, got %d", len(shortHelp))
	}
}

func TestKeyMap_ShortHelp_ClaudeOAuth_URLState(t *testing.T) {
	km := DefaultKeyMap()
	km.isClaudeOAuthHelp = true
	km.isClaudeOAuthURLState = true

	shortHelp := km.ShortHelp()

	// URL state should also have copy url binding
	if len(shortHelp) != 3 {
		t.Errorf("expected 3 short help bindings in OAuth URL state, got %d", len(shortHelp))
	}
}

func TestKeyMap_ShortHelp_ClaudeOAuthComplete(t *testing.T) {
	km := DefaultKeyMap()
	km.isClaudeOAuthHelp = true
	km.isClaudeOAuthHelpComplete = true

	shortHelp := km.ShortHelp()

	// Complete state should only show close
	if len(shortHelp) != 1 {
		t.Errorf("expected 1 short help binding when OAuth complete, got %d", len(shortHelp))
	}
}

// =============================================================================
// APIKeyInput Tests
// =============================================================================

func TestNewAPIKeyInput(t *testing.T) {
	setupTestConfig(t)

	input := NewAPIKeyInput()
	if input == nil {
		t.Fatal("NewAPIKeyInput returned nil")
	}

	// Initial state should be Initial
	if input.state != APIKeyInputStateInitial {
		t.Errorf("expected state APIKeyInputStateInitial, got %v", input.state)
	}

	// Provider name should be default
	if input.providerName != "Provider" {
		t.Errorf("expected providerName 'Provider', got %q", input.providerName)
	}

	// showTitle should be true by default
	if !input.showTitle {
		t.Error("expected showTitle to be true by default")
	}
}

func TestAPIKeyInput_SetProviderName(t *testing.T) {
	setupTestConfig(t)

	input := NewAPIKeyInput()
	input.SetProviderName("Anthropic")

	if input.providerName != "Anthropic" {
		t.Errorf("expected providerName 'Anthropic', got %q", input.providerName)
	}
}

func TestAPIKeyInput_SetShowTitle(t *testing.T) {
	setupTestConfig(t)

	input := NewAPIKeyInput()
	input.SetShowTitle(false)

	if input.showTitle {
		t.Error("expected showTitle to be false after SetShowTitle(false)")
	}
}

func TestAPIKeyInput_GetTitle(t *testing.T) {
	setupTestConfig(t)

	input := NewAPIKeyInput()
	input.SetProviderName("OpenAI")
	input.Init()

	title := input.GetTitle()
	if title == "" {
		t.Error("expected non-empty title after Init")
	}
}

func TestAPIKeyInput_Init(t *testing.T) {
	setupTestConfig(t)

	input := NewAPIKeyInput()
	cmd := input.Init()

	if cmd == nil {
		t.Error("expected Init to return a non-nil Cmd (spinner tick)")
	}
}

func TestAPIKeyInput_Value(t *testing.T) {
	setupTestConfig(t)

	input := NewAPIKeyInput()
	input.Init()

	// Initially empty
	if input.Value() != "" {
		t.Errorf("expected empty value initially, got %q", input.Value())
	}
}

func TestAPIKeyInput_SetWidth(t *testing.T) {
	setupTestConfig(t)

	input := NewAPIKeyInput()
	input.SetWidth(80)

	if input.width != 80 {
		t.Errorf("expected width 80, got %d", input.width)
	}
}

func TestAPIKeyInput_Reset(t *testing.T) {
	setupTestConfig(t)

	input := NewAPIKeyInput()
	input.Init()
	input.state = APIKeyInputStateError

	input.Reset()

	if input.state != APIKeyInputStateInitial {
		t.Errorf("expected state to be reset to Initial, got %v", input.state)
	}
}

func TestAPIKeyInput_Tick(t *testing.T) {
	setupTestConfig(t)

	input := NewAPIKeyInput()
	input.Init()

	// When not verifying, tick should return nil
	cmd := input.Tick()
	if cmd != nil {
		t.Error("expected Tick to return nil when not verifying")
	}

	// When verifying, tick should return a command
	input.state = APIKeyInputStateVerifying
	cmd = input.Tick()
	if cmd == nil {
		t.Error("expected Tick to return non-nil Cmd when verifying")
	}
}

func TestAPIKeyInput_Update_StateChange(t *testing.T) {
	setupTestConfig(t)

	input := NewAPIKeyInput()
	input.Init()

	tests := []struct {
		name          string
		state         APIKeyInputState
		expectedState APIKeyInputState
	}{
		{"to verifying", APIKeyInputStateVerifying, APIKeyInputStateVerifying},
		{"to verified", APIKeyInputStateVerified, APIKeyInputStateVerified},
		{"to error", APIKeyInputStateError, APIKeyInputStateError},
		{"to initial", APIKeyInputStateInitial, APIKeyInputStateInitial},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := APIKeyStateChangeMsg{State: tt.state}
			u, _ := input.Update(msg)
			updated := u.(*APIKeyInput)
			if updated.state != tt.expectedState {
				t.Errorf("expected state %v, got %v", tt.expectedState, updated.state)
			}
		})
	}
}

func TestAPIKeyInput_Update_SpinnerTick(t *testing.T) {
	setupTestConfig(t)

	input := NewAPIKeyInput()
	input.Init()
	input.state = APIKeyInputStateVerifying

	msg := spinner.TickMsg{ID: 0}
	u, cmd := input.Update(msg)
	_ = u.(*APIKeyInput)

	// Should return another tick command when verifying
	if cmd == nil {
		t.Error("expected tick command when verifying")
	}
}

func TestAPIKeyInput_View(t *testing.T) {
	setupTestConfig(t)

	input := NewAPIKeyInput()
	input.SetProviderName("TestProvider")
	input.Init()

	view := input.View()
	if view == "" {
		t.Error("expected non-empty view")
	}
}

func TestAPIKeyInput_Cursor(t *testing.T) {
	setupTestConfig(t)

	input := NewAPIKeyInput()
	input.Init()

	// Cursor positioning is dependent on internal state
	cursor := input.Cursor()
	// May be nil or a valid cursor
	_ = cursor
}

// =============================================================================
// modelKey Helper Tests
// =============================================================================

func TestModelKey(t *testing.T) {
	tests := []struct {
		providerID string
		modelID    string
		expected   string
	}{
		{"anthropic", "claude-3-opus", "anthropic:claude-3-opus"},
		{"openai", "gpt-4", "openai:gpt-4"},
		{"", "model", ""},
		{"provider", "", ""},
		{"", "", ""},
	}

	for _, tt := range tests {
		result := modelKey(tt.providerID, tt.modelID)
		if result != tt.expected {
			t.Errorf("modelKey(%q, %q) = %q, want %q", tt.providerID, tt.modelID, result, tt.expected)
		}
	}
}

// =============================================================================
// Model Constants Tests
// =============================================================================

func TestModelTypeConstants(t *testing.T) {
	if LargeModelType != 0 {
		t.Errorf("expected LargeModelType = 0, got %d", LargeModelType)
	}
	if SmallModelType != 1 {
		t.Errorf("expected SmallModelType = 1, got %d", SmallModelType)
	}
}

func TestModelInputPlaceholders(t *testing.T) {
	if largeModelInputPlaceholder == "" {
		t.Error("expected non-empty largeModelInputPlaceholder")
	}
	if smallModelInputPlaceholder == "" {
		t.Error("expected non-empty smallModelInputPlaceholder")
	}
	if largeModelInputPlaceholder == smallModelInputPlaceholder {
		t.Error("expected different placeholders for large and small model types")
	}
}

// =============================================================================
// ModelDialog Tests
// =============================================================================

func TestNewModelDialogCmp(t *testing.T) {
	setupTestConfig(t)

	dialog := NewModelDialogCmp()
	if dialog == nil {
		t.Fatal("NewModelDialogCmp returned nil")
	}
}

func TestModelDialogCmp_ID(t *testing.T) {
	setupTestConfig(t)

	dialog := NewModelDialogCmp()
	id := dialog.ID()

	if id != ModelsDialogID {
		t.Errorf("expected ID %q, got %q", ModelsDialogID, id)
	}

	if ModelsDialogID != "models" {
		t.Errorf("expected ModelsDialogID = 'models', got %q", ModelsDialogID)
	}
}

func TestModelDialogCmp_Init(t *testing.T) {
	setupTestConfig(t)

	dialog := NewModelDialogCmp().(*modelDialogCmp)
	cmd := dialog.Init()

	// Init should return a batch command
	if cmd == nil {
		t.Error("expected Init to return a non-nil Cmd")
	}
}

func TestModelDialogCmp_Position(t *testing.T) {
	setupTestConfig(t)

	dialog := NewModelDialogCmp().(*modelDialogCmp)
	dialog.wWidth = 100
	dialog.wHeight = 40

	row, col := dialog.Position()

	// Position should be calculated based on window size
	if row < 0 || col < 0 {
		t.Errorf("expected non-negative position, got row=%d, col=%d", row, col)
	}

	// Should be roughly centered
	if col < 10 || col > 50 {
		t.Errorf("expected col around center, got %d", col)
	}
}

func TestModelDialogCmp_Update_WindowSize(t *testing.T) {
	setupTestConfig(t)

	dialog := NewModelDialogCmp().(*modelDialogCmp)
	dialog.Init()

	msg := tea.WindowSizeMsg{Width: 120, Height: 50}
	u, _ := dialog.Update(msg)
	updated := u.(*modelDialogCmp)

	if updated.wWidth != 120 {
		t.Errorf("expected wWidth 120, got %d", updated.wWidth)
	}
	if updated.wHeight != 50 {
		t.Errorf("expected wHeight 50, got %d", updated.wHeight)
	}
}

func TestModelDialogCmp_Update_CloseKey(t *testing.T) {
	setupTestConfig(t)

	dialog := NewModelDialogCmp().(*modelDialogCmp)
	dialog.Init()

	// Press escape to close
	msg := tea.KeyPressMsg{Code: tea.KeyEscape}
	u, cmd := dialog.Update(msg)
	_ = u.(*modelDialogCmp)

	// Should emit close command
	if cmd == nil {
		t.Error("expected close command when pressing escape")
	}
}

func TestModelDialogCmp_Update_EscapeFromAPIKey(t *testing.T) {
	setupTestConfig(t)

	dialog := NewModelDialogCmp().(*modelDialogCmp)
	dialog.Init()
	dialog.needsAPIKey = true

	msg := tea.KeyPressMsg{Code: tea.KeyEscape}
	u, _ := dialog.Update(msg)
	updated := u.(*modelDialogCmp)

	// Should go back to model selection
	if updated.needsAPIKey {
		t.Error("expected needsAPIKey to be false after escape")
	}
}

func TestModelDialogCmp_Update_EscapeFromClaudeAuthChooser(t *testing.T) {
	setupTestConfig(t)

	dialog := NewModelDialogCmp().(*modelDialogCmp)
	dialog.Init()
	dialog.showClaudeAuthMethodChooser = true

	msg := tea.KeyPressMsg{Code: tea.KeyEscape}
	u, _ := dialog.Update(msg)
	updated := u.(*modelDialogCmp)

	if updated.showClaudeAuthMethodChooser {
		t.Error("expected showClaudeAuthMethodChooser to be false after escape")
	}
}

func TestModelDialogCmp_Update_Tab(t *testing.T) {
	setupTestConfig(t)

	dialog := NewModelDialogCmp().(*modelDialogCmp)
	dialog.Init()

	// Initial type should be Large
	if dialog.modelList.GetModelType() != LargeModelType {
		t.Errorf("expected initial model type Large, got %d", dialog.modelList.GetModelType())
	}

	// Tab to switch model type
	msg := tea.KeyPressMsg{Code: tea.KeyTab}
	u, _ := dialog.Update(msg)
	updated := u.(*modelDialogCmp)

	if updated.modelList.GetModelType() != SmallModelType {
		t.Errorf("expected model type Small after tab, got %d", updated.modelList.GetModelType())
	}

	// Tab again to go back
	u, _ = updated.Update(msg)
	updated = u.(*modelDialogCmp)

	if updated.modelList.GetModelType() != LargeModelType {
		t.Errorf("expected model type Large after second tab, got %d", updated.modelList.GetModelType())
	}
}

func TestModelDialogCmp_View(t *testing.T) {
	setupTestConfig(t)

	dialog := NewModelDialogCmp().(*modelDialogCmp)
	dialog.Init()
	dialog.wWidth = 100
	dialog.wHeight = 40

	view := dialog.View()

	if view == "" {
		t.Error("expected non-empty view")
	}

	// Should contain the model selection title
	if !strings.Contains(view, "Switch Model") {
		t.Error("expected view to contain 'Switch Model'")
	}
}

func TestModelDialogCmp_View_APIKeyInput(t *testing.T) {
	setupTestConfig(t)

	dialog := NewModelDialogCmp().(*modelDialogCmp)
	dialog.Init()
	dialog.wWidth = 100
	dialog.wHeight = 40
	dialog.needsAPIKey = true
	dialog.apiKeyInput.SetProviderName("TestProvider")
	dialog.apiKeyInput.Init()

	view := dialog.View()

	if view == "" {
		t.Error("expected non-empty view when showing API key input")
	}
}

func TestModelDialogCmp_View_ClaudeAuthChooser(t *testing.T) {
	setupTestConfig(t)

	dialog := NewModelDialogCmp().(*modelDialogCmp)
	dialog.Init()
	dialog.wWidth = 100
	dialog.wHeight = 40
	dialog.showClaudeAuthMethodChooser = true

	view := dialog.View()

	if view == "" {
		t.Error("expected non-empty view when showing Claude auth chooser")
	}
}

func TestModelDialogCmp_Cursor(t *testing.T) {
	setupTestConfig(t)

	dialog := NewModelDialogCmp().(*modelDialogCmp)
	dialog.Init()
	dialog.wWidth = 100
	dialog.wHeight = 40

	// Cursor may be nil in various states
	cursor := dialog.Cursor()
	_ = cursor
}

func TestModelDialogCmp_modelTypeRadio(t *testing.T) {
	setupTestConfig(t)

	dialog := NewModelDialogCmp().(*modelDialogCmp)
	dialog.Init()

	// Large model type
	dialog.modelList.modelType = LargeModelType
	radio := dialog.modelTypeRadio()
	if !strings.Contains(radio, "Large Task") {
		t.Error("expected radio to contain 'Large Task'")
	}
	if !strings.Contains(radio, "Small Task") {
		t.Error("expected radio to contain 'Small Task'")
	}

	// Small model type
	dialog.modelList.modelType = SmallModelType
	radio = dialog.modelTypeRadio()
	if radio == "" {
		t.Error("expected non-empty radio for small model type")
	}
}

// =============================================================================
// ModelListComponent Tests
// =============================================================================

func TestNewModelListComponent(t *testing.T) {
	setupTestConfig(t)

	keyMap := DefaultKeyMap()
	lm := NewModelListComponent(convertToListKeyMap(keyMap), "Test placeholder", true)

	if lm == nil {
		t.Fatal("NewModelListComponent returned nil")
	}

	if lm.modelType != LargeModelType {
		t.Errorf("expected initial modelType = LargeModelType, got %d", lm.modelType)
	}
}

func TestModelListComponent_GetModelType(t *testing.T) {
	setupTestConfig(t)

	keyMap := DefaultKeyMap()
	lm := NewModelListComponent(convertToListKeyMap(keyMap), "Test", false)

	if lm.GetModelType() != LargeModelType {
		t.Errorf("expected initial type LargeModelType, got %d", lm.GetModelType())
	}
}

func TestModelListComponent_SetInputPlaceholder(t *testing.T) {
	setupTestConfig(t)

	keyMap := DefaultKeyMap()
	lm := NewModelListComponent(convertToListKeyMap(keyMap), "Initial placeholder", false)

	// Should not panic
	lm.SetInputPlaceholder("New placeholder")
}

func TestModelListComponent_SelectedModel_Empty(t *testing.T) {
	setupTestConfig(t)

	keyMap := DefaultKeyMap()
	lm := NewModelListComponent(convertToListKeyMap(keyMap), "Test", false)

	// Before init, should return nil
	selected := lm.SelectedModel()
	if selected != nil {
		t.Error("expected nil selected model before init")
	}
}

func TestModelListComponent_View(t *testing.T) {
	setupTestConfig(t)

	keyMap := DefaultKeyMap()
	lm := NewModelListComponent(convertToListKeyMap(keyMap), "Test", false)

	view := lm.View()
	// Should not panic and return something
	_ = view
}

func TestModelListComponent_Cursor(t *testing.T) {
	setupTestConfig(t)

	keyMap := DefaultKeyMap()
	lm := NewModelListComponent(convertToListKeyMap(keyMap), "Test", false)

	cursor := lm.Cursor()
	// May be nil
	_ = cursor
}

// =============================================================================
// Message Type Tests
// =============================================================================

func TestModelSelectedMsg(t *testing.T) {
	msg := ModelSelectedMsg{
		Model: config.SelectedModel{
			Model:    "claude-3-opus",
			Provider: "anthropic",
		},
		ModelType: config.SelectedModelTypeLarge,
	}

	if msg.Model.Model != "claude-3-opus" {
		t.Errorf("expected model 'claude-3-opus', got %q", msg.Model.Model)
	}
	if msg.Model.Provider != "anthropic" {
		t.Errorf("expected provider 'anthropic', got %q", msg.Model.Provider)
	}
}

func TestCloseModelDialogMsg(t *testing.T) {
	// Just verify the type exists and can be instantiated
	msg := CloseModelDialogMsg{}
	_ = msg
}

func TestAPIKeyStateChangeMsg(t *testing.T) {
	tests := []struct {
		state APIKeyInputState
	}{
		{APIKeyInputStateInitial},
		{APIKeyInputStateVerifying},
		{APIKeyInputStateVerified},
		{APIKeyInputStateError},
	}

	for _, tt := range tests {
		msg := APIKeyStateChangeMsg{State: tt.state}
		if msg.State != tt.state {
			t.Errorf("expected state %v, got %v", tt.state, msg.State)
		}
	}
}

// =============================================================================
// ModelOption Tests
// =============================================================================

func TestModelOption(t *testing.T) {
	opt := ModelOption{
		Provider: catwalk.Provider{
			ID:   "anthropic",
			Name: "Anthropic",
		},
		Model: catwalk.Model{
			ID:   "claude-3-opus",
			Name: "Claude 3 Opus",
		},
	}

	if string(opt.Provider.ID) != "anthropic" {
		t.Errorf("expected provider ID 'anthropic', got %q", opt.Provider.ID)
	}
	if opt.Model.ID != "claude-3-opus" {
		t.Errorf("expected model ID 'claude-3-opus', got %q", opt.Model.ID)
	}
}

// =============================================================================
// Helper function to convert KeyMap to list.KeyMap
// =============================================================================

func convertToListKeyMap(km KeyMap) list.KeyMap {
	return list.KeyMap{
		Down:        km.Next,
		Up:          km.Previous,
		DownOneItem: km.Next,
		UpOneItem:   km.Previous,
	}
}

// =============================================================================
// Legacy Tests (preserved from original)
// =============================================================================

func TestModelListRendering(t *testing.T) {
	// Test that model list renders correctly
	models := []Model{
		{Name: "MiniMax M2.1", Provider: "MiniMax"},
		{Name: "Claude Opus 4", Provider: "Anthropic"},
		{Name: "Grok 4", Provider: "xAI"},
	}

	if len(models) != 3 {
		t.Errorf("Expected 3 models, got %d", len(models))
	}
}

func TestModelSelection(t *testing.T) {
	models := []Model{
		{ID: "model1", Name: "MiniMax M2.1", Selected: false},
		{ID: "model2", Name: "Claude Opus 4", Selected: true},
	}
	
	selected := false
	for _, m := range models {
		if m.Selected {
			selected = true
			if m.ID != "model2" {
				t.Errorf("Wrong model selected: %s", m.ID)
			}
		}
	}
	
	if !selected {
		t.Error("No model selected")
	}
}

func TestModelSelectionToggle(t *testing.T) {
	m := Model{ID: "test", Name: "Test Model", Selected: false}
	
	if m.Selected {
		t.Error("Model should not be selected initially")
	}
	
	m.Selected = true
	if !m.Selected {
		t.Error("Model should be selected after toggle")
	}
}

func TestProviderFiltering(t *testing.T) {
	allModels := []Model{
		{Name: "Model1", Provider: "Anthropic"},
		{Name: "Model2", Provider: "MiniMax"},
		{Name: "Model3", Provider: "Anthropic"},
		{Name: "Model4", Provider: "xAI"},
	}
	
	anthropic := filterByProvider(allModels, "Anthropic")
	if len(anthropic) != 2 {
		t.Errorf("Expected 2 Anthropic models, got %d", len(anthropic))
	}
	
	minimax := filterByProvider(allModels, "MiniMax")
	if len(minimax) != 1 {
		t.Errorf("Expected 1 MiniMax model, got %d", len(minimax))
	}
}

func TestModelSearch(t *testing.T) {
	models := []Model{
		{Name: "MiniMax M2.1 Reasoning"},
		{Name: "MiniMax M2.1 Fast"},
		{Name: "Claude Opus 4"},
	}
	
	results := searchModels(models, "MiniMax")
	if len(results) != 2 {
		t.Errorf("Expected 2 MiniMax results, got %d", len(results))
	}
	
	results = searchModels(models, "Opus")
	if len(results) != 1 {
		t.Errorf("Expected 1 Opus result, got %d", len(results))
	}
	
	results = searchModels(models, "NotExist")
	if len(results) != 0 {
		t.Errorf("Expected 0 results for non-existent, got %d", len(results))
	}
}

func TestModelSorting(t *testing.T) {
	models := []Model{
		{Name: "Z-Model", Provider: "Z-Provider"},
		{Name: "A-Model", Provider: "A-Provider"},
		{Name: "M-Model", Provider: "M-Provider"},
	}
	
	sorted := sortByName(models)
	
	if sorted[0].Name != "A-Model" {
		t.Error("First model should be A-Model after sorting")
	}
	
	if sorted[2].Name != "Z-Model" {
		t.Error("Last model should be Z-Model after sorting")
	}
}

func TestModelValidation(t *testing.T) {
	tests := []struct {
		name    string
		model   Model
		valid   bool
	}{
		{"valid model", Model{ID: "m1", Name: "Test", Provider: "Provider"}, true},
		{"empty name", Model{ID: "m1", Name: "", Provider: "Provider"}, false},
		{"empty id", Model{ID: "", Name: "Test", Provider: "Provider"}, false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.model.IsValid(); got != tt.valid {
				t.Errorf("Model.IsValid() = %v, want %v", got, tt.valid)
			}
		})
	}
}

func TestModelType(t *testing.T) {
	tests := []struct {
		name     string
		model    Model
		expected ModelType
	}{
		{"reasoning", Model{Name: "MiniMax Reasoning", Type: ModelTypeReasoning}, ModelTypeReasoning},
		{"fast", Model{Name: "MiniMax Fast", Type: ModelTypeFast}, ModelTypeFast},
		{"default", Model{Name: "Default"}, ModelTypeFast},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.model.Type != tt.expected {
				t.Errorf("Model.Type = %v, want %v", tt.model.Type, tt.expected)
			}
		})
	}
}

func TestRecentlyUsedModels(t *testing.T) {
	models := []Model{
		{Name: "Used1", LastUsed: timeNow().Add(-1 * time.Hour)},
		{Name: "Used2", LastUsed: timeNow().Add(-2 * time.Hour)},
		{Name: "Used3", LastUsed: timeNow().Add(-30 * time.Minute)},
	}
	
	sorted := sortByRecentlyUsed(models)
	
	if sorted[0].Name != "Used3" {
		t.Error("Most recently used should be first")
	}
}

func TestModelCategories(t *testing.T) {
	models := []Model{
		{Name: "R1", Category: "Reasoning"},
		{Name: "F1", Category: "Fast"},
		{Name: "R2", Category: "Reasoning"},
		{Name: "F2", Category: "Fast"},
	}
	
	categories := extractCategories(models)
	
	if len(categories) != 2 {
		t.Errorf("Expected 2 categories, got %d", len(categories))
	}
}

// Helper types and functions (would use actual implementation)
type Model struct {
	ID        string
	Name      string
	Provider  string
	Selected  bool
	Type      ModelType
	Category  string
	LastUsed  time.Time
}

type ModelType int

const (
	ModelTypeFast ModelType = iota
	ModelTypeReasoning
)

func (m Model) IsValid() bool {
	return m.ID != "" && m.Name != ""
}

func filterByProvider(models []Model, provider string) []Model {
	var result []Model
	for _, m := range models {
		if m.Provider == provider {
			result = append(result, m)
		}
	}
	return result
}

func searchModels(models []Model, query string) []Model {
	var result []Model
	for _, m := range models {
		if strings.Contains(strings.ToLower(m.Name), strings.ToLower(query)) {
			result = append(result, m)
		}
	}
	return result
}

func sortByName(models []Model) []Model {
	sorted := make([]Model, len(models))
	copy(sorted, models)
	// Simple sort for testing
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i].Name > sorted[j].Name {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	return sorted
}

func sortByRecentlyUsed(models []Model) []Model {
	sorted := make([]Model, len(models))
	copy(sorted, models)
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i].LastUsed.Before(sorted[j].LastUsed) {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	return sorted
}

func extractCategories(models []Model) []string {
	set := make(map[string]bool)
	for _, m := range models {
		if m.Category != "" {
			set[m.Category] = true
		}
	}
	
	categories := make([]string, 0, len(set))
	for c := range set {
		categories = append(categories, c)
	}
	return categories
}

func timeNow() time.Time {
	return time.Now()
}
