package blueprint_types

// 目前的类型需要跟proto中的类型一致
const (
	FieldKindUnknown   FieldKind = 0
	FieldKindDouble    FieldKind = 1  // double
	FieldKindFloat     FieldKind = 2  // float
	FieldKindInt64     FieldKind = 3  // int64
	FieldKindUInt64    FieldKind = 4  // uint64
	FieldKindInt32     FieldKind = 5  // int32
	FieldKindFixed64   FieldKind = 6  // fixed64
	FieldKindFixed32   FieldKind = 7  // fixed32
	FieldKindBool      FieldKind = 8  // bool
	FieldKindString    FieldKind = 9  // string
	FieldKindMessage   FieldKind = 11 // message
	FieldKindBytes     FieldKind = 12 // bytes
	FieldKindUInt32    FieldKind = 13 // uint32
	FieldKindEnum      FieldKind = 14 // enum
	FieldKindMMEObject FieldKind = 16 // MME对象类型
	FieldKindMap       FieldKind = 18 // map
	FieldKindXMap      FieldKind = 19 // xmap
	FieldKindRepeated  FieldKind = 20 // repeated
)

// FieldKindString 字段类型字符串常量
const (
	FieldKindUnknownString   string = "unknown"
	FieldKindDoubleString    string = "double"
	FieldKindFloatString     string = "float"
	FieldKindInt64String     string = "int64"
	FieldKindUInt64String    string = "uint64"
	FieldKindInt32String     string = "int32"
	FieldKindFixed64String   string = "fixed64"
	FieldKindFixed32String   string = "fixed32"
	FieldKindBoolString      string = "bool"
	FieldKindStringString    string = "string"
	FieldKindBytesString     string = "bytes"
	FieldKindUInt32String    string = "uint32"
	FieldKindEnumString      string = "enum"
	FieldKindMMEObjectString string = "MMEObject"
	FieldKindMapString       string = "map"
	FieldKindXMapString      string = "xmap"
	FieldKindRepeatedString  string = "repeated"
	FieldKindMessageString   string = "message"
)

var FieldKindStringMap = map[FieldKind]string{
	FieldKindUnknown:   FieldKindUnknownString,
	FieldKindDouble:    FieldKindDoubleString,
	FieldKindFloat:     FieldKindFloatString,
	FieldKindInt64:     FieldKindInt64String,
	FieldKindUInt64:    FieldKindUInt64String,
	FieldKindInt32:     FieldKindInt32String,
	FieldKindFixed64:   FieldKindFixed64String,
	FieldKindFixed32:   FieldKindFixed32String,
	FieldKindBool:      FieldKindBoolString,
	FieldKindString:    FieldKindStringString,
	FieldKindBytes:     FieldKindBytesString,
	FieldKindUInt32:    FieldKindUInt32String,
	FieldKindEnum:      FieldKindEnumString,
	FieldKindMMEObject: FieldKindMMEObjectString,
	FieldKindMap:       FieldKindMapString,
	FieldKindXMap:      FieldKindXMapString,
	FieldKindRepeated:  FieldKindRepeatedString,
	FieldKindMessage:   FieldKindMessageString,
}

var FieldKindStringToKindMap = map[string]FieldKind{
	FieldKindUnknownString:   FieldKindUnknown,
	FieldKindDoubleString:    FieldKindDouble,
	FieldKindFloatString:     FieldKindFloat,
	FieldKindInt64String:     FieldKindInt64,
	FieldKindUInt64String:    FieldKindUInt64,
	FieldKindInt32String:     FieldKindInt32,
	FieldKindFixed64String:   FieldKindFixed64,
	FieldKindFixed32String:   FieldKindFixed32,
	FieldKindBoolString:      FieldKindBool,
	FieldKindStringString:    FieldKindString,
	FieldKindBytesString:     FieldKindBytes,
	FieldKindUInt32String:    FieldKindUInt32,
	FieldKindEnumString:      FieldKindEnum,
	FieldKindMMEObjectString: FieldKindMMEObject,
	FieldKindMapString:       FieldKindMap,
	FieldKindXMapString:      FieldKindXMap,
	FieldKindRepeatedString:  FieldKindRepeated,
	FieldKindMessageString:   FieldKindMessage,
}
