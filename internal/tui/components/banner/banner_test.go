package banner

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	b := New()
	if b == nil {
		t.Fatal("New() returned nil")
	}
}

func TestBannerInit(t *testing.T) {
	b := New()
	cmd := b.Init()
	if cmd != nil {
		t.Error("Init() should return nil when banner is hidden")
	}
}

func TestBannerShow(t *testing.T) {
	b := New().(*bannerCmp)

	// Test success banner
	msg := ShowBannerMsg{
		Type:    BannerSuccess,
		Message: "Test success",
		Timeout: 2 * time.Second,
	}

	_, cmd := b.Update(msg)
	if cmd == nil {
		t.Error("Update with ShowBannerMsg should return a command")
	}

	if !b.visible {
		t.Error("Banner should be visible after ShowBannerMsg")
	}

	if b.message != "Test success" {
		t.Errorf("Expected message 'Test success', got '%s'", b.message)
	}

	if b.bannerType != BannerSuccess {
		t.Errorf("Expected BannerSuccess type, got %v", b.bannerType)
	}
}

func TestBannerHide(t *testing.T) {
	b := New().(*bannerCmp)

	// First show the banner
	showMsg := ShowBannerMsg{
		Type:    BannerSuccess,
		Message: "Test",
		Timeout: 100 * time.Millisecond,
	}
	b.Update(showMsg)

	// Then hide it
	hideMsg := HideBannerMsg{}
	_, cmd := b.Update(hideMsg)

	if b.visible {
		t.Error("Banner should be hidden after HideBannerMsg")
	}

	if cmd != nil {
		t.Error("HideBannerMsg should not return a command")
	}
}

func TestBannerAutoHide(t *testing.T) {
	b := New().(*bannerCmp)

	// Show banner with timeout
	showMsg := ShowBannerMsg{
		Type:    BannerError,
		Message: "Test error",
		Timeout: 50 * time.Millisecond,
	}

	b.Update(showMsg)

	// Wait for timeout
	time.Sleep(100 * time.Millisecond)

	// Process timeout message
	timeoutMsg := hideTimeoutMsg{id: b.id}
	_, _ = b.Update(timeoutMsg)

	if b.visible {
		t.Error("Banner should auto-hide after timeout")
	}
}

func TestBannerView(t *testing.T) {
	b := New().(*bannerCmp)

	// Hidden banner should render empty
	if b.View() != "" {
		t.Error("Hidden banner should render empty string")
	}

	// Show banner
	showMsg := ShowBannerMsg{
		Type:    BannerInfo,
		Message: "Info message",
		Timeout: 2 * time.Second,
	}
	b.Update(showMsg)

	view := b.View()
	if view == "" {
		t.Error("Visible banner should render non-empty string")
	}
}

func TestBannerTypes(t *testing.T) {
	testCases := []struct {
		name       string
		bannerType BannerType
	}{
		{"Success", BannerSuccess},
		{"Error", BannerError},
		{"Info", BannerInfo},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b := New().(*bannerCmp)
			msg := ShowBannerMsg{
				Type:    tc.bannerType,
				Message: "Test " + tc.name,
				Timeout: 1 * time.Second,
			}

			b.Update(msg)

			if b.bannerType != tc.bannerType {
				t.Errorf("Expected type %v, got %v", tc.bannerType, b.bannerType)
			}
		})
	}
}

func TestBannerSetSize(t *testing.T) {
	b := New().(*bannerCmp)

	cmd := b.SetSize(100, 50)
	if cmd != nil {
		t.Error("SetSize should return nil")
	}

	if b.width != 100 {
		t.Errorf("Expected width 100, got %d", b.width)
	}
}

func TestBannerRaceConditions(t *testing.T) {
	b := New()

	// Concurrent show/hide operations
	done := make(chan bool)

	go func() {
		for i := 0; i < 100; i++ {
			msg := ShowBannerMsg{
				Type:    BannerSuccess,
				Message: "Test",
				Timeout: 1 * time.Millisecond,
			}
			b.Update(msg)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			b.Update(HideBannerMsg{})
		}
		done <- true
	}()

	<-done
	<-done
}

func TestBannerAgentCompletionSuccess(t *testing.T) {
	b := New().(*bannerCmp)

	msg := AgentCompletionMsg{
		Success: true,
		Message: "Task completed successfully",
	}

	_, cmd := b.Update(msg)
	if cmd == nil {
		t.Error("AgentCompletionMsg should trigger banner display")
	}

	if !b.visible {
		t.Error("Banner should be visible after AgentCompletionMsg")
	}

	if b.bannerType != BannerSuccess {
		t.Error("Successful completion should show success banner")
	}
}

func TestBannerAgentCompletionError(t *testing.T) {
	b := New().(*bannerCmp)

	msg := AgentCompletionMsg{
		Success: false,
		Message: "Task failed with error",
	}

	_, cmd := b.Update(msg)
	if cmd == nil {
		t.Error("AgentCompletionMsg should trigger banner display")
	}

	if !b.visible {
		t.Error("Banner should be visible after AgentCompletionMsg")
	}

	if b.bannerType != BannerError {
		t.Error("Failed completion should show error banner")
	}
}
