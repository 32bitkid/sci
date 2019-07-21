package resource

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
)

type Number uint16
type RID uint16

type Resource interface {
	ID() RID
	Type() Type
	Bytes() []uint8
}

type resourceHeader struct {
	ID               RID
	CompressedSize   uint16
	DecompressedSize uint16
	CompressionMethod
}

const InvalidRID = RID(0xFFFF)

func ParsePayloadFrom(r io.Reader, decomp DecompressorLUT) (RID, []byte, error) {
	src := bufio.NewReader(r)

	var header resourceHeader
	err := binary.Read(src, binary.LittleEndian, &header)
	if err != nil {
		return 0xFFFF, nil, err
	}

	decompressor, ok := decomp[header.CompressionMethod]
	if !ok {
		return InvalidRID, nil, fmt.Errorf("unhandled compression type: %d", header.CompressionMethod)
	}

	buffer := make([]uint8, header.DecompressedSize)
	if err := decompressor(src, buffer, header.CompressedSize, header.DecompressedSize); err != nil {
		return InvalidRID, nil, err
	}
	return header.ID, buffer, nil
}
