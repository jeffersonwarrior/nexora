// Package banner provides an animated notification banner component for success/error/info messages.
package banner

import (
	"image/color"
	"sync"
	"sync/atomic"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/nexora/nexora/internal/tui/components/core"
	"github.com/nexora/nexora/internal/tui/styles"
	"github.com/nexora/nexora/internal/tui/util"
)

// BannerType represents the type of banner notification
type BannerType int

const (
	BannerSuccess BannerType = iota
	BannerError
	BannerInfo
)

// Default timeout for auto-dismiss
const defaultTimeout = 3 * time.Second

// ShowBannerMsg triggers the banner to show with a message
type ShowBannerMsg struct {
	Type    BannerType
	Message string
	Timeout time.Duration
}

// HideBannerMsg triggers the banner to hide
type HideBannerMsg struct{}

// AgentCompletionMsg is sent when an agent completes a task
type AgentCompletionMsg struct {
	Success bool
	Message string
}

// hideTimeoutMsg is an internal message for auto-dismiss
type hideTimeoutMsg struct {
	id int
}

// Internal ID management for timeout messages
var lastID int64

func nextID() int {
	return int(atomic.AddInt64(&lastID, 1))
}

// Banner is the public interface for the banner component
type Banner interface {
	util.Model
	core.Sizeable
}

// bannerCmp implements the Banner interface
type bannerCmp struct {
	mu         sync.RWMutex
	width      int
	visible    bool
	message    string
	bannerType BannerType
	id         int
}

// New creates a new banner component
func New() Banner {
	return &bannerCmp{
		visible: false,
		id:      nextID(),
	}
}

// Init initializes the banner component
func (b *bannerCmp) Init() tea.Cmd {
	return nil
}

// Update handles incoming messages
func (b *bannerCmp) Update(msg tea.Msg) (util.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ShowBannerMsg:
		b.mu.Lock()
		b.visible = true
		b.message = msg.Message
		b.bannerType = msg.Type
		b.id = nextID()
		timeout := msg.Timeout
		if timeout == 0 {
			timeout = defaultTimeout
		}
		b.mu.Unlock()

		// Return command to auto-hide after timeout
		return b, b.hideAfter(timeout)

	case HideBannerMsg:
		b.mu.Lock()
		b.visible = false
		b.mu.Unlock()
		return b, nil

	case hideTimeoutMsg:
		// Only hide if the ID matches (prevents stale timeout messages)
		b.mu.Lock()
		if msg.id == b.id {
			b.visible = false
		}
		b.mu.Unlock()
		return b, nil

	case AgentCompletionMsg:
		// Convert agent completion to banner message
		var bannerType BannerType
		if msg.Success {
			bannerType = BannerSuccess
		} else {
			bannerType = BannerError
		}

		showMsg := ShowBannerMsg{
			Type:    bannerType,
			Message: msg.Message,
			Timeout: defaultTimeout,
		}
		return b.Update(showMsg)
	}

	return b, nil
}

// View renders the banner
func (b *bannerCmp) View() string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if !b.visible {
		return ""
	}

	t := styles.CurrentTheme()

	// Choose colors and icons based on banner type
	var (
		bgColor color.Color
		fgColor color.Color
		icon    string
	)

	switch b.bannerType {
	case BannerSuccess:
		bgColor = t.Success
		fgColor = t.BgBase
		icon = "✓"
	case BannerError:
		bgColor = t.Error
		fgColor = t.BgBase
		icon = "✗"
	case BannerInfo:
		bgColor = t.Info
		fgColor = t.BgBase
		icon = "ℹ"
	}

	// Create banner style
	bannerStyle := lipgloss.NewStyle().
		Background(bgColor).
		Foreground(fgColor).
		Bold(true).
		Padding(0, 2).
		Width(b.width)

	content := icon + " " + b.message

	return bannerStyle.Render(content)
}

// SetSize sets the banner width
func (b *bannerCmp) SetSize(width, height int) tea.Cmd {
	b.mu.Lock()
	b.width = width
	b.mu.Unlock()
	return nil
}

// GetSize returns the banner dimensions
func (b *bannerCmp) GetSize() (int, int) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if !b.visible {
		return b.width, 0
	}
	return b.width, 1
}

// hideAfter creates a command that hides the banner after a timeout
func (b *bannerCmp) hideAfter(d time.Duration) tea.Cmd {
	b.mu.RLock()
	id := b.id
	b.mu.RUnlock()
	return tea.Tick(d, func(time.Time) tea.Msg {
		return hideTimeoutMsg{id: id}
	})
}
