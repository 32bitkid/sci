package resources

import (
	"bytes"
	"encoding/binary"
	"log"
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

func NewCursor(resource Content) error {
	r := bytes.NewReader(resource)
	var cursor Cursor
	if err := binary.Read(r, binary.LittleEndian, &cursor); err != nil {
		return err
	}
	println(cursor.X, cursor.Y)
	for _, v := range cursor.Transparency {
		log.Printf("%016b", v)
	}
	log.Println("----")
	for _, v := range cursor.Color {
		log.Printf("%016b", v)
	}
	return nil
}
