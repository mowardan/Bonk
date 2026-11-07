package assets

import "embed"

//go:embed sounds/pain/*.mp3
var PainFS embed.FS

//go:embed sounds/sexy/*.mp3
var SexyFS embed.FS

//go:embed sounds/halo/*.mp3
var HaloFS embed.FS

//go:embed sounds/lizard/*.mp3
var LizardFS embed.FS
