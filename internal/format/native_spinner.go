// Package format provides native text formatting utilities without Charm dependencies.
package format

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

// NativeSpinner wraps the native spinner implementation for non-interactive mode.
type NativeSpinner struct {
	done    chan struct{}
	ctx     context.Context
	cancel  context.CancelFunc
	anim    *NativeAnim
	output  io.Writer
}

// NativeAnim provides an animated spinner without Charm dependencies.
type NativeAnim struct {
	width          int
	step           atomic.Int64
	ellipsisStep   atomic.Int64
	label          string
	runes          []string
	ellipsisFrames []string
	startTime      time.Time
	initialized    atomic.Bool
	settings       NativeSettings
}

// NativeSettings defines settings for the native animation.
type NativeSettings struct {
	Size       int
	Label      string
	LabelColor string
	ColorStart string
	ColorEnd   string
}

// Default native settings.
var (
	defaultRunes        = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	defaultEllipsis    = []string{".", "..", "...", ""}
	defaultLabelColor  = "\033[36m" // Cyan
	defaultSpinnerColor = "\033[35m" // Magenta
	resetColor         = "\033[0m"
)

// NewNativeSpinner creates a new spinner with the given message using native implementation.
func NewNativeSpinner(ctx context.Context, cancel context.CancelFunc, settings NativeSettings) *NativeSpinner {
	anim := NewNativeAnim(settings)
	
	return &NativeSpinner{
		done:   make(chan struct{}, 1),
		ctx:    ctx,
		cancel: cancel,
		anim:   anim,
		output: os.Stderr,
	}
}

// NewNativeAnim creates a new native Anim instance with the specified settings.
func NewNativeAnim(settings NativeSettings) *NativeAnim {
	if settings.Size < 1 {
		settings.Size = 10
	}
	if settings.LabelColor == "" {
		settings.LabelColor = defaultLabelColor
	}
	if settings.ColorStart == "" {
		settings.ColorStart = defaultSpinnerColor
	}
	
	anim := &NativeAnim{
		width:          settings.Size,
		label:          settings.Label,
		runes:          defaultRunes,
		ellipsisFrames: defaultEllipsis,
		startTime:      time.Now(),
		settings:       settings,
	}
	
	return anim
}

// Start begins the spinner animation.
func (s *NativeSpinner) Start() {
	go func() {
		defer close(s.done)
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		
		for {
			select {
			case <-s.ctx.Done():
				return
			case <-ticker.C:
				s.render()
			}
		}
	}()
}

// Stop ends the spinner animation.
func (s *NativeSpinner) Stop() {
	s.cancel()
	<-s.done
	// Clear the line
	fmt.Fprintf(s.output, "\r\033[K")
}

// SetLabel updates the label text.
func (s *NativeSpinner) SetLabel(label string) {
	s.anim.SetLabel(label)
}

// render writes the current animation frame to the output.
func (s *NativeSpinner) render() {
	frame := s.anim.View()
	fmt.Fprintf(s.output, "\r%s", frame)
}

// SetLabel updates the label text.
func (a *NativeAnim) SetLabel(newLabel string) {
	a.label = newLabel
}

// View renders the current state of the animation.
func (a *NativeAnim) View() string {
	step := int(a.step.Load())
	ellipsisStep := int(a.ellipsisStep.Load())
	
	var builder strings.Builder
	
	// Add spinner character
	spinnerChar := a.runes[step%len(a.runes)]
	builder.WriteString(a.settings.ColorStart)
	builder.WriteString(spinnerChar)
	builder.WriteString(resetColor)
	builder.WriteString(" ")
	
	// Add label
	if a.label != "" {
		builder.WriteString(a.settings.LabelColor)
		builder.WriteString(a.label)
		builder.WriteString(resetColor)
		
		// Add ellipsis animation after initialization
		if time.Since(a.startTime) >= time.Second {
			if ellipsisStep/4 < len(a.ellipsisFrames) {
				builder.WriteString(a.ellipsisFrames[ellipsisStep/4])
			}
		}
	}
	
	return builder.String()
}

// Step advances the animation one step.
func (a *NativeAnim) Step() {
	current := a.step.Add(1)
	if time.Since(a.startTime) >= time.Second {
		a.ellipsisStep.Add(1)
	}
	
	// Ensure step stays within bounds
	if int(current) >= len(a.runes) {
		a.step.Store(0)
	}
	
	// Ensure ellipsis step stays within bounds  
	ellipsisStep := int(a.ellipsisStep.Load())
	if ellipsisStep >= len(a.ellipsisFrames)*4 {
		a.ellipsisStep.Store(0)
	}
}

// Width returns the total width of the animation.
func (a *NativeAnim) Width() int {
	width := len(a.runes[0]) + 1 // spinner + space
	if a.label != "" {
		width += len(a.label) + len(a.ellipsisFrames[0])
	}
	return width
}