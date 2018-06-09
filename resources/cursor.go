package resources

import (
	"encoding/binary"
	"io"
)

type Point struct {
	X int16
	Y int16
}

type Cursor struct {
	Point
	Transparency [16]uint16
	Color        [16]uint16
}

func (c Cursor) String() string {
	str := ""
	for i := 0; i < 256; i++ {
		x := uint16(i & 0xF)
		y := uint16(i >> 4)

		tr := (c.Transparency[y] >> x) & 1 == 1
		clr := (c.Color[y] >> x) & 1 == 1
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

func ReadCursor(reader io.Reader) (cursor Cursor, err error) {
	err = binary.Read(reader, binary.LittleEndian, &cursor)
	return
}
