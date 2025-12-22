package mmeobj

import (
	"fmt"

	mmeobject "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types/mme_obj"
)

// Module 模块定义
type Module struct {
	*MMEObject
	Mechanisms map[Index]*Mechanism
}

func NewModule() *Module {
	return &Module{
		MMEObject:  NewMMEObject(mmeobject.ObjectTypeModule),
		Mechanisms: make(map[Index]*Mechanism),
	}
}

func (m *Module) LinkMechanisms(mechanisms []*Mechanism) {
	mechanismMap := make(map[string]*Mechanism)
	for _, mechanism := range mechanisms {
		mechanismMap[mechanism.Name] = mechanism
	}
	for i := range m.Fields {
		field := m.Fields[i]
		if field.Type.IsMMEObjectType() {
			if mechanism, ok := mechanismMap[field.Type.GetTypeName()]; ok {
				m.Mechanisms[Index(field.Number)] = mechanism
			} else {
				panic(fmt.Sprintf("module %s link mechanism %s not found", m.Name, field.Type.GetTypeName()))
			}
		}
	}
}
