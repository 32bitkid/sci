package screen

import (
	"image"
)

type Scaler5x6 struct {
	*Ditherer
	TextureMap
}

func (s Scaler5x6) NewPic(bounds image.Rectangle) Pic {
	normBounds := image.Rect(0, 0, bounds.Dx(), bounds.Dy())
	scaledBounds := image.Rect(0, 0, bounds.Dx()*5, bounds.Dy()*6)

	ditherer := s.Ditherer
	if ditherer == nil {
		ditherer = DefaultDitherers.EGA
	}

	priorityDitherer := &Ditherer{
		Palette:  DefaultPalettes.Depth,
		DitherFn: noDither,
	}

	controlDitherer := &Ditherer{
		Palette:  DefaultPalettes.EGA,
		DitherFn: noDither,
	}

	fillBuffer := make([]uint8, bounds.Dx()*bounds.Dy())
	for i, max := 0, len(fillBuffer); i < max; i++ {
		fillBuffer[i] = 15
	}

	return &picLayers{
		visual: &buffer5x6{
			Paletted:   image.NewPaletted(scaledBounds, s.Ditherer.Palette),
			Ditherer:   ditherer,
			TextureMap: s.TextureMap,

			fillBuffer: fillBuffer,
			bounds:     normBounds,
		},
		priority: &buffer1x1{
			Paletted: image.NewPaletted(normBounds, priorityDitherer.Palette),
			Ditherer: priorityDitherer,
		},
		control: &buffer1x1{
			Paletted: image.NewPaletted(normBounds, controlDitherer.Palette),
			Ditherer: controlDitherer,
		},
	}
}

type buffer5x6 struct {
	*image.Paletted
	*Ditherer
	TextureMap

	fillBuffer []uint8
	bounds     image.Rectangle
	stack      []point
}

func (b buffer5x6) Clear(color uint8) {
	for i, max := 0, len(b.Pix); i < max; i++ {
		y := i / b.Stride / 320
		x := i % b.Stride / 320
		b.Pix[i] = b.DitherAt(x, y, color)
	}
}

func (b buffer5x6) Image() image.Image {
	return b.Paletted
}

func (b buffer5x6) plot(x, y int, color uint8) {
	b.fillBuffer[b.bounds.Dx()*y+x] = b.DitherAt(x, y, color)

	px, py := x*5, y*6
	texture, ok := b.TextureMap[color]

	if !ok || texture == nil {
		c := b.DitherAt(x, y, color)
		for h := 0; h < 6; h++ {
			for w := 0; w < 5; w++ {
				dx, dy := px+w, py+h
				offset := dy*b.Stride + dx
				b.Pix[offset] = c

			}
		}
	} else {
		yTexLen := len(texture)
		yti := y % yTexLen
		xTexLen := len(texture[yti])
		xti := x % xTexLen
		tex := texture[yti][xti]

		c1, c2 := b.GetMapping(color)

		for h := 0; h < 6; h++ {
			for w := 0; w < 5; w++ {
				dx, dy := px+w, py+h
				offset := dy*b.Stride + dx
				if tex[h][w] {
					b.Pix[offset] = c1
				} else {
					b.Pix[offset] = c2
				}

			}
		}
	}
}

func (b buffer5x6) Line(x1, y1, x2, y2 int, color uint8) {
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
			b.plot(left, y, color)
		}
	case top == bottom:
		swapIf(&right, &left, right > left)
		for x := right; x <= left; x++ {
			b.plot(x, top, color)
		}
	default:
		// bresenham
		dx, dy := right-left, bottom-top
		stepX, stepY := ((dx>>15)<<1)+1, ((dy>>15)<<1)+1

		dx, dy = absInt(dx)<<1, absInt(dy)<<1

		b.plot(left, top, color)
		b.plot(right, bottom, color)

		if dx > dy {
			fraction := dy - (dx >> 1)
			for left != right {
				if fraction >= 0 {
					top += stepY
					fraction -= dx
				}
				left += stepX
				fraction += dy
				b.plot(left, top, color)
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
				b.plot(left, top, color)
			}
		}
	}
}

func (b buffer5x6) Pattern(cx, cy, size int, isRect bool, isSolid bool, seed uint8, color uint8) {
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
			for px := left; px < right; px++ {
				fill := isSolid || noise[noiseIndex%len(noise)]
				if fill {
					b.plot(px, py, color)
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

			for x := 0; x < size; x++ {
				px := left + x
				if px >= 320 {
					break
				}
				pixel := ((row >> (size - (x + 1))) & 1) == 1
				if pixel {
					fill := isSolid || noise[noiseIndex%len(noise)]
					if fill {
						b.plot(px, py, color)
					}
					noiseIndex++
				}
			}
		}
	}
}

func (b buffer5x6) isLegal(p point, legalColor uint8) bool {
	return b.fillBuffer[p.y*b.bounds.Dx()+p.x] == legalColor
}

func (b buffer5x6) Fill(cx, cy int, legalColor uint8, color uint8) {
	var (
		p     point
		stack = b.stack
	)

	// initial
	stack = append(stack, point{cx, cy})

	for len(stack) > 0 {
		p, stack = stack[0], stack[1:]

		var (
			x, y = p.x, p.y
		)

		if !b.isLegal(p, legalColor) {
			continue
		}

		b.plot(x, y, color)

		if down := (point{x, y + 1}); down.y < 190 {
			if b.isLegal(down, legalColor) {
				stack = append(stack, down)
			}
		}

		if up := (point{x, y - 1}); up.y >= 0 {
			if b.isLegal(up, legalColor) {
				stack = append(stack, up)
			}
		}

		// flood right
		for dx := x + 1; dx < 320; dx++ {
			right := point{dx, y}
			if !b.isLegal(right, legalColor) {
				break
			}

			b.plot(right.x, right.y, color)
			if down := (point{dx, y + 1}); down.y < 190 {
				if b.isLegal(down, legalColor) {
					stack = append(stack, down)
				}
			}
			if up := (point{dx, y - 1}); up.y >= 0 {
				if b.isLegal(up, legalColor) {
					stack = append(stack, up)
				}
			}
		}

		// flood left
		for dx := x - 1; dx >= 0; dx-- {
			left := point{dx, y}
			if !b.isLegal(left, legalColor) {
				break
			}

			b.plot(left.x, left.y, color)
			if down := (point{dx, y + 1}); down.y < 190 {
				if b.isLegal(down, legalColor) {
					stack = append(stack, down)
				}
			}
			if up := (point{dx, y - 1}); up.y >= 0 {
				if b.isLegal(up, legalColor) {
					stack = append(stack, up)
				}
			}
		}
	}
}
