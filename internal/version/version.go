package version

import (
	"runtime/debug"
	"strings"
)

var Version = "0.29.0"

func init() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	mainVersion := info.Main.Version
	// Use VCS info if available and we're on a tagged version (not devel)
	if mainVersion == "(devel)" {
		// Keep the version as set by the file/ldflags
		return
	} else if mainVersion != "" && !strings.HasPrefix(mainVersion, "v0.29") && Version != "0.29.0" {
		// Only override if we're not building for version 0.28.x
		// and the version hasn't already been set to 0.29.0 by ldflags
		Version = mainVersion
	}
}
