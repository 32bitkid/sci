package resources

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/32bitkid/bitreader"
	"image"
	"math/rand"
)

type ditherFn func(x, y int, c1, c2 uint8) uint8

var noDither = func(x, y int, c1, _ uint8) uint8 { return c1 }
var dither5050 ditherFn = func(x, y int, c1, c2 uint8) uint8 {
	if (x&1)^(y&1) == 0 {
		return c1
	}
	return c2
}

type picReader struct {
	bits bitreader.BitReader
}

// getAbsCoords gets reads an absolute position from the
// bit-stream. The format is 24-bits long:
//
// bits  |
//  0-3  | high "byte" of x-position
//  4-7  | high "byte" of y-position
//  8-15 | low byte of x-position
// 16-23 | low byte of y-position
//
func (p picReader) getAbsCoords() (int, int, error) {
	code, err := p.bits.Read32(24)
	if err != nil {
		return 0, 0, err
	}
	x := ((code & 0xF00000) >> 12) | ((code & 0xFF00) >> 8)
	y := ((code & 0x0F0000) >> 8) | ((code & 0x00FF) >> 0)
	return int(x), int(y), nil
}

// getRelCoords2 reads a medium length delta from the bit-stream.
// The total payload is 16-bits long:
//
// bits |
// 0-7  | y-delta
// 8-15 | x-delta
//
func (p picReader) getRelCoords2(x1, y1 int) (int, int, error) {
	dy, err := p.bits.Read8(8)
	if err != nil {
		return 0, 0, err
	}

	dx, err := p.bits.Read8(8)
	if err != nil {
		return 0, 0, err
	}

	x2, y2 := x1, y1

	if dy&0x80 != 0 {
		y2 -= int(dy & 0x7F)
	} else {
		y2 += int(dy & 0x7F)
	}

	if dx&0x80 != 0 {
		x2 -= 128 - int(dx&0x7F)
	} else {
		x2 += int(dx & 0x7F)
	}

	return x2, y2, nil
}

// getRelCoords1 reads a medium length delta from the bit-stream.
// The total payload is 8-bits long:
//
// bits |
// 0-3  | y-delta
// 4-7  | x-delta
//
func (p picReader) getRelCoords1(x, y int) (int, int, error) {
	xSign, err := p.bits.Read1()
	if err != nil {
		return 0, 0, err
	}
	dx, err := p.bits.Read8(3)
	if err != nil {
		return 0, 0, err
	}

	ySign, err := p.bits.Read1()
	if err != nil {
		return 0, 0, err
	}
	dy, err := p.bits.Read8(3)
	if err != nil {
		return 0, 0, err
	}

	if xSign {
		x -= int(dx)
	} else {
		x += int(dx)
	}

	if ySign {
		y -= int(dy)
	} else {
		y += int(dy)
	}

	return x, y, nil
}

// Picture Op-Codes
type pOpCode uint8

const (
	pOpSetColor         pOpCode = 0xf0
	pOpDisableVisual            = 0xf1
	pOpSetPriority              = 0xf2
	pOpDisablePriority          = 0xf3
	pOpShortPatterns            = 0xf4
	pOpMediumLines              = 0xf5
	pOpLongLines                = 0xf6
	pOpShortLines               = 0xf7
	pOpFill                     = 0xf8
	pOpSetPattern               = 0xf9
	pOpAbsolutePatterns         = 0xfa
	pOpSetControl               = 0xfb
	pOpDisableControl           = 0xfc
	pOpMediumPatterns           = 0xfd
	pOpOPX                      = 0xfe
	pOpDone                     = 0xff
)

// Extended Picture Op-Codes
type pOpxCode uint8

const (
	pOpxUpdatePaletteEntries pOpxCode = 0x00
	pOpxSetPalette                    = 0x01
)

// picPalette is an array of 40 uint8 values, which is actually a tuple of two 4-bit EGA colors.
type picPalette [40]uint8

var defaultPalette = picPalette{
	0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77,
	0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0x88,
	0x88, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x88,
	0x88, 0xf9, 0xfa, 0xfb, 0xfc, 0xfd, 0xfe, 0xff,
	0x08, 0x91, 0x2a, 0x3b, 0x4c, 0x5d, 0x6e, 0x88,
}

type picDrawMode uint

func (mode *picDrawMode) Set(flag picDrawMode, value bool) {
	if value {
		*mode |= flag
	} else {
		*mode &= ^flag
	}
}

func (mode picDrawMode) Has(flag picDrawMode) bool {
	return mode&flag == flag
}

const (
	picDrawVisual   picDrawMode = 1
	picDrawPriority             = 2
	picDrawControl              = 4
)

type picState struct {
	palettes [4]picPalette
	drawMode picDrawMode

	color        uint8
	priorityCode uint8
	controlCode  uint8

	patternCode    uint8
	patternTexture uint8

	visual   *image.Paletted
	priority *image.Paletted
	control  *image.Paletted
	aux      *image.Paletted

	unditherer Unditherer

	debugFn func(*picState, ...interface{})
	fillStack []point
}

func (s *picState) debugger(params ...interface{}) {
	if s.debugFn != nil {
		s.debugFn(s, params...)
	}
}

func (s *picState) fill(cx, cy int) {
	switch {
	case s.drawMode.Has(picDrawVisual):
		// TODO this doesn't seem like it should happen. either legalColor shouldn't be 15 or color shouldn't be 15/15
		if s.color == 255 {
			return
		}
		c1, c2 := s.unditherer.color(s.color)
		fill(cx, cy, 0xf, s.visual, c1, c2, dither5050, s.fillStack)
		s.debugger()
	case s.drawMode.Has(picDrawPriority):
		if s.priorityCode == 0 {
			return
		}
		c := s.priorityCode
		fill(cx, cy, 0x0, s.priority, c, c, noDither, s.fillStack)
	case s.drawMode.Has(picDrawControl):
		if s.controlCode == 0 {
			return
		}
		c := s.controlCode
		fill(cx, cy, 0x0, s.control, c, c, noDither, s.fillStack)
	default:
		return
	}
}

func (s *picState) line(x1, y1, x2, y2 int) {
	if s.drawMode.Has(picDrawVisual) {
		c1, c2 := s.unditherer.color(s.color)
		line(x1, y1, x2, y2, s.visual, c1, c2, dither5050)
		s.debugger()
	}
	if s.drawMode.Has(picDrawPriority) {
		c := s.priorityCode
		line(x1, y1, x2, y2, s.priority, c, c, noDither)
	}
	if s.drawMode.Has(picDrawControl) {
		c := s.controlCode
		line(x1, y1, x2, y2, s.control, c, c, noDither)
	}
}

func (s *picState) drawPattern(cx, cy int) {
	size := int(s.patternCode & 0x7)
	isRect := s.patternCode&0x10 != 0
	solid := s.patternCode&0x20 == 0

	if s.drawMode.Has(picDrawVisual) {
		c1, c2 := s.unditherer.color(s.color)
		drawPattern(cx, cy, size, isRect, solid, s.visual, c1, c2, dither5050)
		s.debugger()
	}
	if s.drawMode.Has(picDrawPriority) {
		c := s.priorityCode
		drawPattern(cx, cy, size, isRect, solid, s.priority, c, c, noDither)
	}
	if s.drawMode.Has(picDrawControl) {
		c := s.controlCode
		drawPattern(cx, cy, size, isRect, solid, s.control, c, c, noDither)
	}
}

type PicOptions struct {
	*Colors
	DebugFn func(*picState, ...interface{})
}

func ReadPic(
	resource *Resource,
	options ...*PicOptions,
) (image.Image, error) {
	r := picReader{
		bitreader.NewReader(bufio.NewReader(bytes.NewReader(resource.bytes))),
	}

	var debugFn func(*picState, ...interface{})
	palette := RetroEGAColors.Palette
	unditherer := RetroEGAColors.Unditherer

	for _, opts := range options {
		if opts == nil {
			continue
		}

		if opts.Colors != nil && opts.Palette != nil {
			palette = opts.Palette
		}

		if opts.Colors != nil && opts.Unditherer != nil {
			unditherer = opts.Unditherer
		}

		if opts.DebugFn != nil {
			debugFn = opts.DebugFn
		}
	}

	var state = picState{
		visual:   image.NewPaletted(image.Rect(0, 0, 320, 190), palette),
		priority: image.NewPaletted(image.Rect(0, 0, 320, 190), Depth16Palette),
		control:  image.NewPaletted(image.Rect(0, 0, 320, 190), Depth16Palette),
		aux:      image.NewPaletted(image.Rect(0, 0, 320, 190), Depth16Palette),
		drawMode: picDrawVisual | picDrawPriority,
		palettes: [...]picPalette{
			defaultPalette,
			defaultPalette,
			defaultPalette,
			defaultPalette,
		},
		unditherer: unditherer,
		debugFn:    debugFn,
		fillStack:  make([]point, 0, 320*190),
	}

	for i := 0; i < (320 * 190); i++ {
		state.visual.Pix[i] = 0xf
	}

opLoop:
	for {
		op, err := r.bits.Read8(8)
		if err != nil {
			return nil, err
		}

		switch pOpCode(op) {
		case pOpSetColor:
			code, err := r.bits.Read8(8)
			if err != nil {
				return nil, err
			}

			pal := code / 40
			index := code % 40
			color := state.palettes[pal][index]
			state.color = color
			state.drawMode.Set(picDrawVisual, true)
		case pOpDisableVisual:
			state.drawMode.Set(picDrawVisual, false)

		case pOpSetPriority:
			code, err := r.bits.Read8(8)
			if err != nil {
				return nil, err
			}
			state.priorityCode = code & 0xF
			state.drawMode.Set(picDrawPriority, true)
		case pOpDisablePriority:
			state.drawMode.Set(picDrawPriority, false)

		case pOpSetControl:
			code, err := r.bits.Read8(8)
			if err != nil {
				return nil, err
			}
			state.controlCode = code & 0xf
			state.drawMode.Set(picDrawControl, true)
		case pOpDisableControl:
			state.drawMode.Set(picDrawControl, false)

		case pOpShortLines:
			x1, y1, err := r.getAbsCoords()
			if err != nil {
				return nil, err
			}
			for {
				if peek, err := r.bits.Peek8(8); err != nil {
					return nil, err
				} else if peek >= 0xf0 {
					break
				}

				x2, y2, err := r.getRelCoords1(x1, y1)
				if err != nil {
					return nil, err
				}

				state.line(x1, y1, x2, y2)
				x1, y1 = x2, y2
			}
		case pOpMediumLines:
			x1, y1, err := r.getAbsCoords()
			if err != nil {
				return nil, err
			}
			for {
				if peek, err := r.bits.Peek8(8); err != nil {
					return nil, err
				} else if peek >= 0xf0 {
					break
				}

				x2, y2, err := r.getRelCoords2(x1, y1)
				if err != nil {
					return nil, err
				}

				state.line(x1, y1, x2, y2)
				x1, y1 = x2, y2
			}
		case pOpLongLines:
			x1, y1, err := r.getAbsCoords()
			if err != nil {
				return nil, err
			}
			for {
				if peek, err := r.bits.Peek8(8); err != nil {
					return nil, err
				} else if peek >= 0xf0 {
					break
				}

				x2, y2, err := r.getAbsCoords()
				if err != nil {
					return nil, err
				}

				state.line(x1, y1, x2, y2)
				x1, y1 = x2, y2
			}

		case pOpFill:
			for {
				if peek, err := r.bits.Peek8(8); err != nil {
					return nil, err
				} else if peek >= 0xf0 {
					break
				}

				x, y, err := r.getAbsCoords()
				if err != nil {
					return nil, err
				}

				state.fill(x, y)
			}

		case pOpSetPattern:
			code, err := r.bits.Read8(8)
			if err != nil {
				return nil, err
			}
			state.patternCode = code & 0x3f
		case pOpShortPatterns:
			if state.patternCode&0x20 != 0 {
				texture, err := r.bits.Read8(8)
				if err != nil {
					return nil, err
				}
				state.patternTexture = texture >> 1
			}

			x, y, err := r.getAbsCoords()
			if err != nil {
				return nil, err
			}
			state.drawPattern(x, y)

			for {
				if peek, err := r.bits.Peek8(8); err != nil {
					return nil, err
				} else if peek >= 0xf0 {
					break
				}

				if state.patternCode&0x20 != 0 {
					texture, err := r.bits.Read8(8)
					if err != nil {
						return nil, err
					}
					state.patternTexture = texture >> 1
				}

				x, y, err = r.getRelCoords1(x, y)
				if err != nil {
					return nil, err
				}
				state.drawPattern(x, y)
			}
		case pOpMediumPatterns:
			if state.patternCode&0x20 != 0 {
				texture, err := r.bits.Read8(8)
				if err != nil {
					return nil, err
				}
				state.patternTexture = texture >> 1
			}

			x, y, err := r.getAbsCoords()
			if err != nil {
				return nil, err
			}
			state.drawPattern(x, y)

			for {
				if peek, err := r.bits.Peek8(8); err != nil {
					return nil, err
				} else if peek >= 0xf0 {
					break
				}

				if state.patternCode&0x20 != 0 {
					texture, err := r.bits.Read8(8)
					if err != nil {
						return nil, err
					}
					state.patternTexture = texture >> 1
				}

				x, y, err = r.getRelCoords2(x, y)
				if err != nil {
					return nil, err
				}
				state.drawPattern(x, y)
			}
		case pOpAbsolutePatterns:
			for {
				if peek, err := r.bits.Peek8(8); err != nil {
					return nil, err
				} else if peek >= 0xf0 {
					break
				}

				if state.patternCode&0x20 != 0 {
					texture, err := r.bits.Read8(8)
					if err != nil {
						return nil, err
					}
					state.patternTexture = texture >> 1
				}

				x, y, err := r.getAbsCoords()
				if err != nil {
					return nil, err
				}
				state.drawPattern(x, y)
			}

		case pOpOPX:
			opx, err := r.bits.Read8(8)
			if err != nil {
				return nil, err
			}
			switch pOpxCode(opx) {
			case pOpxUpdatePaletteEntries:
				for {
					if peek, err := r.bits.Peek8(8); err != nil {
						return nil, err
					} else if peek >= 0xf0 {
						break
					}

					index, err := r.bits.Read8(8)
					if err != nil {
						return nil, err
					}

					color, err := r.bits.Read8(8)
					if err != nil {
						return nil, err
					}
					state.palettes[index/40][index%40] = color
				}

			case pOpxSetPalette:
				var pal struct {
					I uint8
					P picPalette
				}
				err := binary.Read(r.bits, binary.LittleEndian, &pal)
				if err != nil {
					return nil, err
				}
				state.palettes[pal.I] = pal.P
			case pOpxCode(0x02):
				var pal struct {
					I uint8
					P picPalette
				}
				err := binary.Read(r.bits, binary.LittleEndian, &pal)
				if err != nil {
					return nil, err
				}
				// TODO this looks like a palette, but not sure
				// what its supposed to be used for...
				//state.palettes[pal.I] = pal.P
			case pOpxCode(0x03), pOpxCode(0x05):
				// not sure what this byte is for...
				r.bits.Skip(8)
			case pOpxCode(0x04), pOpxCode(0x06):
			default:
				return nil, fmt.Errorf("unhandled opx 0x%02x", opx)
			}
		case pOpDone:
			break opLoop
		default:
			return nil, fmt.Errorf("unhandled op 0x%02x", op)
		}

	}
	return state.visual, nil
}

type point struct{ x, y int }

func (p point) isLegal(dst *image.Paletted, legalColor uint8) bool {
	idx := p.y*dst.Stride + p.x
	return dst.Pix[idx] == legalColor
}

func fill(cx, cy int, legalColor uint8, dst *image.Paletted, c1, c2 uint8, dither ditherFn, stack []point) {
	var (
		p      point
		stride = dst.Stride
	)

	// initial
	stack = append(stack, point{cx, cy})

	for len(stack) > 0 {
		p, stack = stack[0], stack[1:]

		var (
			x, y = p.x, p.y
			i    = y*stride + x
		)

		if !p.isLegal(dst, legalColor) {
			continue
		}

		dst.Pix[i] = dither(x, y, c1, c2)

		if down := (point{x, y + 1}); down.y < 190 {
			if down.isLegal(dst, legalColor) {
				stack = append(stack, down)
			}
		}

		if up := (point{x, y - 1}); up.y >= 0 {
			if up.isLegal(dst, legalColor) {
				stack = append(stack, up)
			}
		}

		// flood right
		for dx := x + 1; dx < 320; dx++ {
			var i = y*stride + dx
			if dst.Pix[i] != legalColor {
				break
			}

			dst.Pix[i] = dither(dx, y, c1, c2)
			if down := (point{dx, y + 1}); down.y < 190 {
				if down.isLegal(dst, legalColor) {
					stack = append(stack, down)
				}
			}
			if up := (point{dx, y - 1}); up.y >= 0 {
				if up.isLegal(dst, legalColor) {
					stack = append(stack, up)
				}
			}
		}

		// flood left
		for dx := x - 1; dx >= 0; dx-- {
			var i = y*stride + dx
			if dst.Pix[i] != legalColor {
				break
			}

			dst.Pix[i] = dither(dx, y, c1, c2)
			if down := (point{dx, y + 1}); down.y < 190 {
				if down.isLegal(dst, legalColor) {
					stack = append(stack, down)
				}
			}
			if up := (point{dx, y - 1}); up.y >= 0 {
				if up.isLegal(dst, legalColor) {
					stack = append(stack, up)
				}
			}
		}
	}
}

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

func drawPattern(cx, cy int, size int, isRect, isSolid bool, dst *image.Paletted, c1, c2 uint8, dither ditherFn) {
	if isRect {
		for y := -size; y <= size; y++ {
			if cy+y < 0 || cy+y >= 190 {
				continue
			}

			offset := (cy + y) * dst.Stride
			for x := -size; x <= size+1; x++ {
				if cx+x < 0 || cx+x >= 320 {
					continue
				}
				if isSolid || rand.Float64() < 0.5 {
					dst.Pix[offset+cx+x] = dither(cx+x, y, c1, c2)
				}
			}
		}
	} else {
		r2 := size * size
		for y := -size; y <= size; y++ {
			if cy+y < 0 || cy+y >= 190 {
				continue
			}

			offset := (cy + y) * dst.Stride
			sx := sqrts[r2-y*y]
			for x := -sx; x <= sx; x++ {
				if cx+x < 0 || cx+x >= 320 {
					continue
				}
				if isSolid || rand.Float64() < 0.5 {
					dst.Pix[offset+cx+x] = dither(cx+x, y, c1, c2)
				}
			}
		}
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
func clip(v *int, min, max int) {
	switch {
	case *v < min:
		*v = min
	case *v > max:
		*v = max
	}
}

func line(left, top, right, bottom int, dst *image.Paletted, c1, c2 uint8, dither ditherFn) {
	clip(&left, 0, 319)
	clip(&top, 0, 189)
	clip(&right, 0, 319)
	clip(&bottom, 0, 189)

	switch {
	case left == right:
		swapIf(&top, &bottom, top > bottom)
		for y := top; y <= bottom; y++ {
			dst.Pix[y*dst.Stride+left] = dither(left, y, c1, c2)
		}
	case top == bottom:
		swapIf(&right, &left, right > left)
		for x := right; x <= left; x++ {
			dst.Pix[top*dst.Stride+x] = dither(x, top, c1, c2)
		}
	default:
		// bresenham
		dx, dy := right-left, bottom-top
		stepX, stepY := ((dx>>15)<<1)+1, ((dy>>15)<<1)+1

		dx, dy = absInt(dx)<<1, absInt(dy)<<1

		dst.Pix[top*dst.Stride+left] = dither(left, top, c1, c2)
		dst.Pix[bottom*dst.Stride+right] = dither(right, bottom, c1, c2)

		if dx > dy {
			fraction := dy - (dx >> 1)
			for left != right {
				if fraction >= 0 {
					top += stepY
					fraction -= dx
				}
				left += stepX
				fraction += dy
				dst.Pix[top*dst.Stride+left] = dither(left, top, c1, c2)
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
				dst.Pix[top*dst.Stride+left] = dither(left, top, c1, c2)
			}
		}
	}
}
