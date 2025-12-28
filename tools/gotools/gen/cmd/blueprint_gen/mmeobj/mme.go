package mmeobj

import (
	"maps"

	types "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types"
	mmeobject "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types/mme_obj"
)

type Index int32

type MMEObject struct {
	Name       string
	ObjectType mmeobject.ObjectType
	Fields     []*types.Field
	Settings   map[string]any
	SourceFile string // 源文件路径（用于 Definition 接口）
}

func NewMMEObject(objectType mmeobject.ObjectType) *MMEObject {
	return &MMEObject{
		ObjectType: objectType,
		Fields:     make([]*types.Field, 0),
		Settings:   make(map[string]any),
	}
}

func (m *MMEObject) GetObjectType() mmeobject.ObjectType {
	return m.ObjectType
}

func (m *MMEObject) GetName() string {
	return m.Name
}

func (m *MMEObject) GetFields() []*types.Field {
	return m.Fields
}

// GetKind 实现 Definition 接口
// 根据 ObjectType 转换为 DefinitionKind
func (m *MMEObject) GetKind() types.DefinitionKind {
	switch m.ObjectType {
	case mmeobject.ObjectTypeEntity:
		return types.DefKindEntity
	case mmeobject.ObjectTypeManager:
		return types.DefKindManager
	case mmeobject.ObjectTypeModule:
		return types.DefKindModule
	case mmeobject.ObjectTypeMechanism:
		return types.DefKindMechanism
	default:
		return types.DefKindUnknown
	}
}

// GetFile 实现 Definition 接口
// 返回定义所在的源文件路径
func (m *MMEObject) GetFile() string {
	return m.SourceFile
}

func (m *MMEObject) GetSettings() map[string]any {
	return m.Settings
}

func (m *MMEObject) SetSettings(settings map[string]any) {
	maps.Copy(m.Settings, settings)
}

func (m *MMEObject) AddField(field *types.Field) {
	m.Fields = append(m.Fields, field)
}

func (m *MMEObject) RemoveField(field *types.Field) {
	for i, f := range m.Fields {
		if f.Number == field.Number {
			m.Fields = append(m.Fields[:i], m.Fields[i+1:]...)
			break
		}
	}
}

func (m *MMEObject) RangeFields(f func(field *types.Field) bool) {
	for _, field := range m.Fields {
		if !f(field) {
			break
		}
	}
}

func (m *MMEObject) HasMapField() bool {
	for _, field := range m.Fields {
		if field.IsMapField() {
			return true
		}
	}
	return false
}

func (m *MMEObject) HasXMapField() bool {
	for _, field := range m.Fields {
		if field.IsXMapField() {
			return true
		}
	}
	return false
}

// field 中存在xmap，且xmap的Value类型是基础类型
func (m *MMEObject) HasXMapValueIsBaseType() bool {
	for _, field := range m.Fields {
		if field.IsXMapField() {
			if field.Type.ValueType.IsFieldBaseType() {
				return true
			}
		}
	}
	return false
}

// field 中存在map，且map的Value类型是基础类型
func (m *MMEObject) HasMapValueIsBaseType() bool {
	for _, field := range m.Fields {
		if field.IsMapField() {
			if field.Type.ValueType.IsFieldBaseType() {
				return true
			}
		}
	}
	return false
}

// field 中存在xmap，且xmap的Value类型是 MMEObject 或 Message
func (m *MMEObject) HasXMapValueIsMMEObjectOrMessage() (bool, *types.Field) {
	for _, field := range m.Fields {
		if field.IsXMapField() {
			if field.Type.IsXMapValueMMEObject() {
				return true, field
			}
		}
	}
	return false, nil
}

// field 中存在map，且map的Value类型是 MMEObject 或 Message
func (m *MMEObject) HasMapValueIsMMEObjectOrMessage() (bool, *types.Field) {
	for _, field := range m.Fields {
		if field.IsMapField() {
			if field.Type.IsMapValueMMEObject() {
				return true, field
			}
		}
	}
	return false, nil
}
