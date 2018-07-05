package screen

import (
	"image"
	"image/color"
)

var (
	red   = color.RGBA{R: 0xFF, G: 0x99, B: 0x99, A: 0xff}
	green = color.RGBA{G: 0xFF, R: 0x99, B: 0x99, A: 0xff}
	blue  = color.RGBA{B: 0xFF, R: 0x99, G: 0x99, A: 0xff}
)

func rgbMul(a, b color.Color, _ float64) color.Color {
	r1, g1, b1, _ := a.RGBA()
	r2, g2, b2, _ := b.RGBA()
	return color.RGBA{
		R: uint8((r1 * r2 / 0xffff) >> 8),
		G: uint8((g1 * g2 / 0xffff) >> 8),
		B: uint8((b1 * b2 / 0xffff) >> 8),
		A: 0xFF,
	}
}

func RenderToCRT(src image.Image) image.Image {
	srcRect := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, srcRect.Dx()*6, srcRect.Dy()*6))
	for sy, dy := srcRect.Min.Y, 0; sy < srcRect.Max.Y; sy, dy = sy+1, dy+6 {
		for sx, dx := srcRect.Min.X, 0; sx < srcRect.Max.X; sx, dx = sx+1, dx+6 {
			lc := src.At(clamp(sx-1, srcRect.Min.X, srcRect.Max.X), sy)
			c := src.At(sx, sy)
			rc := src.At(clamp(sx+1, srcRect.Min.X, srcRect.Max.X), sy)
			for i := 0; i < 36; i++ {
				ix, iy := i%6, i/6
				co := c

				// Bleed
				switch ix {
				case 0:
					co = rgbMix(lc, c, 3.0/6.0)
				case 1:
					co = rgbMix(lc, c, 4.0/6.0)
				case 2:
					co = rgbMix(lc, c, 5.0/6.0)
				case 4:
					co = rgbMix(c, rc, 1.0/6.0)
				case 5:
					co = rgbMix(c, rc, 2.0/6.0)
				}

				// Scan-lines
				switch iy {
				case 0:
					co = darken(co, 0.7)
				case 1:
					co = darken(co, 0.2)
				case 4:
					co = darken(co, 0.1)
				case 5:
					co = darken(co, 0.4)
				}

				const smVal = 0.2

				// Shadow Mask
				switch iy % 2 {
				case 0:
					switch ix {
					case 0, 1:
						co = rgbMul(co, red, smVal)
					case 2, 3:
						co = rgbMul(co, green, smVal)
					case 4, 5:
						co = rgbMul(co, blue, smVal)
					}
				case 1:
					switch ix {
					case 3, 4:
						co = rgbMul(co, red, smVal)
					case 0, 5:
						co = rgbMul(co, green, smVal)
					case 1, 2:
						co = rgbMul(co, blue, smVal)
					}
				}

				dst.Set(dx+ix, dy+iy, co)
			}
		}
	}

	return dst
}
