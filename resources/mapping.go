package resources

import (
	"bufio"
	"compress/lzw"
	"encoding/binary"
	"fmt"
	"io"
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
	return "Type(INVALID)"
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

func (m *Mapping) GetResourceFile() string {
	return fmt.Sprintf("RESOURCE.%03d", m.File)
}


func (m Mapping) LoadFrom(file io.ReadSeeker) (*Resource, error) {
	src := bufio.NewReader(file)

	if _, err := file.Seek(int64(m.Offset), io.SeekStart); err != nil {
		return nil, err
	}

	var header Header
	err := binary.Read(src, binary.LittleEndian, &header)
	if err != nil {
		return nil, err
	}

	switch header.CompressionMethod {
	case CompressionNone:
		buffer := make([]uint8, header.DecompressedSize)
		if _, err := io.ReadFull(src, buffer); err != nil {
			return nil, err
		}
		return &Resource{
			m,
			header,
			buffer,
		}, nil
	case CompressionLZW:
		buffer := make([]uint8, header.DecompressedSize)
		r := lzw.NewReader(src, lzw.LSB, 8)
		if _, err := io.ReadFull(r, buffer); err != nil {
			return nil, err
		}
		return &Resource{
			m,
			header,
			buffer,
		}, nil
	case CompressionHuffman:
		buffer := make([]uint8, header.DecompressedSize)
		if err := huffman(src, buffer); err != nil {
			return nil, err
		}
		return &Resource{
			m,
			header,
			buffer,
		}, nil
	}

	return nil, fmt.Errorf("cannot read goo %d", header.CompressionMethod)
}

type CompressionMethod uint16

const (
	CompressionNone    CompressionMethod = 0
	CompressionLZW                       = 1
	CompressionHuffman                   = 2
)

type Header struct {
	ID               uint16
	CompressedSize   uint16
	DecompressedSize uint16
	CompressionMethod
}

type Resource struct {
	details Mapping
	header  Header
	bytes   []uint8
}

const idEndToken uint16 = (1 << 16) - 1
const tailEndToken uint32 = (1 << 32) - 1

func ParseResourceMap(r io.Reader) ([]Mapping, error) {
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
