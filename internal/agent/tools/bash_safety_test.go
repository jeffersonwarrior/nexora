package tools

import (
	"context"
	"testing"

	"github.com/nexora/nexora/internal/shell"
)

func TestBlockFuncs(t *testing.T) {
	blockers := blockFuncs()

	tests := []struct {
		name        string
		command     []string
		shouldBlock bool
		reason      string
	}{
		// ========================================
		// rm -rf Tests
		// ========================================
		{
			name:        "block rm -rf",
			command:     []string{"rm", "-rf", "/"},
			shouldBlock: true,
			reason:      "recursive force removal is extremely dangerous",
		},
		{
			name:        "block rm -fr",
			command:     []string{"rm", "-fr", "/tmp/test"},
			shouldBlock: true,
			reason:      "alternate flag order is still dangerous",
		},
		{
			name:        "block rm -r -f",
			command:     []string{"rm", "-r", "-f", "file.txt"},
			shouldBlock: true,
			reason:      "separate flags should also be blocked",
		},
		{
			name:        "allow rm single file",
			command:     []string{"rm", "file.txt"},
			shouldBlock: false,
			reason:      "removing single file is safe",
		},
		{
			name:        "allow rm -r (without force)",
			command:     []string{"rm", "-r", "directory/"},
			shouldBlock: false,
			reason:      "recursive without force requires confirmation",
		},

		// ========================================
		// Process Kill Tests
		// ========================================
		{
			name:        "block pkill nexora",
			command:     []string{"pkill", "nexora"},
			shouldBlock: true,
			reason:      "killing nexora would terminate the agent",
		},
		{
			name:        "block pkill tmux",
			command:     []string{"pkill", "tmux"},
			shouldBlock: true,
			reason:      "killing tmux would destroy all sessions",
		},
		{
			name:        "block killall nexora",
			command:     []string{"killall", "NEXORA"},
			shouldBlock: true,
			reason:      "case-insensitive match should catch variants",
		},
		{
			name:        "block killall tmux",
			command:     []string{"killall", "-9", "tmux"},
			shouldBlock: true,
			reason:      "force kill of tmux should be blocked",
		},
		{
			name:        "allow pkill other process",
			command:     []string{"pkill", "firefox"},
			shouldBlock: false,
			reason:      "killing other processes is allowed",
		},

		// ========================================
		// Init/Systemd Kill Tests
		// ========================================
		{
			name:        "block kill PID 1",
			command:     []string{"kill", "-9", "1"},
			shouldBlock: true,
			reason:      "killing init/systemd would crash the system",
		},
		{
			name:        "block kill -1 (all processes)",
			command:     []string{"kill", "-1"},
			shouldBlock: true,
			reason:      "kill -1 sends signal to all processes",
		},
		{
			name:        "allow kill other PID",
			command:     []string{"kill", "-9", "12345"},
			shouldBlock: false,
			reason:      "killing specific non-critical PID is allowed",
		},

		// ========================================
		// Disk Format/Wipe Tests
		// ========================================
		{
			name:        "block mkfs",
			command:     []string{"mkfs", "/dev/sda"},
			shouldBlock: true,
			reason:      "formatting disk would destroy data",
		},
		{
			name:        "block mkfs.ext4",
			command:     []string{"mkfs.ext4", "/dev/sdb1"},
			shouldBlock: true,
			reason:      "specific filesystem format is also dangerous",
		},
		{
			name:        "block fdisk",
			command:     []string{"fdisk", "/dev/sda"},
			shouldBlock: true,
			reason:      "partition manipulation can destroy data",
		},
		{
			name:        "block dd",
			command:     []string{"dd", "if=/dev/zero", "of=/dev/sda"},
			shouldBlock: true,
			reason:      "dd can wipe entire disks",
		},
		{
			name:        "block shred",
			command:     []string{"shred", "-n", "10", "/dev/sda"},
			shouldBlock: true,
			reason:      "shred securely wipes data",
		},

		// ========================================
		// Fork Bomb Tests
		// ========================================
		{
			name:        "block classic fork bomb",
			command:     []string{"bash", "-c", ":(){ :|:& };:"},
			shouldBlock: true,
			reason:      "classic bash fork bomb would crash system",
		},
		{
			name:        "block while true loop",
			command:     []string{"bash", "-c", "while true; do echo bomb; done"},
			shouldBlock: true,
			reason:      "infinite loop can consume resources",
		},
		{
			name:        "allow normal while loop",
			command:     []string{"bash", "-c", "while [ $i -lt 10 ]; do echo $i; done"},
			shouldBlock: false,
			reason:      "finite while loop is safe",
		},

		// ========================================
		// Git Force Push Tests
		// ========================================
		{
			name:        "block git push --force",
			command:     []string{"git", "push", "--force"},
			shouldBlock: true,
			reason:      "force push can overwrite remote history",
		},
		{
			name:        "block git push -f",
			command:     []string{"git", "push", "-f"},
			shouldBlock: true,
			reason:      "short form of force push should also block",
		},
		{
			name:        "allow normal git push",
			command:     []string{"git", "push", "origin", "main"},
			shouldBlock: false,
			reason:      "normal git push is safe",
		},

		// ========================================
		// Dangerous chmod Tests
		// ========================================
		{
			name:        "block chmod 777",
			command:     []string{"chmod", "777", "file.txt"},
			shouldBlock: true,
			reason:      "world-writable permissions are insecure",
		},
		{
			name:        "block chmod 000",
			command:     []string{"chmod", "000", "file.txt"},
			shouldBlock: true,
			reason:      "removing all permissions can lock out access",
		},
		{
			name:        "allow chmod 644",
			command:     []string{"chmod", "644", "file.txt"},
			shouldBlock: false,
			reason:      "standard file permissions are safe",
		},

		// ========================================
		// System Directory Protection Tests
		// ========================================
		{
			name:        "block rm on /bin",
			command:     []string{"rm", "/bin/ls"},
			shouldBlock: true,
			reason:      "removing system binaries breaks the system",
		},
		{
			name:        "block rm on /etc",
			command:     []string{"rm", "/etc/passwd"},
			shouldBlock: true,
			reason:      "removing critical config files is dangerous",
		},
		{
			name:        "block mv on /usr/bin",
			command:     []string{"mv", "/usr/bin/python", "/tmp/python"},
			shouldBlock: true,
			reason:      "moving system binaries breaks dependencies",
		},
		{
			name:        "allow rm in /tmp",
			command:     []string{"rm", "/tmp/test.txt"},
			shouldBlock: false,
			reason:      "removing temporary files is safe",
		},
		{
			name:        "allow rm in /home",
			command:     []string{"rm", "/home/user/file.txt"},
			shouldBlock: false,
			reason:      "removing user files is allowed",
		},

		// ========================================
		// Safe Commands (should not block)
		// ========================================
		{
			name:        "allow ls",
			command:     []string{"ls", "-la"},
			shouldBlock: false,
			reason:      "listing files is read-only and safe",
		},
		{
			name:        "allow git status",
			command:     []string{"git", "status"},
			shouldBlock: false,
			reason:      "checking status is read-only",
		},
		{
			name:        "allow cat",
			command:     []string{"cat", "file.txt"},
			shouldBlock: false,
			reason:      "reading files is safe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocked := false
			for _, blocker := range blockers {
				if blocker(tt.command) {
					blocked = true
					break
				}
			}

			if blocked != tt.shouldBlock {
				if tt.shouldBlock {
					t.Errorf("Expected command to be BLOCKED but it was ALLOWED: %v\nReason: %s",
						tt.command, tt.reason)
				} else {
					t.Errorf("Expected command to be ALLOWED but it was BLOCKED: %v\nReason: %s",
						tt.command, tt.reason)
				}
			}
		})
	}
}

func TestBlockFuncsIntegration(t *testing.T) {
	// Test with actual shell instance to ensure blockers integrate correctly
	blockers := blockFuncs()

	sh := shell.NewShell(&shell.Options{
		BlockFuncs: blockers,
	})

	// Use background context for tests
	ctx := context.Background()

	// Test that a blocked command actually fails
	_, _, err := sh.Exec(ctx, "rm -rf /test")
	if err == nil {
		t.Error("Expected rm -rf to be blocked, but it executed")
	}
	if err != nil && !containsString(err.Error(), "not allowed for security reasons") {
		t.Errorf("Expected security error message, got: %v", err)
	}

	// Test that a safe command still works
	_, _, err = sh.Exec(ctx, "echo 'test'")
	if err != nil {
		t.Errorf("Expected echo to work, but got error: %v", err)
	}
}

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
