package blueprint_gen

// FieldOption 字段选项
type FieldOption struct {
	Access string // access=all/s/c
}

// FieldType 字段类型定义
type FieldType struct {
	BaseType   string // int32, int64, string, bool, float, double
	IsMap      bool
	IsXMap     bool
	IsRepeated bool
	KeyType    string // map/xmap 的 key 类型
	ValueType  string // map/xmap 的 value 类型，或基础类型
}

// Field 字段定义
type Field struct {
	Name    string
	Type    FieldType
	Number  int32 // 字段编号
	Options FieldOption
	Comment string
}

// MechanismDataField 机制数据字段
type MechanismDataField struct {
	Field
}

// Request 请求定义
type Request struct {
	Name    string
	Fields  []Field
	Rsp     *Response
	Comment string
}

// Response 响应定义
type Response struct {
	Fields  []Field
	Comment string
}

// Notify 通知定义
type Notify struct {
	Name    string
	Fields  []Field
	Comment string
}

// Mechanism 机制定义
type Mechanism struct {
	Name       string
	DataFields []MechanismDataField
	Settings   map[string]interface{}
	Requests   []Request
	Notifies   []Notify
}

// ModuleMechanismRef Module 中引用的 Mechanism
type ModuleMechanismRef struct {
	MechanismName string
	FieldName     string
	Number        int32
	Settings      map[string]interface{}
}

// Module 模块定义
type Module struct {
	Name       string
	Mechanisms []ModuleMechanismRef
}

// ManagerField Manager 中的字段（通常是 map<key, Module>）
type ManagerField struct {
	Name       string
	KeyType    string
	ModuleName string
	Number     int32
	IsXMap     bool
}

// Manager 管理器定义
type Manager struct {
	Name   string
	Fields []ManagerField
}

// EntityField Entity 中的字段（通常是 Manager）
type EntityField struct {
	Name        string
	ManagerName string
	Number      int32
	Options     FieldOption
}

// Entity 实体定义
type Entity struct {
	Name   string
	Fields []EntityField
}

// HeadFileConfig 头文件配置
type HeadFileConfig struct {
	EntityFields              map[string]interface{}
	ModuleFields              map[string]interface{}
	ModuleStorageOption       []string
	MechanismFields           map[string]interface{}
	MechanismDataFieldOptions map[string]FieldOptionDefinition
	CommonDataStructs         []DataStruct
}

// FieldOptionDefinition 字段选项定义
type FieldOptionDefinition struct {
	Default     string
	Enum        string
	Description string
	Choices     []string
}

// DataStruct 通用数据结构
type DataStruct struct {
	Name   string
	Fields []Field
}

// NetWallMessage NetWall 消息定义
type NetWallMessage struct {
	Name    string
	Fields  []Field
	Rsp     *Response // 仅用于 Request
	Comment string
}

// NetWall NetWall 定义
type NetWall struct {
	Name     string
	Requests []NetWallMessage
	Notifies []NetWallMessage
}

// NetWallFile NetWall 文件内容
type NetWallFile struct {
	NetWall     NetWall
	DataStructs []DataStruct
}

// BlueprintData 解析后的蓝图数据
type BlueprintData struct {
	HeadFile   *HeadFileConfig
	Entities   []Entity
	Managers   []Manager
	Modules    []Module
	Mechanisms []Mechanism
	NetWalls   []NetWallFile
}
