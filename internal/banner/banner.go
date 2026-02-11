// Package banner converts text into ASCII-art banner strings using bitmap font data.
package banner

import (
	"strings"

	mcolor "github.com/qraqras/misaki-banner/internal/color"
	mfont "github.com/qraqras/misaki-banner/internal/font"
)

// ShadowMode selects the shadow rendering style.
type ShadowMode string

const (
	ShadowNone    ShadowMode = ""        // `██ `
	ShadowOutline ShadowMode = "outline" // `██╗`
	ShadowSolid   ShadowMode = "solid"   // `░░▄`
)

// Options controls how the banner is rendered.
type Options struct {
	Shadow   ShadowMode // shadow rendering style
	Color    string     // text color (RGB format "r,g,b" or preset name)
	Gradient bool       // enable gradient effect (light to dark)
}

// glyphInfo holds bitmap and width information for a single glyph.
type glyphInfo struct {
	bitmap [][]bool
	width  int
}

// Generate creates an ASCII-art banner string from the given text.
// If text contains newlines, each line is rendered separately and joined.
func Generate(face *mfont.Face, text string, opts Options) string {
	// Handle multi-line text
	if strings.Contains(text, "\n") {
		var parts []string
		for _, line := range strings.Split(text, "\n") {
			if line == "" {
				continue
			}
			parts = append(parts, Generate(face, line, opts))
		}
		return strings.Join(parts, "\n\n")
	}

	runes := []rune(text)
	if len(runes) == 0 {
		return ""
	}

	height := face.FontSize()

	// Collect trimmed bitmaps
	var glyphs []glyphInfo
	for _, r := range runes {
		bm := face.RuneBitmap(r)
		w := 0
		if len(bm) > 0 {
			w = len(bm[0])
		}
		glyphs = append(glyphs, glyphInfo{bitmap: bm, width: w})
	}

	// Calculate total width
	totalWidth := 0
	for _, g := range glyphs {
		totalWidth += g.width
	}

	// Build a combined 2D bool grid
	grid := make([][]bool, height)
	for y := 0; y < height; y++ {
		grid[y] = make([]bool, totalWidth)
		xOff := 0
		for _, g := range glyphs {
			for x := 0; x < g.width; x++ {
				if y < len(g.bitmap) && x < len(g.bitmap[y]) && g.bitmap[y][x] {
					grid[y][xOff+x] = true
				}
			}
			xOff += g.width
		}
	}

	switch opts.Shadow {
	case ShadowOutline:
		lines := renderWithCharSet(grid, height, totalWidth, opts, getCharSet(ShadowOutline))
		return trimBlankLines(lines)
	case ShadowSolid:
		lines := renderWithCharSet(grid, height, totalWidth, opts, getCharSet(ShadowSolid))
		return trimBlankLines(lines)
	default:
		lines := renderWithCharSet(grid, height, totalWidth, opts, getCharSet(ShadowNone))
		return trimBlankLines(lines)
	}
}

// colorPixel returns the string wrapped with the ANSI color for the given pixel.
// It calculates the gradient based on the pixel's position across the entire string.
func colorPixel(s string, y, x int, totalWidth int, opts Options, baseColor mcolor.RGB, hasColor bool) string {
	if !hasColor {
		return s
	}

	var color mcolor.RGB

	if opts.Gradient {
		// Gradient mode: left to right, slightly brighter to slightly darker
		if totalWidth <= 1 {
			color = baseColor
		} else {
			// Normalize x position to [0, 1]
			t := float64(x) / float64(totalWidth-1)
			if t < 0 {
				t = 0
			}
			if t > 1 {
				t = 1
			}

			// Create a natural gradient by shifting hue only
			// Left: +18 degrees hue
			// Center: 0 degrees hue (base color)
			// Right: -18 degrees hue
			hueDelta := 18.0 - (36.0 * t) // +18 to -18
			lightDelta := 0.0             // No lightness change
			color = mcolor.ShiftColor(baseColor, hueDelta, lightDelta)
		}
	} else {
		// Single color mode
		color = baseColor
	}

	return color.ANSI() + s + mcolor.Reset
}

// charSet defines the character set for different rendering modes.
type charSet struct {
	textOn          string // character for main text pixels
	textOff         string // character for empty pixels
	shadowLeftAbove string // left && above
	shadowLeftDiag  string // left && diagonal
	shadowLeft      string // left only
	shadowAboveDiag string // above && diagonal
	shadowAbove     string // above only
	shadowDiag      string // diagonal only
}

// getCharSet returns the character set for the given shadow mode.
func getCharSet(mode ShadowMode) charSet {
	switch mode {
	case ShadowOutline:
		return charSet{
			textOn:          "██",
			textOff:         "  ",
			shadowLeftAbove: "╔═",
			shadowLeftDiag:  "║ ",
			shadowLeft:      "╗ ",
			shadowAboveDiag: "══",
			shadowAbove:     "╚═",
			shadowDiag:      "╝ ",
		}
	case ShadowSolid:
		return charSet{
			textOn:          "░░",
			textOff:         "  ",
			shadowLeftAbove: "█▀",
			shadowLeftDiag:  "█ ",
			shadowLeft:      "▄ ",
			shadowAboveDiag: "▀▀",
			shadowAbove:     " ▀",
			shadowDiag:      "▀ ",
		}
	default:
		return charSet{
			textOn:  "██",
			textOff: "  ",
		}
	}
}

// renderWithCharSet renders a glyph grid with the given character set.
func renderWithCharSet(grid [][]bool, h, w int, opts Options, chars charSet) []string {
	// For non-shadow modes, use simple rendering
	if chars.shadowLeftAbove == "" {
		return renderSimple(grid, h, w, opts, chars)
	}
	// For shadow modes, use shadow rendering
	return renderShadow(grid, h, w, opts, chars)
}

// renderSimple renders a glyph grid without shadow effects.
func renderSimple(grid [][]bool, height, width int, opts Options, chars charSet) []string {
	lines := make([]string, height)
	for y := 0; y < height; y++ {
		var sb strings.Builder
		for x := 0; x < width; x++ {
			if grid[y][x] {
				// Parse base color
				var baseColor mcolor.RGB
				hasColor := false
				if opts.Color != "" {
					var err error
					baseColor, err = mcolor.ParseColor(opts.Color)
					if err == nil {
						hasColor = true
					}
				}
				sb.WriteString(colorPixel(chars.textOn, y, x, width, opts, baseColor, hasColor))
			} else {
				sb.WriteString(chars.textOff)
			}
		}
		lines[y] = sb.String()
	}
	return lines
}

// renderShadow renders a glyph grid with shadow effects as a string array.
func renderShadow(grid [][]bool, h, w int, opts Options, chars charSet) []string {
	isOn := func(y, x int) bool {
		if y < 0 || y >= h || x < 0 || x >= w {
			return false
		}
		return grid[y][x]
	}

	// Extended canvas: +1 row for bottom shadow, +1 col for right shadow
	outH := h + 1
	outW := w + 1

	lines := make([]string, outH)
	for y := 0; y < outH; y++ {
		var sb strings.Builder
		for x := 0; x < outW; x++ {
			if isOn(y, x) {
				// Parse base color
				var baseColor mcolor.RGB
				hasColor := false
				if opts.Color != "" {
					var err error
					baseColor, err = mcolor.ParseColor(opts.Color)
					if err == nil {
						hasColor = true
					}
				}
				sb.WriteString(colorPixel(chars.textOn, y, x, w, opts, baseColor, hasColor))
				continue
			}

			left := isOn(y, x-1)       // left
			above := isOn(y-1, x)      // above
			diagonal := isOn(y-1, x-1) // diagonal (above-left)

			// Get color from adjacent pixel for shadow
			var baseColor mcolor.RGB
			hasColor := false
			if opts.Color != "" {
				var err error
				baseColor, err = mcolor.ParseColor(opts.Color)
				if err == nil {
					hasColor = true
				}
			}

			var shadowStr string
			switch {
			case left && above:
				shadowStr = chars.shadowLeftAbove
				shadowStr = colorPixel(shadowStr, y, x-1, w, opts, baseColor, hasColor)
			case left && diagonal:
				shadowStr = chars.shadowLeftDiag
				shadowStr = colorPixel(shadowStr, y, x-1, w, opts, baseColor, hasColor)
			case left:
				shadowStr = chars.shadowLeft
				shadowStr = colorPixel(shadowStr, y, x-1, w, opts, baseColor, hasColor)
			case above && diagonal:
				shadowStr = chars.shadowAboveDiag
				shadowStr = colorPixel(shadowStr, y-1, x, w, opts, baseColor, hasColor)
			case above:
				shadowStr = chars.shadowAbove
				shadowStr = colorPixel(shadowStr, y-1, x, w, opts, baseColor, hasColor)
			case diagonal:
				shadowStr = chars.shadowDiag
				shadowStr = colorPixel(shadowStr, y-1, x-1, w, opts, baseColor, hasColor)
			default:
				shadowStr = chars.textOff
			}
			sb.WriteString(shadowStr)
		}
		lines[y] = sb.String()
	}
	return lines
}

func trimBlankLines(lines []string) string {
	// Trim trailing blank lines
	for len(lines) > 0 {
		if strings.TrimSpace(lines[len(lines)-1]) == "" {
			lines = lines[:len(lines)-1]
		} else {
			break
		}
	}
	// Trim leading blank lines
	for len(lines) > 0 {
		if strings.TrimSpace(lines[0]) == "" {
			lines = lines[1:]
		} else {
			break
		}
	}
	return strings.Join(lines, "\n")
}
