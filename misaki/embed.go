// Package misaki provides the embedded font data.
package misaki

import _ "embed"

//go:embed misaki_gothic.ttf
var GothicTTF []byte

//go:embed misaki_gothic_2nd.ttf
var Gothic2ndTTF []byte

//go:embed misaki_mincho.ttf
var MinchoTTF []byte
