package servicezone

import (
	mmeobj "gitee.com/orbit-w/orbit/app/mme"
	"gitee.com/orbit-w/orbit/app/proto/mme"
)

type IContext interface {
	LoadRefs(refs []*mme.EntityRef) ([]mmeobj.IEntity, error)
	Load(id int64, entityType mme.EntityType) (mmeobj.IEntity, error)
}

type ServiceZoneContext struct {
	*ServiceZone
}

func NewServiceZoneContext(zone *ServiceZone) IContext {
	return &ServiceZoneContext{
		ServiceZone: zone,
	}
}
