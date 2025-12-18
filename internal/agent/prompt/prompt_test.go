package prompt

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/nexora/nexora/internal/config"
)

func TestPromptDataEnvironment(t *testing.T) {
	// Create a test config
	tempDir := t.TempDir()
	cfg, err := config.Load(tempDir, tempDir, false)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Create prompt with fixed time
	fixedTime := time.Date(2025, 12, 17, 22, 46, 11, 0, time.UTC)
	p, err := NewPrompt("test", "{{.DateTime}}", WithTimeFunc(func() time.Time {
		return fixedTime
	}), WithWorkingDir(tempDir))
	if err != nil {
		t.Fatalf("Failed to create prompt: %v", err)
	}

	// Get prompt data
	ctx := context.Background()
	data, err := p.promptData(ctx, "test-provider", "test-model", *cfg)
	if err != nil {
		t.Fatalf("Failed to get prompt data: %v", err)
	}

	// Verify new fields exist
	t.Run("DateTime", func(t *testing.T) {
		if data.DateTime == "" {
			t.Error("DateTime should not be empty")
		}
		if !strings.Contains(data.DateTime, "2025") {
			t.Errorf("DateTime should contain year 2025, got: %s", data.DateTime)
		}
	})

	t.Run("CurrentUser", func(t *testing.T) {
		if data.CurrentUser == "" {
			t.Error("CurrentUser should not be empty")
		}
	})

	t.Run("LocalIP", func(t *testing.T) {
		// Should be either a valid IP or "unavailable"
		if data.LocalIP == "" {
			t.Error("LocalIP should not be empty")
		}
	})

	t.Run("RuntimeVersions", func(t *testing.T) {
		// These should be set (either version or "not installed")
		if data.PythonVersion == "" {
			t.Error("PythonVersion should not be empty")
		}
		if data.NodeVersion == "" {
			t.Error("NodeVersion should not be empty")
		}
		if data.GoVersion == "" {
			t.Error("GoVersion should not be empty")
		}
	})

	t.Run("ShellType", func(t *testing.T) {
		if data.ShellType == "" {
			t.Error("ShellType should not be empty")
		}
	})

	t.Run("SystemResources", func(t *testing.T) {
		if data.MemoryInfo == "" {
			t.Error("MemoryInfo should not be empty")
		}
		if data.DiskInfo == "" {
			t.Error("DiskInfo should not be empty")
		}
	})

	t.Run("Architecture", func(t *testing.T) {
		if data.Architecture == "" {
			t.Error("Architecture should not be empty")
		}
		t.Logf("Architecture: %s", data.Architecture)
	})

	t.Run("ContainerType", func(t *testing.T) {
		if data.ContainerType == "" {
			t.Error("ContainerType should not be empty")
		}
		t.Logf("ContainerType: %s", data.ContainerType)
	})

	t.Run("TerminalInfo", func(t *testing.T) {
		if data.TerminalInfo == "" {
			t.Error("TerminalInfo should not be empty")
		}
		t.Logf("TerminalInfo: %s", data.TerminalInfo)
	})

	t.Run("NetworkStatus", func(t *testing.T) {
		if data.NetworkStatus == "" {
			t.Error("NetworkStatus should not be empty")
		}
		t.Logf("NetworkStatus: %s", data.NetworkStatus)
	})

	t.Run("ActiveServices", func(t *testing.T) {
		if data.ActiveServices == "" {
			t.Log("ActiveServices is empty (expected when NEXORA_FULL_ENV not set)")
		} else {
			t.Logf("ActiveServices: %s", data.ActiveServices)
		}
	})
}

func TestPromptBuildWithEnvironment(t *testing.T) {
	tempDir := t.TempDir()
	cfg, err := config.Load(tempDir, tempDir, false)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	template := `## Environment
Date/Time: {{.DateTime}}
OS: {{.Platform}} ({{.Architecture}}){{if ne .ContainerType "none"}} - Running in {{.ContainerType}}{{end}}
User: {{.CurrentUser}}
Python: {{.PythonVersion}}
Memory: {{.MemoryInfo}}
Terminal: {{.TerminalInfo}}
Network: {{.NetworkStatus}}
Services: {{.ActiveServices}}
`

	p, err := NewPrompt("test", template, WithWorkingDir(tempDir))
	if err != nil {
		t.Fatalf("Failed to create prompt: %v", err)
	}

	ctx := context.Background()
	output, err := p.Build(ctx, "test-provider", "test-model", *cfg)
	if err != nil {
		t.Fatalf("Failed to build prompt: %v", err)
	}

	// Verify output contains expected sections
	if !strings.Contains(output, "## Environment") {
		t.Error("Output should contain Environment section")
	}
	if !strings.Contains(output, "Date/Time:") {
		t.Error("Output should contain Date/Time")
	}
	if !strings.Contains(output, "OS:") {
		t.Error("Output should contain OS")
	}
	if !strings.Contains(output, "User:") {
		t.Error("Output should contain User")
	}

	t.Logf("Generated prompt output:\n%s", output)
}

func TestEnvironmentHelpers(t *testing.T) {
	ctx := context.Background()

	t.Run("getCurrentUser", func(t *testing.T) {
		user := getCurrentUser()
		if user == "" || user == "unknown" {
			t.Skip("Unable to get current user")
		}
		t.Logf("Current user: %s", user)
	})

	t.Run("getLocalIP", func(t *testing.T) {
		ip := getLocalIP(ctx)
		t.Logf("Local IP: %s", ip)
		if ip == "" {
			t.Error("LocalIP should not be empty")
		}
	})

	t.Run("getRuntimeVersion", func(t *testing.T) {
		python := getRuntimeVersion(ctx, "python3 --version")
		t.Logf("Python: %s", python)

		node := getRuntimeVersion(ctx, "node --version")
		t.Logf("Node: %s", node)

		goVer := getRuntimeVersion(ctx, "go version")
		t.Logf("Go: %s", goVer)
	})

	t.Run("getMemoryInfo", func(t *testing.T) {
		mem := getMemoryInfo(ctx)
		t.Logf("Memory: %s", mem)
		if mem == "" {
			t.Error("MemoryInfo should not be empty")
		}
	})

	t.Run("getDiskInfo", func(t *testing.T) {
		cwd, _ := os.Getwd()
		disk := getDiskInfo(ctx, cwd)
		t.Logf("Disk: %s", disk)
		if disk == "" {
			t.Error("DiskInfo should not be empty")
		}
	})

	t.Run("getArchitecture", func(t *testing.T) {
		arch := getArchitecture()
		t.Logf("Architecture: %s", arch)
		if arch == "" {
			t.Error("Architecture should not be empty")
		}
	})

	t.Run("detectContainer", func(t *testing.T) {
		container := detectContainer(ctx)
		t.Logf("Container: %s", container)
		if container == "" {
			t.Error("ContainerType should not be empty")
		}
	})

	t.Run("getTerminalInfo", func(t *testing.T) {
		terminal := getTerminalInfo(ctx)
		t.Logf("Terminal: %s", terminal)
		if terminal == "" {
			t.Error("TerminalInfo should not be empty")
		}
	})

	t.Run("getNetworkStatus", func(t *testing.T) {
		network := getNetworkStatus(ctx)
		t.Logf("Network: %s", network)
		if network == "" {
			t.Error("NetworkStatus should not be empty")
		}
	})

	t.Run("detectActiveServices", func(t *testing.T) {
		services := detectActiveServices(ctx)
		t.Logf("Services: %s", services)
		// Services may be empty depending on system state
	})
}
