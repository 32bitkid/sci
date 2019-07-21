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

type ViewMapping struct{ Mapping }

func (view ViewMapping) Render() (View, error) {
	res, err := view.Resource()
	if err != nil {
		return nil, err
	}
	return NewView(res.Bytes())
}

type TextMapping struct{ Mapping }

func (view TextMapping) GetText() (Text, error) {
	res, err := view.Resource()
	if err != nil {
		return nil, err
	}
	return NewText(res.Bytes())
}

type FontMapping struct{ Mapping }

func (view FontMapping) GetFont() (*Font, error) {
	res, err := view.Resource()
	if err != nil {
		return nil, err
	}
	return NewFont(res.Bytes())
}

type CursorMapping struct{ Mapping }

func (view CursorMapping) GetCursor() (Cursor, error) {
	res, err := view.Resource()
	if err != nil {
		return Cursor{}, err
	}
	return NewCursor(res.Bytes())
}
