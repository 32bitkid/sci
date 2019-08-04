package screen

import "image"

type Scaler1x1 struct {
	VisualDitherer *Ditherer
}

func (s Scaler1x1) NewPic(bounds image.Rectangle) Pic {
	return Pic{
		Visual:   s.newVisual(bounds),
		Priority: s.newPriority(bounds),
		Control:  s.newControl(bounds),
	}
}

func (s Scaler1x1) newVisual(r image.Rectangle) Buffer {
	palette := EGAPalette
	if s.VisualDitherer != nil {
		dPal := s.VisualDitherer.Palette
		if dPal != nil {
			palette = dPal
		}
	}
	return &buffer1x1{
		Paletted: image.NewPaletted(r, palette),
		ditherer: s.VisualDitherer,
		ditherFn: dither5050,
	}
}

func (s Scaler1x1) newControl(r image.Rectangle) Buffer {
	return &buffer1x1{
		Paletted: image.NewPaletted(r, Depth16Palette),
		ditherFn: noDither,
	}
}

func (s Scaler1x1) newPriority(r image.Rectangle) Buffer {
	return &buffer1x1{
		Paletted: image.NewPaletted(r, EGAPalette),
		ditherFn: noDither,
	}
}

type buffer1x1 struct {
	*image.Paletted
	ditherer *Ditherer
	ditherFn
	stack []point
}

type ditherFn func(x, y int, c1, c2 uint8) uint8

var noDither = func(x, y int, c1, _ uint8) uint8 { return c1 }
var dither5050 ditherFn = func(x, y int, c1, c2 uint8) uint8 {
	if (x&1)^(y&1) == 0 {
		return c1
	}
	return c2
}

func (buf *buffer1x1) Clear(c uint8) {
	for i, max := 0, len(buf.Paletted.Pix); i < max; i++ {
		buf.Paletted.Pix[i] = c
	}
}

func (buf *buffer1x1) Image() *image.Paletted {
	return buf.Paletted
}

func (buf *buffer1x1) dither(x, y int, color uint8) uint8 {
	c1, c2 := buf.ditherer.Get(color)
	if buf.ditherFn != nil {
		return buf.ditherFn(x, y, c1, c2)
	}
	return color
}

func (buf *buffer1x1) Line(x1, y1, x2, y2 int, color uint8) {
	left, top, right, bottom := x1, y1, x2, y2
	clip(&left, 0, 319)
	clip(&top, 0, 189)
	clip(&right, 0, 319)
	clip(&bottom, 0, 189)

	switch {
	case left == right:
		swapIf(&top, &bottom, top > bottom)
		for y := top; y <= bottom; y++ {
			buf.Pix[y*buf.Stride+left] = buf.dither(left, y, color)
		}
	case top == bottom:
		swapIf(&right, &left, right > left)
		for x := right; x <= left; x++ {
			buf.Pix[top*buf.Stride+x] = buf.dither(x, top, color)
		}
	default:
		// bresenham
		dx, dy := right-left, bottom-top
		stepX, stepY := ((dx>>15)<<1)+1, ((dy>>15)<<1)+1

		dx, dy = absInt(dx)<<1, absInt(dy)<<1

		buf.Pix[top*buf.Stride+left] = buf.dither(left, top, color)
		buf.Pix[bottom*buf.Stride+right] = buf.dither(right, bottom, color)

		if dx > dy {
			fraction := dy - (dx >> 1)
			for left != right {
				if fraction >= 0 {
					top += stepY
					fraction -= dx
				}
				left += stepX
				fraction += dy
				buf.Pix[top*buf.Stride+left] = buf.dither(left, top, color)
			}
		} else {
			fraction := dx - (dy >> 1)
			for top != bottom {
				if fraction >= 0 {
					left += stepX
					fraction -= dy
				}
				top += stepY
				fraction += dx
				buf.Pix[top*buf.Stride+left] = buf.dither(left, top, color)
			}
		}
	}
}

func (buf *buffer1x1) Pattern(cx, cy, size int, isRect, isSolid bool, seed uint8, color uint8) {
	patternIndex := vectorPatternTextureOffset[seed]
	if isRect {
		for y := -size; y <= size; y++ {
			if cy+y < 0 || cy+y >= 190 {
				continue
			}

			offset := (cy + y) * buf.Stride
			for x := -size; x <= size+1; x++ {
				if cx+x < 0 || cx+x >= 320 {
					continue
				}
				if isSolid || vectorPatternTextures[patternIndex%len(vectorPatternTextures)] {
					buf.Pix[offset+cx+x] = buf.dither(cx+x, y, color)
				}
				patternIndex++
			}
		}
	} else {
		r2 := size * size
		for y := -size; y <= size; y++ {
			if cy+y < 0 || cy+y >= 190 {
				continue
			}

			offset := (cy + y) * buf.Stride
			sx := sqrts[r2-y*y]
			for x := -sx; x <= sx; x++ {
				if cx+x < 0 || cx+x >= 320 {
					continue
				}
				if isSolid || vectorPatternTextures[patternIndex%len(vectorPatternTextures)] {
					buf.Pix[offset+cx+x] = buf.dither(cx+x, y, color)
				}
				patternIndex++
			}
		}
	}
}

func (buf *buffer1x1) isLegal(p point, legalColor uint8) bool {
	idx := p.y*buf.Stride + p.x
	return buf.Pix[idx] == legalColor
}

func (buf *buffer1x1) Fill(cx, cy int, legalColor uint8, color uint8) {
	var (
		p      point
		stride = buf.Stride
		stack  = buf.stack
	)

	// initial
	stack = append(stack, point{cx, cy})

	for len(stack) > 0 {
		p, stack = stack[0], stack[1:]

		var (
			x, y = p.x, p.y
			i    = y*stride + x
		)

		if !buf.isLegal(p, legalColor) {
			continue
		}

		buf.Pix[i] = buf.dither(x, y, color)

		if down := (point{x, y + 1}); down.y < 190 {
			if buf.isLegal(down, legalColor) {
				stack = append(stack, down)
			}
		}

		if up := (point{x, y - 1}); up.y >= 0 {
			if buf.isLegal(up, legalColor) {
				stack = append(stack, up)
			}
		}

		// flood right
		for dx := x + 1; dx < 320; dx++ {
			var i = y*stride + dx
			if buf.Pix[i] != legalColor {
				break
			}

			buf.Pix[i] = buf.dither(dx, y, color)
			if down := (point{dx, y + 1}); down.y < 190 {
				if buf.isLegal(down, legalColor) {
					stack = append(stack, down)
				}
			}
			if up := (point{dx, y - 1}); up.y >= 0 {
				if buf.isLegal(up, legalColor) {
					stack = append(stack, up)
				}
			}
		}

		// flood left
		for dx := x - 1; dx >= 0; dx-- {
			var i = y*stride + dx
			if buf.Pix[i] != legalColor {
				break
			}

			buf.Pix[i] = buf.dither(dx, y, color)
			if down := (point{dx, y + 1}); down.y < 190 {
				if buf.isLegal(down, legalColor) {
					stack = append(stack, down)
				}
			}
			if up := (point{dx, y - 1}); up.y >= 0 {
				if buf.isLegal(up, legalColor) {
					stack = append(stack, up)
				}
			}
		}
	}
}

type point struct{ x, y int }

// (0..(7*7)) => i => int(math.Round(math.Sqrt(float64(i))))
var sqrts = [50]int{
	0, 1, 1, 2, 2, 2, 2,
	3, 3, 3, 3, 3, 3, 4,
	4, 4, 4, 4, 4, 4, 4,
	5, 5, 5, 5, 5, 5, 5,
	5, 5, 5, 6, 6, 6, 6,
	6, 6, 6, 6, 6, 6, 6,
	6, 7, 7, 7, 7, 7, 7,
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func swapIf(a, b *int, cond bool) {
	if cond {
		*a, *b = *b, *a
	}
}

func clip(v *int, min, max int) {
	switch {
	case *v < min:
		*v = min
	case *v > max:
		*v = max
	}
}
