// Package misaki provides the embedded font data.
package misaki

import _ "embed"

//go:embed misaki_gothic.ttf
var GothicTTF []byte
