package cmd

import (
	"fmt"
	"runtime"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/term"
	"github.com/nexora/nexora/internal/version"
	"github.com/spf13/cobra"
)

var aboutCmd = &cobra.Command{
	Use:   "about",
	Short: "Display information about Nexora",
	Long:  `Display version, platform, license, and community information about Nexora.`,
	Run: func(cmd *cobra.Command, args []string) {
		output := formatAboutInfo()
		cmd.Println(output)
	},
}

func formatAboutInfo() string {
	if !term.IsTerminal(1) {
		// Non-TTY: plain text output
		return formatPlainAbout()
	}
	// TTY: styled output with lipgloss
	return formatStyledAbout()
}

func formatPlainAbout() string {
	return fmt.Sprintf(`Nexora %s
AI-Powered CLI Agent

Platform: %s/%s
Go Version: %s
License: MIT

Community:
  Discord:   https://discord.gg/GCyC6qT79M
  Twitter/X: https://x.com/i/communities/2004598673062216166/
  Reddit:    r/Zackor

Repository:
  GitHub:    https://github.com/jeffersonwarrior/nexora
  Releases:  https://github.com/jeffersonwarrior/nexora/releases
`,
		version.Display(),
		runtime.GOOS,
		runtime.GOARCH,
		runtime.Version(),
	)
}

func formatStyledAbout() string {
	// Styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")).
		MarginTop(1)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Width(12)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("255"))

	linkStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("33")).
		Underline(true)

	// Title
	title := titleStyle.Render(fmt.Sprintf("Nexora %s", version.Display()))
	subtitle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("AI-Powered CLI Agent")

	// Platform info
	platformInfo := fmt.Sprintf("%s%s",
		labelStyle.Render("Platform:"),
		valueStyle.Render(fmt.Sprintf(" %s/%s", runtime.GOOS, runtime.GOARCH)),
	)
	goVersionInfo := fmt.Sprintf("%s%s",
		labelStyle.Render("Go Version:"),
		valueStyle.Render(" "+runtime.Version()),
	)
	licenseInfo := fmt.Sprintf("%s%s",
		labelStyle.Render("License:"),
		valueStyle.Render(" MIT"),
	)

	// Community section
	communityHeader := headerStyle.Render("Community:")
	discordLink := fmt.Sprintf("  %s %s",
		labelStyle.Render("Discord:"),
		linkStyle.Render("https://discord.gg/GCyC6qT79M"),
	)
	twitterLink := fmt.Sprintf("  %s %s",
		labelStyle.Render("Twitter/X:"),
		linkStyle.Render("https://x.com/i/communities/2004598673062216166/"),
	)
	redditLink := fmt.Sprintf("  %s %s",
		labelStyle.Render("Reddit:"),
		valueStyle.Render("r/Zackor"),
	)

	// Repository section
	repoHeader := headerStyle.Render("Repository:")
	githubLink := fmt.Sprintf("  %s %s",
		labelStyle.Render("GitHub:"),
		linkStyle.Render("https://github.com/jeffersonwarrior/nexora"),
	)
	releasesLink := fmt.Sprintf("  %s %s",
		labelStyle.Render("Releases:"),
		linkStyle.Render("https://github.com/jeffersonwarrior/nexora/releases"),
	)

	// Combine all sections
	return fmt.Sprintf(`%s
%s

%s
%s
%s

%s
%s
%s
%s

%s
%s
%s
`,
		title,
		subtitle,
		platformInfo,
		goVersionInfo,
		licenseInfo,
		communityHeader,
		discordLink,
		twitterLink,
		redditLink,
		repoHeader,
		githubLink,
		releasesLink,
	)
}
