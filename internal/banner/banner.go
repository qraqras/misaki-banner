// Package banner converts text into ASCII-art banner strings using bitmap font data.
package banner

import (
	"strings"

	mfont "github.com/qraqras/mskbanner/internal/font"
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
	// ShadowBox uses box-drawing characters (╗║╚═╝).
	ShadowBox ShadowMode = "box"
	// ShadowThin uses light-shade block (░░) for a soft shadow.
	ShadowThin ShadowMode = "thin"
)

// Options controls how the banner is rendered.
type Options struct {
	OnChar      string     // character(s) for filled pixels
	OffChar     string     // character(s) for empty pixels
	Shadow      ShadowMode // shadow rendering style
	ShadowColor string     // ANSI color for shadow (e.g. "240" for gray)
}

// DefaultOptions returns default rendering options.
func DefaultOptions() Options {
	return Options{
		OnChar:  DefaultOnChar,
		OffChar: DefaultOffChar,
	}
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
	type glyphInfo struct {
		bitmap [][]bool
		width  int
	}
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
	case ShadowBox:
		lines := boxShadowLines(grid, height, totalWidth, opts)
		return trimBlankLines(lines)
	case ShadowThin:
		lines := thinShadowLines(grid, height, totalWidth, opts)
		return trimBlankLines(lines)
	default:
		lines := gridLines(grid, height, totalWidth, opts)
		return trimBlankLines(lines)
	}
}

// gridLines renders a glyph grid as a string array, one per row.
func gridLines(grid [][]bool, height, width int, opts Options) []string {
	lines := make([]string, height)
	for y := 0; y < height; y++ {
		var sb strings.Builder
		for x := 0; x < width; x++ {
			if grid[y][x] {
				sb.WriteString(opts.OnChar)
			} else {
				sb.WriteString(opts.OffChar)
			}
		}
		lines[y] = sb.String()
	}
	return lines
}

// boxShadowLines renders a glyph grid with box-drawing shadow as a string array.
func boxShadowLines(grid [][]bool, h, w int, opts Options) []string {
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
				sb.WriteString(opts.OnChar)
				continue
			}

			L := isOn(y, x-1)   // left
			A := isOn(y-1, x)   // above
			D := isOn(y-1, x-1) // diagonal (above-left)

			switch {
			case L && A:
				sb.WriteString("╔═")
			case L:
				if D {
					sb.WriteString("║ ")
				} else {
					sb.WriteString("╗ ")
				}
			case A:
				if D {
					sb.WriteString("══")
				} else {
					sb.WriteString("╚═")
				}
			case D:
				sb.WriteString("╝ ")
			default:
				sb.WriteString(opts.OffChar)
			}
		}
		lines[y] = sb.String()
	}
	return lines
}

// thinShadowLines renders text with MEDIUM SHADE and shadow with FULL BLOCK (dimmed).
func thinShadowLines(grid [][]bool, h, w int, opts Options) []string {
	isOn := func(y, x int) bool {
		if y < 0 || y >= h || x < 0 || x >= w {
			return false
		}
		return grid[y][x]
	}

	textChar := "▒▒" // MEDIUM SHADE for text

	// Extended canvas: +1 row for bottom shadow, +1 col for right shadow
	outH := h + 1
	outW := w + 1

	lines := make([]string, outH)
	for y := 0; y < outH; y++ {
		var sb strings.Builder
		for x := 0; x < outW; x++ {
			if isOn(y, x) {
				sb.WriteString(textChar)
				continue
			}

			L := isOn(y, x-1)   // left
			A := isOn(y-1, x)   // above
			D := isOn(y-1, x-1) // diagonal (above-left)

			switch {
			case L && A:
				sb.WriteString("█▀") // L-shape: left + top
			case L:
				if D {
					sb.WriteString("█ ") // vertical shadow (full block)
				} else {
					sb.WriteString("▄ ") // vertical shadow (left half)
				}
			case A:
				if D {
					sb.WriteString("▀▀") // horizontal shadow (full block)
				} else {
					sb.WriteString(" ▀") // horizontal shadow (right half)
				}
			case D:
				sb.WriteString("▀ ") // corner only (upper-left quarter)
			default:
				sb.WriteString(opts.OffChar)
			}
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
