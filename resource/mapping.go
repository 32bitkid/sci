package resource

import "image"

type Mapping interface {
	Type() Type
	Number() Number

	Resource() (Resource, error)
}

type PictureMapping struct{ Mapping }

func (pic PictureMapping) Render(options ...PicOptions) (*image.Paletted, error) {
	res, err := pic.Resource()
	if err != nil {
		return nil, err
	}
	return NewPic(res.Bytes(), options...)
}
