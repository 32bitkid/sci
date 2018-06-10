package resources

import (
	"image/color"
)

type egaColor struct {
	Color uint16
}

func (ega egaColor) RGBA() (r, g, b, a uint32) {
	rn := uint32(ega.Color >> 8 & 0xf)
	gn := uint32(ega.Color >> 4 & 0xf)
	bn := uint32(ega.Color & 0xf)
	r = rn<<28 | rn<<24 | rn<<20 | rn<<16 | rn<<12 | rn<<8 | rn<<4 | rn<<0
	g = gn<<28 | gn<<24 | gn<<20 | gn<<16 | gn<<12 | gn<<8 | gn<<4 | gn<<0
	b = bn<<28 | bn<<24 | bn<<20 | bn<<16 | bn<<12 | bn<<8 | bn<<4 | rn<<0
	a = 0xFFFFFFFF
	return
}

var egaPalette = []color.Color{
	egaColor{0x000},
	egaColor{0x00A},
	egaColor{0x0A0},
	egaColor{0x0AA},
	egaColor{0xA00},
	egaColor{0xA0A},
	egaColor{0xA50},
	egaColor{0xAAA},

	egaColor{0x555},
	egaColor{0x55F},
	egaColor{0x5F5},
	egaColor{0x5FF},
	egaColor{0xF55},
	egaColor{0xF5F},
	egaColor{0xFF5},
	egaColor{0xFFF},
}