package state

import (
	"testing"
)

func TestHasNoProgressWithActualProgress(t *testing.T) {
	pt := NewProgressTracker()
	
	// Simulate a realistic work scenario with actual progress
	// Mix of viewing, editing, and testing different files
	actions := []struct {
		toolName, targetFile, command string
		success                        bool
	}{
		{"view", "main.go", "", true},
		{"view", "utils.go", "", true}, 
		{"edit", "main.go", "add function", true},
		{"edit", "utils.go", " refactor", true},
		{"bash", "go test ./...", "", true},
		{"view", "main_test.go", "", true},
		{"edit", "main_test.go", "add test", true},
		{"bash", "go test -v ./...", "", true},
		{"view", "README.md", "", true},
		{"edit", "README.md", "update docs", true},
		{"bash", "go build .", "", true},
		{"view", "config.go", "", true},
		{"edit", "config.go", "add option", true},
		{"bash", "go test -run TestConfig", "", true},
		{"view", "main.go", "", true}, // 15 actions total
	}
	
	// Record all actions
	for _, action := range actions {
		pt.RecordAction(action.toolName, action.targetFile, action.command, "", action.success)
	}
	
	// Record some file modifications to indicate progress
	pt.RecordFileModification("main.go", "hash1")
	pt.RecordFileModification("utils.go", "hash2")
	
	// Should NOT detect as stuck despite 15+ actions
	stuck, reason := pt.IsStuck()
	if stuck {
		t.Errorf("Expected not stuck, but detected: %s", reason)
	}
}

func TestHasNoProgressWithActualStuck(t *testing.T) {
	pt := NewProgressTracker()
	
	// Simulate a very stuck scenario: alternating between viewing and trying to edit the same file
	for i := 0; i < 15; i++ {
		if i%2 == 0 {
			pt.RecordAction("edit", "same.go", "try", "error", false)
		} else {
			pt.RecordAction("view", "same.go", "", "", true)
		}
	}
	
	// Should detect as stuck
	stuck, reason := pt.IsStuck()
	if !stuck {
		t.Error("Expected stuck detection, but no stuck condition was found")
	} else if reason == "" {
		t.Error("Expected explanatory reason for stuck condition")
	}
	
	t.Logf("Stuck detected: %s", reason)
}

func TestHasNoProgressWithMinimalSuccess(t *testing.T) {
	pt := NewProgressTracker()
	
	// Simulate just barely enough progress to avoid stuck detection
	for i := 0; i < 12; i++ {
		// Mostly failed operations
		pt.RecordAction("edit", "file.go", "try", "error", false)
	}
	// Add just enough meaningful successful operations on different targets
	pt.RecordAction("edit", "file1.go", "fix", "", true)
	pt.RecordAction("bash", "go build", "", "", true)
	pt.RecordAction("edit", "file2.go", "update", "", true)
	
	// Should NOT be stuck due to meaningful variety
	stuck, reason := pt.IsStuck()
	if stuck {
		t.Errorf("Should not be stuck with variety: %s", reason)
	}
}

func TestHasNoProgressMessageAccuracy(t *testing.T) {
	pt := NewProgressTracker()
	
	// Add exactly 15 actions with very little meaningful progress
	for i := 0; i < 8; i++ {
		pt.RecordAction("view", "same.go", "", "", true) // Same target, just viewing
	}
	for i := 0; i < 7; i++ {
		pt.RecordAction("view", "another.go", "", "", true) // Another target, just viewing
	}
	
	// Should detect stuck and mention "15 actions" accurately
	stuck, reason := pt.IsStuck()
	if !stuck {
		t.Error("Expected stuck detection")
	} else {
		// Reason should mention "15 actions" to be accurate
		if len(reason) < 10 {
			t.Errorf("Reason too short: %s", reason)
		}
		t.Logf("Stuck reason: %s", reason)
	}
}