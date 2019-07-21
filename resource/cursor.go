package resource

import (
	"bytes"
	"encoding/binary"
)

func NewCursor(b []byte) (cursor Cursor, err error) {
	reader := bytes.NewReader(b)
	err = binary.Read(reader, binary.LittleEndian, &cursor)
	return
}

type HotSpot struct {
	X int16
	Y int16
}

type Cursor struct {
	HotSpot
	Transparency [16]uint16
	Color        [16]uint16
}

func (c Cursor) String() string {
	str := ""
	for i := 0; i < 256; i++ {
		x := uint16(i & 0xF)
		y := uint16(i >> 4)

		tr := (c.Transparency[y]>>x)&1 == 1
		clr := (c.Color[y]>>x)&1 == 1
		switch {
		case tr:
			str += " "
		case clr:
			str += "\u2588"
		default:
			str += "\u2591"
		}

		if x == 15 {
			str += "\n"
		}
	}
	return str
}
