package version

import (
	"runtime/debug"
	"strings"
)

var Version = "v0.29.3"

// Display returns a sanitized version string for UI display.
// Strips any +suffix (e.g., +dirty, +dev) from the version.
func Display() string {
	if idx := strings.Index(Version, "+"); idx != -1 {
		return Version[:idx]
	}
	return Version
}

func init() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	mainVersion := info.Main.Version
	// Use VCS info only if it's a clean tagged version (vX.Y.Z format)
	if mainVersion == "(devel)" || mainVersion == "" {
		// Keep the version as set by the file/ldflags
		return
	}
	// Only use build info version if it's a clean semver tag (not pseudo-version)
	if strings.Count(mainVersion, "-") == 0 || strings.HasPrefix(mainVersion, "v0.29") {
		Version = mainVersion
	}
}
