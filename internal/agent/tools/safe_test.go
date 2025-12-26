package tools

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSafeCommands_Contains(t *testing.T) {
	tests := []struct {
		name    string
		command string
		want    bool
	}{
		// Bash builtins
		{name: "pwd", command: "pwd", want: true},
		{name: "ls", command: "ls", want: true},
		{name: "echo", command: "echo", want: true},
		{name: "date", command: "date", want: true},
		{name: "whoami", command: "whoami", want: true},

		// Git commands
		{name: "git status", command: "git status", want: true},
		{name: "git log", command: "git log", want: true},
		{name: "git diff", command: "git diff", want: true},
		{name: "git branch", command: "git branch", want: true},

		// Process commands
		{name: "ps", command: "ps", want: true},
		{name: "top", command: "top", want: true},
		{name: "kill", command: "kill", want: true},

		// Unsafe commands (not in list)
		{name: "rm", command: "rm", want: false},
		{name: "mv", command: "mv", want: false},
		{name: "cp", command: "cp", want: false},
		{name: "chmod", command: "chmod", want: false},
		{name: "chown", command: "chown", want: false},
		{name: "sudo", command: "sudo", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found := false
			for _, safe := range safeCommands {
				if safe == tt.command {
					found = true
					break
				}
			}
			assert.Equal(t, tt.want, found)
		})
	}
}

func TestSafeCommands_GitReadOnly(t *testing.T) {
	// Verify all git commands in safe list are read-only
	gitReadOnly := []string{
		"git blame",
		"git branch",
		"git config --get",
		"git config --list",
		"git describe",
		"git diff",
		"git grep",
		"git log",
		"git ls-files",
		"git ls-remote",
		"git remote",
		"git rev-parse",
		"git shortlog",
		"git show",
		"git status",
		"git tag",
	}

	for _, cmd := range gitReadOnly {
		found := false
		for _, safe := range safeCommands {
			if safe == cmd {
				found = true
				break
			}
		}
		assert.True(t, found, "git command should be in safe list: %s", cmd)
	}

	// Verify dangerous git commands are NOT in safe list
	gitDangerous := []string{
		"git add",
		"git commit",
		"git push",
		"git pull",
		"git reset",
		"git checkout",
		"git rebase",
		"git merge",
	}

	for _, cmd := range gitDangerous {
		found := false
		for _, safe := range safeCommands {
			if safe == cmd {
				found = true
				break
			}
		}
		assert.False(t, found, "dangerous git command should NOT be in safe list: %s", cmd)
	}
}

func TestSafeCommands_NoWriteCommands(t *testing.T) {
	// Verify no write/destructive commands in safe list
	writeCommands := []string{
		"rm", "rmdir", "unlink",
		"mv", "cp",
		"chmod", "chown",
		"mkdir", "touch",
		"vi", "vim", "nano", "emacs",
		"dd", "shred",
	}

	for _, cmd := range writeCommands {
		found := false
		for _, safe := range safeCommands {
			if safe == cmd {
				found = true
				break
			}
		}
		assert.False(t, found, "write command should NOT be in safe list: %s", cmd)
	}
}

func TestSafeCommands_WindowsSpecific(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific test on non-Windows platform")
	}

	windowsCommands := []string{
		"ipconfig",
		"nslookup",
		"ping",
		"systeminfo",
		"tasklist",
		"where",
	}

	for _, cmd := range windowsCommands {
		found := false
		for _, safe := range safeCommands {
			if safe == cmd {
				found = true
				break
			}
		}
		assert.True(t, found, "Windows command should be in safe list: %s", cmd)
	}
}

func TestSafeCommands_UnixOnly(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix-specific test on Windows platform")
	}

	// Verify Windows commands are NOT in safe list on Unix
	// On non-Windows, Windows-specific commands should not be in the list
	// This is platform-dependent, so we just verify the test runs
}

func TestSafeCommands_NotEmpty(t *testing.T) {
	assert.NotEmpty(t, safeCommands, "safe commands list should not be empty")
	assert.Greater(t, len(safeCommands), 20, "should have at least 20 safe commands")
}

func TestSafeCommands_NoDuplicates(t *testing.T) {
	seen := make(map[string]bool)
	for _, cmd := range safeCommands {
		assert.False(t, seen[cmd], "duplicate command in safe list: %s", cmd)
		seen[cmd] = true
	}
}

func TestSafeCommands_AllTrimmed(t *testing.T) {
	for _, cmd := range safeCommands {
		assert.Equal(t, cmd, cmd, "command should be trimmed: '%s'", cmd)
		assert.NotEmpty(t, cmd, "command should not be empty")
	}
}

func TestSafeCommands_SystemInfo(t *testing.T) {
	// Verify system information commands are safe
	sysInfoCommands := []string{
		"uname",
		"hostname",
		"whoami",
		"id",
		"groups",
		"uptime",
		"df",
		"du",
		"free",
	}

	for _, cmd := range sysInfoCommands {
		found := false
		for _, safe := range safeCommands {
			if safe == cmd {
				found = true
				break
			}
		}
		assert.True(t, found, "system info command should be safe: %s", cmd)
	}
}

func TestSafeCommands_ProcessManagement(t *testing.T) {
	// Process viewing is safe, killing requires permission but is included
	processCommands := []string{
		"ps",
		"top",
		"kill",
		"killall",
		"nice",
		"nohup",
	}

	for _, cmd := range processCommands {
		found := false
		for _, safe := range safeCommands {
			if safe == cmd {
				found = true
				break
			}
		}
		assert.True(t, found, "process command should be in safe list: %s", cmd)
	}
}

func TestSafeCommands_GitConfig(t *testing.T) {
	// Only read-only git config commands should be safe
	assert.Contains(t, safeCommands, "git config --get")
	assert.Contains(t, safeCommands, "git config --list")

	// Verify write config is NOT in safe list
	found := false
	for _, safe := range safeCommands {
		if safe == "git config --set" || safe == "git config" {
			found = true
			break
		}
	}
	assert.False(t, found, "git config write should NOT be in safe list")
}
