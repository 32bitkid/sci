package resource

import (
	"github.com/32bitkid/sci/decompression"
)

type RID uint16

type Resource interface {
	ID() RID
	Type() Type
	Bytes() []uint8
}

type Header struct {
	ID               RID
	CompressedSize   uint16
	DecompressedSize uint16
	decompression.Method
}
