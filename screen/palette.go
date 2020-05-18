package screen

import (
	"image/color"
)

type Ditherer struct {
	color.Palette
	ColorMapping
	DitherFn
}

func (d *Ditherer) unpack(c uint8) (uint8, uint8) {
	if d == nil {
		return c & 0xF, c >> 4
	}
	return d.GetMapping(c)
}

func (d *Ditherer) DitherAt(x, y int, c uint8) uint8 {
	ditherFn := d.DitherFn
	if ditherFn == nil {
		ditherFn = dither5050
	}
	c1, c2 := d.unpack(c)
	return ditherFn(x, y, c1, c2)
}

type ColorMapping map[uint8]struct{ c1, c2 uint8 }

func (u ColorMapping) GetMapping(c uint8) (uint8, uint8) {
	if e, ok := u[c]; ok {
		return e.c1, e.c2
	}
	return c & 0xF, c >> 4
}

var DefaultPalettes = struct {
	Depth   color.Palette
	EGA     color.Palette
	DB32EGA color.Palette
	EGACOM  color.Palette
}{
	Depth: color.Palette{
		color.Gray{Y: 0x00},
		color.Gray{Y: 0x11},
		color.Gray{Y: 0x22},
		color.Gray{Y: 0x33},
		color.Gray{Y: 0x44},
		color.Gray{Y: 0x55},
		color.Gray{Y: 0x66},
		color.Gray{Y: 0x77},

		color.Gray{Y: 0x88},
		color.Gray{Y: 0x99},
		color.Gray{Y: 0xAA},
		color.Gray{Y: 0xBB},
		color.Gray{Y: 0xCC},
		color.Gray{Y: 0xDD},
		color.Gray{Y: 0xEE},
		color.Gray{Y: 0xFF},
	},
	EGA: color.Palette{
		rgb24Color(0x000000),
		rgb24Color(0x0000AA),
		rgb24Color(0x00AA00),
		rgb24Color(0x00AAAA),
		rgb24Color(0xAA0000),
		rgb24Color(0xAA00AA),
		rgb24Color(0xAA5500),
		rgb24Color(0xAAAAAA),

		rgb24Color(0x555555),
		rgb24Color(0x5555FF),
		rgb24Color(0x55FF55),
		rgb24Color(0x55FFFF),
		rgb24Color(0xFF5555),
		rgb24Color(0xFF55FF),
		rgb24Color(0xFFFF55),
		rgb24Color(0xFFFFFF),
	},
	DB32EGA: color.Palette{
		rgb24Color(0x000000),
		rgb24Color(0x3f3f74),
		rgb24Color(0x4b692f),
		rgb24Color(0x306082),
		rgb24Color(0xac3232),
		rgb24Color(0x45283c),
		rgb24Color(0x8f563b),
		rgb24Color(0x847e87),

		rgb24Color(0x323c39),
		rgb24Color(0x639bff),
		rgb24Color(0x6abe30),
		rgb24Color(0x5fcde4),
		rgb24Color(0xd95763),
		rgb24Color(0xd77bba),
		rgb24Color(0xfbf236),
		rgb24Color(0xffffff),
	},
	EGACOM: color.Palette{
		0x0: rgb(24, 24, 24),
		0x1: rgb(44, 66, 103),
		0x2: rgb(83, 138, 106),
		0x3: rgb(87, 110, 84),
		0x4: rgb(123, 45, 47),
		0x5: rgb(157, 68, 106),
		0x6: rgb(108, 75, 55),
		0x7: rgb(148, 153, 158),

		0x8: rgb(82, 87, 92),
		0x9: rgb(56, 102, 139),
		0xa: rgb(99, 180, 101),
		0xb: rgb(130, 232, 232),
		0xc: rgb(208, 64, 67),
		0xd: rgb(235, 114, 114),
		0xe: rgb(230, 196, 57),
		0xf: rgb(238, 247, 237),
	},
}

var DefaultDitherers = struct {
	EGA              *Ditherer
	DB32EGA          *Ditherer
	EGACOM           *Ditherer
	ExtendedDitherer *Ditherer
}{
	EGA:     &Ditherer{Palette: DefaultPalettes.EGA},
	DB32EGA: &Ditherer{Palette: DefaultPalettes.DB32EGA},
	EGACOM:  &Ditherer{Palette: DefaultPalettes.EGACOM},
	ExtendedDitherer: &Ditherer{
		Palette: color.Palette{
			0x00: rgb24Color(0x000000), // black
			0x01: rgb24Color(0x3f3f74), // deep-koamaru
			0x02: rgb24Color(0x4b692f), // dell
			0x03: rgb24Color(0x306082), // venice-blue
			0x04: rgb24Color(0xac3232), // brown
			0x05: rgb24Color(0x76428a), // clairvoyant
			0x06: rgb24Color(0x8f563b), // rope
			0x07: rgb24Color(0x847e87), // topaz

			0x08: rgb24Color(0x323c39), // opal
			0x09: rgb24Color(0x639bff), // cornflower
			0x0a: rgb24Color(0x6abe30), // christi
			0x0b: rgb24Color(0x5fcde4), // viking
			0x0c: rgb24Color(0xd95763), // mandy
			0x0d: rgb24Color(0xd77bba), // plum
			0x0e: rgb24Color(0xfbf236), // golden-fizz
			0x0f: rgb24Color(0xffffff), // white

			0x10: rgb24Color(0x222034), // valhalla
			0x11: rgb24Color(0x524b24), // verdigris
			0x12: rgb24Color(0x183041), // ???
			0x13: rgb24Color(0x45283c), // loulou
			0x14: rgb24Color(0x663931), // oiled-cedar
			0x15: rgb24Color(0x595652), // smokey-ash
			0x16: rgb24Color(0xd9a066), // twine
			0x17: rgb24Color(0x9badb7), // heather

			0x18: rgb24Color(0x5b6ee1), // royal blue
			0x19: rgb24Color(0xcbdbfc), // light-steel-blue
			0x1a: rgb24Color(0x8f974a), // rainforest
			0x1b: rgb24Color(0x99e550), // atlantis
			0x1c: rgb24Color(0xdf7126), // tahiti-gold
			0x1d: rgb24Color(0x8a6f30), // stinger
			0x1e: rgb24Color(0xeec39a), // pancho
			0x1f: rgb24Color(0x696a6a), // dim-gray

			// Mixes
			0x20: rgbMix(DefaultPalettes.DB32EGA[0xc], DefaultPalettes.DB32EGA[0xe], 1.0/3.0),
			0x21: rgbMix(DefaultPalettes.DB32EGA[0xc], DefaultPalettes.DB32EGA[0xe], 2.0/3.0),
			0x22: rgbMix(DefaultPalettes.DB32EGA[0xe], DefaultPalettes.DB32EGA[0xf], 0.5),
			0x23: rgbMix(DefaultPalettes.DB32EGA[0x0], DefaultPalettes.DB32EGA[0x8], 1.0/3.0),
			0x24: rgbMix(DefaultPalettes.DB32EGA[0x0], DefaultPalettes.DB32EGA[0x8], 2.0/3.0),
			0x25: rgbMix(DefaultPalettes.DB32EGA[0x4], DefaultPalettes.DB32EGA[0x8], 2.0/4.0),
			0x26: rgbMix(DefaultPalettes.DB32EGA[0x4], DefaultPalettes.DB32EGA[0x8], 2.0/3.0),
			0x27: rgbMix(DefaultPalettes.DB32EGA[0xc], DefaultPalettes.DB32EGA[0x8], 1.0/3.0),
			0x28: rgbMix(DefaultPalettes.DB32EGA[0xc], DefaultPalettes.DB32EGA[0x8], 2.0/3.0),
			0x29: rgbMix(DefaultPalettes.DB32EGA[0x6], DefaultPalettes.DB32EGA[0xc], 1.0/3.0),
			0x2a: rgbMix(DefaultPalettes.DB32EGA[0x6], DefaultPalettes.DB32EGA[0xc], 2.0/3.0),
			0x2b: rgbMix(rgb24Color(0x9badb7), DefaultPalettes.DB32EGA[0xf], 1.5/3.0),
			0x2c: rgbMix(DefaultPalettes.DB32EGA[0x8], DefaultPalettes.DB32EGA[0x9], 1.0/3.0),
			0x2d: rgbMix(DefaultPalettes.DB32EGA[0x8], DefaultPalettes.DB32EGA[0x9], 2.0/3.0),
			0x2e: rgbMix(DefaultPalettes.DB32EGA[0x0], DefaultPalettes.DB32EGA[0x4], 1.0/3.0),
			0x2f: rgbMix(DefaultPalettes.DB32EGA[0x0], DefaultPalettes.DB32EGA[0x4], 2.0/3.0),

			0x30: rgbMix(DefaultPalettes.DB32EGA[0x4], DefaultPalettes.DB32EGA[0xc], 1.0/3.0),
			0x31: rgbMix(DefaultPalettes.DB32EGA[0x4], DefaultPalettes.DB32EGA[0xc], 2.0/3.0),
			0x32: rgb24Color(0xcd69ca),
			0x33: rgbMix(DefaultPalettes.DB32EGA[0xb], DefaultPalettes.DB32EGA[0xf], 1.0/3.0),
			0x34: rgbMix(DefaultPalettes.DB32EGA[0xb], DefaultPalettes.DB32EGA[0xf], 2.0/3.0),
			0x35: rgbMix(DefaultPalettes.DB32EGA[0x3], DefaultPalettes.DB32EGA[0xb], 1.25/3.0),
			0x36: rgbMix(DefaultPalettes.DB32EGA[0x3], DefaultPalettes.DB32EGA[0xb], 1.75/3.0),
			0x37: rgbMix(DefaultPalettes.DB32EGA[0x2], DefaultPalettes.DB32EGA[0xb], 1.0/3.0),
			0x38: rgbMix(DefaultPalettes.DB32EGA[0x2], DefaultPalettes.DB32EGA[0x8], 1.5/3.0),
			0x39: rgbMix(DefaultPalettes.DB32EGA[0xd], DefaultPalettes.DB32EGA[0xf], 1.5/3.0),
			0x3a: rgbMix(rgb24Color(0xeec39a), DefaultPalettes.DB32EGA[0xe], 1.5/3.0),
			0x3b: rgbMix(DefaultPalettes.DB32EGA[0x8], rgb24Color(0x524b24), 1.5/3.0),
			0x3c: rgbMix(DefaultPalettes.DB32EGA[0x0], DefaultPalettes.DB32EGA[0xa], 1.0/3.0),
			0x3d: rgbMix(DefaultPalettes.DB32EGA[0x5], DefaultPalettes.DB32EGA[0x9], 2.0/3.0),
			0x3e: rgbMix(DefaultPalettes.DB32EGA[0x8], DefaultPalettes.DB32EGA[0xa], 1.25/3.0),
			0x3f: rgbMix(DefaultPalettes.DB32EGA[0x8], DefaultPalettes.DB32EGA[0xa], 1.75/3.0),

			0x40: rgbMix(DefaultPalettes.DB32EGA[0x8], DefaultPalettes.DB32EGA[0xe], 1.5/3.0),
			0x41: rgbMix(DefaultPalettes.DB32EGA[0x8], DefaultPalettes.DB32EGA[0xe], 2.5/3.0),
			0x42: rgbMix(DefaultPalettes.DB32EGA[0x5], DefaultPalettes.DB32EGA[0xb], 1.0/3.0),
			0x43: rgbMix(DefaultPalettes.DB32EGA[0x5], DefaultPalettes.DB32EGA[0xb], 2.0/3.0),
			0x44: rgbMix(DefaultPalettes.DB32EGA[0xb], DefaultPalettes.DB32EGA[0xd], 0.75/3.0),
			0x45: rgbMix(DefaultPalettes.DB32EGA[0xb], DefaultPalettes.DB32EGA[0xd], 1.75/3.0),
			0x46: rgbMix(DefaultPalettes.DB32EGA[0x4], DefaultPalettes.DB32EGA[0x7], 1.5/3.0),
			0x47: rgbMix(rgb24Color(0x9badb7), DefaultPalettes.DB32EGA[0xc], 1.75/3.0),
			0x48: rgbMix(DefaultPalettes.DB32EGA[0x8], DefaultPalettes.DB32EGA[0xb], 1.25/3.0),
			0x49: rgbMix(DefaultPalettes.DB32EGA[0x8], DefaultPalettes.DB32EGA[0xb], 1.75/3.0),
			0x4a: rgbMix(DefaultPalettes.DB32EGA[0x6], DefaultPalettes.DB32EGA[0x7], 1.0/3.0),
			0x4b: rgbMix(DefaultPalettes.DB32EGA[0x6], DefaultPalettes.DB32EGA[0x7], 2.0/3.0),
			0x4c: rgbMix(DefaultPalettes.DB32EGA[0x6], DefaultPalettes.DB32EGA[0xe], 2.0/3.0),
			0x4d: rgbMix(DefaultPalettes.DB32EGA[0x3], DefaultPalettes.DB32EGA[0xd], 1.25/3.0),
			0x4e: rgbMix(DefaultPalettes.DB32EGA[0x3], DefaultPalettes.DB32EGA[0xd], 1.75/3.0),
			0x4f: rgbMix(rgb24Color(0x5b6ee1), DefaultPalettes.DB32EGA[0xb], 2.25/3.0),
		},
		ColorMapping: ColorMapping{
			0x01: {0x10, 0x10}, 0x10: {0x10, 0x10},
			0x02: {0x3c, 0x24}, 0x20: {0x24, 0x3c},
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
			0x23: {0x02, 0x37}, 0x32: {0x37, 0x02},
			0x28: {0x38, 0x38}, 0x82: {0x38, 0x38},
			0x2c: {0x11, 0x1d}, 0xc2: {0x1d, 0x11},
			0x34: {0x15, 0x1f}, 0x43: {0x1f, 0x15},
			0x3b: {0x35, 0x36}, 0xb3: {0x36, 0x35},
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
			//0x6f: {0x1e, 0x17}, 0xf6: {0x17, 0x1e},
			0x78: {0x15, 0x15}, 0x87: {0x15, 0x15},
			0x7c: {0x06, 0x47}, 0xc7: {0x47, 0x06},
			0x7e: {0x3a, 0x3a}, 0xe7: {0x3a, 0x3a},
			//0x7f: {0x17, 0x2b}, 0xf7: {0x2b, 0x17},
			0x8a: {0x3f, 0x3e}, 0xa8: {0x3e, 0x3f},
			0x8b: {0x48, 0x49}, 0xb8: {0x49, 0x48},
			0x8e: {0x40, 0x41}, 0xe8: {0x41, 0x40},
			0x98: {0x2c, 0x2d}, 0x89: {0x2d, 0x2c},
			0xaa: {0x3f, 0x3f},
			0xac: {0x1d, 0x1d}, 0xca: {0x1d, 0x1d},
			0xbc: {0x17, 0x1f}, 0xcb: {0x1f, 0x17},
			0xbd: {0x44, 0x45}, 0xdb: {0x45, 0x44},
			//0xbf: {0x33, 0x34}, 0xfb: {0x34, 0x33},
			0xc8: {0x27, 0x28}, 0x8c: {0x28, 0x27},
			0xce: {0x20, 0x21}, 0xec: {0x21, 0x20},
			//0xcf: {0x0d, 0x1e}, 0xfc: {0x1e, 0x0d},
			//0xdf: {0x39, 0x39}, 0xfd: {0x39, 0x39},
			0xde: {0x1e, 0x0c}, 0xed: {0x0c, 0x1e},
			//0xef: {0x22, 0x22}, 0xfe: {0x22, 0x22},
			//0x3f: {0x17, 0x17}, 0xf3: {0x17, 0x17},
			0x59: {0x3d, 0x3d}, 0x95: {0x3d, 0x3d},
		},
	},
}

func NewUnditherer(pal color.Palette) *Ditherer {
	return NewMixDitherer(pal, 0.5)
}

func NewMixDitherer(pal color.Palette, ratio float64) *Ditherer {
	newPal := make(color.Palette, len(pal), 256)
	mapping := ColorMapping{}
	copy(newPal, pal)
	for a := uint8(0); a < 0xf; a++ {
		for b := a + 1; b < 0xf; b++ {
			idx1 := uint8(len(newPal))
			newPal = append(newPal, rgbMix(pal[a], pal[b], ratio))
			idx2 := uint8(len(newPal))
			newPal = append(newPal, rgbMix(pal[b], pal[a], ratio))
			mapping[a<<4|b] = struct{ c1, c2 uint8 }{idx2, idx1}
			mapping[b<<4|a] = struct{ c1, c2 uint8 }{idx1, idx2}
		}
	}
	return &Ditherer{
		Palette:      newPal,
		ColorMapping: mapping,
	}
}

func NewAdaptiveDithering(in color.Palette, lower, upper float64) *Ditherer {
	pal := make([]color.Color, 16, 256)
	copy(pal, in)
	mapping := ColorMapping{}
	for c1 := 0; c1 < 0x0f; c1++ {
		for c2 := c1 + 1; c2 < 0x0f; c2++ {
			r1, g1, b1, _ := pal[c1].RGBA()
			r2, g2, b2, _ := pal[c2].RGBA()

			lum1 := (299*r1 + 587*g1 + 114*b1) / 1000
			lum2 := (299*r2 + 587*g2 + 114*b2) / 1000
			var dLum uint32
			if lum1 > lum2 {
				dLum = lum1 - lum2
			} else {
				dLum = lum2 - lum1
			}

			if dLum >= uint32(0xffff*upper) {
				m1 := rgbMix(pal[c1], pal[c2], 2.0/7.0)
				m2 := rgbMix(pal[c1], pal[c2], 5.0/7.0)
				idx := uint8(len(pal))
				pal = append(append(pal, m1), m2)
				mapping[uint8(c1<<4|c2)] = struct{ c1, c2 uint8 }{idx + 1, idx}
				mapping[uint8(c2<<4|c1)] = struct{ c1, c2 uint8 }{idx, idx + 1}
			} else if dLum >= uint32(0xffff*lower) {
				m1 := rgbMix(pal[c1], pal[c2], 1.0/3.0)
				m2 := rgbMix(pal[c1], pal[c2], 2.0/3.0)
				idx := uint8(len(pal))
				pal = append(append(pal, m1), m2)
				mapping[uint8(c1<<4|c2)] = struct{ c1, c2 uint8 }{idx + 1, idx}
				mapping[uint8(c2<<4|c1)] = struct{ c1, c2 uint8 }{idx, idx + 1}
			} else {
				m1 := rgbMix(pal[c1], pal[c2], 0.5)
				idx := uint8(len(pal))
				pal = append(pal, m1)
				mapping[uint8(c1<<4|c2)] = struct{ c1, c2 uint8 }{idx, idx}
				mapping[uint8(c2<<4|c1)] = struct{ c1, c2 uint8 }{idx, idx}
			}
		}
	}

	return &Ditherer{
		Palette:      pal,
		ColorMapping: mapping,
	}
}
