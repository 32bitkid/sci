package sci

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/32bitkid/sci/decompression"
	"github.com/32bitkid/sci/resource"
	"io"
	"os"
	"path"
)

type diskMapping struct {
	resourceType resource.Type
	number       resource.Number
	file         uint8
	offset       uint32

	rootPath string

	cache         resource.Resource
	decompressors decompression.LUT
}

func (dr *diskMapping) Type() resource.Type     { return dr.resourceType }
func (dr *diskMapping) Number() resource.Number { return dr.number }

func (dr *diskMapping) Stat() (*resource.Header, error) {
	resourceFn := path.Join(dr.rootPath, fmt.Sprintf("RESOURCE.%03d", dr.file))
	var file *os.File

	file, err := os.Open(resourceFn)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if _, err = file.Seek(int64(dr.offset), io.SeekStart); err != nil {
		return nil, err
	}

	var header resource.Header
	if err = binary.Read(file, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	return &header, nil
}

func (dr *diskMapping) Resource() (resource.Resource, error) {
	if dr.cache != nil {
		return dr.cache, nil
	}

	resourceFn := path.Join(dr.rootPath, fmt.Sprintf("RESOURCE.%03d", dr.file))
	var file *os.File

	file, err := os.Open(resourceFn)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if _, err = file.Seek(int64(dr.offset), io.SeekStart); err != nil {
		return nil, err
	}

	resourceID, payload, err := parsePayloadFrom(file, dr.decompressors)
	if err != nil {
		return nil, err
	}

	dr.cache = &cachedResource{
		id:           resourceID,
		resourceType: dr.resourceType,
		payload:      payload,
	}

	return dr.cache, nil
}

type cachedResource struct {
	id           resource.RID
	resourceType resource.Type
	payload      []uint8
}

func (res cachedResource) ID() resource.RID    { return res.id }
func (res cachedResource) Type() resource.Type { return res.resourceType }
func (res cachedResource) Bytes() []byte       { return res.payload }


func parsePayloadFrom(src io.Reader, lut decompression.LUT) (resource.RID, []byte, error) {
	var header resource.Header
	err := binary.Read(src, binary.LittleEndian, &header)
	if err != nil {
		const invalidRID = resource.RID(0xFFFF)
		return invalidRID, nil, err
	}

	decompressor, ok := lut[header.Method]
	if !ok {
		return header.ID, nil, fmt.Errorf("unhandled compression type: %d", header.Method)
	}

	var buffer bytes.Buffer
	if err := decompressor(src, &buffer, header.CompressedSize, header.DecompressedSize); err != nil {
		return header.ID, nil, err
	}
	return header.ID, buffer.Bytes(), nil
}