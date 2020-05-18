package resource

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
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

			sprite.Pixels = bitmap
			sprite.Mirrored = mirrored
			group = append(group, sprite)
		}
		view = append(view, group)
	}

	return view, nil
}

type View []SpriteGroup

type SpriteGroup []Sprite

type SpriteHeader struct {
	Width    uint16
	Height   uint16
	X        int8
	Y        int8
	KeyColor uint8
}

type Sprite struct {
	SpriteHeader
	Mirrored bool
	Pixels   []uint8
}

func (group SpriteGroup) GIF(d *screen.Ditherer) *gif.GIF {
	palette := make(color.Palette, 256)
	for i := 0; i < 256; i++ {
		palette[i] = color.Gray{}
	}
	copy(palette, d.Palette)

	palette[0xff] = color.RGBA{0xdd, 0xdd, 0xdd, 0xff}

	var images []*image.Paletted
	var delays []int
	var dispose []byte

	// FIXME this bounds checking isn't correct
	rect := image.Rectangle{}
	for _, s := range group {
		if int(s.X) < rect.Min.X {
			rect.Min.X = int(s.X)
		}

		if int(s.Y) < rect.Min.Y {
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
	rect = rect.Sub(offset)
	rect.Max.X *= 5
	rect.Max.Y *= 6

	for _, s := range group {
		img := image.NewPaletted(rect, palette)
		for i := range img.Pix {
			img.Pix[i] = 0xff
		}

		stride := int(s.Width)
		ox := int(s.X) - offset.X
		oy := int(s.Y) - offset.Y

		for y := 0; y < int(s.Height); y++ {
			for x := 0; x < stride; x += 1 {
				ax := ox + x
				if s.Mirrored {
					ax = stride - 1 - x
				}
				srcPix := s.Pixels[x+(y*stride)]
				if srcPix == s.KeyColor {
					continue
				}
				for dy := (oy + y) * 6; dy < ((oy+y)*6)+6; dy++ {
					for dx := ax * 5; dx < (ax*5)+5; dx++ {
						img.Pix[dx+(dy*img.Stride)] = srcPix
					}
				}

			}
		}

		images = append(images, img)
		delays = append(delays, 20)
		dispose = append(dispose, gif.DisposalBackground)
	}

	return &gif.GIF{
		Image:           images,
		Delay:           delays,
		Disposal:        dispose,
		BackgroundIndex: 255,
	}
}
