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
}

const idEndToken uint16 = (1 << 16) - 1
const tailEndToken uint32 = (1 << 32) - 1

// LoadMap parses the mapping file that exists in the Root folder.
func (res *Root) LoadMapping() error {
	r, err := os.Open(path.Join(res.Path, "RESOURCE.MAP"))
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

		baseRes := &diskMapping{
			resourceType: resource.Type(id >> 11),
			number:       resource.Number(id & ((1 << 11) - 1)),
			file:         uint8(tail >> 26),
			offset:       uint32(tail & ((1 << 26) - 1)),

			rootPath: res.Path,
		}

		switch baseRes.resourceType {
		case resource.TypePic:
			res.Mapping = append(res.Mapping, resource.PictureMapping{baseRes})
		default:
			res.Mapping = append(res.Mapping, baseRes)
		}
	}

	return nil
}
