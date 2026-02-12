package font

import (
	"testing"
)

func TestNewFace_ValidFonts(t *testing.T) {
	fonts := []FontName{FontMisakiGothic, FontMisakiGothic2nd, FontMisakiMincho}
	for _, name := range fonts {
		t.Run(string(name), func(t *testing.T) {
			face, err := NewFace(name)
			if err != nil {
				t.Fatalf("NewFace(%q) returned error: %v", name, err)
			}
			if face.FontSize() != misakiFontSize {
				t.Errorf("FontSize() = %d, want %d", face.FontSize(), misakiFontSize)
			}
		})
	}
}

func TestNewFace_InvalidFont(t *testing.T) {
	_, err := NewFace("nonexistent_font")
	if err == nil {
		t.Error("NewFace(\"nonexistent_font\") expected error, got nil")
	}
}

func TestRuneBitmap_ASCII(t *testing.T) {
	face, err := NewFace(FontMisakiGothic2nd)
	if err != nil {
		t.Fatalf("NewFace failed: %v", err)
	}

	bm := face.RuneBitmap('A')
	if len(bm) != misakiFontSize {
		t.Fatalf("RuneBitmap('A') height = %d, want %d", len(bm), misakiFontSize)
	}

	// 'A' should have some pixels set
	hasPixel := false
	for _, row := range bm {
		for _, v := range row {
			if v {
				hasPixel = true
				break
			}
		}
	}
	if !hasPixel {
		t.Error("RuneBitmap('A') returned blank bitmap")
	}
}

func TestRuneBitmap_Japanese(t *testing.T) {
	face, err := NewFace(FontMisakiGothic2nd)
	if err != nil {
		t.Fatalf("NewFace failed: %v", err)
	}

	bm := face.RuneBitmap('あ')
	if len(bm) != misakiFontSize {
		t.Fatalf("RuneBitmap('あ') height = %d, want %d", len(bm), misakiFontSize)
	}

	// 'あ' should have some pixels set
	hasPixel := false
	for _, row := range bm {
		for _, v := range row {
			if v {
				hasPixel = true
				break
			}
		}
	}
	if !hasPixel {
		t.Error("RuneBitmap('あ') returned blank bitmap")
	}
}

func TestRuneBitmap_Space(t *testing.T) {
	face, err := NewFace(FontMisakiGothic2nd)
	if err != nil {
		t.Fatalf("NewFace failed: %v", err)
	}

	bm := face.RuneBitmap(' ')
	if len(bm) != misakiFontSize {
		t.Fatalf("RuneBitmap(' ') height = %d, want %d", len(bm), misakiFontSize)
	}

	// Space should be blank
	for _, row := range bm {
		for _, v := range row {
			if v {
				t.Error("RuneBitmap(' ') has pixels set, expected blank")
				return
			}
		}
	}
}

func TestAdvance(t *testing.T) {
	face, err := NewFace(FontMisakiGothic2nd)
	if err != nil {
		t.Fatalf("NewFace failed: %v", err)
	}

	// Half-width ASCII should have advance <= 8
	advA := face.Advance('A')
	if advA <= 0 || advA > misakiFontSize {
		t.Errorf("Advance('A') = %d, expected 1-%d", advA, misakiFontSize)
	}

	// Full-width Japanese should have advance == 8
	advKana := face.Advance('あ')
	if advKana != misakiFontSize {
		t.Errorf("Advance('あ') = %d, want %d", advKana, misakiFontSize)
	}
}
