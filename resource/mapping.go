package resource

import "github.com/32bitkid/sci/screen"

type Mapping interface {
	Type() Type
	Number() Number

	Resource() (Resource, error)
}

type PictureMapping struct{ Mapping }

func (pic PictureMapping) Render(options ...PicOptions) (screen.Pic, error) {
	res, err := pic.Resource()
	if err != nil {
		return screen.Pic{}, err
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

func (view TextMapping) Text() (Text, error) {
	res, err := view.Resource()
	if err != nil {
		return nil, err
	}
	return NewText(res.Bytes())
}

type FontMapping struct{ Mapping }

func (view FontMapping) Font() (*Font, error) {
	res, err := view.Resource()
	if err != nil {
		return nil, err
	}
	return NewFont(res.Bytes())
}

type CursorMapping struct{ Mapping }

func (view CursorMapping) Cursor() (Cursor, error) {
	res, err := view.Resource()
	if err != nil {
		return Cursor{}, err
	}
	return NewCursor(res.Bytes())
}
