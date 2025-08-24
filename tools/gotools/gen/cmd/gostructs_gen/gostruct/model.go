package gostruct

// GoField 表示Go结构体字段信息
type GoField struct {
	Name      string
	ProtoName string
	Type      string
	ProtoType string
	Number    int32
	Kind      FieldKind
	Comment   string

	IsOptional bool
	IsOneof    bool

	MapInfo      *MapInfo
	RepeatedInfo *RepeatedInfo

	OneofName  string
	OneofIndex int32

	Tags *FieldTags
}

func (f *GoField) IsMapType() bool {
	return f.Kind == FieldKindMap
}

func (f *GoField) IsRepeatedType() bool {
	return f.Kind == FieldKindRepeated
}

func (f *GoField) IsPointerType() bool {
	return f.Kind == FieldKindMessage
}

// FieldTags 字段标签配置
type FieldTags struct {
	JSON     *JSONTag
	BSON     *BsonTag
	Protobuf *ProtobufTag
	Custom   []CustomTag
}

// JSONTag JSON标签配置
type JSONTag struct {
	Name      string
	OmitEmpty bool
	Ignore    bool
}

// BsonTag BSON 标签配置
type BsonTag struct {
	Name      string
	OmitEmpty bool
	Ignore    bool
}

// ProtobufTag Protobuf标签配置
type ProtobufTag struct {
	Type   string
	Number int32
	Name   string
	Rule   string
}

// CustomTag 自定义标签
type CustomTag struct {
	Key   string
	Value string
}

type MapInfo struct {
	KeyType   string    // 键的类型
	KeyKind   FieldKind // 键的Kind
	ValueType string    // 值的类型
	ValueKind FieldKind // 值的Kind
}

// Map 的值是不是 Message，如果是Message，则一定是结构体指针
func (m *MapInfo) IsPointerValueType() bool {
	return m.ValueKind == FieldKindMessage
}

type RepeatedInfo struct {
	ValueType string    // 值的类型
	ValueKind FieldKind // 值的Kind
}

func (r *RepeatedInfo) IsPointerValueType() bool {
	return r.ValueKind == FieldKindMessage
}
