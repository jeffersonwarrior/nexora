package tools

import (
	"context"
	"os/exec"
)

// getFDCmd returns an fd command if available, or nil if not found
func getFDCmd(ctx context.Context) *exec.Cmd {
	if _, err := exec.LookPath("fd"); err != nil {
		// Try fdfind on Debian/Ubuntu
		if _, err := exec.LookPath("fdfind"); err != nil {
			return nil
		}
		return exec.CommandContext(ctx, "fdfind")
	}
	return exec.CommandContext(ctx, "fd")
}

// isFDInstalled checks if fd or fdfind is available
func isFDInstalled() bool {
	if _, err := exec.LookPath("fd"); err == nil {
		return true
	}
	if _, err := exec.LookPath("fdfind"); err == nil {
		return true
	}
	return false
}
