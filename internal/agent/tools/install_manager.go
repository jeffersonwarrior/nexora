package tools

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// ExternalTool represents a command-line tool that might be used by Nexora
type ExternalTool struct {
	Name        string
	Description string
	InstallCmd  map[string][]string // OS -> command to install
	CheckCmd    []string            // Command to check if tool is installed
	Priority    int                 // Higher = more important
	AIRelevant  bool                // Whether this tool is particularly useful for AI operations
}

var essentialTools = []ExternalTool{
	{
		Name:        "ripgrep",
		Description: "Fast file search tool, much faster than grep",
		InstallCmd: map[string][]string{
			"darwin":     {"brew", "install", "ripgrep"},
			"linux":      {"apt", "install", "ripgrep"}, // Deb/Ubuntu
			"linux-apt":  {"apt", "install", "ripgrep"}, // Deb/Ubuntu variant
			"linux-yum":  {"yum", "install", "ripgrep"}, // RHEL/CentOS
			"linux-dnf":  {"dnf", "install", "ripgrep"}, // Fedora
			"linux-arch": {"pacman", "-S", "ripgrep"},   // Arch Linux
			"windows":    {"winget", "install", "BurntSushi.ripgrep"},
		},
		CheckCmd:   []string{"rg", "--version"},
		Priority:   100,
		AIRelevant: true,
	},
	{
		Name:        "fd",
		Description: "Fast find replacement for finding files and directories",
		InstallCmd: map[string][]string{
			"darwin":     {"brew", "install", "fd"},
			"linux":      {"apt", "install", "fd-find"}, // Deb/Ubuntu, provides 'fdfind'
			"linux-apt":  {"apt", "install", "fd-find"}, // Deb/Ubuntu variant
			"linux-yum":  {"yum", "install", "fd"},      // RHEL/CentOS
			"linux-dnf":  {"dnf", "install", "fd"},      // Fedora
			"linux-arch": {"pacman", "-S", "fd"},        // Arch Linux
			"windows":    {"winget", "install", "sharkdp.fd"},
		},
		CheckCmd:   []string{"fd", "--version"},
		Priority:   95,
		AIRelevant: true,
	},
	{
		Name:        "delta",
		Description: "Better diff viewer with syntax highlighting",
		InstallCmd: map[string][]string{
			"darwin":     {"brew", "install", "git-delta"},
			"linux":      {"apt", "install", "git-delta"},
			"linux-apt":  {"apt", "install", "git-delta"},
			"linux-yum":  {"yum", "install", "git-delta"},
			"linux-dnf":  {"dnf", "install", "git-delta"},
			"linux-arch": {"pacman", "-S", "git-delta"},
			"windows":    {"winget", "install", "dandavison.delta"},
		},
		CheckCmd:   []string{"delta", "--version"},
		Priority:   80,
		AIRelevant: true,
	},
	{
		Name:        "bat",
		Description: "Cat with wings - syntax highlighting and git integration",
		InstallCmd: map[string][]string{
			"darwin":     {"brew", "install", "bat"},
			"linux":      {"apt", "install", "bat"},
			"linux-apt":  {"apt", "install", "bat"},
			"linux-yum":  {"yum", "install", "bat"},
			"linux-dnf":  {"dnf", "install", "bat"},
			"linux-arch": {"pacman", "-S", "bat"},
			"windows":    {"winget", "install", "sharkdp.bat"},
		},
		CheckCmd:   []string{"bat", "--version"},
		Priority:   75,
		AIRelevant: false,
	},
	{
		Name:        "exa",
		Description: "Modern replacement for ls",
		InstallCmd: map[string][]string{
			"darwin":     {"brew", "install", "exa"},
			"linux":      {"apt", "install", "exa"},
			"linux-apt":  {"apt", "install", "exa"},
			"linux-yum":  {"yum", "install", "exa"},
			"linux-dnf":  {"dnf", "install", "exa"},
			"linux-arch": {"pacman", "-S", "exa"},
			"windows":    {"winget", "install", "eza.eza"}, // eza is the maintained fork
		},
		CheckCmd:   []string{"exa", "--version"},
		Priority:   70,
		AIRelevant: false,
	},
	{
		Name:        "jq",
		Description: "Command-line JSON processor",
		InstallCmd: map[string][]string{
			"darwin":     {"brew", "install", "jq"},
			"linux":      {"apt", "install", "jq"},
			"linux-apt":  {"apt", "install", "jq"},
			"linux-yum":  {"yum", "install", "jq"},
			"linux-dnf":  {"dnf", "install", "jq"},
			"linux-arch": {"pacman", "-S", "jq"},
			"windows":    {"winget", "install", "jqlang.jq"},
		},
		CheckCmd:   []string{"jq", "--version"},
		Priority:   90,
		AIRelevant: true,
	},
	{
		Name:        "fzf",
		Description: "Command-line fuzzy finder",
		InstallCmd: map[string][]string{
			"darwin":     {"brew", "install", "fzf"},
			"linux":      {"apt", "install", "fzf"},
			"linux-apt":  {"apt", "install", "fzf"},
			"linux-yum":  {"yum", "install", "fzf"},
			"linux-dnf":  {"dnf", "install", "fzf"},
			"linux-arch": {"pacman", "-S", "fzf"},
			"windows":    {"winget", "install", "junegunn.fzf"},
		},
		CheckCmd:   []string{"fzf", "--version"},
		Priority:   85,
		AIRelevant: true,
	},
	{
		Name:        "sd",
		Description: "Intuitive find & replace CLI (sed alternative)",
		InstallCmd: map[string][]string{
			"darwin":     {"brew", "install", "sd"},
			"linux":      {"apt", "install", "sd"},
			"linux-apt":  {"apt", "install", "sd"},
			"linux-yum":  {"yum", "install", "sd"},
			"linux-dnf":  {"dnf", "install", "sd"},
			"linux-arch": {"pacman", "-S", "sd"},
			"windows":    {"winget", "install", "chmln.sd"},
		},
		CheckCmd:   []string{"sd", "--version"},
		Priority:   60,
		AIRelevant: true,
	},
	{
		Name:        "tokei",
		Description: "Code statistics tool - great for understanding codebases",
		InstallCmd: map[string][]string{
			"darwin":     {"brew", "install", "tokei"},
			"linux":      {"apt", "install", "tokei"},
			"linux-apt":  {"apt", "install", "tokei"},
			"linux-yum":  {"yum", "install", "tokei"},
			"linux-dnf":  {"dnf", "install", "tokei"},
			"linux-arch": {"pacman", "-S", "tokei"},
			"windows":    {"cargo", "install", "tokei"}, // Via cargo
		},
		CheckCmd:   []string{"tokei", "--version"},
		Priority:   50,
		AIRelevant: true,
	},
	{
		Name:        "hexyl",
		Description: "Hex viewer - useful for binary file inspection",
		InstallCmd: map[string][]string{
			"darwin":     {"brew", "install", "hexyl"},
			"linux":      {"apt", "install", "hexyl"},
			"linux-apt":  {"apt", "install", "hexyl"},
			"linux-yum":  {"yum", "install", "hexyl"},
			"linux-dnf":  {"dnf", "install", "hexyl"},
			"linux-arch": {"pacman", "-S", "hexyl"},
			"windows":    {"winget", "install", "sharkdp.hexyl"},
		},
		CheckCmd:   []string{"hexyl", "--version"},
		Priority:   40,
		AIRelevant: false,
	},
}

// InstallManager handles the checking and installation of essential tools
type InstallManager struct {
	tools []ExternalTool
}

// NewInstallManager creates a new installation manager
func NewInstallManager() *InstallManager {
	return &InstallManager{
		tools: essentialTools,
	}
}

// GetOSType returns the OS type with package manager detection
func GetOSType() string {
	osType := runtime.GOOS
	if osType == "linux" {
		// Try to detect the package manager
		if _, err := exec.LookPath("apt"); err == nil {
			return "linux-apt"
		}
		if _, err := exec.LookPath("yum"); err == nil {
			return "linux-yum"
		}
		if _, err := exec.LookPath("dnf"); err == nil {
			return "linux-dnf"
		}
		if _, err := exec.LookPath("pacman"); err == nil {
			return "linux-arch"
		}
		return "linux" // fallback
	}
	return osType
}

// IsToolInstalled checks if a tool is installed
func (im *InstallManager) IsToolInstalled(ctx context.Context, tool ExternalTool) bool {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, tool.CheckCmd[0], tool.CheckCmd[1:]...)
	err := cmd.Run()
	return err == nil
}

// InstallTool installs a tool if not already installed
func (im *InstallManager) InstallTool(ctx context.Context, tool ExternalTool) error {
	if im.IsToolInstalled(ctx, tool) {
		return nil // Already installed
	}

	osType := GetOSType()
	installCmd, exists := tool.InstallCmd[osType]
	if !exists {
		// Try generic OS type
		installCmd, exists = tool.InstallCmd[runtime.GOOS]
		if !exists {
			return fmt.Errorf("no installation command available for %s on %s", tool.Name, osType)
		}
	}

	slog.Info("Installing tool", "tool", tool.Name, "os", osType)

	// Create a subprocess with timeout
	cmdCtx, cancel := context.WithTimeout(ctx, 300*time.Second) // 5 minute timeout
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, installCmd[0], installCmd[1:]...)

	// Run without showing output to user unless in debug mode
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install %s: %w", tool.Name, err)
	}

	slog.Info("Successfully installed tool", "tool", tool.Name)
	return nil
}

// CheckAndInstallAIRelevantTools checks and installs the most important tools for AI operations
func (im *InstallManager) CheckAndInstallAIRelevantTools(ctx context.Context) error {
	slog.Info("Checking for AI-relevant tools")

	aiTools := make([]ExternalTool, 0)
	for _, tool := range im.tools {
		if tool.AIRelevant && tool.Priority >= 80 {
			aiTools = append(aiTools, tool)
		}
	}

	for _, tool := range aiTools {
		if !im.IsToolInstalled(ctx, tool) {
			slog.Info("AI-relevant tool not found, attempting to install", "tool", tool.Name, "description", tool.Description)
			if err := im.InstallTool(ctx, tool); err != nil {
				slog.Warn("Failed to install AI-relevant tool", "tool", tool.Name, "error", err)
				// Continue with other tools even if one fails
			}
		}
	}

	return nil
}

// GetUnavailableAIRelevantTools returns a list of AI-relevant tools that are not installed
func (im *InstallManager) GetUnavailableAIRelevantTools(ctx context.Context) []ExternalTool {
	var unavailable []ExternalTool
	for _, tool := range im.tools {
		if tool.AIRelevant && !im.IsToolInstalled(ctx, tool) {
			unavailable = append(unavailable, tool)
		}
	}
	return unavailable
}

// GenerateSuggestions returns a formatted list of suggested tools
func (im *InstallManager) GenerateSuggestions(ctx context.Context) string {
	var suggestions strings.Builder
	suggestions.WriteString("\n# Essential Tools for Enhanced AI-Assisted Development\n\n")

	// Group by priority
	highPriority := make([]ExternalTool, 0)
	mediumPriority := make([]ExternalTool, 0)
	lowPriority := make([]ExternalTool, 0)

	for _, tool := range im.tools {
		if tool.AIRelevant {
			if tool.Priority >= 90 {
				highPriority = append(highPriority, tool)
			} else if tool.Priority >= 70 {
				mediumPriority = append(mediumPriority, tool)
			} else {
				lowPriority = append(lowPriority, tool)
			}
		}
	}

	addToolList := func(title string, tools []ExternalTool) {
		if len(tools) == 0 {
			return
		}
		suggestions.WriteString(fmt.Sprintf("## %s\n\n", title))
		for _, tool := range tools {
			installed := im.IsToolInstalled(ctx, tool)
			status := "✅ Installed"
			if !installed {
				status = "❌ Missing"
			}
			suggestions.WriteString(fmt.Sprintf("- **%s** - %s %s\n", tool.Name, tool.Description, status))
		}
		suggestions.WriteString("\n")
	}

	addToolList("🔥 High Priority (Essential for AI Operations)", highPriority)
	addToolList("⚡ Medium Priority (Significant Enhancement)", mediumPriority)
	addToolList("💡 Lower Priority (Nice to Have)", lowPriority)

	suggestions.WriteString("## Installation Commands\n\n")

	osType := GetOSType()
	suggestions.WriteString(fmt.Sprintf("For your OS (%s):\n\n", osType))

	for _, tool := range im.tools {
		if tool.AIRelevant {
			if installCmd, exists := tool.InstallCmd[osType]; exists {
				suggestions.WriteString(fmt.Sprintf("- **%s**: `%s`\n", tool.Name, strings.Join(installCmd, " ")))
			} else if genericCmd, exists := tool.InstallCmd[runtime.GOOS]; exists {
				suggestions.WriteString(fmt.Sprintf("- **%s**: `%s`\n", tool.Name, strings.Join(genericCmd, " ")))
			}
		}
	}

	return suggestions.String()
}
