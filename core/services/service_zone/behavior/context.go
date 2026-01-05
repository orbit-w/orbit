package servicezone_behavior

import (
	"gitee.com/orbit-w/orbit/internal/game/mme_agent/entities"
	"gitee.com/orbit-w/orbit/pkg/proto/mme"
)

type IContext interface {
	AddEntity(entity entities.IEntity)
	SetEntity(entity entities.IEntity)
	RemoveEntity(targetId int64)
	LoadRefs(refs []*mme.EntityRef) ([]entities.IEntity, error)
	Load(id int64, entityType mme.EntityType) (entities.IEntity, error)
	Persist(entities ...entities.IEntity) // 异步持久化实体数据
}
