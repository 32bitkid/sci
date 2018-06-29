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

func lerpui32(v0, v1 uint32, t float64) uint32 {
	return uint32(float64(v0)*(1-t) + float64(v1)*t)
}

func rgbmix(c1, c2 color.Color, t float64) color.Color {
	r1, g1, b1, _ := c1.RGBA()
	r2, g2, b2, _ := c2.RGBA()
	return color.RGBA{
		R: uint8(lerpui32(r1, r2, t) >> 8),
		G: uint8(lerpui32(g1, g2, t) >> 8),
		B: uint8(lerpui32(b1, b2, t) >> 8),
		A: 0xff,
	}
}

var extendedPalette = []color.Color{
	0x00: rgb24Color{0x000000}, // black
	0x01: rgb24Color{0x3f3f74}, // deep-koamaru
	0x02: rgb24Color{0x4b692f}, // dell
	0x03: rgb24Color{0x306082}, // venice-blue
	0x04: rgb24Color{0xac3232}, // brown
	0x05: rgb24Color{0x76428a}, // clairvoyant
	0x06: rgb24Color{0x8f563b}, // rope
	0x07: rgb24Color{0x847e87}, // topaz

	0x08: rgb24Color{0x323c39}, // opal
	0x09: rgb24Color{0x639bff}, // cornflower
	0x0a: rgb24Color{0x6abe30}, // christi
	0x0b: rgb24Color{0x5fcde4}, // viking
	0x0c: rgb24Color{0xd95763}, // mandy
	0x0d: rgb24Color{0xd77bba}, // plum
	0x0e: rgb24Color{0xfbf236}, // golden-fizz
	0x0f: rgb24Color{0xffffff}, // white

	0x10: rgb24Color{0x222034}, // valhalla
	0x11: rgb24Color{0x524b24}, // verdigris
	0x12: rgb24Color{0x183041}, // ???
	0x13: rgb24Color{0x45283c}, // loulou
	0x14: rgb24Color{0x663931}, // oiled-cedar
	0x15: rgb24Color{0x595652}, // smokey-ash
	0x16: rgb24Color{0xd9a066}, // twine
	0x17: rgb24Color{0x9badb7}, // heather

	0x18: rgb24Color{0x5b6ee1}, // royal blue
	0x19: rgb24Color{0xcbdbfc}, // light-steel-blue
	0x1a: rgb24Color{0x8f974a}, // rainforest
	0x1b: rgb24Color{0x99e550}, // atlantis
	0x1c: rgb24Color{0xdf7126}, // tahiti-gold
	0x1d: rgb24Color{0x8a6f30}, // stinger
	0x1e: rgb24Color{0xeec39a}, // pancho
	0x1f: rgb24Color{0x696a6a}, //  dim-gray

	0x20: rgbmix(db16Palette[0xc], db16Palette[0xe], 1.0/3.0),
	0x21: rgbmix(db16Palette[0xc], db16Palette[0xe], 2.0/3.0),
	0x22: rgbmix(db16Palette[0xe], db16Palette[0xf], 0.5),
	0x23: rgbmix(db16Palette[0x0], db16Palette[0x8], 1.0/3.0),
	0x24: rgbmix(db16Palette[0x0], db16Palette[0x8], 2.0/3.0),
	0x25: rgbmix(db16Palette[0x4], db16Palette[0x8], 2.0/4.0),
	0x26: rgbmix(db16Palette[0x4], db16Palette[0x8], 2.0/3.0),
	0x27: rgbmix(db16Palette[0xc], db16Palette[0x8], 1.0/3.0),
	0x28: rgbmix(db16Palette[0xc], db16Palette[0x8], 2.0/3.0),
	0x29: rgbmix(db16Palette[0x6], db16Palette[0xc], 1.0/3.0),
	0x2a: rgbmix(db16Palette[0x6], db16Palette[0xc], 2.0/3.0),
	0x2b: rgbmix(rgb24Color{0x9badb7}, db16Palette[0xf], 1.5/3.0),
	0x2c: rgbmix(db16Palette[0x8], db16Palette[0x9], 1.0/3.0),
	0x2d: rgbmix(db16Palette[0x8], db16Palette[0x9], 2.0/3.0),
	0x2e: rgbmix(db16Palette[0x0], db16Palette[0x4], 1.0/3.0),
	0x2f: rgbmix(db16Palette[0x0], db16Palette[0x4], 2.0/3.0),

	0x30: rgbmix(db16Palette[0x4], db16Palette[0xc], 1.0/3.0),
	0x31: rgbmix(db16Palette[0x4], db16Palette[0xc], 2.0/3.0),
	0x32: rgb24Color{0xcd69ca},
	0x33: rgbmix(db16Palette[0xb], db16Palette[0xf], 1.0/3.0),
	0x34: rgbmix(db16Palette[0xb], db16Palette[0xf], 2.0/3.0),
	0x35: rgbmix(db16Palette[0x3], db16Palette[0xb], 1.25/3.0),
	0x36: rgbmix(db16Palette[0x3], db16Palette[0xb], 1.75/3.0),
	0x37: rgbmix(db16Palette[0x2], db16Palette[0xb], 1.0/3.0),
	0x38: rgbmix(db16Palette[0x2], db16Palette[0x8], 1.5/3.0),
	0x39: rgbmix(db16Palette[0xd], db16Palette[0xf], 1.5/3.0),
	0x3a: rgbmix(rgb24Color{0xeec39a}, db16Palette[0xe], 1.5/3.0),
	0x3b: rgbmix(db16Palette[0x8], rgb24Color{0x524b24}, 1.5/3.0),
	0x3c: rgbmix(db16Palette[0x0], db16Palette[0xa], 1.0/3.0),
	0x3d: color.Black,
	0x3e: rgbmix(db16Palette[0x8], db16Palette[0xa], 1.25/3.0),
	0x3f: rgbmix(db16Palette[0x8], db16Palette[0xa], 1.75/3.0),

	0x40: rgbmix(db16Palette[0x8], db16Palette[0xe], 1.5/3.0),
	0x41: rgbmix(db16Palette[0x8], db16Palette[0xe], 2.5/3.0),
	0x42: rgbmix(db16Palette[0x5], db16Palette[0xb], 1.0/3.0),
	0x43: rgbmix(db16Palette[0x5], db16Palette[0xb], 2.0/3.0),
	0x44: rgbmix(db16Palette[0xb], db16Palette[0xd], 0.75/3.0),
	0x45: rgbmix(db16Palette[0xb], db16Palette[0xd], 1.75/3.0),
	0x46: rgbmix(db16Palette[0x4], db16Palette[0x7], 1.5/3.0),
	0x47: rgbmix(db16Palette[0x7], db16Palette[0xc], 1.5/3.0),

	0x48: rgbmix(db16Palette[0x8], db16Palette[0xb], 1.25/3.0),
	0x49: rgbmix(db16Palette[0x8], db16Palette[0xb], 1.75/3.0),

	0x4a: rgbmix(db16Palette[0x6], db16Palette[0x7], 1.0/3.0),
	0x4b: rgbmix(db16Palette[0x6], db16Palette[0x7], 2.0/3.0),
	0x4c: rgbmix(db16Palette[0x6], db16Palette[0xe], 2.0/3.0),
	0x4d: rgbmix(db16Palette[0x3], db16Palette[0xd], 1.25/3.0),
	0x4e: rgbmix(db16Palette[0x3], db16Palette[0xd], 1.75/3.0),
	0x4f: rgbmix(rgb24Color{0x5b6ee1}, db16Palette[0xb], 2.25/3.0),
}

var extendedUnditherer = map[uint8]struct{ c1, c2 uint8 }{
	0x01: {0x10, 0x10}, 0x10: {0x10, 0x10},
	0x03: {0x12, 0x12}, 0x30: {0x12, 0x12},
	0x04: {0x2e, 0x2f}, 0x40: {0x2f, 0x2e},
	0x05: {0x13, 0x10}, 0x50: {0x10, 0x13},
	0x06: {0x13, 0x14}, 0x60: {0x14, 0x13},
	0x08: {0x23, 0x24}, 0x80: {0x24, 0x23},
	0x0b: {0x0c, 0x0c}, 0xb0: {0x0c, 0x0c},
	0x0e: {0x11, 0x1d}, 0xe0: {0x1d, 0x11},
	0x16: {0x10, 0x14}, 0x61: {0x14, 0x10},
	0x19: {0x03, 0x18}, 0x91: {0x18, 0x03},
	0x1b: {0x18, 0x4f}, 0xb1: {0x4f, 0x18},
	0x34: {0x15, 0x1f}, 0x43: {0x1f, 0x15},
	0x3b: {0x35, 0x36}, 0xf3: {0x36, 0x35},
	0x3d: {0x4d, 0x4e}, 0xd3: {0x4e, 0x4d},
	0x3e: {0x1a, 0x17}, 0xe3: {0x17, 0x1a},
	0x47: {0x46, 0x46}, 0x74: {0x46, 0x46},
	0x48: {0x25, 0x26}, 0x84: {0x26, 0x25},
	0x4c: {0x30, 0x31}, 0xc4: {0x31, 0x30},
	0x5b: {0x42, 0x43}, 0xb5: {0x43, 0x42},
	0x5d: {0x32, 0x32}, 0xd5: {0x32, 0x32},
	0x67: {0x4a, 0x4b}, 0x76: {0x4b, 0x4a},
	0x68: {0x11, 0x14}, 0x86: {0x14, 0x11},
	0x6c: {0x29, 0x2a}, 0xc6: {0x2a, 0x29},
	0x6e: {0x16, 0x4c}, 0xe6: {0x16, 0x4c},
	0x6f: {0x1e, 0x17}, 0xf6: {0x17, 0x1e},
	0x78: {0x15, 0x15}, 0x87: {0x15, 0x15},
	0x7c: {0x47, 0x46}, 0xc7: {0x46, 0x47},
	0x7e: {0x3a, 0x3a}, 0xe7: {0x3a, 0x3a},
	0x7f: {0x17, 0x2b}, 0xf7: {0x2b, 0x17},
	0x8b: {0x48, 0x49}, 0xb8: {0x49, 0x48},
	0x8e: {0x40, 0x41}, 0xe8: {0x41, 0x40},
	0x98: {0x2c, 0x2d}, 0x89: {0x2d, 0x2c},
	0xbc: {0x17, 0x1f}, 0xcb: {0x1f, 0x17},
	0xbd: {0x44, 0x45}, 0xdb: {0x45, 0x44},
	0xbf: {0x33, 0x34}, 0xfb: {0x34, 0x33},
	0xc8: {0x27, 0x28}, 0x8c: {0x28, 0x27},
	0xce: {0x20, 0x21}, 0xec: {0x21, 0x20},
	0xdf: {0x39, 0x39}, 0xfd: {0x39, 0x39},
	0xef: {0x22, 0x22}, 0xfe: {0x22, 0x22},
	0x2c: {0x11, 0x1d}, 0xc2: {0x1d, 0x11},
	0xac: {0x1d, 0x1d}, 0xca: {0x1d, 0x1d},

	0x02: {0x3c, 0x24}, 0x20: {0x24, 0x3c},
	0x28: {0x38, 0x38}, 0x82: {0x38, 0x38},
	0x8a: {0x3f, 0x3e}, 0xa8: {0x3e, 0x3f},
	0xaa: {0x3f, 0x3f},
	0x23: {0x02, 0x37}, 0x32: {0x37, 0x02},

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
