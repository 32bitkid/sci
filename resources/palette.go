package resources

import (
	"image/color"
)

type rgb24Color struct {
	RGB uint32
}

func (ega rgb24Color) RGBA() (r, g, b, a uint32) {
	rb := uint32(ega.RGB >> 16 & 0xFF)
	gb := uint32(ega.RGB >> 8 & 0xFF)
	bb := uint32(ega.RGB & 0xFF)

	r = rb<<8 | rb
	g = gb<<8 | gb
	b = bb<<8 | bb
	a = 0xFFFFFF
	return
}

var egaPalette = []color.Color{
	rgb24Color{0x000000},
	rgb24Color{0x0000AA},
	rgb24Color{0x00AA00},
	rgb24Color{0x00AAAA},
	rgb24Color{0xAA0000},
	rgb24Color{0xAA00AA},
	rgb24Color{0xAA5500},
	rgb24Color{0xAAAAAA},

	rgb24Color{0x555555},
	rgb24Color{0x5555FF},
	rgb24Color{0x55FF55},
	rgb24Color{0x55FFFF},
	rgb24Color{0xFF5555},
	rgb24Color{0xFF55FF},
	rgb24Color{0xFFFF55},
	rgb24Color{0xFFFFFF},
}

var gray16Palette = []color.Color{
	color.Gray{0x00},
	color.Gray{0x11},
	color.Gray{0x22},
	color.Gray{0x33},
	color.Gray{0x44},
	color.Gray{0x55},
	color.Gray{0x66},
	color.Gray{0x77},

	color.Gray{0x88},
	color.Gray{0x99},
	color.Gray{0xAA},
	color.Gray{0xBB},
	color.Gray{0xCC},
	color.Gray{0xDD},
	color.Gray{0xEE},
	color.Gray{0xFF},
}
