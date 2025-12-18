package filepathext

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestSmartJoin(t *testing.T) {
	tests := []struct {
		name     string
		one      string
		two      string
		expected string
	}{
		{
			name:     "both relative paths",
			one:      "foo",
			two:      "bar",
			expected: filepath.Join("foo", "bar"),
		},
		{
			name:     "second path absolute",
			one:      "foo",
			two:      "/absolute/path",
			expected: "/absolute/path",
		},
		{
			name:     "first path absolute, second relative",
			one:      "/absolute",
			two:      "relative",
			expected: filepath.Join("/absolute", "relative"),
		},
		{
			name:     "empty first path",
			one:      "",
			two:      "bar",
			expected: "bar",
		},
		{
			name:     "empty second path",
			one:      "foo",
			two:      "",
			expected: "foo",
		},
		{
			name:     "both empty",
			one:      "",
			two:      "",
			expected: "",
		},
	}

	// Add Windows-specific tests
	if runtime.GOOS == "windows" {
		tests = append(tests, []struct {
			name     string
			one      string
			two      string
			expected string
		}{
			{
				name:     "windows drive letter absolute",
				one:      "foo",
				two:      "C:\\absolute\\path",
				expected: "C:\\absolute\\path",
			},
			{
				name:     "windows UNC path",
				one:      "foo",
				two:      "\\\\server\\share",
				expected: "\\\\server\\share",
			},
		}...)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SmartJoin(tt.one, tt.two)
			if result != tt.expected {
				t.Errorf("SmartJoin(%q, %q) = %q, want %q", tt.one, tt.two, result, tt.expected)
			}
		})
	}
}

func TestSmartIsAbs(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "unix absolute path",
			path:     "/absolute/path",
			expected: true,
		},
		{
			name:     "unix relative path",
			path:     "relative/path",
			expected: false,
		},
		{
			name:     "unix relative with dots",
			path:     "../relative",
			expected: false,
		},
		{
			name:     "empty path",
			path:     "",
			expected: false,
		},
		{
			name:     "single slash",
			path:     "/",
			expected: true,
		},
		{
			name:     "relative current dir",
			path:     "./foo",
			expected: false,
		},
	}

	// Add Windows-specific tests
	if runtime.GOOS == "windows" {
		tests = append(tests, []struct {
			name     string
			path     string
			expected bool
		}{
			{
				name:     "windows drive letter",
				path:     "C:\\path",
				expected: true,
			},
			{
				name:     "windows drive letter forward slash",
				path:     "C:/path",
				expected: true,
			},
			{
				name:     "windows UNC path",
				path:     "\\\\server\\share",
				expected: true,
			},
			{
				name:     "windows relative with backslash",
				path:     "relative\\path",
				expected: false,
			},
			{
				name:     "unix-style absolute on windows",
				path:     "/absolute/path",
				expected: true,
			},
		}...)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SmartIsAbs(tt.path)
			if result != tt.expected {
				t.Errorf("SmartIsAbs(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

// TestSmartJoinWithAbsolutePaths tests that SmartJoin correctly handles
// when the second path is absolute
func TestSmartJoinWithAbsolutePaths(t *testing.T) {
	testCases := []struct {
		base     string
		path     string
		absolute bool
	}{
		{"/home/user", "/etc/config", true},
		{"/home/user", "relative/path", false},
		{"relative", "/absolute", true},
		{"relative", "also/relative", false},
	}

	for _, tc := range testCases {
		result := SmartJoin(tc.base, tc.path)
		if tc.absolute {
			if result != tc.path {
				t.Errorf("SmartJoin(%q, %q) = %q, expected %q (should use absolute path)",
					tc.base, tc.path, result, tc.path)
			}
		} else {
			expected := filepath.Join(tc.base, tc.path)
			if result != expected {
				t.Errorf("SmartJoin(%q, %q) = %q, expected %q",
					tc.base, tc.path, result, expected)
			}
		}
	}
}

// TestCrossPlatformPaths tests that paths work correctly across platforms
func TestCrossPlatformPaths(t *testing.T) {
	// These should work on both Unix and Windows
	testCases := []struct {
		one        string
		two        string
		shouldJoin bool
	}{
		{"base", "relative", true},
		{"base", "/absolute", false},
		{"/absolute", "relative", true},
		{"", "path", true},
	}

	for _, tc := range testCases {
		result := SmartJoin(tc.one, tc.two)
		isAbs := SmartIsAbs(tc.two)

		if !tc.shouldJoin && !isAbs {
			t.Errorf("Expected %q to be absolute, but SmartIsAbs returned false", tc.two)
		}

		if tc.shouldJoin && isAbs && tc.two != "" {
			t.Errorf("Expected %q to be relative, but SmartIsAbs returned true", tc.two)
		}

		if !tc.shouldJoin && result != tc.two {
			t.Errorf("SmartJoin(%q, %q) = %q, expected absolute path %q",
				tc.one, tc.two, result, tc.two)
		}
	}
}
