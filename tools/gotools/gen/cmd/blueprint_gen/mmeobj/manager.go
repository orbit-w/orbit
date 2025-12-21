package mmeobj

import mmeobject "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types/mme_obj"

// Manager 管理器定义
type Manager struct {
	*MMEObject
}

func NewManager() *Manager {
	return &Manager{
		MMEObject: NewMMEObject(mmeobject.ObjectTypeManager),
	}
}
