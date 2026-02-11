package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/qraqras/misaki-banner/internal/banner"
	mfont "github.com/qraqras/misaki-banner/internal/font"
)

func main() {
	shadow := flag.String("shadow", "", "shadow style: outline (box-drawing) or solid (shading)")
	fontName := flag.String("font", "misaki_gothic_2nd", "font name: misaki_gothic, misaki_gothic_2nd, or misaki_mincho")
	color := flag.String("color", "", "text color: preset (c,m,y) or hex (#RRGGBB/RRGGBB) or RGB (r,g,b)")
	gradient := flag.Bool("gradient", false, "enable gradient effect (from light to dark, top-left to bottom-right)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <text>\n\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	text := strings.Join(flag.Args(), " ")
	if text == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Replace literal \n with newline
	text = strings.ReplaceAll(text, `\n`, "\n")

	var font mfont.FontName
	switch *fontName {
	case "misaki_gothic":
		font = mfont.FontMisakiGothic
	case "misaki_gothic_2nd":
		font = mfont.FontMisakiGothic2nd
	case "misaki_mincho":
		font = mfont.FontMisakiMincho
	default:
		fmt.Fprintf(os.Stderr, "Unknown font: %s (use misaki_gothic, misaki_gothic_2nd, or misaki_mincho)\n", *fontName)
		os.Exit(1)
	}

	face, err := mfont.NewFace(font)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var shadowMode banner.ShadowMode
	switch *shadow {
	case "outline":
		shadowMode = banner.ShadowOutline
	case "solid":
		shadowMode = banner.ShadowSolid
	case "":
		shadowMode = banner.ShadowNone
	default:
		fmt.Fprintf(os.Stderr, "Unknown shadow mode: %s (use outline or solid)\n", *shadow)
		os.Exit(1)
	}

	opts := banner.Options{
		Shadow:   shadowMode,
		Color:    *color,
		Gradient: *gradient,
	}

	result := banner.Generate(face, text, opts)
	fmt.Println(result)
}
