package blueprint_gen

import (
	types "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types"
	mmeobject "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types/mme_obj"
)

type MMEObjectBase interface {
	GetFields() []*types.Field
	GetName() string
	HasMapField() bool
	HasXMapField() bool
	GetObjectType() mmeobject.ObjectType
}

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

func (m *MMEObject) HasMapField() bool {
	return hasMapField(m.Fields)
}

func (m *MMEObject) HasXMapField() bool {
	hasXMap, _ := hasXMapField(m.Fields)
	return hasXMap
}
