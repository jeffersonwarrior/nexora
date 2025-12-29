package delegation

import (
	"testing"
)

func TestMaxConcurrentDelegates(t *testing.T) {
	if MaxConcurrentDelegates != 10 {
		t.Errorf("expected MaxConcurrentDelegates=10, got %d", MaxConcurrentDelegates)
	}
}

func TestCanSpawnDelegate_UnderLimit(t *testing.T) {
	tests := []int{0, 1, 5, 9}
	for _, activeCount := range tests {
		err := CanSpawnDelegate(activeCount)
		if err != nil {
			t.Errorf("CanSpawnDelegate(%d) should allow spawn under limit, got error: %v", activeCount, err)
		}
	}
}

func TestCanSpawnDelegate_AtLimit(t *testing.T) {
	err := CanSpawnDelegate(10)
	if err == nil {
		t.Error("CanSpawnDelegate(10) should return error at limit, got nil")
	}
	expectedMsg := "maximum concurrent delegates reached (10/10)"
	if err != nil && err.Error() != expectedMsg {
		t.Errorf("expected error message %q, got %q", expectedMsg, err.Error())
	}
}

func TestCanSpawnDelegate_OverLimit(t *testing.T) {
	err := CanSpawnDelegate(15)
	if err == nil {
		t.Error("CanSpawnDelegate(15) should return error over limit, got nil")
	}
}
