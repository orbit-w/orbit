package blueprint_gen

import (
	"fmt"

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

// FieldOption 字段选项（保留用于兼容，但应该使用 types.FieldOption）
type FieldOption struct {
	Access string // access=all/s/c
}

// Request 请求定义
type Request struct {
	Name    string
	Fields  []*types.Field
	Rsp     *Response
	Comment string
}

// Response 响应定义
type Response struct {
	Fields  []*types.Field
	Comment string
}

// Notify 通知定义
type Notify struct {
	Name    string
	Fields  []*types.Field
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

func NewMechanism() *Mechanism {
	return &Mechanism{
		Fields:   make([]*types.Field, 0),
		Settings: make(map[string]any),
		Requests: make([]Request, 0),
		Notifies: make([]Notify, 0),
	}
}

func (m *Mechanism) HasMapField() bool {
	return hasMapField(m.Fields)
}

func (m *Mechanism) HasXMapField() bool {
	hasXMap, _ := hasXMapField(m.Fields)
	return hasXMap
}

func (m *Mechanism) GetFields() []*types.Field {
	return m.Fields
}

// Module 模块定义
type Module struct {
	Name     string
	Settings map[string]any
	Fields   []*types.Field // 使用 types.Field 替代 ModuleMechanismRef
}

func NewModule() *Module {
	return &Module{
		Settings: make(map[string]any),
		Fields:   make([]*types.Field, 0),
	}
}

func (m *Module) HasMapField() bool {
	return hasMapField(m.Fields)
}

func (m *Module) HasXMapField() bool {
	hasXMap, _ := hasXMapField(m.Fields)
	return hasXMap
}

func (m *Module) GetFields() []*types.Field {
	return m.Fields
}

// Manager 管理器定义
type Manager struct {
	Name   string
	Fields []*types.Field // 使用 types.Field 替代 ManagerField
}

func NewManager() *Manager {
	return &Manager{
		Fields: make([]*types.Field, 0),
	}
}

func (m *Manager) HasMapField() bool {
	return hasMapField(m.Fields)
}

func (m *Manager) HasXMapField() bool {
	hasXMap, _ := hasXMapField(m.Fields)
	return hasXMap
}

func (m *Manager) GetFields() []*types.Field {
	return m.Fields
}

// Entity 实体定义
type Entity struct {
	Name   string
	Fields []*types.Field // 使用 types.Field 替代 EntityField
}

func NewEntity() *Entity {
	return &Entity{
		Fields: make([]*types.Field, 0),
	}
}

func (e *Entity) HasMapField() bool {
	return hasMapField(e.Fields)
}

func (e *Entity) HasXMapField() bool {
	hasXMap, _ := hasXMapField(e.Fields)
	return hasXMap
}

func (e *Entity) GetFields() []*types.Field {
	return e.Fields
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
	Fields []*types.Field
}

// NetWallMessage NetWall 消息定义
type NetWallMessage struct {
	Name    string
	Fields  []*types.Field
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
	Entities      []*Entity
	Managers      []*Manager
	Modules       []*Module
	Mechanisms    []*Mechanism
	NetWalls      []NetWallFile
	ObjectTypeMap map[string]ObjectType
}

func NewBlueprintContext() *BlueprintContext {
	return &BlueprintContext{
		Entities:      make([]*Entity, 0),
		Managers:      make([]*Manager, 0),
		Modules:       make([]*Module, 0),
		Mechanisms:    make([]*Mechanism, 0),
		NetWalls:      make([]NetWallFile, 0),
		ObjectTypeMap: make(map[string]ObjectType),
	}
}

func (ctx *BlueprintContext) AddEntity(entity *Entity) {
	ctx.Entities = append(ctx.Entities, entity)
	ctx.ObjectTypeMap[entity.Name] = ObjectTypeEntity
}

func (ctx *BlueprintContext) AddManager(manager *Manager) {
	ctx.Managers = append(ctx.Managers, manager)
	ctx.ObjectTypeMap[manager.Name] = ObjectTypeManager
}

func (ctx *BlueprintContext) AddModule(module *Module) {
	ctx.Modules = append(ctx.Modules, module)
	ctx.ObjectTypeMap[module.Name] = ObjectTypeModule
}

func (ctx *BlueprintContext) AddMechanism(mechanism *Mechanism) {
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

// CheckFields 检查Entity所有字段是否符合要求
func (ctx *BlueprintContext) CheckEntityFields() {
	for _, entity := range ctx.Entities {
		for _, field := range entity.Fields {
			if !field.IsMMEObjectType() {
				panic(fmt.Sprintf("entity %s 的 %s 字段类型不支持，必须是 MMEObject 类型", entity.Name, field.Name))
			}
			fieldName := field.GetTypeName()
			t, ok := ctx.ObjectTypeMap[fieldName]
			if !ok {
				panic(fmt.Sprintf("entity %s 的 %s 字段类型不支持，必须是 MMEObject 类型", entity.Name, field.Name))
			}

			if t != ObjectTypeManager {
				panic(fmt.Sprintf("entity %s 的 %s 字段类型不支持，必须是 Manager 类型", entity.Name, field.Name))
			}
		}
	}
}

// SetHeadFile 设置头文件配置
func (ctx *BlueprintContext) SetHeadFile(headFile *HeadFileConfig) {
	ctx.HeadFile = headFile
}

func hasMapField(fields []*types.Field) bool {
	for _, field := range fields {
		if field.IsMapField() {
			return true
		}
	}
	return false
}

func hasXMapField(fields []*types.Field) (bool, *types.Field) {
	for _, field := range fields {
		if field.IsXMapField() {
			return true, field
		}
	}
	return false, nil
}

// field 中存在xmap，且xmap的Value类型是基础类型
func hasXMapValueIsBaseType(fields []*types.Field) (bool, *types.Field) {
	for _, field := range fields {
		if field.IsXMapField() {
			if field.Type.ValueType.IsFieldBaseType() {
				return true, field
			}
		}
	}
	return false, nil
}

// field 中存在map，且map的Value类型是基础类型
func hasMapValueIsBaseType(fields []*types.Field) (bool, *types.Field) {
	for _, field := range fields {
		if field.IsMapField() {
			if field.Type.ValueType.IsFieldBaseType() {
				return true, field
			}
		}
	}
	return false, nil
}

// field 中存在xmap，且xmap的Value类型是 MMEObject 或 Message
func hasXMapValueIsMMEObjectOrMessage(fields []*types.Field) (bool, *types.Field) {
	for _, field := range fields {
		if field.IsXMapField() {
			if field.Type.IsXMapValueMMEObject() {
				return true, field
			}
		}
	}
	return false, nil
}

// field 中存在map，且map的Value类型是 MMEObject 或 Message
func hasMapValueIsMMEObjectOrMessage(fields []*types.Field) (bool, *types.Field) {
	for _, field := range fields {
		if field.IsMapField() {
			if field.Type.IsMapValueMMEObject() {
				return true, field
			}
		}
	}
	return false, nil
}
