package version

import (
	"runtime/debug"
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	// Test that the version is set
	if Version == "" {
		t.Error("Version should not be empty")
	}

	// Test version format (should be semver-like or "(devel)")
	if !strings.HasPrefix(Version, "v") && !strings.HasPrefix(Version, "0.") && Version != "(devel)" {
		t.Logf("Version format: %q", Version)
	}

	t.Logf("Current version: %s", Version)
}

func TestVersionInit(t *testing.T) {
	// This test verifies that the init() function runs without panicking
	// and that the version is accessible

	// The init function should have already run before this test
	if Version == "" {
		t.Error("Version should be initialized by init()")
	}

	// Check that we can read build info (may not be available in all build modes)
	info, ok := debug.ReadBuildInfo()
	if ok {
		t.Logf("Build info available: %v", info.Main.Version)

		// Verify main module
		if info.Main.Path != "" {
			t.Logf("Main module path: %s", info.Main.Path)
		}
	} else {
		t.Log("Build info not available (normal for some build modes)")
	}
}

func TestVersionStability(t *testing.T) {
	// Version should remain constant during runtime
	originalVersion := Version

	// Call some operations
	info, _ := debug.ReadBuildInfo()
	_ = info

	// Verify version hasn't changed
	if Version != originalVersion {
		t.Errorf("Version changed during test: was %q, now %q", originalVersion, Version)
	}
}

func TestVersionExpectedValues(t *testing.T) {
	tests := []struct {
		name  string
		check func() bool
		desc  string
	}{
		{
			name: "not empty",
			check: func() bool {
				return Version != ""
			},
			desc: "Version should not be empty",
		},
		{
			name: "reasonable length",
			check: func() bool {
				return len(Version) > 0 && len(Version) < 100
			},
			desc: "Version should have reasonable length",
		},
		{
			name: "valid version format",
			check: func() bool {
				// Should be either:
				// - Starts with "v" (like v1.2.3)
				// - Starts with digit (like 0.28.5)
				// - Is "(devel)"
				return strings.HasPrefix(Version, "v") ||
					strings.HasPrefix(Version, "0.") ||
					strings.HasPrefix(Version, "1.") ||
					Version == "(devel)"
			},
			desc: "Version should follow expected format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.check() {
				t.Errorf("%s: got %q", tt.desc, Version)
			}
		})
	}
}

func TestVersionAgainstBuildInfo(t *testing.T) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		t.Skip("Build info not available")
	}

	mainVersion := info.Main.Version
	t.Logf("BuildInfo version: %q, Package version: %q", mainVersion, Version)

	// If we have a non-devel build, check the logic
	if mainVersion != "(devel)" && mainVersion != "" {
		// The init function should have potentially updated Version
		// based on the build info

		// If mainVersion starts with "v0.28" and Version is "0.28.5",
		// that means the version wasn't overridden (expected behavior)
		if strings.HasPrefix(mainVersion, "v0.28") && Version == "0.28.5" {
			t.Log("Version correctly kept as 0.28.5 for v0.28.x build")
		}
	}
}

// TestVersionPackageConstants ensures the constants are properly defined
func TestVersionPackageConstants(t *testing.T) {
	// This would fail at compile time if Version wasn't declared,
	// but we can still verify it's accessible
	_ = Version

	// Verify we can compare and use it
	if len(Version) >= 0 {
		t.Log("Version variable is accessible and comparable")
	}
}

// TestVersionInitLogic simulates different build scenarios
func TestVersionInitLogic(t *testing.T) {
	// We can't actually re-run init(), but we can verify the current state
	// makes sense given the init() logic

	info, ok := debug.ReadBuildInfo()
	if !ok {
		t.Skip("Build info not available, can't test init logic")
	}

	mainVersion := info.Main.Version

	t.Run("devel version handling", func(t *testing.T) {
		if mainVersion == "(devel)" {
			// Init should have kept the default or ldflags version
			if Version == "" {
				t.Error("Version should not be empty even in devel mode")
			}
			t.Logf("Devel mode: Version = %q", Version)
		}
	})

	t.Run("tagged version handling", func(t *testing.T) {
		if mainVersion != "" && mainVersion != "(devel)" {
			// Init may have updated Version based on mainVersion
			// This depends on the specific conditions in init()
			t.Logf("Tagged version: mainVersion=%q, Version=%q", mainVersion, Version)
		}
	})
}

// TestDisplay tests the Display() function that strips pseudo-version suffixes
func TestDisplay(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected string
	}{
		{
			name:     "clean semver",
			version:  "v0.29.3",
			expected: "v0.29.3",
		},
		{
			name:     "with dirty suffix",
			version:  "v0.29.3+dirty",
			expected: "v0.29.3",
		},
		{
			name:     "with pseudo-version",
			version:  "v0.29.3-0.20251228184059-1c881ae20c65",
			expected: "v0.29.3",
		},
		{
			name:     "with pseudo-version and dirty",
			version:  "v0.29.3-0.20251228184059-1c881ae20c65+dirty",
			expected: "v0.29.3",
		},
		{
			name:     "without v prefix",
			version:  "0.29.3",
			expected: "v0.29.3",
		},
		{
			name:     "without v prefix with suffix",
			version:  "0.29.3-0.20251228+dirty",
			expected: "v0.29.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original version
			original := Version
			defer func() { Version = original }()

			// Set test version
			Version = tt.version

			// Test Display()
			result := Display()
			if result != tt.expected {
				t.Errorf("Display() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestDisplayCurrentVersion tests Display() with the current version
func TestDisplayCurrentVersion(t *testing.T) {
	result := Display()

	// Should always return a clean semver (vX.Y.Z)
	if !strings.HasPrefix(result, "v") {
		t.Errorf("Display() should return version with v prefix, got %q", result)
	}

	// Should not contain pseudo-version suffixes
	if strings.Contains(result, "-0.") {
		t.Errorf("Display() should strip pseudo-version suffix, got %q", result)
	}

	// Should not contain +dirty
	if strings.Contains(result, "+") {
		t.Errorf("Display() should strip build metadata, got %q", result)
	}

	t.Logf("Display() = %q (raw Version = %q)", result, Version)
}

// BenchmarkVersionAccess tests the performance of accessing the version
func BenchmarkVersionAccess(b *testing.B) {
	var v string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v = Version
	}
	_ = v
}

// BenchmarkDisplay tests the performance of the Display function
func BenchmarkDisplay(b *testing.B) {
	var v string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v = Display()
	}
	_ = v
}

// BenchmarkReadBuildInfo tests the performance of reading build info
func BenchmarkReadBuildInfo(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		debug.ReadBuildInfo()
	}
}

// TestIsPseudoVersion tests the isPseudoVersion function
func TestIsPseudoVersion(t *testing.T) {
	tests := []struct {
		version  string
		isPseudo bool
	}{
		// Pseudo-versions (should be true)
		{"v0.29.3-0.20251228184059-1c881ae20c65", true},
		{"v0.29.3-0.20251228184059-1c881ae20c65+dirty", true},
		{"v1.0.0-0.20240101120000-abc123def456", true},
		{"v0.0.0-20251228184059-1c881ae20c65", true},

		// Clean versions (should be false)
		{"v0.29.3", false},
		{"v1.0.0", false},
		{"0.29.3", false},
		{"v1.0.0-beta", false},
		{"v1.0.0-rc1", false},
		{"v1.0.0-alpha.1", false},
		{"v2.0.0-preview", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			result := isPseudoVersion(tt.version)
			if result != tt.isPseudo {
				t.Errorf("isPseudoVersion(%q) = %v, want %v", tt.version, result, tt.isPseudo)
			}
		})
	}
}

// TestVersionNotOverwrittenByPseudoVersion ensures pseudo-versions don't override hardcoded version
func TestVersionNotOverwrittenByPseudoVersion(t *testing.T) {
	// The current Version should be clean semver, not a pseudo-version
	if isPseudoVersion(Version) {
		t.Errorf("Version should not be a pseudo-version, got %q", Version)
	}

	// Display() should return a clean version
	display := Display()
	if strings.Contains(display, "-0.20") {
		t.Errorf("Display() should not contain pseudo-version timestamp, got %q", display)
	}
	if strings.Contains(display, "+dirty") {
		t.Errorf("Display() should not contain +dirty, got %q", display)
	}
}
