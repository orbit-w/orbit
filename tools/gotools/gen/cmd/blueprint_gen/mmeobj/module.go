package mmeobj

import mmeobject "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types/mme_obj"

// Module 模块定义
type Module struct {
	*MMEObject
}

func NewModule() *Module {
	return &Module{
		MMEObject: NewMMEObject(mmeobject.ObjectTypeModule),
	}
}
