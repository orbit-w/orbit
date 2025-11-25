package servicezone_behavior

import (
	mmeobj "gitee.com/orbit-w/orbit/app/mme"
	"gitee.com/orbit-w/orbit/app/proto/mme"
)

type IContext interface {
	AddEntity(entity mmeobj.IEntity)
	SetEntity(entity mmeobj.IEntity)
	RemoveEntity(targetId int64)
	LoadRefs(refs []*mme.EntityRef) ([]mmeobj.IEntity, error)
	Load(id int64, entityType mme.EntityType) (mmeobj.IEntity, error)
}
