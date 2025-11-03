package blueprint_gen

import (
	types "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types"
)

// ObjectType MME 对象类型
type ObjectType string

const (
	ObjectTypeEntity    ObjectType = "Entity"
	ObjectTypeManager   ObjectType = "Manager"
	ObjectTypeModule    ObjectType = "Module"
	ObjectTypeMechanism ObjectType = "Mechanism"
)

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
	Name     string
	Fields   []*types.Field
	Settings map[string]any
	Requests []Request
	Notifies []Notify
}

// Module 模块定义
type Module struct {
	Name       string
	Mechanisms []*types.Field // 使用 types.Field 替代 ModuleMechanismRef
}

// Manager 管理器定义
type Manager struct {
	Name   string
	Fields []*types.Field // 使用 types.Field 替代 ManagerField
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
	EntityFields              map[string]any
	ModuleFields              map[string]any
	ModuleStorageOption       []string
	MechanismFields           map[string]any
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

// BlueprintContext 解析后的蓝图数据
type BlueprintContext struct {
	HeadFile      *HeadFileConfig
	Entities      []Entity
	Managers      []Manager
	Modules       []Module
	Mechanisms    []Mechanism
	NetWalls      []NetWallFile
	ObjectTypeMap map[string]ObjectType
}

func NewBlueprintContext() *BlueprintContext {
	return &BlueprintContext{
		Entities:      make([]Entity, 0),
		Managers:      make([]Manager, 0),
		Modules:       make([]Module, 0),
		Mechanisms:    make([]Mechanism, 0),
		NetWalls:      make([]NetWallFile, 0),
		ObjectTypeMap: make(map[string]ObjectType),
	}
}

func (ctx *BlueprintContext) AddEntity(entity Entity) {
	ctx.Entities = append(ctx.Entities, entity)
	ctx.ObjectTypeMap[entity.Name] = ObjectTypeEntity
}

func (ctx *BlueprintContext) AddManager(manager Manager) {
	ctx.Managers = append(ctx.Managers, manager)
	ctx.ObjectTypeMap[manager.Name] = ObjectTypeManager
}

func (ctx *BlueprintContext) AddModule(module Module) {
	ctx.Modules = append(ctx.Modules, module)
	ctx.ObjectTypeMap[module.Name] = ObjectTypeModule
}

func (ctx *BlueprintContext) AddMechanism(mechanism Mechanism) {
	ctx.Mechanisms = append(ctx.Mechanisms, mechanism)
	ctx.ObjectTypeMap[mechanism.Name] = ObjectTypeMechanism
}

func (ctx *BlueprintContext) AddNetWall(netWall NetWall) {
	ctx.NetWalls = append(ctx.NetWalls, NetWallFile{
		NetWall: netWall,
	})
}

// AddNetWallFile 添加 NetWallFile
func (ctx *BlueprintContext) AddNetWallFile(netWallFile NetWallFile) {
	ctx.NetWalls = append(ctx.NetWalls, netWallFile)
}

// SetHeadFile 设置头文件配置
func (ctx *BlueprintContext) SetHeadFile(headFile *HeadFileConfig) {
	ctx.HeadFile = headFile
}
