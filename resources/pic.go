package resources

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/32bitkid/bitreader"
	"image"
	"math"
	"math/rand"
	"time"
)

type ditherFn func(a, b uint8) uint8

func create5050Dither() ditherFn {
	state := false
	return func(a, b uint8) (val uint8) {
		val = a
		if state {
			val = b
		}
		state = !state
		return
	}
}

func createNoiseDither() ditherFn {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return func(a, b uint8) uint8 {
		if r.Float64() < 0.5 {
			return b
		}
		return a
	}
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

type pOpxCode uint8

const (
	pOpxSetPalette pOpxCode = 0x01
)

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
	col1     uint8
	col2     uint8
	palettes [4]picPalette
	drawMode picDrawMode

	priorityCode uint8
	controlCode  uint8

	patternCode    uint8
	patternTexture uint8

	visual   *image.Paletted
	priority *image.Paletted
	control  *image.Paletted
	aux      *image.Paletted
}

func ReadPic(resource *Resource) (image.Image, error) {
	r := picReader{
		bitreader.NewReader(bufio.NewReader(bytes.NewReader(resource.bytes))),
	}

	var state = picState{
		visual:   image.NewPaletted(image.Rect(0, 0, 320, 200), egaPalette),
		priority: image.NewPaletted(image.Rect(0, 0, 320, 200), gray16Palette),
		control:  image.NewPaletted(image.Rect(0, 0, 320, 200), egaPalette),
		aux:      image.NewPaletted(image.Rect(0, 0, 320, 200), egaPalette),
		drawMode: picDrawVisual | picDrawPriority,
		palettes: [...]picPalette{
			defaultPalette,
			defaultPalette,
			defaultPalette,
			defaultPalette,
		},
	}

	for i := 0; i < (320 * 200); i++ {
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
			state.col1 = (color >> 4) & 0x0F
			state.col2 = (color >> 0) & 0x0F
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
			state.drawMode.Set(picDrawControl, true)
			state.controlCode = code & 0xf

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

				x, y, err := r.getRelCoords1(x, y)
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

				x, y, err := r.getRelCoords2(x, y)
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
			case pOpxCode(0x00):
				for {
					if peek, err := r.bits.Peek8(8); err != nil {
						return nil, err
					} else if peek >= 0xf0 {
						break
					}

					code, err := r.bits.Read16(16)
					if err != nil {
						return nil, err
					}
					index := code >> 8
					color := uint8(code & 0xff)
					state.palettes[index/40][color%40] = color
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

func (s *picState) fill(x, y int) {
	fmt.Printf("filling at (%d, %d)\n", x, y)
}

func (s *picState) line(x1, y1, x2, y2 int) {
	if s.drawMode.Has(picDrawVisual) {
		line(s.visual, x1, y1, x2, y2, s.col1, s.col2)
	}
	if s.drawMode.Has(picDrawPriority) {
		line(s.priority, x1, y1, x2, y2, s.priorityCode, s.priorityCode)
	}
	if s.drawMode.Has(picDrawControl) {
		line(s.control, x1, y1, x2, y2, s.controlCode, s.controlCode)
	}
}

func (s *picState) drawPattern(cx, cy int) {
	size := int(s.patternCode & 0x7)
	isRect := s.patternCode&0x10 != 0
	solid := s.patternCode&0x20 == 0
	dst := s.visual
	drawPattern(dst, cx, cy, size, s.col1, s.col2, isRect, solid)
}

// (0..256) => i => int(math.Round(math.Sqrt(float64(i) + 0.5)))
var sqrt = [256]int{
	0x01, 0x01, 0x02, 0x02, 0x02, 0x02, 0x03, 0x03,
	0x03, 0x03, 0x03, 0x03, 0x04, 0x04, 0x04, 0x04,
	0x04, 0x04, 0x04, 0x04, 0x05, 0x05, 0x05, 0x05,
	0x05, 0x05, 0x05, 0x05, 0x05, 0x05, 0x06, 0x06,
	0x06, 0x06, 0x06, 0x06, 0x06, 0x06, 0x06, 0x06,
	0x06, 0x06, 0x07, 0x07, 0x07, 0x07, 0x07, 0x07,
	0x07, 0x07, 0x07, 0x07, 0x07, 0x07, 0x07, 0x07,
	0x08, 0x08, 0x08, 0x08, 0x08, 0x08, 0x08, 0x08,
	0x08, 0x08, 0x08, 0x08, 0x08, 0x08, 0x08, 0x08,
	0x09, 0x09, 0x09, 0x09, 0x09, 0x09, 0x09, 0x09,
	0x09, 0x09, 0x09, 0x09, 0x09, 0x09, 0x09, 0x09,
	0x09, 0x09, 0x0a, 0x0a, 0x0a, 0x0a, 0x0a, 0x0a,
	0x0a, 0x0a, 0x0a, 0x0a, 0x0a, 0x0a, 0x0a, 0x0a,
	0x0a, 0x0a, 0x0a, 0x0a, 0x0a, 0x0a, 0x0b, 0x0b,
	0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b,
	0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b,
	0x0b, 0x0b, 0x0b, 0x0b, 0x0c, 0x0c, 0x0c, 0x0c,
	0x0c, 0x0c, 0x0c, 0x0c, 0x0c, 0x0c, 0x0c, 0x0c,
	0x0c, 0x0c, 0x0c, 0x0c, 0x0c, 0x0c, 0x0c, 0x0c,
	0x0c, 0x0c, 0x0c, 0x0c, 0x0d, 0x0d, 0x0d, 0x0d,
	0x0d, 0x0d, 0x0d, 0x0d, 0x0d, 0x0d, 0x0d, 0x0d,
	0x0d, 0x0d, 0x0d, 0x0d, 0x0d, 0x0d, 0x0d, 0x0d,
	0x0d, 0x0d, 0x0d, 0x0d, 0x0d, 0x0d, 0x0e, 0x0e,
	0x0e, 0x0e, 0x0e, 0x0e, 0x0e, 0x0e, 0x0e, 0x0e,
	0x0e, 0x0e, 0x0e, 0x0e, 0x0e, 0x0e, 0x0e, 0x0e,
	0x0e, 0x0e, 0x0e, 0x0e, 0x0e, 0x0e, 0x0e, 0x0e,
	0x0e, 0x0e, 0x0f, 0x0f, 0x0f, 0x0f, 0x0f, 0x0f,
	0x0f, 0x0f, 0x0f, 0x0f, 0x0f, 0x0f, 0x0f, 0x0f,
	0x0f, 0x0f, 0x0f, 0x0f, 0x0f, 0x0f, 0x0f, 0x0f,
	0x0f, 0x0f, 0x0f, 0x0f, 0x0f, 0x0f, 0x0f, 0x0f,
	0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10,
	0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10,
}

func drawPattern(dst *image.Paletted, cx, cy int, size int, col1, col2 uint8, isRect, isSolid bool) {
	dither := create5050Dither()
	if !isSolid {
		dither = createNoiseDither()
	}

	if isRect {
		for y := -size; y <= size; y++ {
			for x := -size; x <= size; x++ {
				col := dither(col1, col2)
				dst.Set(cx+x, cy+y, dst.Palette[col])
			}
		}
	} else {
		r2 := size * size
		for y := -size; y <= size; y++ {
			sx := sqrt[r2-y*y]
			for x := -sx; x <= sx; x++ {
				col := dither(col1, col2)
				dst.Set(cx+x, cy+y, dst.Palette[col])
			}
		}
	}
}

func line(dst *image.Paletted, x1, y1, x2, y2 int, col1, col2 uint8) {
	dx := x2 - x1
	dy := y2 - y1

	dither := create5050Dither()

	switch {
	case dx == 0 && dy == 0:
		dst.Set(x1, y1, dst.Palette[col1])
	case dx == 0:
		i0, i1 := y1, y2
		if i0 > i1 {
			i0, i1 = i1, i0
		}
		for i := i0; i < i1; i++ {
			col := dither(col1, col2)
			dst.Set(x1, i, dst.Palette[col])
		}
	case dy == 0:
		i0, i1 := x1, x2
		if i0 > i1 {
			i0, i1 = i1, i0
		}
		for i := i0; i < i1; i++ {
			col := dither(col1, col2)
			dst.Set(i, y1, dst.Palette[col])
		}
	default:
		// bresenham
		dErr := math.Abs(float64(dy) / float64(dx))
		err := float64(0)
		xDir := ((dx >> 63) << 1) + 1
		yDir := ((dy >> 63) << 1) + 1

		for x, y := x1, y1; x != x2; x += xDir {
			col := dither(col1, col2)
			dst.Set(x, y, dst.Palette[col])
			err += dErr
			if err >= 0.5 {
				y += yDir
				err -= 1
			}
		}
		// last pixel
		dst.Set(x2, y2, dst.Palette[dither(col1, col2)])
	}
}
