package screen

import "image/color"
import clr "github.com/lucasb-eyer/go-colorful"

func rgbMix(c1, c2 color.Color, t float64) color.Color {
	clr1, _ := clr.MakeColor(c1)
	clr2, _ := clr.MakeColor(c2)
	if (clr1.R == clr1.G && clr1.G == clr1.B) || (clr2.R == clr2.G && clr2.G == clr2.B) {
		return clr1.BlendRgb(clr2, t).Clamped()
	}
	return clr1.BlendLab(clr2, t).Clamped()
}

func lighten(src color.Color, p float64) color.Color {
	srcColor, _ := clr.MakeColor(src)
	h, c, l := srcColor.Hcl()
	return clr.Hcl(h, c, l + p).Clamped()
}

func darken(src color.Color, p float64) color.Color {
	srcColor, _ := clr.MakeColor(src)
	h, c, l := srcColor.Hcl()
	return clr.Hcl(h, c, l - p).Clamped()
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
