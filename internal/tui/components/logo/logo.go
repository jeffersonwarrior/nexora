// Package logo renders a Nexora wordmark in a stylized way.
package logo

import (
	"image/color"
	"math/rand"
	"strings"
	"sync"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/charmbracelet/x/exp/slice"
	"github.com/nexora/cli/internal/tui/styles"
)

// letterform represents a letterform. It can be stretched horizontally by
// a given amount via the boolean argument.
type letterform func(bool) string

const diag = `╱`

const nexora = "Powered by NEXORA"

var (
	randSourceMu sync.Mutex
	randSource   *rand.Rand
)

func init() {
	randSource = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func cachedRandN(n int) int {
	randSourceMu.Lock()
	defer randSourceMu.Unlock()
	if n <= 0 {
		return 0
	}
	return randSource.Intn(n)
}

// Opts are the options for rendering the Nexora title art.
type Opts struct {
	FieldColor   color.Color // diagonal lines
	TitleColorA  color.Color // left gradient ramp point
	TitleColorB  color.Color // right gradient ramp point
	CharmColor   color.Color // NEXORA text color
	VersionColor color.Color // Version text color
	Width        int         // width of the rendered logo, used for truncation
}

// Render renders the Nexora logo.
func Render(version string, compact bool, o Opts) string {
	fg := func(c color.Color, s string) string {
		return lipgloss.NewStyle().Foreground(c).Render(s)
	}

	// Simple one-line logo: /// NEXORA <v> ///
	logoText := "/// NEXORA " + version + " ///"
	logoText = fg(o.CharmColor, logoText)

	if compact {
		return logoText
	}

	// For the wide version, just center it on the screen
	if o.Width > 0 {
		padding := max(0, (o.Width-lipgloss.Width(logoText))/2)
		logoText = strings.Repeat(" ", padding) + logoText
	}
	return logoText
}

// SmallRender renders a smaller version of the Nexora logo, suitable for
// smaller windows or sidebar usage.
func SmallRender(width int) string {
	t := styles.CurrentTheme()
	logoText := t.S().Base.Foreground(t.Primary).Render("/// NEXORA")
	if width > 0 {
		remainingWidth := width - lipgloss.Width(logoText) - 3 // 3 for " //"
		if remainingWidth > 0 {
			logoText += " //"
			lines := strings.Repeat("╱", remainingWidth)
			logoText += t.S().Base.Foreground(t.Secondary).Render(lines)
		}
	}
	return logoText
}

// renderWord renders letterforms to fork a word. stretchIndex is the index of
// the letter to stretch, or -1 if no letter should be stretched.
func renderWord(spacing int, stretchIndex int, letterforms ...letterform) string {
	if spacing < 0 {
		spacing = 0
	}

	renderedLetterforms := make([]string, len(letterforms))

	// pick one letter randomly to stretch
	for i, letter := range letterforms {
		renderedLetterforms[i] = letter(i == stretchIndex)
	}

	if spacing > 0 {
		// Add spaces between the letters and render.
		renderedLetterforms = slice.Intersperse(renderedLetterforms, strings.Repeat(" ", spacing))
	}
	return strings.TrimSpace(
		lipgloss.JoinHorizontal(lipgloss.Top, renderedLetterforms...),
	)
}

// letterC renders the letter C in a stylized way. It takes an integer that
// determines how many cells to stretch the letter. If the stretch is less than
// 1, defaults to no stretching.
func letterC(stretch bool) string {
	// Here's what we're making:
	//
	// ▄▀▀▀▀
	// █
	//	▀▀▀▀

	left := heredoc.Doc(`
		▄
		█
	`)
	right := heredoc.Doc(`
		▀

		▀
	`)
	return joinLetterform(
		left,
		stretchLetterformPart(right, letterformProps{
			stretch:    stretch,
			width:      4,
			minStretch: 7,
			maxStretch: 12,
		}),
	)
}

// letterH renders the letter H in a stylized way. It takes an integer that
// determines how many cells to stretch the letter. If the stretch is less than
// 1, defaults to no stretching.
func letterH(stretch bool) string {
	// Here's what we're making:
	//
	// █   █
	// █▀▀▀█
	// ▀   ▀

	side := heredoc.Doc(`
		█
		█
		▀`)
	middle := heredoc.Doc(`

		▀
	`)
	return joinLetterform(
		side,
		stretchLetterformPart(middle, letterformProps{
			stretch:    stretch,
			width:      3,
			minStretch: 8,
			maxStretch: 12,
		}),
		side,
	)
}

// letterR renders the letter R in a stylized way. It takes an integer that
// determines how many cells to stretch the letter. If the stretch is less than
// 1, defaults to no stretching.
func letterR(stretch bool) string {
	// Here's what we're making:
	//
	// █▀▀▀▄
	// █▀▀▀▄
	// ▀   ▀

	left := heredoc.Doc(`
		█
		█
		▀
	`)
	center := heredoc.Doc(`
		▀
		▀
	`)
	right := heredoc.Doc(`
		▄
		▄
		▀
	`)
	return joinLetterform(
		left,
		stretchLetterformPart(center, letterformProps{
			stretch:    stretch,
			width:      3,
			minStretch: 7,
			maxStretch: 12,
		}),
		right,
	)
}

// letterSStylized renders the letter S in a stylized way, more so than
// [letterS]. It takes an integer that determines how many cells to stretch the
// letter. If the stretch is less than 1, defaults to no stretching.
func letterSStylized(stretch bool) string {
	// Here's what we're making:
	//
	// ▄▀▀▀▀▀
	// ▀▀▀▀▀█
	// ▀▀▀▀▀

	left := heredoc.Doc(`
		▄
		▀
		▀
	`)
	center := heredoc.Doc(`
		▀
		▀
		▀
	`)
	right := heredoc.Doc(`
		▀
		█
	`)
	return joinLetterform(
		left,
		stretchLetterformPart(center, letterformProps{
			stretch:    stretch,
			width:      3,
			minStretch: 7,
			maxStretch: 12,
		}),
		right,
	)
}

// letterU renders the letter U in a stylized way. It takes an integer that
// determines how many cells to stretch the letter. If the stretch is less than
// 1, defaults to no stretching.
func letterU(stretch bool) string {
	// Here's what we're making:
	//
	// █   █
	// █   █
	//	▀▀▀

	side := heredoc.Doc(`
		█
		█
	`)
	middle := heredoc.Doc(`


		▀
	`)
	return joinLetterform(
		side,
		stretchLetterformPart(middle, letterformProps{
			stretch:    stretch,
			width:      3,
			minStretch: 7,
			maxStretch: 12,
		}),
		side,
	)
}

func joinLetterform(letters ...string) string {
	return lipgloss.JoinHorizontal(lipgloss.Top, letters...)
}

// letterformProps defines letterform stretching properties.
// for readability.
type letterformProps struct {
	width      int
	minStretch int
	maxStretch int
	stretch    bool
}

// stretchLetterformPart is a helper function for letter stretching. If randomize
// is false the minimum number will be used.
func stretchLetterformPart(s string, p letterformProps) string {
	if p.maxStretch < p.minStretch {
		p.minStretch, p.maxStretch = p.maxStretch, p.minStretch
	}
	n := p.width
	if p.stretch {
		n = cachedRandN(p.maxStretch-p.minStretch) + p.minStretch //nolint:gosec
	}
	parts := make([]string, n)
	for i := range parts {
		parts[i] = s
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}
