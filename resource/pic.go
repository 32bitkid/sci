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

func NewPic(b []byte, options ...PicOptions) (screen.Pic, error) {
	var debugFn DebugCallback
	var scaler screen.Scaler = &screen.Scaler1x1{
		VisualDitherer: screen.DefaultDitherers.EGA,
	}
	for _, opts := range options {
		if opts.Scaler != nil {
			scaler = opts.Scaler
		}

		if opts.DebugFn != nil {
			debugFn = opts.DebugFn
		}
	}

	return readPic(b, scaler, debugFn)
}

type PicOptions struct {
	screen.Scaler
	DebugFn DebugCallback
}

type DebugCallback func(*PicState)

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

func (mode *picDrawMode) set(flag picDrawMode, enabled bool) {
	if enabled {
		*mode |= flag
	} else {
		*mode &= ^flag
	}
}

func (mode picDrawMode) has(flag picDrawMode) bool {
	return mode&flag == flag
}

type PicState struct {
	screen.Pic

	palettes [4]picPalette
	drawMode picDrawMode

	color        uint8
	priorityCode uint8
	controlCode  uint8

	patternCode    uint8
	patternTexture uint8

	debugFn DebugCallback
}

func (s *PicState) debugger() {
	if s.debugFn != nil {
		s.debugFn(s)
	}
}

func (s *PicState) fill(cx, cy int) {
	switch {
	case s.drawMode.has(picDrawVisual):
		if s.color == 255 {
			// FIXME this fill occurs but it doesn't make any sense.
			//  It's asking for a solid white fill, but that should be a noop if legalColor is always 15.
			return
		}
		s.Visual.Fill(cx, cy, 0xf, s.color)
		s.debugger()
	case s.drawMode.has(picDrawPriority):
		if s.priorityCode == 0 {
			return
		}
		s.Priority.Fill(cx, cy, 0x0, s.priorityCode)
	case s.drawMode.has(picDrawControl):
		if s.controlCode == 0 {
			return
		}
		s.Priority.Fill(cx, cy, 0x0, s.controlCode)
	default:
		return
	}
}

func (s *PicState) line(x1, y1, x2, y2 int) {
	if s.drawMode.has(picDrawVisual) {
		s.Visual.Line(x1, y1, x2, y2, s.color)
		s.debugger()
	}
	if s.drawMode.has(picDrawPriority) {
		c := s.priorityCode
		s.Priority.Line(x1, y1, x2, y2, c)
	}
	if s.drawMode.has(picDrawControl) {
		c := s.controlCode
		s.Control.Line(x1, y1, x2, y2, c)
	}
}

func (s *PicState) drawPattern(cx, cy int) {
	size := int(s.patternCode & 0x7)
	isRect := s.patternCode&0x10 != 0
	isSolid := s.patternCode&0x20 == 0

	if s.drawMode.has(picDrawVisual) {
		s.Visual.Pattern(cx, cy, size, isRect, isSolid, s.patternTexture, s.color)
		s.debugger()
	}
	if s.drawMode.has(picDrawPriority) {
		c := s.priorityCode
		s.Priority.Pattern(cx, cy, size, isRect, isSolid, s.patternTexture, c)
	}
	if s.drawMode.has(picDrawControl) {
		c := s.controlCode
		s.Control.Pattern(cx, cy, size, isRect, isSolid, s.patternTexture, c)
	}
}

func readPic(
	payload []byte,
	scaler screen.Scaler,
	debugFn DebugCallback,
) (screen.Pic, error) {
	r := picReader{
		bitreader.NewReader(bufio.NewReader(bytes.NewReader(payload))),
	}

	bounds := image.Rect(0, 0, 320, 190)

	var state = PicState{
		Pic:      scaler.NewPic(bounds),
		drawMode: picDrawVisual | picDrawPriority,
		palettes: [4]picPalette{
			defaultPalette,
			defaultPalette,
			defaultPalette,
			defaultPalette,
		},
		debugFn: debugFn,
	}

	state.Visual.Clear(0xF)
	state.Control.Clear(0x0)
	state.Priority.Clear(0x0)

opLoop:
	for {
		op, err := r.bits.Read8(8)
		if err != nil {
			return screen.Pic{}, err
		}

		switch pOpCode(op) {
		case pOpSetColor:
			code, err := r.bits.Read8(8)
			if err != nil {
				return screen.Pic{}, err
			}

			pal := code / 40
			index := code % 40
			state.color = state.palettes[pal][index]
			state.drawMode.set(picDrawVisual, true)
		case pOpDisableVisual:
			state.drawMode.set(picDrawVisual, false)

		case pOpSetPriority:
			code, err := r.bits.Read8(8)
			if err != nil {
				return screen.Pic{}, err
			}
			state.priorityCode = code & 0xF
			state.drawMode.set(picDrawPriority, true)
		case pOpDisablePriority:
			state.drawMode.set(picDrawPriority, false)

		case pOpSetControl:
			code, err := r.bits.Read8(8)
			if err != nil {
				return screen.Pic{}, err
			}
			state.controlCode = code & 0xf
			state.drawMode.set(picDrawControl, true)
		case pOpDisableControl:
			state.drawMode.set(picDrawControl, false)

		// Lines
		case pOpShortLines:
			x1, y1, err := r.getAbsCoords()
			if err != nil {
				return screen.Pic{}, err
			}
			for {
				if peek, err := r.bits.Peek8(8); err != nil {
					return screen.Pic{}, err
				} else if peek >= 0xf0 {
					break
				}

				x2, y2, err := r.getRelCoords1(x1, y1)
				if err != nil {
					return screen.Pic{}, err
				}

				state.line(x1, y1, x2, y2)
				x1, y1 = x2, y2
			}
		case pOpMediumLines:
			x1, y1, err := r.getAbsCoords()
			if err != nil {
				return screen.Pic{}, err
			}
			for {
				if peek, err := r.bits.Peek8(8); err != nil {
					return screen.Pic{}, err
				} else if peek >= 0xf0 {
					break
				}

				x2, y2, err := r.getRelCoords2(x1, y1)
				if err != nil {
					return screen.Pic{}, err
				}

				state.line(x1, y1, x2, y2)
				x1, y1 = x2, y2
			}
		case pOpLongLines:
			x1, y1, err := r.getAbsCoords()
			if err != nil {
				return screen.Pic{}, err
			}
			for {
				if peek, err := r.bits.Peek8(8); err != nil {
					return screen.Pic{}, err
				} else if peek >= 0xf0 {
					break
				}

				x2, y2, err := r.getAbsCoords()
				if err != nil {
					return screen.Pic{}, err
				}

				state.line(x1, y1, x2, y2)
				x1, y1 = x2, y2
			}

		// Fills
		case pOpFill:
			for {
				if peek, err := r.bits.Peek8(8); err != nil {
					return screen.Pic{}, err
				} else if peek >= 0xf0 {
					break
				}

				x, y, err := r.getAbsCoords()
				if err != nil {
					return screen.Pic{}, err
				}

				state.fill(x, y)
			}

		// Patterns
		case pOpSetPattern:
			code, err := r.bits.Read8(8)
			if err != nil {
				return screen.Pic{}, err
			}
			state.patternCode = code & 0x3f
		case pOpShortPatterns:
			if state.patternCode&0x20 != 0 {
				texture, err := r.bits.Read8(8)
				if err != nil {
					return screen.Pic{}, err
				}
				state.patternTexture = texture >> 1
			}

			x, y, err := r.getAbsCoords()
			if err != nil {
				return screen.Pic{}, err
			}
			state.drawPattern(x, y)

			for {
				if peek, err := r.bits.Peek8(8); err != nil {
					return screen.Pic{}, err
				} else if peek >= 0xf0 {
					break
				}

				if state.patternCode&0x20 != 0 {
					texture, err := r.bits.Read8(8)
					if err != nil {
						return screen.Pic{}, err
					}
					state.patternTexture = texture >> 1
				}

				x, y, err = r.getRelCoords1(x, y)
				if err != nil {
					return screen.Pic{}, err
				}
				state.drawPattern(x, y)
			}
		case pOpMediumPatterns:
			if state.patternCode&0x20 != 0 {
				texture, err := r.bits.Read8(8)
				if err != nil {
					return screen.Pic{}, err
				}
				state.patternTexture = texture >> 1
			}

			x, y, err := r.getAbsCoords()
			if err != nil {
				return screen.Pic{}, err
			}
			state.drawPattern(x, y)

			for {
				if peek, err := r.bits.Peek8(8); err != nil {
					return screen.Pic{}, err
				} else if peek >= 0xf0 {
					break
				}

				if state.patternCode&0x20 != 0 {
					texture, err := r.bits.Read8(8)
					if err != nil {
						return screen.Pic{}, err
					}
					state.patternTexture = texture >> 1
				}

				x, y, err = r.getRelCoords2(x, y)
				if err != nil {
					return screen.Pic{}, err
				}
				state.drawPattern(x, y)
			}
		case pOpAbsolutePatterns:
			for {
				if peek, err := r.bits.Peek8(8); err != nil {
					return screen.Pic{}, err
				} else if peek >= 0xf0 {
					break
				}

				if state.patternCode&0x20 != 0 {
					texture, err := r.bits.Read8(8)
					if err != nil {
						return screen.Pic{}, err
					}
					state.patternTexture = texture >> 1
				}

				x, y, err := r.getAbsCoords()
				if err != nil {
					return screen.Pic{}, err
				}
				state.drawPattern(x, y)
			}

		// Extensions
		case pOpOPX:
			opx, err := r.bits.Read8(8)
			if err != nil {
				return screen.Pic{}, err
			}
			switch pOpxCode(opx) {
			case pOpxUpdatePaletteEntries:
				for {
					if peek, err := r.bits.Peek8(8); err != nil {
						return screen.Pic{}, err
					} else if peek >= 0xf0 {
						break
					}

					index, err := r.bits.Read8(8)
					if err != nil {
						return screen.Pic{}, err
					}

					color, err := r.bits.Read8(8)
					if err != nil {
						return screen.Pic{}, err
					}
					state.palettes[index/40][index%40] = color
				}

			case pOpxSetPalette:
				i, err := r.bits.Read8(8)
				if err != nil {
					return screen.Pic{}, err
				}
				err = binary.Read(r.bits, binary.LittleEndian, &state.palettes[i])
				if err != nil {
					return screen.Pic{}, err
				}
			case pOpxCode(0x02):
				// TODO this looks like a palette, but not sure what its supposed to be used for...
				var pal struct {
					I uint8
					P picPalette
				}
				err := binary.Read(r.bits, binary.LittleEndian, &pal)
				if err != nil {
					return screen.Pic{}, err
				}
				// state.palettes[pal.I] = pal.P
			case pOpxCode(0x03), pOpxCode(0x05):
				// TODO not sure what this byte is for...
				if err := r.bits.Skip(8); err != nil {
					return screen.Pic{}, err
				}
			case pOpxCode(0x04), pOpxCode(0x06):
				// TODO not sure what this OP is for, but it appears to have no payload
			case pOpxCode(0x07):
				// TODO not sure what this is for (QfG2 uses this op-code)
				// Vector?
				if err := r.bits.Skip(24); err != nil {
					return screen.Pic{}, err
				}

				var length uint16
				if err := binary.Read(r.bits, binary.LittleEndian, &length); err != nil {
					return screen.Pic{}, err
				}

				// Payload?
				for i := uint(0); i < uint(length); i++ {
					if err := r.bits.Skip(8); err != nil {
						return screen.Pic{}, err
					}
				}

			case pOpxCode(0x08):
				// TODO not sure what this is for (KQ1-sci0 remake uses this op-code)
				for {
					if peek, err := r.bits.Peek8(8); err != nil {
						return screen.Pic{}, err
					} else if peek >= 0xf0 {
						break
					}
					if err := r.bits.Skip(8); err != nil {
						return screen.Pic{}, err
					}
				}
			default:
				return screen.Pic{}, fmt.Errorf("unhandled OPX 0x%02x", opx)
			}

		case pOpDone:
			break opLoop
		default:
			return screen.Pic{}, fmt.Errorf("unhandled OP 0x%02x", op)
		}

	}

	return state.Pic, nil
}
