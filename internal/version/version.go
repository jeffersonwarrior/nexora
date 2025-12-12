package version

import "runtime/debug"

// Build-time parameters set via -ldflags

var Version = "0.26.0"

// A user may install nexora using `go install github.com/nexora/cli@latest`.
// without -ldflags, in which case the version above is unset. As a workaround
// we use the embedded build version that *is* set when using `go install` (and
// is only set for `go install` and not for `go build`).
func init() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	mainVersion := info.Main.Version
	// Only override hardcoded version if git tag is newer than hardcoded version
	if mainVersion != "" && mainVersion != "(devel)" && mainVersion > Version {
		Version = mainVersion
	}
}
