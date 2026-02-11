// Package font handles loading and rasterizing bitmap fonts from TTF.
package font

import (
	"fmt"
	"image"
	"image/draw"

	"github.com/qraqras/misaki-banner/misaki"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// FontName represents available font choices.
type FontName string

const (
	FontMisakiGothic    FontName = "misaki_gothic"
	FontMisakiGothic2nd FontName = "misaki_gothic_2nd"
	FontMisakiMincho    FontName = "misaki_mincho"

	// misakiFontSize is the fixed pixel size for all Misaki fonts (8x8).
	misakiFontSize = 8
)

// fontDef describes a font's embedded data.
type fontDef struct {
	data []byte
}

var fonts = map[FontName]fontDef{
	FontMisakiGothic:    {data: misaki.GothicTTF},
	FontMisakiGothic2nd: {data: misaki.Gothic2ndTTF},
	FontMisakiMincho:    {data: misaki.MinchoTTF},
}

// Face holds a parsed font face ready for rendering.
type Face struct {
	face     font.Face
	fontSize int
}

// FontSize returns the pixel height of this font face.
func (f *Face) FontSize() int {
	return f.fontSize
}

// NewFace creates a new font face for the given font name.
func NewFace(name FontName) (*Face, error) {
	def, ok := fonts[name]
	if !ok {
		return nil, fmt.Errorf("unknown font: %s (available: misaki_gothic, misaki_gothic_2nd, misaki_mincho)", name)
	}

	ft, err := opentype.Parse(def.data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse font: %w", err)
	}

	face, err := opentype.NewFace(ft, &opentype.FaceOptions{
		Size:    misakiFontSize,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create font face: %w", err)
	}

	return &Face{face: face, fontSize: misakiFontSize}, nil
}

// RuneBitmap returns a bitmap (as [][]bool) for the given rune.
// Empty columns on the left/right are trimmed, then 1-cell padding is added
// on each side for consistent spacing.
// true means the pixel is "on".
func (f *Face) RuneBitmap(r rune) [][]bool {
	adv := f.Advance(r)
	metrics := f.face.Metrics()
	ascent := metrics.Ascent.Ceil()

	// Create image sized to the glyph advance x font height (8x8 fixed)
	w := adv
	if w < misakiFontSize {
		w = misakiFontSize
	}
	img := image.NewGray(image.Rect(0, 0, w, misakiFontSize))
	draw.Draw(img, img.Bounds(), image.White, image.Point{}, draw.Src)

	d := &font.Drawer{
		Dst:  img,
		Src:  image.Black,
		Face: f.face,
		Dot:  fixed.P(0, ascent),
	}
	d.DrawString(string(r))

	// Build raw bitmap
	raw := make([][]bool, misakiFontSize)
	for y := 0; y < misakiFontSize; y++ {
		raw[y] = make([]bool, adv)
		for x := 0; x < adv; x++ {
			raw[y][x] = img.GrayAt(x, y).Y < 128
		}
	}

	// Find leftmost and rightmost non-empty columns
	minX, maxX := adv, -1
	for y := 0; y < misakiFontSize; y++ {
		for x := 0; x < adv; x++ {
			if raw[y][x] {
				if x < minX {
					minX = x
				}
				if x > maxX {
					maxX = x
				}
			}
		}
	}

	// If glyph is entirely blank, return a 1-cell blank column
	if maxX < 0 {
		blank := make([][]bool, misakiFontSize)
		for y := 0; y < misakiFontSize; y++ {
			blank[y] = make([]bool, 1)
		}
		return blank
	}

	// Trim to [minX..maxX] then add 1-cell padding on left only
	// Adjacent glyphs each contribute 1 left pad → 2 spaces between chars
	trimW := maxX - minX + 1
	padW := trimW + 1 // +1 left only
	bitmap := make([][]bool, misakiFontSize)
	for y := 0; y < misakiFontSize; y++ {
		bitmap[y] = make([]bool, padW)
		for x := 0; x < trimW; x++ {
			bitmap[y][x+1] = raw[y][minX+x]
		}
	}
	return bitmap
}

// Advance returns the horizontal advance width for the given rune in pixels.
func (f *Face) Advance(r rune) int {
	adv, ok := f.face.GlyphAdvance(r)
	if !ok {
		return misakiFontSize
	}
	return adv.Ceil()
}
