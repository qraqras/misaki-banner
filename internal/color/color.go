package color

import (
	"fmt"
	"math"
	"strconv"
)

// RGB represents a 24-bit color.
type RGB struct {
	R, G, B uint8
}

// ANSI returns the ANSI 24-bit foreground escape sequence for this color.
func (c RGB) ANSI() string {
	return fmt.Sprintf("\033[38;2;%d;%d;%dm", c.R, c.G, c.B)
}

// Reset is the ANSI reset escape sequence.
const Reset = "\033[0m"

// hsl represents a color in HSL space.
type hsl struct {
	H, S, L float64 // H in [0,360), S and L in [0,1]
}

// rgbToHSL converts an RGB color to HSL.
func rgbToHSL(c RGB) hsl {
	r := float64(c.R) / 255.0
	g := float64(c.G) / 255.0
	b := float64(c.B) / 255.0

	max := math.Max(r, math.Max(g, b))
	min := math.Min(r, math.Min(g, b))
	l := (max + min) / 2.0

	if max == min {
		return hsl{0, 0, l}
	}

	d := max - min
	s := d / (1.0 - math.Abs(2.0*l-1.0))

	var h float64
	switch max {
	case r:
		h = math.Mod((g-b)/d+6, 6)
	case g:
		h = (b-r)/d + 2
	case b:
		h = (r-g)/d + 4
	}
	h *= 60

	return hsl{h, s, l}
}

// hslToRGB converts an HSL color to RGB.
func hslToRGB(c hsl) RGB {
	if c.S == 0 {
		v := uint8(math.Round(c.L * 255))
		return RGB{v, v, v}
	}

	C := (1 - math.Abs(2*c.L-1)) * c.S
	X := C * (1 - math.Abs(math.Mod(c.H/60, 2)-1))
	m := c.L - C/2

	var r, g, b float64
	switch {
	case c.H < 60:
		r, g, b = C, X, 0
	case c.H < 120:
		r, g, b = X, C, 0
	case c.H < 180:
		r, g, b = 0, C, X
	case c.H < 240:
		r, g, b = 0, X, C
	case c.H < 300:
		r, g, b = X, 0, C
	default:
		r, g, b = C, 0, X
	}

	return RGB{
		R: uint8(math.Round((r + m) * 255)),
		G: uint8(math.Round((g + m) * 255)),
		B: uint8(math.Round((b + m) * 255)),
	}
}

// ColorPresets is a map of named color presets.
var ColorPresets = map[string]RGB{
	"c": {0, 255, 255},
	"m": {255, 0, 255},
	"y": {255, 255, 0},
}

// ParseColor parses a color string in RGB format "r,g,b", hex format "#RRGGBB"/"RRGGBB", or a preset name.
func ParseColor(s string) (RGB, error) {
	// Strip leading '#' if present
	if len(s) > 0 && s[0] == '#' {
		s = s[1:]
	}

	// Check if it's a preset
	if c, ok := ColorPresets[s]; ok {
		return c, nil
	}

	// Try to parse as hex color (e.g., "ffffff")
	if len(s) == 6 {
		val, err := strconv.ParseUint(s, 16, 32)
		if err == nil {
			return RGB{
				R: uint8((val & 0xFF0000) >> 16),
				G: uint8((val & 0x00FF00) >> 8),
				B: uint8((val & 0x0000FF) >> 0),
			}, nil
		}
	}

	// Try to parse as "r,g,b"
	var r, g, b int
	n, err := fmt.Sscanf(s, "%d,%d,%d", &r, &g, &b)
	if err != nil || n != 3 {
		return RGB{}, fmt.Errorf("invalid color format: %s (use 'RRGGBB', 'r,g,b' or preset name)", s)
	}

	if r < 0 || r > 255 || g < 0 || g > 255 || b < 0 || b > 255 {
		return RGB{}, fmt.Errorf("RGB values must be 0-255")
	}

	return RGB{uint8(r), uint8(g), uint8(b)}, nil
}

// ShiftColor shifts a color in HSL space by adjusting hue and lightness.
// hueDelta is in degrees (-360 to 360), lightnessDelta is a factor (-1.0 to 1.0).
func ShiftColor(c RGB, hueDelta float64, lightnessDelta float64) RGB {
	hslColor := rgbToHSL(c)

	// Adjust hue
	hslColor.H += hueDelta
	if hslColor.H < 0 {
		hslColor.H += 360
	} else if hslColor.H >= 360 {
		hslColor.H -= 360
	}

	// Adjust lightness
	if lightnessDelta > 0 {
		// Brighten
		hslColor.L += (1.0 - hslColor.L) * lightnessDelta
	} else {
		// Darken
		hslColor.L += hslColor.L * lightnessDelta
	}

	if hslColor.L < 0 {
		hslColor.L = 0
	} else if hslColor.L > 1.0 {
		hslColor.L = 1.0
	}

	return hslToRGB(hslColor)
}
