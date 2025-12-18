package prompt

import (
	"context"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

// EnvironmentCache caches environment detection results to avoid repeated shell commands
type EnvironmentCache struct {
	mu         sync.RWMutex
	data       EnvironmentData
	lastUpdate time.Time
	ttl        time.Duration
}

// EnvironmentData holds all cached environment information
type EnvironmentData struct {
	CurrentUser    string
	LocalIP        string
	PythonVersion  string
	NodeVersion    string
	GoVersion      string
	GitUserName    string
	GitUserEmail   string
	MemoryInfo     string
	DiskInfo       string
	Architecture   string
	ContainerType  string
	TerminalInfo   string
	NetworkStatus  string
	ActiveServices string
}

// NewEnvironmentCache creates a new environment cache with the specified TTL
func NewEnvironmentCache(ttl time.Duration) *EnvironmentCache {
	return &EnvironmentCache{
		ttl: ttl,
	}
}

// Get retrieves cached environment data or refreshes if expired
func (c *EnvironmentCache) Get(ctx context.Context, workingDir string, fullEnv bool) (EnvironmentData, error) {
	c.mu.RLock()
	if time.Since(c.lastUpdate) < c.ttl {
		defer c.mu.RUnlock()
		return c.data, nil
	}
	c.mu.RUnlock()

	return c.refresh(ctx, workingDir, fullEnv)
}

// refresh updates the cache with fresh environment data using parallel execution
func (c *EnvironmentCache) refresh(ctx context.Context, workingDir string, fullEnv bool) (EnvironmentData, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock (avoid thundering herd)
	if time.Since(c.lastUpdate) < c.ttl {
		return c.data, nil
	}

	eg, egCtx := errgroup.WithContext(ctx)
	data := EnvironmentData{}

	// Parallel execution of all environment checks
	eg.Go(func() error {
		data.CurrentUser = getCurrentUser()
		return nil
	})

	eg.Go(func() error {
		data.LocalIP = getLocalIP(egCtx)
		return nil
	})

	eg.Go(func() error {
		data.PythonVersion = getRuntimeVersion(egCtx, "python3 --version")
		return nil
	})

	eg.Go(func() error {
		data.NodeVersion = getRuntimeVersion(egCtx, "node --version")
		return nil
	})

	eg.Go(func() error {
		data.GoVersion = getRuntimeVersion(egCtx, "go version")
		return nil
	})

	eg.Go(func() error {
		// Git config is already cached in gitConfigCache, but include in parallel fetch
		gitConfigCache.Do(func() {
			gitConfigCache.userName = getGitConfig(egCtx, "user.name")
			gitConfigCache.userEmail = getGitConfig(egCtx, "user.email")
		})
		data.GitUserName = gitConfigCache.userName
		data.GitUserEmail = gitConfigCache.userEmail
		return nil
	})

	eg.Go(func() error {
		data.MemoryInfo = getMemoryInfo(egCtx)
		return nil
	})

	eg.Go(func() error {
		data.DiskInfo = getDiskInfo(egCtx, workingDir)
		return nil
	})

	eg.Go(func() error {
		data.Architecture = getArchitecture()
		data.ContainerType = detectContainer(egCtx)
		data.TerminalInfo = getTerminalInfo(egCtx)
		return nil
	})

	// Only fetch expensive operations if full env requested
	if fullEnv {
		eg.Go(func() error {
			data.NetworkStatus = getNetworkStatus(egCtx)
			return nil
		})

		eg.Go(func() error {
			data.ActiveServices = detectActiveServices(egCtx)
			return nil
		})
	} else {
		// Use defaults for non-full env mode
		data.NetworkStatus = "online"
		data.ActiveServices = ""
	}

	if err := eg.Wait(); err != nil {
		return EnvironmentData{}, err
	}

	c.data = data
	c.lastUpdate = time.Now()

	return data, nil
}

// Invalidate clears the cache, forcing a refresh on next Get()
func (c *EnvironmentCache) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lastUpdate = time.Time{}
}
