package resource

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"image"

	"github.com/32bitkid/bitreader"
	"github.com/32bitkid/sci/screen"
)

type Pic struct {
	Visual   *image.Paletted
	Priority *image.Paletted
	Control  *image.Paletted
}

func NewPic(b []byte, options ...PicOptions) (Pic, error) {
	var debugFn DebugCallback
	ditherer := screen.EGADitherer
	for _, opts := range options {
		if opts.Ditherer != nil {
			ditherer = opts.Ditherer
		}

		if opts.DebugFn != nil {
			debugFn = opts.DebugFn
		}
	}

	return readPic(b, ditherer, debugFn)
}

type PicOptions struct {
	*screen.Ditherer
	DebugFn DebugCallback
}

type DebugCallback func(*PicState)

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
// The total PayloadBytes is 16-bits long:
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
// The total PayloadBytes is 8-bits long:
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

const (
	picDrawVisual   picDrawMode = 1
	picDrawPriority             = 2
	picDrawControl              = 4
)

func (mode *picDrawMode) Set(flag picDrawMode, enabled bool) {
	if enabled {
		*mode |= flag
	} else {
		*mode &= ^flag
	}
}

func (mode picDrawMode) Has(flag picDrawMode) bool {
	return mode&flag == flag
}

type PicState struct {
	Pic

	palettes [4]picPalette
	drawMode picDrawMode

	color        uint8
	priorityCode uint8
	controlCode  uint8

	patternCode    uint8
	patternTexture uint8

	ditherer  *screen.Ditherer
	debugFn   DebugCallback
	fillStack []point
}

func (s *PicState) debugger() {
	if s.debugFn != nil {
		s.debugFn(s)
	}
}

func (s *PicState) fill(cx, cy int) {
	switch {
	case s.drawMode.Has(picDrawVisual):
		if s.color == 255 {
			// FIXME this fill occurs but it doesn't make any sense.
			//  It's asking for a solid white fill, but that should be a noop if legalColor is always 15.
			return
		}
		c1, c2 := s.ditherer.Get(s.color)
		fill(cx, cy, 0xf, s.Visual, c1, c2, dither5050, s.fillStack)
		s.debugger()
	case s.drawMode.Has(picDrawPriority):
		if s.priorityCode == 0 {
			return
		}
		c := s.priorityCode
		fill(cx, cy, 0x0, s.Priority, c, c, noDither, s.fillStack)
	case s.drawMode.Has(picDrawControl):
		if s.controlCode == 0 {
			return
		}
		c := s.controlCode
		fill(cx, cy, 0x0, s.Control, c, c, noDither, s.fillStack)
	default:
		return
	}
}

func (s *PicState) line(x1, y1, x2, y2 int) {
	if s.drawMode.Has(picDrawVisual) {
		c1, c2 := s.ditherer.Get(s.color)
		line(x1, y1, x2, y2, s.Visual, c1, c2, dither5050)
		s.debugger()
	}
	if s.drawMode.Has(picDrawPriority) {
		c := s.priorityCode
		line(x1, y1, x2, y2, s.Priority, c, c, noDither)
	}
	if s.drawMode.Has(picDrawControl) {
		c := s.controlCode
		line(x1, y1, x2, y2, s.Control, c, c, noDither)
	}
}

func (s *PicState) drawPattern(cx, cy int) {
	size := int(s.patternCode & 0x7)
	isRect := s.patternCode&0x10 != 0
	solid := s.patternCode&0x20 == 0

	var filler FillTextureFn = solidFillTexture
	if !solid {
		filler = newSierraFillTexture(s.patternTexture)
	}

	if s.drawMode.Has(picDrawVisual) {
		c1, c2 := s.ditherer.Get(s.color)
		drawPattern(cx, cy, size, isRect, filler, s.Visual, c1, c2, dither5050)
		s.debugger()
	}
	if s.drawMode.Has(picDrawPriority) {
		c := s.priorityCode
		drawPattern(cx, cy, size, isRect, filler, s.Priority, c, c, noDither)
	}
	if s.drawMode.Has(picDrawControl) {
		c := s.controlCode
		drawPattern(cx, cy, size, isRect, filler, s.Control, c, c, noDither)
	}
}

func readPic(
	payload []byte,
	ditherer *screen.Ditherer,
	debugFn DebugCallback,
) (Pic, error) {
	r := picReader{
		bitreader.NewReader(bufio.NewReader(bytes.NewReader(payload))),
	}

	bounds := image.Rect(0, 0, 320, 190)

	var state = PicState{
		Pic: Pic{
			Visual:   image.NewPaletted(bounds, ditherer.Palette),
			Priority: image.NewPaletted(bounds, screen.Depth16Palette),
			Control:  image.NewPaletted(bounds, screen.EGAPalette),
		},
		drawMode: picDrawVisual | picDrawPriority,
		palettes: [4]picPalette{
			defaultPalette,
			defaultPalette,
			defaultPalette,
			defaultPalette,
		},
		ditherer:  ditherer,
		debugFn:   debugFn,
		fillStack: make([]point, 0, 320*190),
	}

	for i := 0; i < (320 * 190); i++ {
		state.Visual.Pix[i] = 0xF
		state.Priority.Pix[i] = 0x0
		state.Control.Pix[i] = 0x0
	}

opLoop:
	for {
		op, err := r.bits.Read8(8)
		if err != nil {
			return Pic{}, err
		}

		switch pOpCode(op) {
		case pOpSetColor:
			code, err := r.bits.Read8(8)
			if err != nil {
				return Pic{}, err
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
				return Pic{}, err
			}
			state.priorityCode = code & 0xF
			state.drawMode.Set(picDrawPriority, true)
		case pOpDisablePriority:
			state.drawMode.Set(picDrawPriority, false)

		case pOpSetControl:
			code, err := r.bits.Read8(8)
			if err != nil {
				return Pic{}, err
			}
			state.controlCode = code & 0xf
			state.drawMode.Set(picDrawControl, true)
		case pOpDisableControl:
			state.drawMode.Set(picDrawControl, false)

		// Lines
		case pOpShortLines:
			x1, y1, err := r.getAbsCoords()
			if err != nil {
				return Pic{}, err
			}
			for {
				if peek, err := r.bits.Peek8(8); err != nil {
					return Pic{}, err
				} else if peek >= 0xf0 {
					break
				}

				x2, y2, err := r.getRelCoords1(x1, y1)
				if err != nil {
					return Pic{}, err
				}

				state.line(x1, y1, x2, y2)
				x1, y1 = x2, y2
			}
		case pOpMediumLines:
			x1, y1, err := r.getAbsCoords()
			if err != nil {
				return Pic{}, err
			}
			for {
				if peek, err := r.bits.Peek8(8); err != nil {
					return Pic{}, err
				} else if peek >= 0xf0 {
					break
				}

				x2, y2, err := r.getRelCoords2(x1, y1)
				if err != nil {
					return Pic{}, err
				}

				state.line(x1, y1, x2, y2)
				x1, y1 = x2, y2
			}
		case pOpLongLines:
			x1, y1, err := r.getAbsCoords()
			if err != nil {
				return Pic{}, err
			}
			for {
				if peek, err := r.bits.Peek8(8); err != nil {
					return Pic{}, err
				} else if peek >= 0xf0 {
					break
				}

				x2, y2, err := r.getAbsCoords()
				if err != nil {
					return Pic{}, err
				}

				state.line(x1, y1, x2, y2)
				x1, y1 = x2, y2
			}

		// Fills
		case pOpFill:
			for {
				if peek, err := r.bits.Peek8(8); err != nil {
					return Pic{}, err
				} else if peek >= 0xf0 {
					break
				}

				x, y, err := r.getAbsCoords()
				if err != nil {
					return Pic{}, err
				}

				state.fill(x, y)
			}

		// Patterns
		case pOpSetPattern:
			code, err := r.bits.Read8(8)
			if err != nil {
				return Pic{}, err
			}
			state.patternCode = code & 0x3f
		case pOpShortPatterns:
			if state.patternCode&0x20 != 0 {
				texture, err := r.bits.Read8(8)
				if err != nil {
					return Pic{}, err
				}
				state.patternTexture = texture >> 1
			}

			x, y, err := r.getAbsCoords()
			if err != nil {
				return Pic{}, err
			}
			state.drawPattern(x, y)

			for {
				if peek, err := r.bits.Peek8(8); err != nil {
					return Pic{}, err
				} else if peek >= 0xf0 {
					break
				}

				if state.patternCode&0x20 != 0 {
					texture, err := r.bits.Read8(8)
					if err != nil {
						return Pic{}, err
					}
					state.patternTexture = texture >> 1
				}

				x, y, err = r.getRelCoords1(x, y)
				if err != nil {
					return Pic{}, err
				}
				state.drawPattern(x, y)
			}
		case pOpMediumPatterns:
			if state.patternCode&0x20 != 0 {
				texture, err := r.bits.Read8(8)
				if err != nil {
					return Pic{}, err
				}
				state.patternTexture = texture >> 1
			}

			x, y, err := r.getAbsCoords()
			if err != nil {
				return Pic{}, err
			}
			state.drawPattern(x, y)

			for {
				if peek, err := r.bits.Peek8(8); err != nil {
					return Pic{}, err
				} else if peek >= 0xf0 {
					break
				}

				if state.patternCode&0x20 != 0 {
					texture, err := r.bits.Read8(8)
					if err != nil {
						return Pic{}, err
					}
					state.patternTexture = texture >> 1
				}

				x, y, err = r.getRelCoords2(x, y)
				if err != nil {
					return Pic{}, err
				}
				state.drawPattern(x, y)
			}
		case pOpAbsolutePatterns:
			for {
				if peek, err := r.bits.Peek8(8); err != nil {
					return Pic{}, err
				} else if peek >= 0xf0 {
					break
				}

				if state.patternCode&0x20 != 0 {
					texture, err := r.bits.Read8(8)
					if err != nil {
						return Pic{}, err
					}
					state.patternTexture = texture >> 1
				}

				x, y, err := r.getAbsCoords()
				if err != nil {
					return Pic{}, err
				}
				state.drawPattern(x, y)
			}

		// Extensions
		case pOpOPX:
			opx, err := r.bits.Read8(8)
			if err != nil {
				return Pic{}, err
			}
			switch pOpxCode(opx) {
			case pOpxUpdatePaletteEntries:
				for {
					if peek, err := r.bits.Peek8(8); err != nil {
						return Pic{}, err
					} else if peek >= 0xf0 {
						break
					}

					index, err := r.bits.Read8(8)
					if err != nil {
						return Pic{}, err
					}

					color, err := r.bits.Read8(8)
					if err != nil {
						return Pic{}, err
					}
					state.palettes[index/40][index%40] = color
				}

			case pOpxSetPalette:
				i, err := r.bits.Read8(8)
				if err != nil {
					return Pic{}, err
				}
				err = binary.Read(r.bits, binary.LittleEndian, &state.palettes[i])
				if err != nil {
					return Pic{}, err
				}
			case pOpxCode(0x02):
				// TODO this looks like a palette, but not sure what its supposed to be used for...
				var pal struct {
					I uint8
					P picPalette
				}
				err := binary.Read(r.bits, binary.LittleEndian, &pal)
				if err != nil {
					return Pic{}, err
				}
				// state.palettes[pal.I] = pal.P
			case pOpxCode(0x03), pOpxCode(0x05):
				// TODO not sure what this byte is for...
				if err := r.bits.Skip(8); err != nil {
					return Pic{}, err
				}
			case pOpxCode(0x04), pOpxCode(0x06):
				// TODO not sure what this OP is for, but it appears to have no payload
			case pOpxCode(0x07):
				// TODO not sure what this is for (QfG2 uses this op-code)
				// Vector?
				if err := r.bits.Skip(24); err != nil {
					return Pic{}, err
				}

				var length uint16
				if err := binary.Read(r.bits, binary.LittleEndian, &length); err != nil {
					return Pic{}, err
				}

				// Payload?
				for i := uint(0); i < uint(length); i++ {
					if err := r.bits.Skip(8); err != nil {
						return Pic{}, err
					}
				}

			case pOpxCode(0x08):
				// TODO not sure what this is for (KQ1-sci0 remake uses this op-code)
				for {
					if peek, err := r.bits.Peek8(8); err != nil {
						return Pic{}, err
					} else if peek >= 0xf0 {
						break
					}
					if err := r.bits.Skip(8); err != nil {
						return Pic{}, err
					}
				}
			default:
				return Pic{}, fmt.Errorf("unhandled OPX 0x%02x", opx)
			}

		case pOpDone:
			break opLoop
		default:
			return Pic{}, fmt.Errorf("unhandled OP 0x%02x", op)
		}

	}

	return state.Pic, nil
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

func drawPattern(cx, cy int, size int, isRect bool, filler FillTextureFn, dst *image.Paletted, c1, c2 uint8, dither ditherFn) {
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
				if filler() {
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
				if filler() {
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
