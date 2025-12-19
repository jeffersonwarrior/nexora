package format

import (
	"context"
	"testing"
	"time"

	"image/color"

	"github.com/nexora/nexora/internal/tui/components/anim"
	"github.com/stretchr/testify/assert"
)

func TestNewSpinner(t *testing.T) {
	t.Run("creates spinner successfully", func(t *testing.T) {
		// Arrange
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		animSettings := anim.Settings{
			Size:        10,
			Label:       "Loading...",
			LabelColor:  color.White,
			GradColorA:  color.NRGBA{R: 255, G: 0, B: 0, A: 255},
			GradColorB:  color.NRGBA{R: 0, G: 0, B: 255, A: 255},
			CycleColors: true,
		}

		// Act
		spinner := NewSpinner(ctx, cancel, animSettings)

		// Assert
		assert.NotNil(t, spinner)
		assert.NotNil(t, spinner.prog)
		assert.NotNil(t, spinner.done)
	})

	t.Run("handles context cancellation", func(t *testing.T) {
		// Arrange
		ctx, cancel := context.WithCancel(context.Background())

		animSettings := anim.Settings{
			Size:        10,
			Label:       "Test",
			LabelColor:  color.White,
			GradColorA:  color.NRGBA{R: 255, G: 0, B: 0, A: 255},
			GradColorB:  color.NRGBA{R: 0, G: 0, B: 255, A: 255},
			CycleColors: false,
		}

		spinner := NewSpinner(ctx, cancel, animSettings)

		// Act
		spinner.Start()
		time.Sleep(10 * time.Millisecond) // Brief moment for animation to start
		cancel()                          // Cancel context
		spinner.Stop()

		// Assert - should complete without hanging/crashing
		assert.True(t, true) // If we get here, the test passed
	})
}

func TestSpinner_StartStop(t *testing.T) {
	t.Run("starts and stops cleanly", func(t *testing.T) {
		// Arrange
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		animSettings := anim.Settings{
			Size:        10,
			Label:       "Loading...",
			LabelColor:  color.White,
			GradColorA:  color.NRGBA{R: 255, G: 0, B: 0, A: 255},
			GradColorB:  color.NRGBA{R: 0, G: 0, B: 255, A: 255},
			CycleColors: false,
		}

		spinner := NewSpinner(ctx, cancel, animSettings)

		// Act & Assert - should not panic
		spinner.Start()
		time.Sleep(50 * time.Millisecond) // Let it spin briefly
		spinner.Stop()

		assert.True(t, true) // Test passes if no panic occurs
	})
}

func TestSpinner_SetLabel(t *testing.T) {
	t.Run("setLabel is a no-op stub", func(t *testing.T) {
		// Arrange
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		animSettings := anim.Settings{
			Size:        10,
			Label:       "Loading...",
			LabelColor:  color.White,
			GradColorA:  color.NRGBA{R: 255, G: 0, B: 0, A: 255},
			GradColorB:  color.NRGBA{R: 0, G: 0, B: 255, A: 255},
			CycleColors: false,
		}

		spinner := NewSpinner(ctx, cancel, animSettings)

		// Act & Assert - should not panic
		spinner.SetLabel("test label")

		assert.True(t, true) // Test passes if no panic occurs
	})
}

func TestNewSpinnerInterface(t *testing.T) {
	t.Run("creates spinner interface successfully", func(t *testing.T) {
		// Arrange
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		settings := Settings{
			Size:       1,
			Label:      "Loading...",
			LabelColor: "blue",
			ColorStart: "cyan",
			ColorEnd:   "blue",
		}

		// Act
		spinner := NewSpinnerInterface(ctx, cancel, settings)

		// Assert
		assert.NotNil(t, spinner)

		// Verify interface compliance
		assert.Implements(t, (*SpinnerInterface)(nil), spinner)
	})
}

func TestSpinnerInterface_Lifecycle(t *testing.T) {
	t.Run("full lifecycle works", func(t *testing.T) {
		// Arrange
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		settings := Settings{
			Size:       1,
			Label:      "Test",
			LabelColor: "blue",
			ColorStart: "cyan",
			ColorEnd:   "blue",
		}

		spinner := NewSpinnerInterface(ctx, cancel, settings)

		// Act & Assert - full lifecycle
		spinner.Start()
		time.Sleep(50 * time.Millisecond)
		spinner.SetLabel("Updated Label")
		time.Sleep(50 * time.Millisecond)
		spinner.Stop()

		assert.True(t, true) // Test passes if no panic occurs
	})
}
