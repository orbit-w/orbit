package blueprint_gen

import (
	"fmt"
	"strings"

	"gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/mmeobj"
	"gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/pb_gen/net_message"
	types "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types"
	mmeobject "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types/mme_obj"
)

// FieldOption 字段选项（保留用于兼容，但应该使用 types.FieldOption）
type FieldOption struct {
	Access string // access=all/s/c
}

// HeadFileConfig 头文件配置
type HeadFileConfig struct {
	ModuleStorageOption       []string
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

// NetWallMessageType NetWall 消息类型
type NetWallMessageType int

const (
	NetWallMessageTypeRequest NetWallMessageType = iota
	NetWallMessageTypeNotify
	NetWallMessageTypeDataStruct
	NetWallMessageTypeResponse
)

const (
	EnumValueOptionContent = "Content"
)

// EnumValue 枚举值定义
type EnumValue struct {
	Name    string            // 枚举值名称，如 ServiceZoneTypePlay
	Number  int32             // 枚举值编号，如 0
	Comment string            // 枚举值注释
	Options map[string]string // 枚举值选项
}

func NewEnumValue(name string, number int32, comment string) *EnumValue {
	return &EnumValue{
		Name:    name,
		Number:  number,
		Comment: comment,
		Options: make(map[string]string),
	}
}

// ParseComment 解析枚举值注释
func (e *EnumValue) ParseComment() string {
	opt, ok := e.Options[EnumValueOptionContent]
	if !ok {
		panic(fmt.Sprintf("EnumValue %s 的 %s 选项不存在", e.Name, EnumValueOptionContent))
	}
	e.Comment = opt
	return opt
}

// Enum 枚举定义
type Enum struct {
	Name        string       // 枚举名称，如 ServiceZoneType
	Values      []*EnumValue // 枚举值列表
	Comment     string       // 枚举注释
	SourceProto string       // 来源 proto 文件，如 "entities", "managers", "modules", "mechanisms" 或 NetWall 包名
}

func NewEnum(name, comment, sourceProto string) *Enum {
	return &Enum{
		Name:        name,
		Values:      make([]*EnumValue, 0),
		Comment:     comment,
		SourceProto: sourceProto,
	}
}

// NetWallFile NetWall 文件内容
type NetWallFile struct {
	Name        string
	PackageName string
	Requests    []*net_message.NetMessage
	Notifies    []*net_message.NetMessage
	DataStructs []*net_message.NetMessage
	Enums       []*Enum
	NameSpace   map[string]bool
}

func NewNetWallFile() *NetWallFile {
	return &NetWallFile{
		Requests:    make([]*net_message.NetMessage, 0),
		Notifies:    make([]*net_message.NetMessage, 0),
		DataStructs: make([]*net_message.NetMessage, 0),
		Enums:       make([]*Enum, 0),
		NameSpace:   make(map[string]bool),
	}
}

func (n *NetWallFile) GetAllMessages() []*net_message.NetMessage {
	allMessages := make([]*net_message.NetMessage, 0)
	allMessages = append(allMessages, n.Requests...)
	allMessages = append(allMessages, n.Notifies...)
	allMessages = append(allMessages, n.DataStructs...)
	return allMessages
}

func (n *NetWallFile) AddRequest(request *net_message.NetMessage) {
	if n.HasNameSpace(request.GetName()) {
		panic(fmt.Sprintf("NetWallFile %s 的请求 %s 已存在", n.Name, request.GetName()))
	}
	n.Requests = append(n.Requests, request)
	n.NameSpace[request.Name] = true
}

func (n *NetWallFile) AddNotify(notify *net_message.NetMessage) {
	if n.HasNameSpace(notify.GetName()) {
		panic(fmt.Sprintf("NetWallFile %s 的通知 %s 已存在", n.Name, notify.GetName()))
	}
	n.Notifies = append(n.Notifies, notify)
	n.NameSpace[notify.Name] = true
}

func (n *NetWallFile) AddDataStruct(dataStruct *net_message.NetMessage) {
	if n.HasNameSpace(dataStruct.GetName()) {
		panic(fmt.Sprintf("NetWallFile %s 的数据结构 %s 已存在", n.Name, dataStruct.GetName()))
	}
	n.DataStructs = append(n.DataStructs, dataStruct)
	n.NameSpace[dataStruct.Name] = true
}

func (n *NetWallFile) AddEnum(enum *Enum) {
	if n.HasNameSpace(enum.Name) {
		panic(fmt.Sprintf("NetWallFile %s 的枚举 %s 已存在", n.Name, enum.Name))
	}
	n.Enums = append(n.Enums, enum)
	n.NameSpace[enum.Name] = true
}

func (n *NetWallFile) HasNameSpace(name string) bool {
	_, ok := n.NameSpace[name]
	return ok
}

func (n *NetWallFile) GetRequests() []*net_message.NetMessage {
	return n.Requests
}

func (n *NetWallFile) GetNotifies() []*net_message.NetMessage {
	return n.Notifies
}

func (n *NetWallFile) GetDataStructs() []*net_message.NetMessage {
	return n.DataStructs
}

func (n *NetWallFile) GetEnums() []*Enum {
	return n.Enums
}

func (n *NetWallFile) GetNameSpace() map[string]bool {
	return n.NameSpace
}

func (n *NetWallFile) GetPackageName() string {
	return n.PackageName
}

// BlueprintContext 解析后的蓝图数据
type BlueprintContext struct {
	HeadFile            *HeadFileConfig
	Entities            []*mmeobj.Entity
	Managers            []*mmeobj.Manager
	Modules             []*mmeobj.Module
	Mechanisms          []*mmeobj.Mechanism
	NetWalls            []*NetWallFile
	Enums               []*Enum
	ObjectTypeMap       map[string]mmeobject.ObjectType
	NameSpaces          map[string]string // 包名空间,建立对象类型与包名空间的映射
	NameSpaceMap        map[string]bool   // 命名空间检查，enum/struct/table不允许有重复的命名空间
	ImportProtoNameMap  map[string]string // message Name与Proto File名的映射,建立import映射
	PackageProtoNameMap map[string]string // message Name与Proto Package名的映射
}

func NewBlueprintContext() *BlueprintContext {
	return &BlueprintContext{
		Entities:            make([]*mmeobj.Entity, 0),
		Managers:            make([]*mmeobj.Manager, 0),
		Modules:             make([]*mmeobj.Module, 0),
		Mechanisms:          make([]*mmeobj.Mechanism, 0),
		NetWalls:            make([]*NetWallFile, 0),
		Enums:               make([]*Enum, 0),
		ObjectTypeMap:       make(map[string]mmeobject.ObjectType),
		NameSpaces:          make(map[string]string),
		NameSpaceMap:        make(map[string]bool),
		ImportProtoNameMap:  make(map[string]string),
		PackageProtoNameMap: make(map[string]string),
	}
}

func (ctx *BlueprintContext) AddEnum(enum *Enum) {
	if ctx.HasNameSpace(enum.Name) {
		panic(fmt.Sprintf("BlueprintContext 的枚举 %s 已存在", enum.Name))
	}
	ctx.Enums = append(ctx.Enums, enum)
	ctx.NameSpaceMap[enum.Name] = true
}

func (ctx *BlueprintContext) GetEnums() []*Enum {
	return ctx.Enums
}

func (ctx *BlueprintContext) AddEntity(entity *mmeobj.Entity) {
	ctx.Entities = append(ctx.Entities, entity)
	ctx.ObjectTypeMap[entity.Name] = mmeobject.ObjectTypeEntity
}

func (ctx *BlueprintContext) AddManager(manager *mmeobj.Manager) {
	ctx.Managers = append(ctx.Managers, manager)
	ctx.ObjectTypeMap[manager.Name] = mmeobject.ObjectTypeManager
}

func (ctx *BlueprintContext) AddModule(module *mmeobj.Module) {
	ctx.Modules = append(ctx.Modules, module)
	ctx.ObjectTypeMap[module.Name] = mmeobject.ObjectTypeModule
}

func (ctx *BlueprintContext) AddMechanism(mechanism *mmeobj.Mechanism) {
	ctx.Mechanisms = append(ctx.Mechanisms, mechanism)
	ctx.ObjectTypeMap[mechanism.Name] = mmeobject.ObjectTypeMechanism
}

// AddNetWallFile 添加 NetWallFile
func (ctx *BlueprintContext) AddNetWallFile(netWallFile *NetWallFile) {
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

			if t != mmeobject.ObjectTypeManager {
				panic(fmt.Sprintf("entity %s 的 %s 字段类型不支持，必须是 Manager 类型", entity.Name, field.Name))
			}
		}
	}
}

func (ctx *BlueprintContext) GetObjectType(name string) (mmeobject.ObjectType, bool) {
	objType, ok := ctx.ObjectTypeMap[name]
	return objType, ok
}

func (ctx *BlueprintContext) GetNameSpace() map[string]string {
	if len(ctx.NameSpaces) == 0 {
		// 映射 MME 文件中的枚举到 Enum 包
		for _, enum := range ctx.Enums {
			ctx.NameSpaces[enum.Name] = "Enum"
		}

		// 映射 MME 对象（Entities, Managers, Modules, Mechanisms）到 MME 包
		for _, entity := range ctx.Entities {
			ctx.NameSpaces[entity.Name] = "MME"
		}
		for _, manager := range ctx.Managers {
			ctx.NameSpaces[manager.Name] = "MME"
		}
		for _, module := range ctx.Modules {
			ctx.NameSpaces[module.Name] = "MME"
		}
		for _, mechanism := range ctx.Mechanisms {
			ctx.NameSpaces[mechanism.Name] = "MME"
		}

		// 映射 Common DataStructs 到 MME 包
		if ctx.HeadFile != nil {
			for _, ds := range ctx.HeadFile.CommonDataStructs {
				ctx.NameSpaces[ds.Name] = "MME"
			}
		}

		// 映射 NetWall 文件中的消息和数据结构
		for _, netwallFile := range ctx.NetWalls {
			// 映射 NetWall 的 DataStructs
			for _, ds := range netwallFile.DataStructs {
				ctx.NameSpaces[ds.Name] = netwallFile.PackageName
			}
			// 映射 NetWall 的 Requests
			for _, req := range netwallFile.Requests {
				ctx.NameSpaces[req.Name] = netwallFile.PackageName
			}
			// 映射 NetWall 的 Notifies
			for _, notify := range netwallFile.Notifies {
				ctx.NameSpaces[notify.Name] = netwallFile.PackageName
			}
			// 映射 NetWall 的枚举 - 所有枚举都映射到 Enum 包
			for _, enum := range netwallFile.Enums {
				ctx.NameSpaces[enum.Name] = "Enum"
			}
		}
	}
	return ctx.NameSpaces
}

// GetProtoImportByObjectName 根据对象名称获取 proto 导入文件
// name: 对象名称
// 返回: proto 导入文件名
func (ctx *BlueprintContext) GetProtoImportByObjectName(name string) string {
	if len(ctx.ImportProtoNameMap) == 0 {
		for _, netwallFile := range ctx.NetWalls {
			for _, ds := range netwallFile.DataStructs {
				ctx.ImportProtoNameMap[ds.Name] = strings.ToLower(netwallFile.PackageName) + ".proto"
			}
			for _, req := range netwallFile.Requests {
				ctx.ImportProtoNameMap[req.GetName()] = strings.ToLower(netwallFile.PackageName) + ".proto"
				if req.HasResponse() {
					ctx.ImportProtoNameMap[req.GetResponse().GetName()] = strings.ToLower(netwallFile.PackageName) + ".proto"
				}
			}
			for _, notify := range netwallFile.Notifies {
				ctx.ImportProtoNameMap[notify.Name] = strings.ToLower(netwallFile.PackageName) + ".proto"
			}
			// 映射 NetWall 的枚举 - 所有枚举都映射到 common.proto
			for _, enum := range netwallFile.Enums {
				ctx.ImportProtoNameMap[enum.Name] = "common.proto"
			}
		}

		// 检查是否是 Common DataStruct
		if ctx.HeadFile != nil {
			for _, ds := range ctx.HeadFile.CommonDataStructs {
				ctx.ImportProtoNameMap[ds.Name] = "common.proto"
			}
		}

		for _, entity := range ctx.Entities {
			ctx.ImportProtoNameMap[entity.Name] = "entities.proto"
		}
		for _, manager := range ctx.Managers {
			ctx.ImportProtoNameMap[manager.Name] = "managers.proto"
		}
		for _, module := range ctx.Modules {
			ctx.ImportProtoNameMap[module.Name] = "modules.proto"
		}
		for _, mechanism := range ctx.Mechanisms {
			ctx.ImportProtoNameMap[mechanism.Name] = "mechanisms.proto"
		}
		for _, enum := range ctx.Enums {
			ctx.ImportProtoNameMap[enum.Name] = "common.proto"
		}
	}
	return ctx.ImportProtoNameMap[name]
}

// GetPackageProtoName 根据对象名称获取 Proto Package 名
// name: 对象名称
// 返回: Proto Package 名
func (ctx *BlueprintContext) GetPackageProtoName(name string) string {
	if len(ctx.PackageProtoNameMap) == 0 {
		for _, netwallFile := range ctx.NetWalls {
			for _, ds := range netwallFile.DataStructs {
				ctx.PackageProtoNameMap[ds.Name] = netwallFile.PackageName
			}
			for _, req := range netwallFile.Requests {
				ctx.PackageProtoNameMap[req.GetName()] = netwallFile.PackageName
				if req.HasResponse() {
					ctx.PackageProtoNameMap[req.GetResponse().GetName()] = netwallFile.PackageName
				}
			}
			for _, notify := range netwallFile.Notifies {
				ctx.PackageProtoNameMap[notify.Name] = netwallFile.PackageName
			}
			for _, enum := range netwallFile.Enums {
				ctx.PackageProtoNameMap[enum.Name] = "MME"
			}
		}

		if ctx.HeadFile != nil {
			for _, ds := range ctx.HeadFile.CommonDataStructs {
				ctx.PackageProtoNameMap[ds.Name] = "MME"
			}
		}

		for _, entity := range ctx.Entities {
			ctx.PackageProtoNameMap[entity.Name] = "MME"
		}
		for _, manager := range ctx.Managers {
			ctx.PackageProtoNameMap[manager.Name] = "MME"
		}
		for _, module := range ctx.Modules {
			ctx.PackageProtoNameMap[module.Name] = "MME"
		}
		for _, mechanism := range ctx.Mechanisms {
			ctx.PackageProtoNameMap[mechanism.Name] = "MME"
		}
		for _, enum := range ctx.Enums {
			ctx.PackageProtoNameMap[enum.Name] = "MME"
		}
	}
	return ctx.PackageProtoNameMap[name]
}

// SetHeadFile 设置头文件配置
func (ctx *BlueprintContext) SetHeadFile(headFile *HeadFileConfig) {
	ctx.HeadFile = headFile
}

func (ctx *BlueprintContext) HasNameSpace(name string) bool {
	_, ok := ctx.NameSpaces[name]
	return ok
}
