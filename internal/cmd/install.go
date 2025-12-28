package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/nexora/nexora/internal/version"
	"github.com/spf13/cobra"
)

const (
	githubReleasesAPI = "https://api.github.com/repos/jeffersonwarrior/nexora/releases"
	githubRepoURL     = "https://github.com/jeffersonwarrior/nexora.git"
	installUserAgent  = "nexora-installer/1.0"
)

var installCmd = &cobra.Command{
	Use:   "install [version]",
	Short: "Install or update nexora from GitHub releases",
	Long: `Install or update nexora from tagged releases on GitHub.

Downloads the source code for the specified release and builds it using 'go install'.
By default, installs the latest stable release.`,
	Example: `  # Install latest release
  nexora install

  # Install specific version
  nexora install v0.29.3

  # List available versions
  nexora install --list`,
	RunE: runInstall,
}

var (
	installList  bool
	installForce bool
)

func init() {
	installCmd.Flags().BoolVarP(&installList, "list", "l", false, "List available versions")
	installCmd.Flags().BoolVarP(&installForce, "force", "f", false, "Force reinstall even if same version")
}

type ghRelease struct {
	TagName    string    `json:"tag_name"`
	Name       string    `json:"name"`
	Draft      bool      `json:"draft"`
	Prerelease bool      `json:"prerelease"`
	CreatedAt  time.Time `json:"created_at"`
}

func runInstall(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	if installList {
		return listReleases(ctx)
	}

	targetVersion := ""
	if len(args) > 0 {
		targetVersion = args[0]
	}

	return installFromSource(ctx, targetVersion)
}

func listReleases(ctx context.Context) error {
	releases, err := fetchReleases(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("Available releases (current: %s):\n\n", version.Version)
	for _, r := range releases {
		marker := "  "
		if strings.TrimPrefix(r.TagName, "v") == strings.TrimPrefix(version.Version, "v") {
			marker = "* "
		}
		pre := ""
		if r.Prerelease {
			pre = " (pre-release)"
		}
		fmt.Printf("%s%s%s\n", marker, r.TagName, pre)
	}
	return nil
}

func installFromSource(ctx context.Context, targetVersion string) error {
	var release *ghRelease
	var err error

	if targetVersion == "" {
		release, err = fetchLatestRelease(ctx)
		if err != nil {
			return fmt.Errorf("failed to fetch latest release: %w", err)
		}
		targetVersion = release.TagName
	} else {
		// Validate the version exists
		release, err = fetchRelease(ctx, targetVersion)
		if err != nil {
			return fmt.Errorf("failed to fetch release %s: %w", targetVersion, err)
		}
		targetVersion = release.TagName
	}

	currentVersion := strings.TrimPrefix(version.Version, "v")
	releaseVersion := strings.TrimPrefix(targetVersion, "v")

	if currentVersion == releaseVersion && !installForce {
		fmt.Printf("Already at version %s. Use --force to reinstall.\n", targetVersion)
		return nil
	}

	// Create temp directory for clone
	tmpDir, err := os.MkdirTemp("", "nexora-install-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	repoDir := filepath.Join(tmpDir, "nexora")

	fmt.Printf("Installing nexora %s...\n", targetVersion)

	// Clone the repository
	fmt.Printf("Cloning repository...\n")
	cloneCmd := exec.CommandContext(ctx, "git", "clone", "--depth", "1", "--branch", targetVersion, githubRepoURL, repoDir)
	cloneCmd.Stdout = os.Stdout
	cloneCmd.Stderr = os.Stderr
	if err := cloneCmd.Run(); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	// Run go install
	fmt.Printf("Building and installing...\n")
	installCmd := exec.CommandContext(ctx, "go", "install", ".")
	installCmd.Dir = repoDir
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	installCmd.Env = os.Environ()
	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("failed to build: %w", err)
	}

	fmt.Printf("Successfully installed nexora %s\n", targetVersion)
	return nil
}

func fetchReleases(ctx context.Context) ([]ghRelease, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", githubReleasesAPI, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", installUserAgent)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github api returned status %d: %s", resp.StatusCode, string(body))
	}

	var releases []ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, err
	}

	// Filter out drafts
	filtered := make([]ghRelease, 0, len(releases))
	for _, r := range releases {
		if !r.Draft {
			filtered = append(filtered, r)
		}
	}

	return filtered, nil
}

func fetchLatestRelease(ctx context.Context) (*ghRelease, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", githubReleasesAPI+"/latest", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", installUserAgent)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github api returned status %d: %s", resp.StatusCode, string(body))
	}

	var release ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

func fetchRelease(ctx context.Context, tag string) (*ghRelease, error) {
	if !strings.HasPrefix(tag, "v") {
		tag = "v" + tag
	}

	url := fmt.Sprintf("%s/tags/%s", githubReleasesAPI, tag)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", installUserAgent)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("release %s not found", tag)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github api returned status %d: %s", resp.StatusCode, string(body))
	}

	var release ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}
