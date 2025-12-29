package anim

import (
	"image/color"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

func TestNewAnim(t *testing.T) {
	tests := []struct {
		name  string
		opts  Settings
		check func(*Anim) bool
	}{
		{
			name: "default settings",
			opts: Settings{},
			check: func(a *Anim) bool {
				return a != nil && a.cyclingCharWidth == defaultNumCyclingChars
			},
		},
		{
			name: "custom size",
			opts: Settings{Size: 20},
			check: func(a *Anim) bool {
				return a != nil && a.cyclingCharWidth == 20
			},
		},
		{
			name: "custom label",
			opts: Settings{Label: "Loading..."},
			check: func(a *Anim) bool {
				return a != nil && a.labelWidth > 0
			},
		},
		{
			name: "custom colors",
			opts: Settings{
				GradColorA: color.RGBA{R: 255, G: 0, B: 0, A: 255},
				GradColorB: color.RGBA{R: 0, G: 0, B: 255, A: 255},
				LabelColor: color.RGBA{R: 255, G: 255, B: 255, A: 255},
			},
			check: func(a *Anim) bool {
				return a != nil && a.labelColor != nil
			},
		},
		{
			name: "zero size defaults to default",
			opts: Settings{Size: 0},
			check: func(a *Anim) bool {
				return a != nil && a.cyclingCharWidth == defaultNumCyclingChars
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			anim := New(tt.opts)
			if !tt.check(anim) {
				t.Errorf("New() failed validation")
			}
		})
	}
}

func TestAnimInit(t *testing.T) {
	anim := New(Settings{Label: "test"})
	cmd := anim.Init()

	// Init should return a command (Step)
	if cmd == nil {
		t.Error("Init() returned nil, expected a command")
	}
}

func TestAnimView(t *testing.T) {
	tests := []struct {
		name  string
		opts  Settings
		check func(string) bool
	}{
		{
			name: "with label",
			opts: Settings{Label: "test"},
			check: func(s string) bool {
				return len(s) > 0
			},
		},
		{
			name: "without label",
			opts: Settings{},
			check: func(s string) bool {
				return len(s) > 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			anim := New(tt.opts)
			view := anim.View()
			if !tt.check(view) {
				t.Errorf("View() returned invalid output: %q", view)
			}
		})
	}
}

func TestAnimUpdateStepMsg(t *testing.T) {
	anim := New(Settings{Label: "test"})
	initialID := anim.id

	// Test StepMsg for this instance
	msg := StepMsg{id: initialID}
	result, cmd := anim.Update(msg)

	// Result should be the anim itself
	if result != anim {
		t.Error("Update() returned wrong model")
	}

	// Should return a command (Step)
	if cmd == nil {
		t.Error("Update(StepMsg) returned nil, expected a command")
	}

	// Step should have incremented
	if anim.step.Load() != 1 {
		t.Errorf("step not incremented, expected 1, got %d", anim.step.Load())
	}
}

func TestAnimUpdateWrongID(t *testing.T) {
	anim := New(Settings{})
	initialStep := anim.step.Load()

	// Test StepMsg with wrong ID
	msg := StepMsg{id: anim.id + 1}
	result, cmd := anim.Update(msg)

	// Should return unchanged model
	if result != anim {
		t.Error("Update() returned wrong model")
	}

	// Should return nil command
	if cmd != nil {
		t.Error("Update(WrongID) should return nil command")
	}

	// Step should not have changed
	if anim.step.Load() != initialStep {
		t.Error("step should not have changed for wrong ID")
	}
}

func TestAnimUpdateOtherMsg(t *testing.T) {
	anim := New(Settings{})

	// Test with arbitrary message (use nil as unknown message type)
	var msg tea.Msg
	msg = tea.BatchMsg{}
	result, cmd := anim.Update(msg)

	// Should return unchanged model
	if result != anim {
		t.Error("Update() returned wrong model")
	}

	// Should return nil command
	if cmd != nil {
		t.Error("Update(BatchMsg) should return nil command")
	}
}

func TestAnimSetLabel(t *testing.T) {
	tests := []struct {
		name     string
		oldLabel string
		newLabel string
		check    func(*Anim) bool
	}{
		{
			name:     "change label",
			oldLabel: "old",
			newLabel: "new",
			check: func(a *Anim) bool {
				return a.labelWidth > 0
			},
		},
		{
			name:     "empty label",
			oldLabel: "test",
			newLabel: "",
			check: func(a *Anim) bool {
				return a.labelWidth == 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			anim := New(Settings{Label: tt.oldLabel})
			anim.SetLabel(tt.newLabel)
			if !tt.check(anim) {
				t.Errorf("SetLabel() failed validation")
			}
		})
	}
}

func TestAnimWidth(t *testing.T) {
	tests := []struct {
		name  string
		opts  Settings
		check func(int) bool
	}{
		{
			name: "with label",
			opts: Settings{Size: 10, Label: "test"},
			check: func(w int) bool {
				return w > 10
			},
		},
		{
			name: "without label",
			opts: Settings{Size: 10},
			check: func(w int) bool {
				return w == 10
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			anim := New(tt.opts)
			width := anim.Width()
			if !tt.check(width) {
				t.Errorf("Width() returned %d, failed validation", width)
			}
		})
	}
}

func TestAnimStep(t *testing.T) {
	anim := New(Settings{})
	cmd := anim.Step()

	if cmd == nil {
		t.Error("Step() returned nil, expected a command")
	}
}

func TestColorIsUnset(t *testing.T) {
	tests := []struct {
		name     string
		color    color.Color
		expected bool
	}{
		{
			name:     "nil color",
			color:    nil,
			expected: true,
		},
		{
			name:     "color with alpha 0",
			color:    color.RGBA{R: 255, G: 0, B: 0, A: 0},
			expected: true,
		},
		{
			name:     "color with alpha > 0",
			color:    color.RGBA{R: 255, G: 0, B: 0, A: 255},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := colorIsUnset(tt.color)
			if result != tt.expected {
				t.Errorf("colorIsUnset() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMakeGradientRamp(t *testing.T) {
	tests := []struct {
		name  string
		size  int
		stops []color.Color
		check func([]color.Color) bool
	}{
		{
			name: "two colors",
			size: 10,
			stops: []color.Color{
				color.RGBA{R: 255, G: 0, B: 0, A: 255},
				color.RGBA{R: 0, G: 0, B: 255, A: 255},
			},
			check: func(c []color.Color) bool {
				return len(c) == 10
			},
		},
		{
			name: "three colors",
			size: 15,
			stops: []color.Color{
				color.RGBA{R: 255, G: 0, B: 0, A: 255},
				color.RGBA{R: 0, G: 255, B: 0, A: 255},
				color.RGBA{R: 0, G: 0, B: 255, A: 255},
			},
			check: func(c []color.Color) bool {
				return len(c) == 15
			},
		},
		{
			name: "one stop",
			size: 5,
			stops: []color.Color{
				color.RGBA{R: 255, G: 0, B: 0, A: 255},
			},
			check: func(c []color.Color) bool {
				return len(c) == 0
			},
		},
		{
			name:  "no stops",
			size:  5,
			stops: []color.Color{},
			check: func(c []color.Color) bool {
				return len(c) == 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := makeGradientRamp(tt.size, tt.stops...)
			if !tt.check(result) {
				t.Errorf("makeGradientRamp() returned %d colors, unexpected", len(result))
			}
		})
	}
}

func TestAnimConcurrency(t *testing.T) {
	// Simple concurrency test: create multiple anims and update them
	anim1 := New(Settings{Label: "task1"})
	anim2 := New(Settings{Label: "task2"})

	// Both should have different IDs
	if anim1.id == anim2.id {
		t.Error("two Anim instances should have different IDs")
	}

	// Update both with each other's IDs - should not affect the other
	msg1 := StepMsg{id: anim1.id}
	msg2 := StepMsg{id: anim2.id}

	anim1.Update(msg1)
	anim2.Update(msg2)

	step1 := anim1.step.Load()
	step2 := anim2.step.Load()

	if step1 != 1 || step2 != 1 {
		t.Errorf("expected both steps to be 1, got %d and %d", step1, step2)
	}
}

func TestAnimStartTime(t *testing.T) {
	anim := New(Settings{})
	now := time.Now()

	// Start time should be roughly now
	timeDiff := now.Sub(anim.startTime)
	if timeDiff > 100*time.Millisecond || timeDiff < -100*time.Millisecond {
		t.Errorf("startTime not set correctly")
	}
}
