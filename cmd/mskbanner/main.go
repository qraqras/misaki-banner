package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/qraqras/mskbanner/internal/banner"
	mfont "github.com/qraqras/mskbanner/internal/font"
)

func main() {
	onChar := flag.String("on", banner.DefaultOnChar, "character(s) for filled pixels")
	offChar := flag.String("off", banner.DefaultOffChar, "character(s) for empty pixels")
	shadow := flag.String("shadow", "", "shadow style: box (box-drawing) or thin (light-shade)")
	shadowColor := flag.String("shadow-color", "240", "ANSI 256-color code for shadow (e.g. 240=gray, 245=light-gray)")
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

	face, err := mfont.NewFace(mfont.FontMisaki)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var shadowMode banner.ShadowMode
	switch *shadow {
	case "box":
		shadowMode = banner.ShadowBox
	case "thin":
		shadowMode = banner.ShadowThin
	case "":
		shadowMode = banner.ShadowNone
	default:
		fmt.Fprintf(os.Stderr, "Unknown shadow mode: %s (use box or thin)\n", *shadow)
		os.Exit(1)
	}

	opts := banner.Options{
		OnChar:      *onChar,
		OffChar:     *offChar,
		Shadow:      shadowMode,
		ShadowColor: *shadowColor,
	}

	result := banner.Generate(face, text, opts)
	fmt.Println(result)
}
