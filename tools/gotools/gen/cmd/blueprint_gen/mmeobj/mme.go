package mmeobj

import (
	"maps"

	types "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types"
	mmeobject "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types/mme_obj"
)

type MMEObject struct {
	Name       string
	ObjectType mmeobject.ObjectType
	Fields     []*types.Field
	Settings   map[string]any
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

func (m *MMEObject) GetSettings() map[string]any {
	return m.Settings
}

func (m *MMEObject) SetSettings(settings map[string]any) {
	maps.Copy(m.Settings, settings)
}

func (m *MMEObject) AddField(field *types.Field) {
	m.Fields = append(m.Fields, field)
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
