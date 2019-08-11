package resource

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/draw"
	"image/gif"

	"github.com/32bitkid/sci/screen"
)

func NewView(b []byte) (View, error) {
	r := bytes.NewReader(b)

	var header struct {
		Groups   uint16
		Mirrored uint16
		_        uint32
	}

	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	view := make(View, 0, header.Groups)

	groupPointers := make([]uint16, header.Groups)
	if err := binary.Read(r, binary.LittleEndian, &groupPointers); err != nil {
		return nil, err
	}

	for g, groupPointer := range groupPointers {
		if _, err := r.Seek(int64(groupPointer), 0); err != nil {
			return nil, err
		}

		var groupHeader struct {
			Images uint16
			_      [2]byte
		}

		if err := binary.Read(r, binary.LittleEndian, &groupHeader); err != nil {
			return nil, err
		}

		mirrored := header.Mirrored&(1<<uint(g)) == (1 << uint(g))

		group := SpriteGroup{}

		spritePointers := make([]uint16, groupHeader.Images)
		if err := binary.Read(r, binary.LittleEndian, &spritePointers); err != nil {
			return nil, err
		}

		for _, spritePointer := range spritePointers {
			if _, err := r.Seek(int64(spritePointer), 0); err != nil {
				return nil, err
			}

			sprite := Sprite{}

			if err := binary.Read(r, binary.LittleEndian, &sprite.SpriteHeader); err != nil {
				return nil, err
			}

			total := int(sprite.Width * sprite.Height)
			bitmap := make([]uint8, total)
			i := 0
			for i < total {
				var b uint8
				err := binary.Read(r, binary.LittleEndian, &b)
				if err != nil {
					return nil, err
				}
				color := b & 0xF
				repeat := int(b >> 4)
				for r := 0; r < repeat; r++ {
					bitmap[i] = color
					i++
				}
			}

			if mirrored {
				sprite.X = -sprite.X
				stride := sprite.Width
				hStride := stride >> 1
				for y := uint16(0); y < sprite.Height; y++ {
					offs := y * stride
					for x := uint16(0); x < (hStride); x++ {
						a := offs + x
						b := offs + stride - x - 1
						bitmap[a], bitmap[b] = bitmap[b], bitmap[a]
					}
				}
			}

			sprite.Pixels = bitmap
			group = append(group, sprite)
		}
		view = append(view, group)
	}

	return view, nil
}

type View []SpriteGroup

type SpriteGroup []Sprite

func (g SpriteGroup) GIF(palette *screen.Ditherer) *gif.GIF {
	var images []*image.Paletted
	var delays []int
	var dispose []byte

	// FIXME this bounds checking isn't correct
	rect := image.Rectangle{}
	for _, s := range g {
		if rect.Min.X > int(s.X) {
			rect.Min.X = int(s.X)
		}

		if rect.Min.Y > int(s.Y) {
			rect.Min.Y = int(s.Y)
		}

		if rect.Max.X < (int(s.Width) + int(s.X)) {
			rect.Max.X = int(s.Width) + int(s.X)
		}
		if rect.Max.Y < (int(s.Height) + int(s.Y)) {
			rect.Max.Y = int(s.Height) + int(s.Y)
		}
	}

	offset := rect.Min
	rect = rect.Sub(rect.Min)

	for _, s := range g {
		srcRect := image.Rect(
			0, 0,
			int(s.Width), int(s.Height),
		)

		source := &image.Paletted{
			Pix:     s.Pixels,
			Stride:  int(s.Width),
			Rect:    srcRect,
			Palette: palette.Palette,
		}

		mask := image.NewAlpha(srcRect)
		for i := range mask.Pix {
			if s.Pixels[i] != s.KeyColor {
				mask.Pix[i] = 0xff
			}
		}

		img := image.NewPaletted(rect, palette.Palette)
		for i := range img.Pix {
			img.Pix[i] = 0x0
		}

		// TODO actually handle transparency
		draw.DrawMask(
			img,
			image.Rect(
				int(s.X)-offset.X,
				int(s.Y)-offset.Y,
				int(s.Width)+int(s.X)-offset.X,
				int(s.Height)+int(s.Y)-offset.Y,
			),
			source,
			image.ZP,
			mask,
			image.ZP,
			draw.Src,
		)

		images = append(images, img)
		delays = append(delays, 20)
		dispose = append(dispose, gif.DisposalPrevious)
	}

	return &gif.GIF{
		Image:    images,
		Delay:    delays,
		Disposal: dispose,
	}
}

type SpriteHeader struct {
	Width    uint16
	Height   uint16
	X        int8
	Y        uint8
	KeyColor uint8
}

type Sprite struct {
	SpriteHeader
	Pixels []uint8
}
