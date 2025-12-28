package version

import (
	"regexp"
	"runtime/debug"
	"strings"
)

var Version = "v0.29.3"

// semverPattern matches vX.Y.Z or X.Y.Z at the start
var semverPattern = regexp.MustCompile(`^v?(\d+\.\d+\.\d+)`)

// Display returns a sanitized version string for UI display.
// Strips pseudo-version suffixes (e.g., -0.20251228184059-1c881ae20c65)
// and build metadata (e.g., +dirty, +dev) to show only vX.Y.Z
func Display() string {
	if match := semverPattern.FindStringSubmatch(Version); len(match) > 1 {
		// Return with v prefix for consistency
		return "v" + match[1]
	}
	// Fallback: strip +suffix if present
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
	// Pseudo-versions contain timestamp-hash format like: v0.29.3-0.20251228184059-1c881ae20c65
	// Clean tags have no dashes or only pre-release suffix (v1.0.0-beta)
	if isPseudoVersion(mainVersion) {
		// Keep the hardcoded version for pseudo-versions
		return
	}
	Version = mainVersion
}

// isPseudoVersion returns true if the version string is a Go pseudo-version
// Pseudo-versions have format: vX.Y.Z-0.TIMESTAMP-COMMIT or vX.Y.Z-PRE.0.TIMESTAMP-COMMIT
func isPseudoVersion(v string) bool {
	// Pseudo-versions always contain a timestamp in format YYYYMMDDHHMMSS (14 digits)
	// Pattern: -0.20YYMMDDHHMMSS- or similar
	if idx := strings.Index(v, "-0.20"); idx != -1 {
		return true
	}
	// Also catch any version with +dirty or multiple dashes after semver
	parts := strings.SplitN(v, "-", 2)
	if len(parts) < 2 {
		return false // No dash, not pseudo-version
	}
	suffix := parts[1]
	// Pseudo-versions have numeric prefix after first dash (e.g., "0.20251228...")
	if len(suffix) > 0 && suffix[0] >= '0' && suffix[0] <= '9' {
		return true
	}
	return false
}
