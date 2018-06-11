package resources

import (
	"bufio"
	"compress/lzw"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path"
)

type RType uint8

const (
	TypeView RType = iota
	TypePic
	TypeScript
	TypeText
	TypeSound
	TypeMemory
	TypeVocab
	TypeFont
	TypeCursor
	TypePatch
)

func (t RType) String() string {
	switch t {
	case TypeView:
		return "Type(View)"
	case TypePic:
		return "Type(Pic)"
	case TypeScript:
		return "Type(Script)"
	case TypeText:
		return "Type(Text)"
	case TypeSound:
		return "Type(Sound)"
	case TypeMemory:
		return "Type(Memory)"
	case TypeVocab:
		return "Type(Vocab)"
	case TypeFont:
		return "Type(Font)"
	case TypeCursor:
		return "Type(Cursor)"
	case TypePatch:
		return "Type(Patch)"
	}
	return "Type(Invalid)"
}

type RNumber uint16
type RFile uint8
type ROffset uint32

type Mapping struct {
	Type   RType
	Number RNumber
	File   RFile
	Offset ROffset
}

type CompressionMethod uint16

const (
	CompressionNone    CompressionMethod = 0
	CompressionLZW                       = 1
	CompressionHuffman                   = 2
)

type Header struct {
	Id               uint16
	CompressedSize   uint16
	DecompressedSize uint16
	CompressionMethod
}

type Resource struct {
	details Mapping
	header  Header
	bytes   []uint8
}

func (res Mapping) Load(root string) (*Resource, error) {
	filename := path.Join(root, fmt.Sprintf("RESOURCE.%03d", res.File))
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	src := bufio.NewReader(file)

	if _, err := file.Seek(int64(res.Offset), 0); err != nil {
		return nil, err
	}

	var header Header
	binary.Read(src, binary.LittleEndian, &header)

	switch header.CompressionMethod {
	case CompressionNone:
		buffer := make([]uint8, header.DecompressedSize)
		if _, err := io.ReadFull(src, buffer); err != nil {
			return nil, err
		}
		return &Resource{
			res,
			header,
			buffer,
		}, nil
	case CompressionLZW:
		buffer := make([]uint8, header.DecompressedSize)
		r := lzw.NewReader(src, 0, 8)
		defer r.Close()
		if _, err := io.ReadFull(r, buffer); err != nil {
			return nil, err
		}
		return &Resource{
			res,
			header,
			buffer,
		}, nil
	case CompressionHuffman:
		return nil, fmt.Errorf("unsupported decompression: %d", header.CompressionMethod)
	}

	return nil, fmt.Errorf("cannot read goo %d", header.CompressionMethod)
}

func ParseSCI0(root string) ([]Mapping, error) {
	const idEndToken uint16 = (1 << 16) - 1
	const tailEndToken uint32 = (1 << 32) - 1

	r, err := os.Open(path.Join(root, "RESOURCE.MAP"))
	defer r.Close()
	if err != nil {
		return nil, err
	}

	var resources []Mapping
	for {
		var id uint16
		if err := binary.Read(r, binary.LittleEndian, &id); err != nil {
			return nil, err
		}
		var tail uint32
		if err := binary.Read(r, binary.LittleEndian, &tail); err != nil {
			return nil, err
		}

		if id == idEndToken && tail == tailEndToken {
			break
		}

		resources = append(resources, Mapping{
			Type:   RType(id >> 11),
			Number: RNumber(id & ((1 << 11) - 1)),
			File:   RFile(tail >> 26),
			Offset: ROffset(tail & ((1 << 26) - 1)),
		})
	}

	return resources, nil
}
