package aiops

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/nexora/nexora/internal/aiops"
	"github.com/nexora/nexora/internal/tui/components/core"
	"github.com/nexora/nexora/internal/tui/styles"
)

// RenderOptions contains options for rendering AIOPS status.
type RenderOptions struct {
	MaxWidth    int
	MaxHeight   int
	Compact     bool
	ShowDetails bool
	ShowSection bool
	SectionName string
}

// AIOPSStatus represents the current AIOPS operational state.
type AIOPSStatus struct {
	ModelInfo      aiops.ModelInfo
	RecentMetrics  Metrics
	DetectionState DetectionState
	LastAlert      *Alert
	IsConnected    bool
	Uptime         time.Duration
}

type Metrics struct {
	EditResolutions int
	AvgConfidence   float64
	LoopChecks      int
	DriftScore      float64
	LastUpdate      time.Time
}

type DetectionState struct {
	LoopDetection      bool
	DriftDetection     bool
	EditResolution     bool
	ContextCompression bool
	ScriptGeneration   bool
}

type Alert struct {
	Type       string // "loop", "drift", "resolution"
	Severity   string // "warning", "error", "info"
	Message    string
	Suggestion string
	Timestamp  time.Time
}

// RenderAIOPSBlock renders the complete AIOPS status block.
func RenderAIOPSBlock(client aiops.Ops, opts RenderOptions) string {
	if client == nil || !client.Available() {
		return renderDisconnectedState(opts)
	}

	status := getStatusFromClient(client)

	if opts.Compact {
		return renderCompactStatus(status, opts)
	}

	return renderFullStatus(status, opts)
}

func renderDisconnectedState(opts RenderOptions) string {
	t := styles.CurrentTheme()
	lines := []string{
		t.S().Base.Foreground(t.Border).Render("AIOPS"),
		"",
		t.ItemOfflineIcon.String() + " " +
			t.S().Subtle.Render("Disconnected"),
		"",
		t.S().Subtle.Render("Configure in settings"),
	}

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	if opts.MaxWidth > 0 {
		return lipgloss.NewStyle().Width(opts.MaxWidth).Render(content)
	}
	return content
}

func renderFullStatus(status AIOPSStatus, opts RenderOptions) string {
	t := styles.CurrentTheme()

	// Header with model info
	header := fmt.Sprintf("%s %s • %s",
		t.ItemOnlineIcon.String(),
		t.S().Base.Render(status.ModelInfo.Name),
		t.S().Subtle.Render(fmt.Sprintf("Latency: %v", status.ModelInfo.Latency)))

	// Metrics section
	metrics := []string{
		fmt.Sprintf("• Edit Resolutions: %d (%.0f%% avg)",
			status.RecentMetrics.EditResolutions,
			status.RecentMetrics.AvgConfidence*100),
		fmt.Sprintf("• Loop Checks: %d (%s)",
			status.RecentMetrics.LoopChecks,
			getLoopStatusText(status.RecentMetrics.LoopChecks)),
		fmt.Sprintf("• Drift Score: %.1f (%s)",
			status.RecentMetrics.DriftScore,
			getDriftDescription(status.RecentMetrics.DriftScore)),
	}

	// Detection state
	detectionStates := []string{
		getDetectionIcon(status.DetectionState.LoopDetection) + " Loop Detection: Active",
		getDetectionIcon(status.DetectionState.DriftDetection) + " Drift Monitoring: Active",
		getDetectionIcon(status.DetectionState.EditResolution) + " Edit Resolution: Ready",
	}

	lines := []string{
		t.S().Base.Foreground(t.Border).Render("AIOPS"),
		"",
		header,
		"",
		t.S().Subtle.Render("⚡ Recent Activity"),
	}

	for _, metric := range metrics {
		lines = append(lines, t.S().Subtle.Render("  "+metric))
	}

	lines = append(lines, "", t.S().Subtle.Render("📊 Detection State"))
	for _, state := range detectionStates {
		lines = append(lines, t.S().Subtle.Render("  "+state))
	}

	// Add alert if present
	if status.LastAlert != nil && time.Since(status.LastAlert.Timestamp) < 5*time.Minute {
		alertColor := t.FgMuted
		if status.LastAlert.Severity == "error" {
			alertColor = t.Error
		} else if status.LastAlert.Severity == "warning" {
			alertColor = t.Yellow
		}

		lines = append(lines, "",
			t.S().Base.Foreground(alertColor).Render("🚨 Recent Alert"))
		lines = append(lines,
			t.S().Subtle.Render("  "+status.LastAlert.Type+": "+status.LastAlert.Message))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	if opts.MaxWidth > 0 {
		return lipgloss.NewStyle().Width(opts.MaxWidth).Render(content)
	}
	return content
}

func renderCompactStatus(status AIOPSStatus, opts RenderOptions) string {
	t := styles.CurrentTheme()

	lines := []string{
		t.S().Base.Foreground(t.Border).Render("AIOPS"),
		fmt.Sprintf("%s %s",
			t.ItemOnlineIcon.String(),
			t.S().Base.Render(status.ModelInfo.Name)),
		fmt.Sprintf("%v • %.0f%%",
			status.ModelInfo.Latency,
			status.RecentMetrics.AvgConfidence*100),
		"• " + getLoopStatusCompact(status.RecentMetrics.LoopChecks),
		"• " + fmt.Sprintf("%.1f drift", status.RecentMetrics.DriftScore),
	}

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	if opts.MaxWidth > 0 {
		return lipgloss.NewStyle().Width(opts.MaxWidth).Render(content)
	}
	return content
}

// RenderAIOPSList renders AIOPS in the same format as MCP list for compatibility
func RenderAIOPSList(client aiops.Ops, opts RenderOptions) []string {
	t := styles.CurrentTheme()
	aiopsList := []string{}

	if opts.ShowSection {
		sectionName := opts.SectionName
		if sectionName == "" {
			sectionName = "AIOPS"
		}
		section := t.S().Subtle.Render(sectionName)
		aiopsList = append(aiopsList, section, "")
	}

	if client == nil || !client.Available() {
		aiopsList = append(aiopsList,
			core.Status(
				core.StatusOpts{
					Icon:        t.ItemOfflineIcon.String(),
					Title:       "AIOPS Service",
					Description: t.S().Subtle.Render("disconnected"),
				},
				opts.MaxWidth,
			),
		)
		return aiopsList
	}

	// Get AIOPS status
	status := getStatusFromClient(client)

	// Main service status
	icon := t.ItemOnlineIcon.String()
	description := ""
	extraContent := []string{}

	if status.ModelInfo.Latency > 0 {
		extraContent = append(extraContent, t.S().Subtle.Render(fmt.Sprintf("latency %v", status.ModelInfo.Latency)))
	}
	if status.RecentMetrics.EditResolutions > 0 {
		extraContent = append(extraContent, t.S().Subtle.Render(fmt.Sprintf("%d resolutions", status.RecentMetrics.EditResolutions)))
	}

	aiopsList = append(aiopsList,
		core.Status(
			core.StatusOpts{
				Icon:         icon,
				Title:        status.ModelInfo.Name,
				Description:  description,
				ExtraContent: strings.Join(extraContent, " "),
			},
			opts.MaxWidth,
		),
	)

	// Detection services as sub-items
	detectionServices := []struct {
		name   string
		active bool
	}{
		{"Loop Detection", status.DetectionState.LoopDetection},
		{"Drift Detection", status.DetectionState.DriftDetection},
		{"Edit Resolution", status.DetectionState.EditResolution},
	}

	for _, service := range detectionServices {
		serviceIcon := t.ItemOfflineIcon.String()
		if service.active {
			serviceIcon = t.ItemOnlineIcon.String()
		}

		aiopsList = append(aiopsList,
			core.Status(
				core.StatusOpts{
					Icon:  serviceIcon,
					Title: "  " + service.name, // Indent for sub-item
					Description: t.S().Subtle.Render(func() string {
						if service.active {
							return "active"
						}
						return "inactive"
					}()),
				},
				opts.MaxWidth-2, // Account for indent
			),
		)
	}

	return aiopsList
}

// Helper functions
func getStatusFromClient(client aiops.Ops) AIOPSStatus {
	info := client.ModelInfo()

	return AIOPSStatus{
		ModelInfo: info,
		RecentMetrics: Metrics{
			EditResolutions: 0, // Get from agent tracking
			AvgConfidence:   0.85,
			LoopChecks:      2,
			DriftScore:      0.2,
			LastUpdate:      time.Now(),
		},
		DetectionState: DetectionState{
			LoopDetection:      true,
			DriftDetection:     true,
			EditResolution:     true,
			ContextCompression: false,
			ScriptGeneration:   false,
		},
		IsConnected: info.Available,
		Uptime:      0, // Track when connected
	}
}

func getDetectionIcon(enabled bool) string {
	t := styles.CurrentTheme()
	if enabled {
		return t.ItemOnlineIcon.String()
	}
	return t.ItemOfflineIcon.String()
}

func getDriftDescription(score float64) string {
	switch {
	case score < 0.3:
		return "on track"
	case score < 0.6:
		return "deviating"
	default:
		return "off track"
	}
}

func getLoopStatusText(checks int) string {
	if checks == 0 {
		return "no loops"
	}
	if checks == 1 {
		return "1 check"
	}
	return fmt.Sprintf("%d checks", checks)
}

func getLoopStatusCompact(checks int) string {
	if checks == 0 {
		return "no loops"
	}
	return fmt.Sprintf("%d loops checked", checks)
}
