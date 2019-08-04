package screen

import (
	"image"
)

type Pic struct {
	Visual   Buffer
	Priority Buffer
	Control  Buffer
}

type Scaler interface {
	NewPic(bounds image.Rectangle) Pic
}
