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
	ResourceTypeView RType = iota
	ResourceTypePic
	ResourceTypeScript
	ResourceTypeText
	ResourceTypeSound
	ResourceTypeMemory
	ResourceTypeVocab
	ResourceTypeFont
	ResourceTypeCursor
	ResourceTypePatch
)

func (t RType) String() string {
	switch t {
	case ResourceTypeView:
		return "Type(View)"
	case ResourceTypePic:
		return "Type(Pic)"
	case ResourceTypeScript:
		return "Type(Script)"
	case ResourceTypeText:
		return "Type(Text)"
	case ResourceTypeSound:
		return "Type(Sound)"
	case ResourceTypeMemory:
		return "Type(Memory)"
	case ResourceTypeVocab:
		return "Type(Vocab)"
	case ResourceTypeFont:
		return "Type(Font)"
	case ResourceTypeCursor:
		return "Type(Cursor)"
	case ResourceTypePatch:
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

	id   uint16
	root string
}

type Header struct {
	Id                uint16
	CompressedSize    uint16
	DecompressedSize  uint16
	CompressionMethod uint16
}

type Content []byte

func (res Mapping) Load() (Content, error) {
	filename := path.Join(res.root, fmt.Sprintf("RESOURCE.%03d", res.File))
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

	buffer := make(Content, header.DecompressedSize)
	switch header.CompressionMethod {
	case 0:
		if _, err := io.ReadFull(src, buffer); err != nil {
			return nil, err
		}
	case 1:
		r := lzw.NewReader(src, 0, 8)
		defer r.Close()
		if _, err := io.ReadFull(r, buffer); err != nil {
			return nil, err
		}
	case 2:
		_, err := src.ReadByte()
		if err != nil {
			return nil, err
		}

		rawT, err := src.ReadByte()
		if err != nil {
			return nil, err
		}

		_ = int16(rawT) | 0x100

		return nil, fmt.Errorf("unsupported decompression: %d", header.CompressionMethod)
	default:
		return nil, fmt.Errorf("unsupported decompression: %d", header.CompressionMethod)
	}

	return buffer, nil
}

type ResourceMap []Mapping

const idEndToken uint16 = (1 << 16) - 1
const tailEndToken uint32 = (1 << 32) - 1

func ParseSCI0(root string) (ResourceMap, error) {
	r, err := os.Open(path.Join(root, "RESOURCE.MAP"))
	defer r.Close()
	if err != nil {
		return nil, err
	}

	var resources ResourceMap
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
			id:     id,
			root:   root,
		})
	}

	return resources, nil
}
