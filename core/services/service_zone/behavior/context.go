package servicezone_behavior

import (
	mmeobj "gitee.com/orbit-w/orbit/internal/game/mme"
	"gitee.com/orbit-w/orbit/pkg/proto/mme"
)

type IContext interface {
	AddEntity(entity mmeobj.IEntity)
	SetEntity(entity mmeobj.IEntity)
	RemoveEntity(targetId int64)
	LoadRefs(refs []*mme.EntityRef) ([]mmeobj.IEntity, error)
	Load(id int64, entityType mme.EntityType) (mmeobj.IEntity, error)
	Persist(entities ...mmeobj.IEntity) // 异步持久化实体数据
}
