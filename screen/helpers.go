package screen

import "image/color"

func lerpUI32(v0, v1 uint32, t float64) uint32 {
	return uint32(float64(v0)*(1-t) + float64(v1)*t)
}

func rgbMix(c1, c2 color.Color, t float64) color.Color {
	r1, g1, b1, _ := c1.RGBA()
	r2, g2, b2, _ := c2.RGBA()
	return color.NRGBA64{
		R: uint16(lerpUI32(r1, r2, t)),
		G: uint16(lerpUI32(g1, g2, t)),
		B: uint16(lerpUI32(b1, b2, t)),
		A: 0xffff,
	}
}

type rgb24Color uint32

func (rgb24 rgb24Color) RGBA() (r, g, b, a uint32) {
	rb, gb, bb := (rgb24>>16)&0xFF, (rgb24>>8)&0xFF, (rgb24>>0)&0xFF

	r = uint32((rb << 8) | rb)
	g = uint32((gb << 8) | gb)
	b = uint32((bb << 8) | bb)
	a = 0xFFFF
	return
}

func rgb(r, g, b uint8) color.RGBA {
	return color.RGBA{r, g, b, 255}
}
