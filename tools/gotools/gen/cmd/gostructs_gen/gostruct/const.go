package gostruct

const (
	FieldKindScalar FieldKind = iota
	FieldKindMessage
	FieldKindEnum
	FieldKindMap
	FieldKindRepeated
)
