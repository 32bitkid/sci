package resources

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/32bitkid/bitreader"
	"image"
)

type picReader struct {
	bits bitreader.BitReader
}

// format: XXXXYYYYxxxxxxxxyyyyyyyy => XXXXxxxxxxxx, YYYYyyyyyyyy
func (p picReader) getAbsCoords() (int, int, error) {
	code, err := p.bits.Read32(24)
	if err != nil {
		return 0, 0, err
	}
	x := ((code & 0xF00000) >> 12) | ((code & 0xFF00) >> 8)
	y := ((code & 0x0F0000) >> 8) | ((code & 0x00FF) >> 0)
	return int(x), int(y), nil
}

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

func (mode *picDrawMode) Set(flag picDrawMode) {
	*mode |= flag
}

func (mode *picDrawMode) Disable(flag picDrawMode) {
	*mode &= ^flag
}

const (
	picDrawVisual   picDrawMode = 1
	picDrawPriority             = 2
	picDrawControl              = 4
)

func ReadPic(resource *Resource) (image.Image, error) {
	r := picReader{
		bitreader.NewReader(bufio.NewReader(bytes.NewReader(resource.bytes))),
	}

	buffer := image.NewPaletted(image.Rect(0, 0, 320, 200), egaPalette)

	var state = struct {
		col1           uint8
		col2           uint8
		palettes       [4]picPalette
		drawMode       picDrawMode
		priority       uint8
		patternCode    uint8
		patternTexture uint8
		control        uint8
	}{
		drawMode: picDrawVisual | picDrawPriority,
		palettes: [...]picPalette{
			defaultPalette,
			defaultPalette,
			defaultPalette,
			defaultPalette,
		},
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
			state.drawMode.Set(picDrawVisual)
		case pOpDisableVisual:
			state.drawMode.Disable(picDrawVisual)

		case pOpSetPriority:
			code, err := r.bits.Read8(8)
			if err != nil {
				return nil, err
			}
			state.priority = code & 0xF
			state.drawMode.Set(picDrawPriority)
		case pOpDisablePriority:
			state.drawMode.Disable(picDrawPriority)

		case pOpSetControl:
			code, err := r.bits.Read8(8)
			if err != nil {
				return nil, err
			}
			state.control = code & 0xf
		case pOpDisableControl:
			state.drawMode.Disable(picDrawControl)

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

				line(
					buffer,
					x1, y1, x2, y2,
					state.col1, state.col2,
				)

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

				line(
					buffer,
					x1, y1, x2, y2,
					state.col1, state.col2,
				)

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

				line(buffer, x1, y1, x2, y2, state.col1, state.col2)

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

				fill(buffer, x, y, state.col1, state.col2)
			}

		case pOpSetPattern:
			code, err := r.bits.Read8(8)
			if err != nil {
				return nil, err
			}
			state.patternCode = code
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
			drawPattern(
				buffer,
				x, y,
				state.col1, state.col2,
			)

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
				drawPattern(
					buffer,
					x, y,
					state.col1, state.col2,
				)
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
			drawPattern(
				buffer,
				x, y,
				state.col1, state.col2,
			)

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
				drawPattern(
					buffer,
					x, y,
					state.col1, state.col2,
				)
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
				drawPattern(
					buffer,
					x, y,
					state.col1, state.col2,
				)
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
	return buffer, nil
}

func fill(buf *image.Paletted, x, y int, col1, col2 uint8) {
	//fmt.Printf("filling at (%d, %d)\n", x, y)
}

func line(buf *image.Paletted, x1, y1, x2, y2 int, col1, col2 uint8) {
	//fmt.Printf("drawing a line (%d,%d)-(%d,%d)\n", x1, y1, x2, y2)
}

func drawPattern(buf *image.Paletted, x1, y1 int, col1, col2 uint8) {
	//fmt.Printf("drawing a pattern at (%d,%d)\n", x1, y1)
}
