package sci

import (
	"encoding/binary"
	"github.com/32bitkid/sci/resource"
	"os"
	"path"
)

// Root is reference to the root path of a SCI0 game.
type Root struct {
	Path    string
	Mapping []resource.Mapping

	Decompressors resource.DecompressorLUT
}

const idEndToken uint16 = (1 << 16) - 1
const tailEndToken uint32 = (1 << 32) - 1

// LoadMap parses the mapping file that exists in the Root folder.
func (root *Root) LoadMapping() error {
	r, err := os.Open(path.Join(root.Path, "RESOURCE.MAP"))
	if err != nil {
		return err
	}
	defer r.Close()

	for {
		var id uint16
		if err := binary.Read(r, binary.LittleEndian, &id); err != nil {
			return err
		}
		var tail uint32
		if err := binary.Read(r, binary.LittleEndian, &tail); err != nil {
			return err
		}

		if id == idEndToken && tail == tailEndToken {
			break
		}

		// Default to using SCI0 decompressors
		decompressors := root.Decompressors
		if decompressors == nil {
			decompressors = resource.Decompressors.SCI0
		}

		mapping := &diskMapping{
			resourceType: resource.Type(id >> 11),
			number:       resource.Number(id & ((1 << 11) - 1)),
			file:         uint8(tail >> 26),
			offset:       uint32(tail & ((1 << 26) - 1)),

			rootPath:      root.Path,
			decompressors: decompressors,
		}

		switch mapping.resourceType {
		case resource.TypePic:
			root.Mapping = append(root.Mapping, resource.PictureMapping{Mapping: mapping})
		case resource.TypeView:
			root.Mapping = append(root.Mapping, resource.ViewMapping{Mapping: mapping})
		case resource.TypeText:
			root.Mapping = append(root.Mapping, resource.TextMapping{Mapping: mapping})
		case resource.TypeFont:
			root.Mapping = append(root.Mapping, resource.FontMapping{Mapping: mapping})
		default:
			root.Mapping = append(root.Mapping, mapping)
		}
	}

	return nil
}
