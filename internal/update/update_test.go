package update

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckForUpdate_Old(t *testing.T) {
	info, err := Check(t.Context(), "v0.10.0", testClient{"v0.11.0"})
	require.NoError(t, err)
	require.NotNil(t, info)
	require.True(t, info.Available())
}

func TestCheckForUpdate_Beta(t *testing.T) {
	t.Run("current is stable", func(t *testing.T) {
		info, err := Check(t.Context(), "v0.10.0", testClient{"v0.11.0-beta.1"})
		require.NoError(t, err)
		require.NotNil(t, info)
		require.False(t, info.Available())
	})

	t.Run("current is also beta", func(t *testing.T) {
		info, err := Check(t.Context(), "v0.11.0-beta.1", testClient{"v0.11.0-beta.2"})
		require.NoError(t, err)
		require.NotNil(t, info)
		require.True(t, info.Available())
	})

	t.Run("current is beta, latest isn't", func(t *testing.T) {
		info, err := Check(t.Context(), "v0.11.0-beta.1", testClient{"v0.11.0"})
		require.NoError(t, err)
		require.NotNil(t, info)
		require.True(t, info.Available())
	})
}

// TestInfoAvailable verifies update availability logic
func TestInfoAvailable(t *testing.T) {
	tests := []struct {
		name      string
		current   string
		latest    string
		available bool
	}{
		{"same version", "1.0.0", "1.0.0", false},
		{"newer version", "1.0.0", "1.1.0", true},
		{"older version", "1.1.0", "1.0.0", true}, // Different = available
		{"current beta, latest stable", "1.0.0-beta", "1.0.0", true},
		{"current stable, latest beta", "1.0.0", "1.0.0-beta", false},
		{"both beta, same", "1.0.0-beta.1", "1.0.0-beta.1", false},
		{"both beta, different", "1.0.0-beta.1", "1.0.0-beta.2", true},
		{"current rc, latest stable", "1.0.0-rc.1", "1.0.0", true},
		{"current stable, latest rc", "1.0.0", "1.0.0-rc.1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := Info{
				Current: tt.current,
				Latest:  tt.latest,
			}
			if got := info.Available(); got != tt.available {
				t.Errorf("Available() = %v, want %v", got, tt.available)
			}
		})
	}
}

// TestInfoIsDevelopment verifies development version detection
func TestInfoIsDevelopment(t *testing.T) {
	tests := []struct {
		name    string
		current string
		isDev   bool
	}{
		{"devel", "devel", true},
		{"unknown", "unknown", true},
		{"dirty", "v1.0.0-dirty", true},
		{"go install version", "v0.0.0-0.20251231235959-06c807842604", true},
		{"go install no v prefix", "0.0.0-0.20251231235959-06c807842604", true},
		{"stable version", "1.0.0", false},
		{"stable with v", "v1.0.0", false},
		{"beta version", "1.0.0-beta.1", false},
		{"rc version", "1.0.0-rc.1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := Info{Current: tt.current}
			if got := info.IsDevelopment(); got != tt.isDev {
				t.Errorf("IsDevelopment() = %v, want %v", got, tt.isDev)
			}
		})
	}
}

// TestCheck verifies the Check function
func TestCheck(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client := testClient{tag: "v1.2.3"}
		info, err := Check(context.Background(), "v1.0.0", client)

		require.NoError(t, err)
		require.Equal(t, "1.0.0", info.Current)
		require.Equal(t, "1.2.3", info.Latest)
		require.Equal(t, "https://example.org", info.URL)
		require.True(t, info.Available())
	})

	t.Run("same version", func(t *testing.T) {
		client := testClient{tag: "v1.0.0"}
		info, err := Check(context.Background(), "v1.0.0", client)

		require.NoError(t, err)
		require.Equal(t, "1.0.0", info.Current)
		require.Equal(t, "1.0.0", info.Latest)
		require.False(t, info.Available())
	})

	t.Run("no v prefix", func(t *testing.T) {
		client := testClient{tag: "1.2.3"}
		info, err := Check(context.Background(), "1.0.0", client)

		require.NoError(t, err)
		require.Equal(t, "1.0.0", info.Current)
		require.Equal(t, "1.2.3", info.Latest)
	})

	t.Run("client error", func(t *testing.T) {
		client := errorClient{err: errors.New("network error")}
		info, err := Check(context.Background(), "1.0.0", client)

		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to fetch latest release")
		require.Equal(t, "1.0.0", info.Current)
		require.Equal(t, "1.0.0", info.Latest) // Falls back to current
	})
}

// TestGoInstallRegexp verifies the go install version pattern
func TestGoInstallRegexp(t *testing.T) {
	tests := []struct {
		version string
		matches bool
	}{
		{"v0.0.0-0.20251231235959-06c807842604", true},
		{"0.0.0-0.20251231235959-06c807842604", true},
		{"v0.0.0-0.20230101000000-123456789abc", true},
		{"v1.2.3", false},
		{"1.2.3-beta.1", false},
		{"devel", false},
		{"v0.0.0-0.20251231235959", false},                    // Missing commit hash
		{"v0.0.0-0.20251231235959-06c807842604-extra", false}, // Extra suffix
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			if got := goInstallRegexp.MatchString(tt.version); got != tt.matches {
				t.Errorf("goInstallRegexp.MatchString(%q) = %v, want %v",
					tt.version, got, tt.matches)
			}
		})
	}
}

// TestInfoStruct verifies Info struct fields
func TestInfoStruct(t *testing.T) {
	info := Info{
		Current: "1.0.0",
		Latest:  "1.1.0",
		URL:     "https://github.com/example/repo/releases/tag/v1.1.0",
	}

	require.Equal(t, "1.0.0", info.Current)
	require.Equal(t, "1.1.0", info.Latest)
	require.Equal(t, "https://github.com/example/repo/releases/tag/v1.1.0", info.URL)
	require.True(t, info.Available())
	require.False(t, info.IsDevelopment())
}

// TestReleaseStruct verifies Release struct
func TestReleaseStruct(t *testing.T) {
	release := Release{
		TagName: "v1.2.3",
		HTMLURL: "https://github.com/example/repo/releases/tag/v1.2.3",
	}

	require.Equal(t, "v1.2.3", release.TagName)
	require.Equal(t, "https://github.com/example/repo/releases/tag/v1.2.3", release.HTMLURL)
}

// TestDefaultClient verifies the default client is set
func TestDefaultClient(t *testing.T) {
	require.NotNil(t, Default)

	// Verify it implements Client interface
	var _ Client = Default
}

// TestCheckWithContext verifies context handling
func TestCheckWithContext(t *testing.T) {
	t.Run("context not cancelled", func(t *testing.T) {
		ctx := context.Background()
		client := testClient{tag: "v1.0.0"}
		info, err := Check(ctx, "v1.0.0", client)

		require.NoError(t, err)
		require.NotNil(t, info)
	})

	t.Run("context with timeout", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		client := cancelClient{}
		_, err := Check(ctx, "v1.0.0", client)

		// Should get context cancelled error
		require.Error(t, err)
	})
}

type testClient struct{ tag string }

// Latest implements Client.
func (t testClient) Latest(ctx context.Context) (*Release, error) {
	return &Release{
		TagName: t.tag,
		HTMLURL: "https://example.org",
	}, nil
}

type errorClient struct{ err error }

// Latest implements Client.
func (e errorClient) Latest(ctx context.Context) (*Release, error) {
	return nil, e.err
}

type cancelClient struct{}

// Latest implements Client.
func (c cancelClient) Latest(ctx context.Context) (*Release, error) {
	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return &Release{TagName: "v1.0.0", HTMLURL: "https://example.org"}, nil
	}
}
