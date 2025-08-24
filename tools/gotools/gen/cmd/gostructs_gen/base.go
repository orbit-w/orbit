package cmd

// ProtoField 表示 proto message 中的一个字段
type ProtoField struct {
	Name string
	Type string
	Tags map[string]string // 支持多种tag类型，如 "sync":"asset" "json":"omitempty"
}

// ProtoMessage 表示一个 proto message 的信息
type ProtoMessage struct {
	Name      string
	Fields    []ProtoField
	UsesMap   bool
	UsesList  bool
	MapFields []MapFieldInfo
}

// MapFieldInfo 包含 DeltaSyncMap 字段的详细信息
type MapFieldInfo struct {
	Name     string
	SyncTags []string          // sync标签的值列表，用于数据类型
	AllTags  map[string]string // 所有标签
}

// MapKeyTypeInfo 表示 MapKeyType 枚举值信息
type MapKeyTypeInfo struct {
	EnumValue string // 如 "MapKeyType_KEY_STRING"
	TypeName  string // 如 "StringKeyMap"
	MapType   string // 如 "string"
}

// getGenericKeyField 根据键类型获取 GenericKey 中对应的字段名
func getGenericKeyField(keyType string) string {
	switch keyType {
	case "string":
		return "StringKey"
	case "int32":
		return "Int32Key"
	case "int64":
		return "IntKey"
	case "uint32":
		return "Uint32Key"
	case "uint64":
		return "Uint64Key"
	default:
		return "StringKey"
	}
}

// getFullMapGetter 根据 TypeName 获取 FullMap 中对应的 getter 方法名
func getFullMapGetter(typeName string) string {
	switch typeName {
	case "StringKeyMap":
		return "GetStringKeyMap"
	case "Int32KeyMap":
		return "GetInt32KeyMap"
	case "Int64KeyMap":
		return "GetInt64KeyMap"
	case "Uint32KeyMap":
		return "GetUint32KeyMap"
	case "Uint64KeyMap":
		return "GetUint64KeyMap"
	default:
		return "GetStringKeyMap"
	}
}
