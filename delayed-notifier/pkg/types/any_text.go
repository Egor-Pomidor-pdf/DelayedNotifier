package types

type AnyText string

func NewAnyText(text string) AnyText {
	return AnyText(text)
}

func (d AnyText) String() string {
	return string(d)
}

