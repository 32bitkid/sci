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

	cache resource.Resource
}

func (dr diskMapping) Type() resource.Type     { return dr.resourceType }
func (dr diskMapping) Number() resource.Number { return dr.number }

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

	resourceID, payload, err := resource.ParsePayloadFrom(file)
	if err != nil {
		return nil, err
	}

	dr.cache = &loadedResource{
		id:           resourceID,
		resourceType: dr.resourceType,
		bytes:        payload,
	}

	return dr.cache, nil
}

type loadedResource struct {
	id           resource.RID
	resourceType resource.Type
	bytes        []uint8
}

func (res loadedResource) ID() resource.RID    { return res.id }
func (res loadedResource) Type() resource.Type { return res.resourceType }
func (res loadedResource) Bytes() []byte       { return res.bytes }
