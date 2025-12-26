package splash

import "testing"

func TestSplashCreation(t *testing.T) {
	s := New()
	if s == nil {
		t.Fatal("Expected splash")
	}
}
