package resource

type Type uint8

const (
	TypeView Type = iota
	TypePic
	TypeScript
	TypeText
	TypeSound
	TypeMemory
	TypeVocab
	TypeFont
	TypeCursor
	TypePatch
)

func (t Type) String() string {
	switch t {
	case TypeView:
		return "Type(View)"
	case TypePic:
		return "Type(Pic)"
	case TypeScript:
		return "Type(Script)"
	case TypeText:
		return "Type(Text)"
	case TypeSound:
		return "Type(Sound)"
	case TypeMemory:
		return "Type(Memory)"
	case TypeVocab:
		return "Type(Vocab)"
	case TypeFont:
		return "Type(Font)"
	case TypeCursor:
		return "Type(Cursor)"
	case TypePatch:
		return "Type(Patch)"
	}
	return "Type(UNKNOWN)"
}
