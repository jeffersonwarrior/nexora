package prompt

import (
	"cmp"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/nexora/cli/internal/config"
	"github.com/nexora/cli/internal/home"
	"github.com/nexora/cli/internal/shell"
)

// Prompt represents a template-based prompt generator.
type Prompt struct {
	name       string
	template   string
	now        func() time.Time
	platform   string
	workingDir string
}

type PromptDat struct {
	Provider       string
	Model          string
	Config         config.Config
	WorkingDir     string
	IsGitRepo      bool
	Platform       string
	Date           string
	DateTime       string
	GitStatus      string
	ContextFiles   []ContextFile
	CurrentUser    string
	LocalIP        string
	PythonVersion  string
	NodeVersion    string
	GoVersion      string
	ShellType      string
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

type ContextFile struct {
	Path    string
	Content string
}

type Option func(*Prompt)

func WithTimeFunc(fn func() time.Time) Option {
	return func(p *Prompt) {
		p.now = fn
	}
}

func WithPlatform(platform string) Option {
	return func(p *Prompt) {
		p.platform = platform
	}
}

func WithWorkingDir(workingDir string) Option {
	return func(p *Prompt) {
		p.workingDir = workingDir
	}
}

func NewPrompt(name, promptTemplate string, opts ...Option) (*Prompt, error) {
	p := &Prompt{
		name:     name,
		template: promptTemplate,
		now:      time.Now,
	}
	for _, opt := range opts {
		opt(p)
	}
	return p, nil
}

func (p *Prompt) Build(ctx context.Context, provider, model string, cfg config.Config) (string, error) {
	t, err := template.New(p.name).Parse(p.template)
	if err != nil {
		return "", fmt.Errorf("parsing template: %w", err)
	}
	var sb strings.Builder
	d, err := p.promptData(ctx, provider, model, cfg)
	if err != nil {
		return "", err
	}
	if err := t.Execute(&sb, d); err != nil {
		return "", fmt.Errorf("executing template: %w", err)
	}

	return sb.String(), nil
}

func processFile(filePath string) *ContextFile {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}
	return &ContextFile{
		Path:    filePath,
		Content: string(content),
	}
}

func processContextPath(p string, cfg config.Config) []ContextFile {
	var contexts []ContextFile
	fullPath := p
	if !filepath.IsAbs(p) {
		fullPath = filepath.Join(cfg.WorkingDir(), p)
	}
	info, err := os.Stat(fullPath)
	if err != nil {
		return contexts
	}
	if info.IsDir() {
		filepath.WalkDir(fullPath, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() {
				if result := processFile(path); result != nil {
					contexts = append(contexts, *result)
				}
			}
			return nil
		})
	} else {
		result := processFile(fullPath)
		if result != nil {
			contexts = append(contexts, *result)
		}
	}
	return contexts
}

// expandPath expands ~ and environment variables in file paths
func expandPath(path string, cfg config.Config) string {
	path = home.Long(path)
	// Handle environment variable expansion using the same pattern as config
	if strings.HasPrefix(path, "$") {
		if expanded, err := cfg.Resolver().ResolveValue(path); err == nil {
			path = expanded
		}
	}

	return path
}

func (p *Prompt) promptData(ctx context.Context, provider, model string, cfg config.Config) (PromptDat, error) {
	workingDir := cmp.Or(p.workingDir, cfg.WorkingDir())
	platform := cmp.Or(p.platform, runtime.GOOS)

	files := map[string][]ContextFile{}

	for _, pth := range cfg.Options.ContextPaths {
		expanded := expandPath(pth, cfg)
		pathKey := strings.ToLower(expanded)
		if _, ok := files[pathKey]; ok {
			continue
		}
		content := processContextPath(expanded, cfg)
		files[pathKey] = content
	}

	isGit := isGitRepo(cfg.WorkingDir())

	// Gather environment information
	now := p.now()
	currentUser := getCurrentUser()
	localIP := getLocalIP(ctx)
	pythonVer := getRuntimeVersion(ctx, "python3 --version")
	nodeVer := getRuntimeVersion(ctx, "node --version")
	goVer := getRuntimeVersion(ctx, "go version")
	if goVer != "not installed" {
		// Extract just the version number from "go version go1.x.x ..."
		parts := strings.Fields(goVer)
		if len(parts) >= 3 {
			goVer = parts[2]
		}
	}
	shellType := "bash (mvdan/sh)"
	gitUserName := ""
	gitUserEmail := ""
	if isGit {
		gitUserName = getGitConfig(ctx, "user.name")
		gitUserEmail = getGitConfig(ctx, "user.email")
	}
	memInfo := getMemoryInfo(ctx)
	diskInfo := getDiskInfo(ctx, workingDir)

	// New environment detection
	arch := getArchitecture()
	container := detectContainer(ctx)
	terminal := getTerminalInfo(ctx)
	// Lazy-load expensive operations - only if DEBUG env var set
	network := "online"  // Default assumption, skip expensive ping
	services := ""       // Skip expensive service detection by default
	if os.Getenv("NEXORA_FULL_ENV") == "1" {
		network = getNetworkStatus(ctx)
		services = detectActiveServices(ctx)
	}

	data := PromptDat{
		Provider:       provider,
		Model:          model,
		Config:         cfg,
		WorkingDir:     filepath.ToSlash(workingDir),
		IsGitRepo:      isGit,
		Platform:       platform,
		Date:           now.Format("1/2/2006"),
		DateTime:       now.Format("2006-01-02 15:04:05 MST"),
		CurrentUser:    currentUser,
		LocalIP:        localIP,
		PythonVersion:  pythonVer,
		NodeVersion:    nodeVer,
		GoVersion:      goVer,
		ShellType:      shellType,
		GitUserName:    gitUserName,
		GitUserEmail:   gitUserEmail,
		MemoryInfo:     memInfo,
		DiskInfo:       diskInfo,
		Architecture:   arch,
		ContainerType:  container,
		TerminalInfo:   terminal,
		NetworkStatus:  network,
		ActiveServices: services,
	}
	if isGit {
		var err error
		data.GitStatus, err = getGitStatus(ctx, cfg.WorkingDir())
		if err != nil {
			return PromptDat{}, err
		}
	}

	for _, contextFiles := range files {
		data.ContextFiles = append(data.ContextFiles, contextFiles...)
	}
	return data, nil
}

func isGitRepo(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil
}

func getGitStatus(ctx context.Context, dir string) (string, error) {
	sh := shell.NewShell(&shell.Options{
		WorkingDir: dir,
	})
	branch, err := getGitBranch(ctx, sh)
	if err != nil {
		return "", err
	}
	status, err := getGitStatusSummary(ctx, sh)
	if err != nil {
		return "", err
	}
	commits, err := getGitRecentCommits(ctx, sh)
	if err != nil {
		return "", err
	}
	return branch + status + commits, nil
}

func getGitBranch(ctx context.Context, sh *shell.Shell) (string, error) {
	out, _, err := sh.Exec(ctx, "git branch --show-current 2>/dev/null")
	if err != nil {
		return "", nil
	}
	out = strings.TrimSpace(out)
	if out == "" {
		return "", nil
	}
	return fmt.Sprintf("Current branch: %s\n", out), nil
}

func getGitStatusSummary(ctx context.Context, sh *shell.Shell) (string, error) {
	// Reduced from 20 to 5 files for token efficiency
	out, _, err := sh.Exec(ctx, "git status --short 2>/dev/null | head -5")
	if err != nil {
		return "", nil
	}
	out = strings.TrimSpace(out)
	if out == "" {
		return "Status: clean\n", nil
	}
	return fmt.Sprintf("Status:\n%s\n", out), nil
}

func getGitRecentCommits(ctx context.Context, sh *shell.Shell) (string, error) {
	// Reduced from 3 to 2 commits for token efficiency
	out, _, err := sh.Exec(ctx, "git log --oneline -n 2 2>/dev/null")
	if err != nil || out == "" {
		return "", nil
	}
	out = strings.TrimSpace(out)
	return fmt.Sprintf("Recent commits:\n%s\n", out), nil
}

// getCommandOutput runs a command and returns its trimmed output, or fallback on error
func getCommandOutput(ctx context.Context, cmd string, fallback string) string {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return fallback
	}

	execCmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	out, err := execCmd.Output()
	if err != nil {
		return fallback
	}

	result := strings.TrimSpace(string(out))
	if result == "" {
		return fallback
	}
	return result
}

// getCurrentUser returns the current username
func getCurrentUser() string {
	if u, err := user.Current(); err == nil {
		return u.Username
	}
	return "unknown"
}

// getLocalIP attempts to get the local IP address
func getLocalIP(ctx context.Context) string {
	// Try Linux/Unix approach first
	ip := getCommandOutput(ctx, "hostname -I", "")
	if ip != "" {
		// Get first IP
		parts := strings.Fields(ip)
		if len(parts) > 0 {
			return parts[0]
		}
	}

	// Try macOS approach
	ip = getCommandOutput(ctx, "ipconfig getifaddr en0", "")
	if ip != "" {
		return ip
	}

	// Fallback
	return "unavailable"
}

// getRuntimeVersion gets version info for a runtime
func getRuntimeVersion(ctx context.Context, cmd string) string {
	return getCommandOutput(ctx, cmd, "not installed")
}

// getGitConfig gets git configuration value
func getGitConfig(ctx context.Context, key string) string {
	return getCommandOutput(ctx, fmt.Sprintf("git config --get %s", key), "not configured")
}

// getMemoryInfo returns system memory information
func getMemoryInfo(ctx context.Context) string {
	// Try Linux free command
	out := getCommandOutput(ctx, "free -h", "")
	if out != "" {
		lines := strings.Split(out, "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "Mem:") {
				fields := strings.Fields(line)
				if len(fields) >= 7 {
					return fmt.Sprintf("%s total, %s available", fields[1], fields[6])
				}
			}
		}
	}

	// Try macOS vm_stat
	out = getCommandOutput(ctx, "vm_stat", "")
	if out != "" {
		return "available (macOS)"
	}

	return "unavailable"
}

// getDiskInfo returns disk space information for current directory
func getDiskInfo(ctx context.Context, workingDir string) string {
	// Try df command
	cmd := exec.CommandContext(ctx, "df", "-h", workingDir)
	out, err := cmd.Output()
	if err != nil {
		return "unavailable"
	}

	lines := strings.Split(string(out), "\n")
	if len(lines) >= 2 {
		fields := strings.Fields(lines[1])
		if len(fields) >= 4 {
			return fmt.Sprintf("%s total, %s free", fields[1], fields[3])
		}
	}

	return "unavailable"
}

// getArchitecture returns CPU architecture
func getArchitecture() string {
	return runtime.GOARCH
}

// detectContainer returns container type if running in one
func detectContainer(ctx context.Context) string {
	// Check for Docker
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return "Docker"
	}

	// Check cgroup for docker/podman/kubernetes
	if data, err := os.ReadFile("/proc/1/cgroup"); err == nil {
		content := string(data)
		if strings.Contains(content, "docker") {
			return "Docker"
		}
		if strings.Contains(content, "podman") {
			return "Podman"
		}
		if strings.Contains(content, "kubepods") {
			return "Kubernetes"
		}
	}

	// Check for container env vars
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		return "Kubernetes"
	}
	if os.Getenv("container") != "" {
		return os.Getenv("container")
	}

	return "none"
}

// getTerminalInfo returns terminal capabilities
func getTerminalInfo(ctx context.Context) string {
	var info []string

	// Check color support
	term := os.Getenv("TERM")
	if term != "" {
		if strings.Contains(term, "color") || strings.Contains(term, "256") || term == "xterm" {
			info = append(info, "color")
		}
	}

	// Check terminal dimensions
	cols := getCommandOutput(ctx, "tput cols", "")
	lines := getCommandOutput(ctx, "tput lines", "")
	if cols != "" && lines != "" {
		info = append(info, fmt.Sprintf("%sx%s", cols, lines))
	}

	// Check if interactive
	if term != "" && term != "dumb" {
		info = append(info, "interactive")
	}

	if len(info) == 0 {
		return "basic"
	}

	return strings.Join(info, ", ")
}

// getNetworkStatus checks internet connectivity
func getNetworkStatus(ctx context.Context) string {
	// Create a context with short timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// Try to resolve a common DNS (quick check)
	cmd := exec.CommandContext(timeoutCtx, "ping", "-c", "1", "-W", "1", "8.8.8.8")
	out, err := cmd.Output()
	if err == nil && !strings.Contains(string(out), "100% packet loss") {
		return "online"
	}

	// Quick fallback to checking if network interface is up
	out, err = exec.CommandContext(timeoutCtx, "ip", "link", "show").Output()
	if err == nil && strings.Contains(string(out), "state UP") {
		return "network up"
	}

	return "offline or restricted"
}

// detectActiveServices checks for common development services
func detectActiveServices(ctx context.Context) string {
	// Create a context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	var services []string

	// Check for common ports using a single ss/netstat command
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		cmd := exec.CommandContext(timeoutCtx, "sh", "-c", "ss -ln 2>/dev/null | grep -E ':(5432|3306|6379|27017|9200|8080|3000|4200|5000|8000)' || netstat -an 2>/dev/null | grep -E ':(5432|3306|6379|27017|9200|8080|3000|4200|5000|8000)'")
		out, err := cmd.Output()
		if err == nil {
			portMap := map[string]string{
				"5432":  "PostgreSQL",
				"3306":  "MySQL",
				"6379":  "Redis",
				"27017": "MongoDB",
				"9200":  "Elasticsearch",
				"8080":  "HTTP:8080",
				"3000":  "HTTP:3000",
				"4200":  "HTTP:4200",
				"5000":  "HTTP:5000",
				"8000":  "HTTP:8000",
			}

			found := make(map[string]bool)
			for port, name := range portMap {
				if strings.Contains(string(out), ":"+port) && !found[name] {
					services = append(services, name)
					found[name] = true
				}
			}
		}
	}

	// Quick Docker check
	cmd := exec.CommandContext(timeoutCtx, "docker", "info")
	if err := cmd.Run(); err == nil {
		services = append(services, "Docker")
	}

	if len(services) == 0 {
		return "none detected"
	}

	// Limit to first 5 to keep output concise
	if len(services) > 5 {
		services = services[:5]
		return strings.Join(services, ", ") + ", ..."
	}

	return strings.Join(services, ", ")
}

func (p *Prompt) Name() string {
	return p.name
}
