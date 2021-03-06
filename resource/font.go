package resource

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
)

func NewFont(b []byte) (*Font, error) {
	r := bytes.NewReader(b)

	type fontHeader struct {
		_          [2]uint8
		Characters uint16
		LineHeight uint16
	}
	var h fontHeader
	if err := binary.Read(r, binary.LittleEndian, &h); err != nil {
		return nil, err
	}

	type characterHeader struct {
		Width  uint8
		Height uint8
	}

	font := Font{
		Count:      h.Characters,
		LineHeight: h.LineHeight,
		Characters: make([]Character, int(h.Characters)),
	}

	pointers := make([]uint16, h.Characters)
	if err := binary.Read(r, binary.LittleEndian, &pointers); err != nil {
		return nil, err
	}

	for i, offset := range pointers {
		if _, err := r.Seek(int64(offset), 0); err != nil {
			panic(err)
		}

		var ch characterHeader
		if err := binary.Read(r, binary.LittleEndian, &ch); err != nil {
			return nil, err
		}

		bitmapLength := ((ch.Width + 7) >> 3) * ch.Height
		bitmap := make([]uint8, bitmapLength)
		if err := binary.Read(r, binary.LittleEndian, &bitmap); err != nil {
			return nil, err
		}

		font.Characters[i] = Character{
			Width:   ch.Width,
			Height:  ch.Height,
			Bitmaps: bitmap,
		}
	}

	return &font, nil
}

type Font struct {
	Count      uint16
	LineHeight uint16
	Characters []Character
}

type Character struct {
	Width   uint8
	Height  uint8
	Bitmaps []byte
}

func (c Character) String() string {
	result := ""
	bpr := int((c.Width + 7) >> 3)
	for y := 0; y < int(c.Height); y++ {
		line := ""
		for x := 0; x < int(bpr); x++ {
			d := c.Bitmaps[y*bpr+x]
			line += fmt.Sprintf("%08b", d)
		}
		result += line[0:c.Width] + "\n"
	}

	return strings.Replace(strings.Replace(result, "0", " ", -1), "1", "\u2588", -1)
}
