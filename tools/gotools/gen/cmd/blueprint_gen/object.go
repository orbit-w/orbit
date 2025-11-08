package blueprint_gen

import (
	types "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types"
)

type MMEObjectBase interface {
	GetFields() []*types.Field
	GetName() string
	HasMapField() bool
	HasXMapField() bool
}

type MMEObject struct {
	Name     string
	Fields   []*types.Field
	Settings map[string]any
}

func NewMMEObject() *MMEObject {
	return &MMEObject{
		Fields:   make([]*types.Field, 0),
		Settings: make(map[string]any),
	}
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
