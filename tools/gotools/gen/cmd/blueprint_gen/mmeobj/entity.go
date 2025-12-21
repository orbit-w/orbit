package mmeobj

import mmeobject "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types/mme_obj"

// Entity 实体定义
type Entity struct {
	*MMEObject
}

func NewEntity() *Entity {
	return &Entity{
		MMEObject: NewMMEObject(mmeobject.ObjectTypeEntity),
	}
}

func (e *Entity) GetName() string {
	return e.Name
}
