package screen

import (
	"image"
)

type Pic interface {
	Visual() Buffer
	Priority() Buffer
	Control() Buffer
}

type picLayers struct {
	visual   Buffer
	priority Buffer
	control  Buffer
}

func (s picLayers) Visual() Buffer {
	return s.visual
}

func (s picLayers) Priority() Buffer {
	return s.priority
}

func (s picLayers) Control() Buffer {
	return s.control
}

type Scaler interface {
	NewPic(bounds image.Rectangle) Pic
}

type DitherFn func(x, y int, c1, c2 uint8) uint8

var noDither = func(x, y int, c1, _ uint8) uint8 { return c1 }
var dither5050 DitherFn = func(x, y int, c1, c2 uint8) uint8 {
	if (x&1)^(y&1) == 0 {
		return c1
	}
	return c2
}
