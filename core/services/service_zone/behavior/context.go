package servicezone_behavior

import (
	"gitee.com/orbit-w/orbit/internal/game/mme_agent"
	"gitee.com/orbit-w/orbit/pkg/proto/mme"
)

type IContext interface {
	AddEntity(entity mme_agent.IEntity)
	SetEntity(entity mme_agent.IEntity)
	RemoveEntity(targetId int64)
	LoadRefs(refs []*mme.EntityRef) ([]mme_agent.IEntity, error)
	Load(id int64, entityType mme.EntityType) (mme_agent.IEntity, error)
	Persist(entities ...mme_agent.IEntity) // 异步持久化实体数据
}
