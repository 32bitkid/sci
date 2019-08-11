package sci

import (
	"fmt"
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
	decompressors resource.DecompressorLUT
}

func (dr *diskMapping) Type() resource.Type     { return dr.resourceType }
func (dr *diskMapping) Number() resource.Number { return dr.number }

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

	resourceID, payload, err := resource.ParsePayloadFrom(file, dr.decompressors)
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
