package color

import (
	"math"
	"testing"
)

func TestParseColor_Presets(t *testing.T) {
	tests := []struct {
		input string
		want  RGB
	}{
		{"c", RGB{0, 255, 255}},
		{"m", RGB{255, 0, 255}},
		{"y", RGB{255, 255, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseColor(tt.input)
			if err != nil {
				t.Fatalf("ParseColor(%q) returned error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParseColor(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseColor_Hex(t *testing.T) {
	tests := []struct {
		input string
		want  RGB
	}{
		{"ff0000", RGB{255, 0, 0}},
		{"00ff00", RGB{0, 255, 0}},
		{"0000ff", RGB{0, 0, 255}},
		{"ffffff", RGB{255, 255, 255}},
		{"000000", RGB{0, 0, 0}},
		{"ff4444", RGB{255, 68, 68}},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseColor(tt.input)
			if err != nil {
				t.Fatalf("ParseColor(%q) returned error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParseColor(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseColor_HexWithHash(t *testing.T) {
	tests := []struct {
		input string
		want  RGB
	}{
		{"#ff0000", RGB{255, 0, 0}},
		{"#00ff00", RGB{0, 255, 0}},
		{"#0000ff", RGB{0, 0, 255}},
		{"#ffffff", RGB{255, 255, 255}},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseColor(tt.input)
			if err != nil {
				t.Fatalf("ParseColor(%q) returned error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParseColor(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseColor_RGB(t *testing.T) {
	tests := []struct {
		input string
		want  RGB
	}{
		{"255,0,0", RGB{255, 0, 0}},
		{"0,255,0", RGB{0, 255, 0}},
		{"0,0,255", RGB{0, 0, 255}},
		{"128,128,128", RGB{128, 128, 128}},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseColor(tt.input)
			if err != nil {
				t.Fatalf("ParseColor(%q) returned error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParseColor(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseColor_Invalid(t *testing.T) {
	invalids := []string{
		"",
		"not_a_color",
		"256,0,0",
		"-1,0,0",
		"zzzzzz",
		"#gggggg",
	}
	for _, input := range invalids {
		t.Run(input, func(t *testing.T) {
			_, err := ParseColor(input)
			if err == nil {
				t.Errorf("ParseColor(%q) expected error, got nil", input)
			}
		})
	}
}

func TestShiftColor_Roundtrip(t *testing.T) {
	// Shifting by 0 should return the same color (within rounding)
	colors := []RGB{
		{255, 0, 0},
		{0, 255, 0},
		{0, 0, 255},
		{128, 64, 32},
	}
	for _, c := range colors {
		got := ShiftColor(c, 0, 0)
		if absDiff(got.R, c.R) > 1 || absDiff(got.G, c.G) > 1 || absDiff(got.B, c.B) > 1 {
			t.Errorf("ShiftColor(%v, 0, 0) = %v, want ~%v", c, got, c)
		}
	}
}

func TestShiftColor_HueShift(t *testing.T) {
	// Shifting hue by 360 should return the same color
	c := RGB{255, 0, 0}
	got := ShiftColor(c, 360, 0)
	if absDiff(got.R, c.R) > 1 || absDiff(got.G, c.G) > 1 || absDiff(got.B, c.B) > 1 {
		t.Errorf("ShiftColor(%v, 360, 0) = %v, want ~%v", c, got, c)
	}
}

func TestRGB_ANSI(t *testing.T) {
	c := RGB{255, 128, 0}
	want := "\033[38;2;255;128;0m"
	if got := c.ANSI(); got != want {
		t.Errorf("RGB{255,128,0}.ANSI() = %q, want %q", got, want)
	}
}

func absDiff(a, b uint8) uint8 {
	d := int(a) - int(b)
	return uint8(math.Abs(float64(d)))
}
