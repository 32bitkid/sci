package resource

import (
	"bufio"
	"compress/lzw"
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

func ParsePayloadFrom(r io.Reader) (RID, []byte, error) {
	src := bufio.NewReader(r)

	var header resourceHeader
	err := binary.Read(src, binary.LittleEndian, &header)
	if err != nil {
		return 0xFFFF, nil, err
	}

	switch header.CompressionMethod {
	case CompressionNone:
		buffer := make([]uint8, header.DecompressedSize)
		if _, err := io.ReadFull(src, buffer); err != nil {
			return 0xFFFF, nil, err
		}
		return header.ID, buffer, nil
	case CompressionLZW:
		buffer := make([]uint8, header.DecompressedSize)
		lr := io.LimitReader(src, int64(header.CompressedSize))
		r := lzw.NewReader(lr, lzw.LSB, 8)
		if _, err := io.ReadFull(r, buffer); err != nil {
			return 0xFFFF, nil, err
		}
		return header.ID, buffer, nil
	case CompressionHuffman:
		buffer := make([]uint8, header.DecompressedSize)
		lr := io.LimitReader(src, int64(header.CompressedSize))
		if err := huffman(lr, buffer); err != nil {
			return 0xFFFF, nil, err
		}
		return header.ID, buffer, nil
	}

	return 0xFFFF, nil, fmt.Errorf("unhandled compression type: %d", header.CompressionMethod)
}
