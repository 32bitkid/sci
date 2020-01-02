package screen

import (
	"fmt"
	"math/rand"
	"strings"
)

type TextureBlock [6][5]bool
type Texture [][]TextureBlock
type TextureMap map[uint8]Texture

/* Helpers */
func TextureFromTemplate(s string) (tex Texture) {
	var lines [][]rune
	for _, l := range strings.Split(strings.Trim(s, "\n"), "\n") {
		lines = append(lines, []rune(strings.TrimSpace(l)))
	}

	{
		// Validate template
		if len(lines)%6 != 0 {
			println(len(lines))
			panic("invalid template height")
		}

		var expectedLineLen = -1
		for _, line := range lines {
			ll := len(line)
			if ll%5 != 0 {
				panic(fmt.Errorf("invalid template width: %d is not divisible by 5", ll))
			}
			if expectedLineLen == -1 {
				expectedLineLen = ll
			} else if expectedLineLen != ll {
				panic(fmt.Errorf("invalid template width: %d is abnormal.", ll))
			}
		}
	}

	ch := len(lines) / 6
	cw := len(lines[0]) / 5

	for h := 0; h < ch; h++ {
		var row []TextureBlock
		for w := 0; w < cw; w++ {
			var char TextureBlock
			for y := 0; y < 6; y++ {
				line := lines[h*6+y]
				for x := 0; x < 5; x++ {
					tr := line[w*5+x]
					char[y][x] = tr != '_' &&
						tr != '-' &&
						tr != '0'
				}
			}
			row = append(row, char)
		}
		tex = append(tex, row)
	}

	return tex
}

func RandomTexture(w, h uint, size uint) (tex Texture) {
	w *= size
	h *= size

	for y := uint(0); y < h; y++ {
		var row []TextureBlock
		for x := uint(0); x < w; x++ {
			var char TextureBlock
			row = append(row, char)
		}
		tex = append(tex, row)
	}

	for y1 := uint(0); y1 < h*6; y1 += size {
		for x1 := uint(0); x1 < w*5; x1 += size {
			bit := rand.Float64() < 0.5
			for y2 := uint(0); y2 < size; y2++ {
				for x2 := uint(0); x2 < size; x2++ {
					py := y1 + y2
					px := x1 + x2
					tex[py/6][px/5][py%6][px%5] = bit
				}
			}
		}
	}

	return tex
}

var DefaultTextures = struct {
	Dither1 Texture
	Dither2 Texture
	Dither3 Texture
	Dither4 Texture

	Vertical1 Texture
	Vertical2 Texture
	Vertical3 Texture
	Vertical4 Texture

	Horizontal1 Texture
	Horizontal2 Texture
	Horizontal3 Texture
	Horizontal4 Texture
	Horizontal5 Texture

	Random1 Texture
	Random2 Texture
	Random3 Texture
	Random4 Texture
	Random5 Texture
}{
	Dither1: TextureFromTemplate(`
X_X_X_X_X_
_X_X_X_X_X
X_X_X_X_X_
_X_X_X_X_X
X_X_X_X_X_
_X_X_X_X_X
`),

	Dither2: TextureFromTemplate(`
XX__XX__XX__XX__XX__
XX__XX__XX__XX__XX__
__XX__XX__XX__XX__XX
__XX__XX__XX__XX__XX
XX__XX__XX__XX__XX__
XX__XX__XX__XX__XX__
__XX__XX__XX__XX__XX
__XX__XX__XX__XX__XX
XX__XX__XX__XX__XX__
XX__XX__XX__XX__XX__
__XX__XX__XX__XX__XX
__XX__XX__XX__XX__XX
`),

	Dither3: TextureFromTemplate(`
XXX___XXX___XXX___XXX___XXX___
XXX___XXX___XXX___XXX___XXX___
XXX___XXX___XXX___XXX___XXX___
___XXX___XXX___XXX___XXX___XXX
___XXX___XXX___XXX___XXX___XXX
___XXX___XXX___XXX___XXX___XXX
`),

	Dither4: TextureFromTemplate(`
XXXX____XXXX____XXXX____XXXX____XXXX____
XXXX____XXXX____XXXX____XXXX____XXXX____
XXXX____XXXX____XXXX____XXXX____XXXX____
XXXX____XXXX____XXXX____XXXX____XXXX____
____XXXX____XXXX____XXXX____XXXX____XXXX
____XXXX____XXXX____XXXX____XXXX____XXXX
____XXXX____XXXX____XXXX____XXXX____XXXX
____XXXX____XXXX____XXXX____XXXX____XXXX
XXXX____XXXX____XXXX____XXXX____XXXX____
XXXX____XXXX____XXXX____XXXX____XXXX____
XXXX____XXXX____XXXX____XXXX____XXXX____
XXXX____XXXX____XXXX____XXXX____XXXX____
____XXXX____XXXX____XXXX____XXXX____XXXX
____XXXX____XXXX____XXXX____XXXX____XXXX
____XXXX____XXXX____XXXX____XXXX____XXXX
____XXXX____XXXX____XXXX____XXXX____XXXX
XXXX____XXXX____XXXX____XXXX____XXXX____
XXXX____XXXX____XXXX____XXXX____XXXX____
XXXX____XXXX____XXXX____XXXX____XXXX____
XXXX____XXXX____XXXX____XXXX____XXXX____
____XXXX____XXXX____XXXX____XXXX____XXXX
____XXXX____XXXX____XXXX____XXXX____XXXX
____XXXX____XXXX____XXXX____XXXX____XXXX
____XXXX____XXXX____XXXX____XXXX____XXXX
XXXX____XXXX____XXXX____XXXX____XXXX____
XXXX____XXXX____XXXX____XXXX____XXXX____
XXXX____XXXX____XXXX____XXXX____XXXX____
XXXX____XXXX____XXXX____XXXX____XXXX____
____XXXX____XXXX____XXXX____XXXX____XXXX
____XXXX____XXXX____XXXX____XXXX____XXXX
____XXXX____XXXX____XXXX____XXXX____XXXX
____XXXX____XXXX____XXXX____XXXX____XXXX
XXXX____XXXX____XXXX____XXXX____XXXX____
XXXX____XXXX____XXXX____XXXX____XXXX____
XXXX____XXXX____XXXX____XXXX____XXXX____
XXXX____XXXX____XXXX____XXXX____XXXX____
____XXXX____XXXX____XXXX____XXXX____XXXX
____XXXX____XXXX____XXXX____XXXX____XXXX
____XXXX____XXXX____XXXX____XXXX____XXXX
____XXXX____XXXX____XXXX____XXXX____XXXX
XXXX____XXXX____XXXX____XXXX____XXXX____
XXXX____XXXX____XXXX____XXXX____XXXX____
XXXX____XXXX____XXXX____XXXX____XXXX____
XXXX____XXXX____XXXX____XXXX____XXXX____
____XXXX____XXXX____XXXX____XXXX____XXXX
____XXXX____XXXX____XXXX____XXXX____XXXX
____XXXX____XXXX____XXXX____XXXX____XXXX
____XXXX____XXXX____XXXX____XXXX____XXXX
`),

	Vertical1: TextureFromTemplate(`
X_X_X_X_X_
X_X_X_X_X_
X_X_X_X_X_
X_X_X_X_X_
X_X_X_X_X_
X_X_X_X_X_
`),

	Vertical2: TextureFromTemplate(`
XX__XX__XX__XX__XX__
XX__XX__XX__XX__XX__
XX__XX__XX__XX__XX__
XX__XX__XX__XX__XX__
XX__XX__XX__XX__XX__
XX__XX__XX__XX__XX__
`),

	Vertical3: TextureFromTemplate(`
XXX___XXX___XXX___XXX___XXX___
XXX___XXX___XXX___XXX___XXX___
XXX___XXX___XXX___XXX___XXX___
XXX___XXX___XXX___XXX___XXX___
XXX___XXX___XXX___XXX___XXX___
XXX___XXX___XXX___XXX___XXX___
`),

	Vertical4: TextureFromTemplate(`
XXXX____XXXX____XXXX____XXXX____XXXX____
XXXX____XXXX____XXXX____XXXX____XXXX____
XXXX____XXXX____XXXX____XXXX____XXXX____
XXXX____XXXX____XXXX____XXXX____XXXX____
XXXX____XXXX____XXXX____XXXX____XXXX____
XXXX____XXXX____XXXX____XXXX____XXXX____
`),

	Horizontal1: TextureFromTemplate(`
XXXXX
_____
XXXXX
_____
XXXXX
_____
`),

	Horizontal2: TextureFromTemplate(`
XXXXX
XXXXX
_____
_____
XXXXX
XXXXX
_____
_____
XXXXX
XXXXX
_____
_____
`),

	Horizontal3: TextureFromTemplate(`
XXXXX
XXXXX
XXXXX
_____
_____
_____
`),

	Horizontal4: TextureFromTemplate(`
XXXXX
XXXXX
XXXXX
XXXXX
_____
_____
_____
_____
XXXXX
XXXXX
XXXXX
XXXXX
_____
_____
_____
_____
XXXXX
XXXXX
XXXXX
XXXXX
_____
_____
_____
_____
`),

	Horizontal5: TextureFromTemplate(`
XXXXX
XXXXX
XXXXX
XXXXX
XXXXX
_____
_____
_____
_____
_____
XXXXX
XXXXX
XXXXX
XXXXX
XXXXX
_____
_____
_____
_____
_____
XXXXX
XXXXX
XXXXX
XXXXX
XXXXX
_____
_____
_____
_____
_____
XXXXX
XXXXX
XXXXX
XXXXX
XXXXX
_____
_____
_____
_____
_____
XXXXX
XXXXX
XXXXX
XXXXX
XXXXX
_____
_____
_____
_____
_____
XXXXX
XXXXX
XXXXX
XXXXX
XXXXX
_____
_____
_____
_____
_____
`),

	Random1: RandomTexture(32, 20, 1),
	Random2: RandomTexture(16, 10, 2),
	Random3: RandomTexture(10, 7, 3),
	Random4: RandomTexture(8, 5, 4),
	Random5: RandomTexture(6, 4, 5),
}
