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

	r = rb<<24 | rb<<16 | rb<<8 | rb<<0
	g = gb<<24 | gb<<16 | gb<<8 | gb<<0
	b = bb<<24 | bb<<16 | bb<<8 | bb<<0
	a = 0xFFFFFFFF
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

	rgb24Color{0X555555},
	rgb24Color{0X5555FF},
	rgb24Color{0X55FF55},
	rgb24Color{0X55FFFF},
	rgb24Color{0XFF5555},
	rgb24Color{0XFF55FF},
	rgb24Color{0XFFFF55},
	rgb24Color{0XFFFFFF},
}
