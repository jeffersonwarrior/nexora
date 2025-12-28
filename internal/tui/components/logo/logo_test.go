package logo

import (
	"image/color"
	"strings"
	"testing"
)

func TestRender(t *testing.T) {
	tests := []struct {
		name    string
		version string
		compact bool
		opts    Opts
		check   func(string) bool
	}{
		{
			name:    "compact version",
			version: "v0.1.0",
			compact: true,
			opts: Opts{
				CharmColor: color.RGBA{R: 255, G: 0, B: 0, A: 255},
			},
			check: func(s string) bool {
				return strings.Contains(s, "NEXORA") && strings.Contains(s, "v0.1.0")
			},
		},
		{
			name:    "wide version with width",
			version: "v0.2.0",
			compact: false,
			opts: Opts{
				CharmColor: color.RGBA{R: 0, G: 255, B: 0, A: 255},
				Width:      100,
			},
			check: func(s string) bool {
				return strings.Contains(s, "NEXORA") && len(s) > 0
			},
		},
		{
			name:    "no width",
			version: "v0.3.0",
			compact: false,
			opts: Opts{
				CharmColor: color.RGBA{R: 0, G: 0, B: 255, A: 255},
				Width:      0,
			},
			check: func(s string) bool {
				return strings.Contains(s, "NEXORA")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Render(tt.version, tt.compact, tt.opts)
			if !tt.check(result) {
				t.Errorf("Render failed: got %q", result)
			}
		})
	}
}

func TestSmallRender(t *testing.T) {
	tests := []struct {
		name  string
		width int
		check func(string) bool
	}{
		{
			name:  "with width",
			width: 40,
			check: func(s string) bool {
				return strings.Contains(s, "NEXORA") && len(s) > 0
			},
		},
		{
			name:  "zero width",
			width: 0,
			check: func(s string) bool {
				return strings.Contains(s, "NEXORA")
			},
		},
		{
			name:  "small width",
			width: 15,
			check: func(s string) bool {
				return strings.Contains(s, "NEXORA")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SmallRender(tt.width)
			if !tt.check(result) {
				t.Errorf("SmallRender failed: got %q", result)
			}
		})
	}
}

func TestRenderWord(t *testing.T) {
	tests := []struct {
		name        string
		spacing     int
		stretchIdx  int
		letterforms int
		check       func(string) bool
	}{
		{
			name:        "no stretch, no spacing",
			spacing:     0,
			stretchIdx:  -1,
			letterforms: 3,
			check: func(s string) bool {
				return len(s) > 0
			},
		},
		{
			name:        "with spacing",
			spacing:     2,
			stretchIdx:  0,
			letterforms: 2,
			check: func(s string) bool {
				return len(s) > 0
			},
		},
		{
			name:        "negative spacing defaults to 0",
			spacing:     -5,
			stretchIdx:  1,
			letterforms: 2,
			check: func(s string) bool {
				return len(s) > 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create dummy letterforms
			forms := make([]letterform, tt.letterforms)
			for i := range forms {
				forms[i] = func(bool) string { return "X" }
			}
			result := renderWord(tt.spacing, tt.stretchIdx, forms...)
			if !tt.check(result) {
				t.Errorf("renderWord failed: got %q", result)
			}
		})
	}
}

func TestLetterC(t *testing.T) {
	tests := []struct {
		name    string
		stretch bool
		check   func(string) bool
	}{
		{
			name:    "without stretch",
			stretch: false,
			check: func(s string) bool {
				return len(s) > 0
			},
		},
		{
			name:    "with stretch",
			stretch: true,
			check: func(s string) bool {
				return len(s) > 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := letterC(tt.stretch)
			if !tt.check(result) {
				t.Errorf("letterC failed: got %q", result)
			}
		})
	}
}

func TestLetterH(t *testing.T) {
	tests := []struct {
		name    string
		stretch bool
		check   func(string) bool
	}{
		{
			name:    "without stretch",
			stretch: false,
			check: func(s string) bool {
				return len(s) > 0
			},
		},
		{
			name:    "with stretch",
			stretch: true,
			check: func(s string) bool {
				return len(s) > 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := letterH(tt.stretch)
			if !tt.check(result) {
				t.Errorf("letterH failed: got %q", result)
			}
		})
	}
}

func TestLetterR(t *testing.T) {
	tests := []struct {
		name    string
		stretch bool
		check   func(string) bool
	}{
		{
			name:    "without stretch",
			stretch: false,
			check: func(s string) bool {
				return len(s) > 0
			},
		},
		{
			name:    "with stretch",
			stretch: true,
			check: func(s string) bool {
				return len(s) > 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := letterR(tt.stretch)
			if !tt.check(result) {
				t.Errorf("letterR failed: got %q", result)
			}
		})
	}
}

func TestLetterSStylized(t *testing.T) {
	tests := []struct {
		name    string
		stretch bool
		check   func(string) bool
	}{
		{
			name:    "without stretch",
			stretch: false,
			check: func(s string) bool {
				return len(s) > 0
			},
		},
		{
			name:    "with stretch",
			stretch: true,
			check: func(s string) bool {
				return len(s) > 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := letterSStylized(tt.stretch)
			if !tt.check(result) {
				t.Errorf("letterSStylized failed: got %q", result)
			}
		})
	}
}

func TestLetterU(t *testing.T) {
	tests := []struct {
		name    string
		stretch bool
		check   func(string) bool
	}{
		{
			name:    "without stretch",
			stretch: false,
			check: func(s string) bool {
				return len(s) > 0
			},
		},
		{
			name:    "with stretch",
			stretch: true,
			check: func(s string) bool {
				return len(s) > 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := letterU(tt.stretch)
			if !tt.check(result) {
				t.Errorf("letterU failed: got %q", result)
			}
		})
	}
}

func TestCachedRandN(t *testing.T) {
	tests := []struct {
		name string
		n    int
		check func(int) bool
	}{
		{
			name: "positive n",
			n:    10,
			check: func(v int) bool {
				return v >= 0 && v < 10
			},
		},
		{
			name: "zero n",
			n:    0,
			check: func(v int) bool {
				return v == 0
			},
		},
		{
			name: "negative n",
			n:    -5,
			check: func(v int) bool {
				return v == 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cachedRandN(tt.n)
			if !tt.check(result) {
				t.Errorf("cachedRandN(%d) = %d, unexpected", tt.n, result)
			}
		})
	}
}

func TestMax(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int
		expected int
	}{
		{"a > b", 10, 5, 10},
		{"a < b", 5, 10, 10},
		{"a == b", 7, 7, 7},
		{"both negative", -5, -10, -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := max(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("max(%d, %d) = %d, want %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}
