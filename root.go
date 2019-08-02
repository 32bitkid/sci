// Package sci implements access to SCI0-based Sierra On-Line game
// assets and vm.
//
// The Sierra Creative Interpreter version 0 (SCI0) was Sierra On-Line's
// second generation game engine, succeeding the Adventure Game
// Interpreter (AGI). It implemented several upgrades over AGI, namely full
// EGA (320x200x16) support, improved audio/music support.
//
// The SCI0 was succeeded by SCI1, SCI2, and SCI32 engines that extended support
// for even more sound cards, VGA and SVGA graphics, and FVM.

package sci

import (
	"encoding/binary"
	"github.com/32bitkid/sci/resource"
	"os"
	"path"
)

// Root is reference to the root path of a SCI0 game.
type Root struct {
	Decompressors resource.DecompressorLUT
	Path          string
	Mapping       []resource.Mapping
}

func NewSCI0Root(path string) Root {
	return Root{
		Path:          path,
		Decompressors: resource.Decompressors.SCI0,
	}
}

func NewSCI01Root(path string) Root {
	return Root{
		Path:          path,
		Decompressors: resource.Decompressors.SCI01,
	}
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
		case resource.TypeCursor:
			root.Mapping = append(root.Mapping, resource.CursorMapping{Mapping: mapping})
		case resource.TypeFont:
			root.Mapping = append(root.Mapping, resource.FontMapping{Mapping: mapping})
		default:
			root.Mapping = append(root.Mapping, mapping)
		}
	}

	return nil
}
