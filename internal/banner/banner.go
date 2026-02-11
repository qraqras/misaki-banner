// Package banner converts text into ASCII-art banner strings using bitmap font data.
package banner

import (
	"strings"

	mcolor "github.com/qraqras/misaki-banner/internal/color"
	mfont "github.com/qraqras/misaki-banner/internal/font"
)

const (
	// DefaultOnChar is the default character for "on" pixels.
	DefaultOnChar = "██"
	// DefaultOffChar is the default character for "off" pixels.
	DefaultOffChar = "  "
)

// ShadowMode selects the shadow rendering style.
type ShadowMode string

const (
	// ShadowNone disables shadow.
	ShadowNone ShadowMode = ""
	// ShadowOutline uses box-drawing characters (╗║╚═╝).
	ShadowOutline ShadowMode = "outline"
	// ShadowSolid uses shading blocks (░░) for a solid shadow.
	ShadowSolid ShadowMode = "solid"
)

// Options controls how the banner is rendered.
type Options struct {
	OnChar   string     // character(s) for filled pixels
	OffChar  string     // character(s) for empty pixels
	Shadow   ShadowMode // shadow rendering style
	Color    string     // text color (RGB format "r,g,b" or preset name)
	Gradient bool       // enable gradient effect (light to dark)
}

// DefaultOptions returns default rendering options.
func DefaultOptions() Options {
	return Options{
		OnChar:  DefaultOnChar,
		OffChar: DefaultOffChar,
	}
}

// glyphInfo holds bitmap and width information for a single glyph.
type glyphInfo struct {
	bitmap [][]bool
	width  int
}

// Generate creates an ASCII-art banner string from the given text.
// If text contains newlines, each line is rendered separately and joined.
func Generate(face *mfont.Face, text string, opts Options) string {
	if opts.OnChar == "" {
		opts.OnChar = DefaultOnChar
	}
	if opts.OffChar == "" {
		opts.OffChar = DefaultOffChar
	}

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

	// Calculate total width and build charMap (x -> character index)
	totalWidth := 0
	for _, g := range glyphs {
		totalWidth += g.width
	}

	// Build a combined 2D bool grid and track which character each pixel belongs to
	grid := make([][]bool, height)
	charMap := make([][]int, height) // maps (y, x) to character index
	for y := 0; y < height; y++ {
		grid[y] = make([]bool, totalWidth)
		charMap[y] = make([]int, totalWidth)
		xOff := 0
		for ci, g := range glyphs {
			for x := 0; x < g.width; x++ {
				charMap[y][xOff+x] = ci
				if y < len(g.bitmap) && x < len(g.bitmap[y]) && g.bitmap[y][x] {
					grid[y][xOff+x] = true
				}
			}
			xOff += g.width
		}
	}

	// Build per-character gradient data if needed
	var gradientData [][]mcolor.RGB // [charIndex][y] -> color
	if opts.Gradient || opts.Color != "" {
		gradientData = make([][]mcolor.RGB, len(glyphs))
		for ci, g := range glyphs {
			if g.width > 0 && height > 0 {
				gradientData[ci] = make([]mcolor.RGB, height)
			}
		}
	}

	switch opts.Shadow {
	case ShadowOutline:
		lines := outlineShadowLines(grid, charMap, glyphs, height, totalWidth, opts, gradientData)
		return trimBlankLines(lines)
	case ShadowSolid:
		lines := solidShadowLines(grid, charMap, glyphs, height, totalWidth, opts, gradientData)
		return trimBlankLines(lines)
	default:
		lines := gridLines(grid, charMap, glyphs, height, totalWidth, opts, gradientData)
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

// gridLines renders a glyph grid as a string array, one per row.
func gridLines(grid [][]bool, charMap [][]int, glyphs []glyphInfo, height, width int, opts Options, gradientData [][]mcolor.RGB) []string {
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
				sb.WriteString(colorPixel(opts.OnChar, y, x, width, opts, baseColor, hasColor))
			} else {
				sb.WriteString(opts.OffChar)
			}
		}
		lines[y] = sb.String()
	}
	return lines
}

// outlineShadowLines renders a glyph grid with box-drawing shadow as a string array.
func outlineShadowLines(grid [][]bool, charMap [][]int, glyphs []glyphInfo, h, w int, opts Options, gradientData [][]mcolor.RGB) []string {
	isOn := func(y, x int) bool {
		if y < 0 || y >= h || x < 0 || x >= w {
			return false
		}
		return grid[y][x]
	}

	// Extended canvas: +1 row for bottom shadow, +1 col for last char's right shadow
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
				sb.WriteString(colorPixel(opts.OnChar, y, x, w, opts, baseColor, hasColor))
				continue
			}

			L := isOn(y, x-1)   // left
			A := isOn(y-1, x)   // above
			D := isOn(y-1, x-1) // diagonal (above-left)

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
			case L && A:
				shadowStr = "╔═"
				if L {
					shadowStr = colorPixel(shadowStr, y, x-1, w, opts, baseColor, hasColor)
				}
			case L:
				if D {
					shadowStr = "║ "
				} else {
					shadowStr = "╗ "
				}
				shadowStr = colorPixel(shadowStr, y, x-1, w, opts, baseColor, hasColor)
			case A:
				if D {
					shadowStr = "══"
				} else {
					shadowStr = "╚═"
				}
				shadowStr = colorPixel(shadowStr, y-1, x, w, opts, baseColor, hasColor)
			case D:
				shadowStr = "╝ "
				shadowStr = colorPixel(shadowStr, y-1, x-1, w, opts, baseColor, hasColor)
			default:
				shadowStr = opts.OffChar
			}
			sb.WriteString(shadowStr)
		}
		lines[y] = sb.String()
	}
	return lines
}

// solidShadowLines renders text with MEDIUM SHADE and shadow with FULL BLOCK (dimmed).
func solidShadowLines(grid [][]bool, charMap [][]int, glyphs []glyphInfo, h, w int, opts Options, gradientData [][]mcolor.RGB) []string {
	isOn := func(y, x int) bool {
		if y < 0 || y >= h || x < 0 || x >= w {
			return false
		}
		return grid[y][x]
	}

	textChar := "░░" // LIGHT SHADE for text

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
				sb.WriteString(colorPixel(textChar, y, x, w, opts, baseColor, hasColor))
				continue
			}

			L := isOn(y, x-1)   // left
			A := isOn(y-1, x)   // above
			D := isOn(y-1, x-1) // diagonal (above-left)

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
			case L && A:
				shadowStr = "█▀" // L-shape: left + top
				if L {
					shadowStr = colorPixel(shadowStr, y, x-1, w, opts, baseColor, hasColor)
				}
			case L:
				if D {
					shadowStr = "█ " // vertical shadow (full block)
				} else {
					shadowStr = "▄ " // vertical shadow (left half)
				}
				shadowStr = colorPixel(shadowStr, y, x-1, w, opts, baseColor, hasColor)
			case A:
				if D {
					shadowStr = "▀▀" // horizontal shadow (full block)
				} else {
					shadowStr = " ▀" // horizontal shadow (right half)
				}
				shadowStr = colorPixel(shadowStr, y-1, x, w, opts, baseColor, hasColor)
			case D:
				shadowStr = "▀ " // corner only (upper-left quarter)
				shadowStr = colorPixel(shadowStr, y-1, x-1, w, opts, baseColor, hasColor)
			default:
				shadowStr = opts.OffChar
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
