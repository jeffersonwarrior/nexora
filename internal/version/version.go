package version

import (
	"runtime/debug"
	"strings"
)

var Version = "v0.29.2"

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
