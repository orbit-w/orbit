package mmeobj

import (
	"fmt"

	blueprint_types "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types"
	mmeobject "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types/mme_obj"
)

// Manager 管理器定义
type Manager struct {
	*MMEObject
	Modules map[Index]*Module
}

func NewManager() *Manager {
	return &Manager{
		MMEObject: NewMMEObject(mmeobject.ObjectTypeManager),
		Modules:   make(map[Index]*Module),
	}
}

func (m *Manager) LinkModules(modules []*Module) {
	moduleMap := make(map[string]*Module)
	for _, module := range modules {
		moduleMap[module.Name] = module
	}
	for i := range m.Fields {
		field := m.Fields[i]
		fieldType := field.Type
		switch fieldType.GetKind() {
		case blueprint_types.FieldKindXMap:
			valueType := field.GetValueType()
			if valueType.IsMMEObjectType() {
				if module, ok := moduleMap[valueType.GetName()]; ok {
					m.Modules[Index(field.Number)] = module
				} else {
					panic(fmt.Sprintf("manager %s link module %s not found", m.Name, valueType.GetName()))
				}
			}

		case blueprint_types.FieldKindMMEObject:
			if module, ok := moduleMap[fieldType.GetName()]; ok {
				m.Modules[Index(field.Number)] = module
			} else {
				panic(fmt.Sprintf("manager %s link module %s not found", m.Name, fieldType.GetName()))
			}
		}
	}
}
