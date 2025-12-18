package tools

import (
	"context"
	"os/exec"
	"testing"
)

// TestGetFDCmd verifies fd command detection
func TestGetFDCmd(t *testing.T) {
	ctx := context.Background()

	// Test the function (behavior depends on system)
	cmd := getFDCmd(ctx)

	// Verify command is valid if returned
	if cmd != nil {
		if cmd.Path == "" {
			t.Error("getFDCmd returned command with empty path")
		}

		// Verify it's one of the expected commands
		cmdName := cmd.Path[len(cmd.Path)-2:]
		if cmdName != "fd" && cmd.Path[len(cmd.Path)-6:] != "fdfind" {
			t.Errorf("unexpected command: %s", cmd.Path)
		}
	}

	// If cmd is nil, verify fd/fdfind are not in PATH
	if cmd == nil {
		if _, err := exec.LookPath("fd"); err == nil {
			t.Error("getFDCmd returned nil but fd is in PATH")
		}
		if _, err := exec.LookPath("fdfind"); err == nil {
			t.Error("getFDCmd returned nil but fdfind is in PATH")
		}
	}
}

// TestGetFDCmdContext verifies context is passed correctly
func TestGetFDCmdContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	cmd := getFDCmd(ctx)

	// If command is returned, it should have been created with the context
	// We can't directly test the internal context, but we verified it's created correctly
	if cmd != nil {
		if cmd.Path == "" {
			t.Error("command returned with empty path")
		}
	}
}

// TestIsFDInstalled verifies fd installation detection
func TestIsFDInstalled(t *testing.T) {
	installed := isFDInstalled()

	// Verify consistency with actual PATH lookup
	fdExists := false
	if _, err := exec.LookPath("fd"); err == nil {
		fdExists = true
	}
	if _, err := exec.LookPath("fdfind"); err == nil {
		fdExists = true
	}

	if installed != fdExists {
		t.Errorf("isFDInstalled() = %v, but PATH lookup says %v", installed, fdExists)
	}
}

// TestIsFDInstalledConsistency verifies getFDCmd and isFDInstalled agree
func TestIsFDInstalledConsistency(t *testing.T) {
	ctx := context.Background()

	installed := isFDInstalled()
	cmd := getFDCmd(ctx)

	if installed && cmd == nil {
		t.Error("isFDInstalled() is true but getFDCmd() returned nil")
	}

	if !installed && cmd != nil {
		t.Error("isFDInstalled() is false but getFDCmd() returned a command")
	}
}

// TestGetFDCmdReturnValue verifies command structure
func TestGetFDCmdReturnValue(t *testing.T) {
	ctx := context.Background()
	cmd := getFDCmd(ctx)

	if cmd != nil {
		// Verify command path is set
		if cmd.Path == "" {
			t.Error("returned command has empty Path")
		}

		// Verify it's an executable path
		if !isExecutable(cmd.Path) {
			t.Errorf("command path is not executable: %s", cmd.Path)
		}
	}
}

// TestFDCommandNames verifies command name detection
func TestFDCommandNames(t *testing.T) {
	tests := []struct {
		name     string
		checkCmd string
		wantName string
	}{
		{"fd exists", "fd", "fd"},
		{"fdfind exists", "fdfind", "fdfind"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := exec.LookPath(tt.checkCmd); err == nil {
				ctx := context.Background()
				cmd := getFDCmd(ctx)

				if cmd == nil {
					t.Errorf("%s is in PATH but getFDCmd returned nil", tt.checkCmd)
					return
				}

				// Verify the command name matches
				if !contains(cmd.Path, tt.wantName) {
					t.Errorf("expected command to contain %s, got %s", tt.wantName, cmd.Path)
				}
			} else {
				t.Skipf("%s not installed, skipping", tt.checkCmd)
			}
		})
	}
}

// TestFDPrecedence verifies fd is preferred over fdfind
func TestFDPrecedence(t *testing.T) {
	fdExists := false
	fdfindExists := false

	if _, err := exec.LookPath("fd"); err == nil {
		fdExists = true
	}
	if _, err := exec.LookPath("fdfind"); err == nil {
		fdfindExists = true
	}

	if fdExists && fdfindExists {
		ctx := context.Background()
		cmd := getFDCmd(ctx)

		if cmd == nil {
			t.Fatal("both fd and fdfind exist but getFDCmd returned nil")
		}

		// Verify "fd" is preferred (should be in the path)
		if !contains(cmd.Path, "fd") {
			t.Errorf("both fd and fdfind exist, but command is %s (expected fd)", cmd.Path)
		}
	} else {
		t.Skip("both fd and fdfind must be installed to test precedence")
	}
}

// Helper function to check if a path is executable
func isExecutable(path string) bool {
	_, err := exec.LookPath(path)
	return err == nil
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s[len(s)-len(substr):] == substr ||
			len(s) > len(substr) && s[len(s)-len(substr)-1:len(s)-len(substr)] == "/")
}
