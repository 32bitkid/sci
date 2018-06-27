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
	bb := uint32(ega.RGB >> 0 & 0xFF)

	r = rb<<8 | rb
	g = gb<<8 | gb
	b = bb<<8 | bb
	a = 0xFFFF
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

var db16Palette = []color.Color{
	rgb24Color{0x000000},
	rgb24Color{0x3f3f74},
	rgb24Color{0x4b692f},
	rgb24Color{0x306082},
	rgb24Color{0xac3232},
	rgb24Color{0x45283c},
	rgb24Color{0x8f563b},
	rgb24Color{0x847e87},

	rgb24Color{0x323c39},
	rgb24Color{0x639bff},
	rgb24Color{0x6abe30},
	rgb24Color{0x5fcde4},
	rgb24Color{0xd95763},
	rgb24Color{0xd77bba},
	rgb24Color{0xfbf236},
	rgb24Color{0xffffff},
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
