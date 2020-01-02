package screen

import "image"

type Scaler1x1 struct {
	*Ditherer
}

func (s Scaler1x1) NewPic(bounds image.Rectangle) Pic {
	visualDitherer := s.Ditherer
	if s.Ditherer == nil {
		visualDitherer = DefaultDitherers.EGA
	}

	priorityDitherer := &Ditherer{
		Palette:  DefaultPalettes.Depth,
		DitherFn: noDither,
	}
	controlDitherer := &Ditherer{
		Palette:  DefaultPalettes.Depth,
		DitherFn: noDither,
	}

	return &picLayers{
		visual: &buffer1x1{
			Paletted: image.NewPaletted(bounds, visualDitherer.Palette),
			Ditherer: visualDitherer,
		},
		priority: &buffer1x1{
			Paletted: image.NewPaletted(bounds, visualDitherer.Palette),
			Ditherer: priorityDitherer,
		},
		control: &buffer1x1{
			Paletted: image.NewPaletted(bounds, visualDitherer.Palette),
			Ditherer: controlDitherer,
		},
	}
}

type buffer1x1 struct {
	*image.Paletted
	*Ditherer
	stack []point
}

func (buf *buffer1x1) Image() image.Image {
	return buf.Paletted
}

func (buf *buffer1x1) Clear(color uint8) {
	for i, max := 0, len(buf.Pix); i < max; i++ {
		y := i / buf.Stride
		x := i % buf.Stride
		buf.Pix[i] = buf.DitherAt(x, y, color)
	}
}

func (buf *buffer1x1) Line(x1, y1, x2, y2 int, color uint8) {
	var (
		left   = clampInt(0, 319, x1)
		top    = clampInt(0, 189, y1)
		right  = clampInt(0, 319, x2)
		bottom = clampInt(0, 189, y2)
	)

	switch {
	case left == right:
		swapIf(&top, &bottom, top > bottom)
		for y := top; y <= bottom; y++ {
			buf.Pix[y*buf.Stride+left] = buf.DitherAt(left, y, color)
		}
	case top == bottom:
		swapIf(&right, &left, right > left)
		for x := right; x <= left; x++ {
			buf.Pix[top*buf.Stride+x] = buf.DitherAt(x, top, color)
		}
	default:
		// bresenham
		dx, dy := right-left, bottom-top
		stepX, stepY := ((dx>>15)<<1)+1, ((dy>>15)<<1)+1

		dx, dy = absInt(dx)<<1, absInt(dy)<<1

		buf.Pix[top*buf.Stride+left] = buf.DitherAt(left, top, color)
		buf.Pix[bottom*buf.Stride+right] = buf.DitherAt(right, bottom, color)

		if dx > dy {
			fraction := dy - (dx >> 1)
			for left != right {
				if fraction >= 0 {
					top += stepY
					fraction -= dx
				}
				left += stepX
				fraction += dy
				buf.Pix[top*buf.Stride+left] = buf.DitherAt(left, top, color)
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
				buf.Pix[top*buf.Stride+left] = buf.DitherAt(left, top, color)
			}
		}
	}
}

func (buf *buffer1x1) Pattern(cx, cy, size int, isRect, isSolid bool, seed uint8, color uint8) {
	var (
		noiseIndex = noiseOffsets[seed]
		width      = size*2 + 2
		height     = size*2 + 1
		left       = clampInt(0, 320, cx-size)
		top        = clampInt(0, 190, cy-size)
	)

	if left+width > 320 {
		width = 320 - left
	}

	if top+height > 190 {
		height = 190 - top
	}

	if isRect {
		right, bottom := left+width, top+height

		for py := top; py < bottom; py++ {
			offset := py * buf.Stride
			for px := left; px < right; px++ {
				fill := isSolid || noise[noiseIndex%len(noise)]
				if fill {
					buf.Pix[offset+px] = buf.DitherAt(px, py, color)
				}
				noiseIndex++
			}
		}
	} else {
		bitmap := circleBitmaps[size]
		size := len(bitmap)
		for y, row := range bitmap {
			py := top + y
			if py >= 190 {
				break
			}

			offset := py * buf.Stride
			for x := 0; x < size; x++ {
				px := left + x
				if px >= 320 {
					break
				}
				pixel := ((row >> (size - (x + 1))) & 1) == 1
				if pixel {
					fill := isSolid || noise[noiseIndex%len(noise)]
					if fill {
						buf.Pix[offset+px] = buf.DitherAt(px, py, color)
					}
					noiseIndex++
				}
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

		buf.Pix[i] = buf.DitherAt(x, y, color)

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

			buf.Pix[i] = buf.DitherAt(dx, y, color)
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

			buf.Pix[i] = buf.DitherAt(dx, y, color)
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

func clampInt(min, max, i int) int {
	switch {
	case i < min:
		return min
	case i > max:
		return max
	default:
		return i
	}
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
