package testutil

import (
	"os"
	"time"
)

const (
	// Default test timeout durations
	DefaultShortTestTimeout  = 30 * time.Second  // For unit tests
	DefaultMediumTestTimeout = 5 * time.Minute   // For integration tests
	DefaultLongTestTimeout   = 10 * time.Minute  // For full suite tests
	DefaultQATestTimeout     = 5 * time.Minute   // For QA tests
)

// Environment variable names for customizing timeouts
const (
	EnvShortTestTimeout  = "NEXORA_TEST_TIMEOUT_SHORT"
	EnvMediumTestTimeout = "NEXORA_TEST_TIMEOUT_MEDIUM"
	EnvLongTestTimeout   = "NEXORA_TEST_TIMEOUT_LONG"
	EnvQATestTimeout     = "NEXORA_TEST_TIMEOUT_QA"
)

// GetShortTestTimeout returns the timeout for short tests, configurable via environment
func GetShortTestTimeout() time.Duration {
	return getTimeoutFromEnv(EnvShortTestTimeout, DefaultShortTestTimeout)
}

// GetMediumTestTimeout returns the timeout for medium tests, configurable via environment
func GetMediumTestTimeout() time.Duration {
	return getTimeoutFromEnv(EnvMediumTestTimeout, DefaultMediumTestTimeout)
}

// GetLongTestTimeout returns the timeout for long tests, configurable via environment
func GetLongTestTimeout() time.Duration {
	return getTimeoutFromEnv(EnvLongTestTimeout, DefaultLongTestTimeout)
}

// GetQATestTimeout returns the timeout for QA tests, configurable via environment
func GetQATestTimeout() time.Duration {
	return getTimeoutFromEnv(EnvQATestTimeout, DefaultQATestTimeout)
}

// getTimeoutFromEnv parses a timeout duration from an environment variable
func getTimeoutFromEnv(envVar string, defaultValue time.Duration) time.Duration {
	if s := os.Getenv(envVar); s != "" {
		if d, err := time.ParseDuration(s); err == nil {
			return d
		}
	}
	return defaultValue
}