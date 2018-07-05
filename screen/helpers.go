package screen

import "image/color"

func lerpUI32(v0, v1 uint32, t float64) uint32 {
	return uint32(float64(v0)*(1-t) + float64(v1)*t)
}

func rgbMix(c1, c2 color.Color, t float64) color.Color {
	r1, g1, b1, _ := c1.RGBA()
	r2, g2, b2, _ := c2.RGBA()
	return color.RGBA{
		R: uint8(lerpUI32(r1, r2, t) >> 8),
		G: uint8(lerpUI32(g1, g2, t) >> 8),
		B: uint8(lerpUI32(b1, b2, t) >> 8),
		A: 0xff,
	}
}

func darken(c color.Color, p float64) color.Color {
	r, g, b, _ := c.RGBA()
	return color.RGBA{
		R: uint8(uint32(float64(r)*(1-p)) >> 8),
		G: uint8(uint32(float64(g)*(1-p)) >> 8),
		B: uint8(uint32(float64(b)*(1-p)) >> 8),
		A: 255,
	}
}

func clamp(i int, min int, max int) int {
	if i < min {
		return min
	}
	if i > max {
		return max
	}
	return i
}

type rgb24Color struct {
	RGB uint32
}

func (ega rgb24Color) RGBA() (r, g, b, a uint32) {
	r = uint32(ega.RGB>>16&0xFF) * 0xFFFF / 0xFF
	g = uint32(ega.RGB>>8&0xFF) * 0xFFFF / 0xFF
	b = uint32(ega.RGB>>0&0xFF) * 0xFFFF / 0xFF
	a = uint32(0xFFFF)
	return
}
