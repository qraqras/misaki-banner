package banner

import (
	"strings"
	"testing"

	mfont "github.com/qraqras/misaki-banner/internal/font"
)

func newTestFace(t *testing.T) *mfont.Face {
	t.Helper()
	face, err := mfont.NewFace(mfont.FontMisakiGothic2nd)
	if err != nil {
		t.Fatalf("NewFace failed: %v", err)
	}
	return face
}

func TestGenerate_Empty(t *testing.T) {
	face := newTestFace(t)
	result := Generate(face, "", Options{})
	if result != "" {
		t.Errorf("Generate(\"\") = %q, want empty", result)
	}
}

func TestGenerate_BasicOutput(t *testing.T) {
	face := newTestFace(t)
	result := Generate(face, "A", Options{})
	if result == "" {
		t.Fatal("Generate(\"A\") returned empty")
	}
	// Should contain block characters
	if !strings.Contains(result, "██") {
		t.Error("Generate(\"A\") output does not contain ██")
	}
}

func TestGenerate_Japanese(t *testing.T) {
	face := newTestFace(t)
	result := Generate(face, "あ", Options{})
	if result == "" {
		t.Fatal("Generate(\"あ\") returned empty")
	}
	if !strings.Contains(result, "██") {
		t.Error("Generate(\"あ\") output does not contain ██")
	}
}

func TestGenerate_MultiLine(t *testing.T) {
	face := newTestFace(t)
	result := Generate(face, "A\nB", Options{})
	// Multi-line should produce two banner blocks separated by blank line
	parts := strings.Split(result, "\n\n")
	if len(parts) < 2 {
		t.Errorf("Generate(\"A\\nB\") expected 2 blocks separated by blank line, got %d", len(parts))
	}
}

func TestGenerate_ShadowOutline(t *testing.T) {
	face := newTestFace(t)
	result := Generate(face, "A", Options{Shadow: ShadowOutline})
	if result == "" {
		t.Fatal("Generate with ShadowOutline returned empty")
	}
	// Outline mode uses box-drawing characters
	if !strings.Contains(result, "██") {
		t.Error("ShadowOutline output does not contain ██")
	}
}

func TestGenerate_ShadowSolid(t *testing.T) {
	face := newTestFace(t)
	result := Generate(face, "A", Options{Shadow: ShadowSolid})
	if result == "" {
		t.Fatal("Generate with ShadowSolid returned empty")
	}
	// Solid mode uses ░░ for text
	if !strings.Contains(result, "░░") {
		t.Error("ShadowSolid output does not contain ░░")
	}
}

func TestGenerate_WithColor(t *testing.T) {
	face := newTestFace(t)
	result := Generate(face, "A", Options{Color: "c"})
	// Should contain ANSI escape sequence
	if !strings.Contains(result, "\033[38;2;") {
		t.Error("Generate with color does not contain ANSI escape sequence")
	}
	// Should contain reset sequence
	if !strings.Contains(result, "\033[0m") {
		t.Error("Generate with color does not contain ANSI reset sequence")
	}
}

func TestGenerate_WithGradient(t *testing.T) {
	face := newTestFace(t)
	result := Generate(face, "ABC", Options{Color: "c", Gradient: true})
	if !strings.Contains(result, "\033[38;2;") {
		t.Error("Generate with gradient does not contain ANSI escape sequence")
	}
}

func TestGenerate_InvalidColor_NoError(t *testing.T) {
	face := newTestFace(t)
	// Invalid color should not panic, just render without color
	result := Generate(face, "A", Options{Color: "invalid_color"})
	if result == "" {
		t.Fatal("Generate with invalid color returned empty")
	}
	// Should NOT contain ANSI escape since color is invalid
	if strings.Contains(result, "\033[38;2;") {
		t.Error("Generate with invalid color should not contain ANSI escape")
	}
}

func TestTrimBlankLines(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  string
	}{
		{"no blanks", []string{"a", "b"}, "a\nb"},
		{"leading blank", []string{"", "a", "b"}, "a\nb"},
		{"trailing blank", []string{"a", "b", ""}, "a\nb"},
		{"both blanks", []string{"", "a", "b", ""}, "a\nb"},
		{"whitespace only", []string{"  ", "a", "  "}, "a"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := trimBlankLines(tt.input)
			if got != tt.want {
				t.Errorf("trimBlankLines(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
