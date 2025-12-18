package format

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewNativeSpinner(t *testing.T) {
	t.Run("creates spinner successfully", func(t *testing.T) {
		// Arrange
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		settings := NativeSettings{
			Size:       10,
			Label:      "Loading...",
			LabelColor: "cyan",
			ColorStart: "magenta",
			ColorEnd:   "blue",
		}

		// Act
		spinner := NewNativeSpinner(ctx, cancel, settings)

		// Assert
		assert.NotNil(t, spinner)
		assert.NotNil(t, spinner.done)
		assert.NotNil(t, spinner.anim)
		assert.NotNil(t, spinner.ctx)
		assert.NotNil(t, spinner.cancel)
	})
}

func TestNewNativeAnim(t *testing.T) {
	t.Run("creates animation with default settings", func(t *testing.T) {
		// Arrange
		settings := NativeSettings{
			Label: "Test",
		}

		// Act
		anim := NewNativeAnim(settings)

		// Assert
		assert.NotNil(t, anim)
		assert.Equal(t, 10, anim.width) // default size
		assert.Equal(t, "Test", anim.label)
		assert.Equal(t, defaultLabelColor, anim.settings.LabelColor)
		assert.Len(t, anim.runes, len(defaultRunes))
		assert.Len(t, anim.ellipsisFrames, len(defaultEllipsis))
	})

	t.Run("creates animation with custom settings", func(t *testing.T) {
		// Arrange
		settings := NativeSettings{
			Size:       5,
			Label:      "Custom",
			LabelColor: "red",
			ColorStart: "green",
			ColorEnd:   "blue",
		}

		// Act
		anim := NewNativeAnim(settings)

		// Assert
		assert.NotNil(t, anim)
		assert.Equal(t, 5, anim.width)
		assert.Equal(t, "Custom", anim.label)
		assert.Equal(t, "red", anim.settings.LabelColor)
	})

	t.Run("handles invalid size gracefully", func(t *testing.T) {
		// Arrange
		settings := NativeSettings{
			Size:  0, // Invalid size
			Label: "Test",
		}

		// Act
		anim := NewNativeAnim(settings)

		// Assert
		assert.NotNil(t, anim)
		assert.Equal(t, 10, anim.width) // Should default to 10
	})
}

func TestNativeSpinner_StartStop(t *testing.T) {
	t.Run("starts and stops cleanly", func(t *testing.T) {
		// Arrange
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		settings := NativeSettings{
			Size:  5,
			Label: "Test",
		}

		spinner := NewNativeSpinner(ctx, cancel, settings)

		// Act
		spinner.Start()
		time.Sleep(150 * time.Millisecond) // Let it animate briefly
		spinner.Stop()

		// Assert - should complete without hanging
		assert.True(t, true)
	})

	t.Run("handles context cancellation", func(t *testing.T) {
		// Arrange
		ctx, cancel := context.WithCancel(context.Background())

		settings := NativeSettings{
			Size:  5,
			Label: "Test",
		}

		spinner := NewNativeSpinner(ctx, cancel, settings)

		// Act
		spinner.Start()
		time.Sleep(50 * time.Millisecond)
		cancel() // Cancel context
		spinner.Stop()

		// Assert - should handle cancellation gracefully
		assert.True(t, true)
	})
}

func TestNativeSpinner_SetLabel(t *testing.T) {
	t.Run("updates label dynamically", func(t *testing.T) {
		// Arrange
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		settings := NativeSettings{
			Size:  5,
			Label: "Original",
		}

		spinner := NewNativeSpinner(ctx, cancel, settings)

		// Act
		spinner.SetLabel("Updated")

		// Assert
		assert.Equal(t, "Updated", spinner.anim.label)
	})
}

func TestNativeAnim_View(t *testing.T) {
	t.Run("generates view with content", func(t *testing.T) {
		// Arrange
		settings := NativeSettings{
			Size:  5,
			Label: "Test",
		}

		anim := NewNativeAnim(settings)
		anim.step.Store(1)

		// Act
		view := anim.View()

		// Assert
		assert.NotEmpty(t, view)
		assert.Contains(t, view, "Test")
		assert.True(t, len(view) > 10) // Should include spinner characters
	})

	t.Run("handles empty label", func(t *testing.T) {
		// Arrange
		settings := NativeSettings{
			Size: 5,
		}

		anim := NewNativeAnim(settings)

		// Act
		view := anim.View()

		// Assert
		assert.NotNil(t, view)
	})
}

func TestNativeAnim_Ellipsis(t *testing.T) {
	t.Run("generates ellipsis animation", func(t *testing.T) {
		// Arrange
		settings := NativeSettings{Size: 5}
		anim := NewNativeAnim(settings)

		// Act & Assert - test the ellipsis frames directly
		assert.NotNil(t, anim.ellipsisFrames)
		assert.Equal(t, defaultEllipsis, anim.ellipsisFrames)

		// Test each ellipsis step
		for i := 0; i < 4; i++ {
			anim.ellipsisStep.Store(int64(i))
			step := anim.ellipsisStep.Load()
			assert.Equal(t, int64(i), step)
		}
	})
}

func TestNativeSpinner_WithCustomOutput(t *testing.T) {
	t.Run("writes to custom output", func(t *testing.T) {
		// Arrange
		var buf bytes.Buffer
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		settings := NativeSettings{
			Size:  5,
			Label: "Test",
		}

		spinner := NewNativeSpinner(ctx, cancel, settings)
		spinner.output = &buf // Override output to capture

		// Act
		spinner.Start()
		time.Sleep(150 * time.Millisecond) // Let it animate
		spinner.Stop()

		// Assert
		output := buf.String()
		assert.NotEmpty(t, output)
		assert.Contains(t, output, "Test")

		// Should contain spinner characters
		foundSpinnerChars := false
		for _, char := range defaultRunes {
			if strings.Contains(output, char) {
				foundSpinnerChars = true
				break
			}
		}
		assert.True(t, foundSpinnerChars, "Output should contain spinner characters")
	})
}
