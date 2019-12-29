package screen

import "image"

type Buffer interface {
	Clear(color uint8)
	Image() *image.Paletted

	Line(x1, y1, x2, y2 int, color uint8)
	Pattern(cx, cy, size int, isRect bool, isSolid bool, seed uint8, color uint8)
	Fill(cx, cy int, legalColor uint8, color uint8)
}
