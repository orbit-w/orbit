package servicezone

import (
	"gitee.com/orbit-w/orbit/app/proto/enum"
)

// Subscriber 订阅者信息
type Subscriber struct {
	UUID           string                                 // 订阅者唯一标识
	Strategy       ISubscribeStrategy                     // 订阅策略
	Entities       map[int64]enum.EntityType              // 已订阅的 Entities ID 列表, 用于快速遍历, 保证订阅顺序
	EntitiesByType map[enum.EntityType]map[int64]struct{} // 已订阅的 Entities 类型和 Entity 列表, Key 为 Entity 类型, Value 为 Entity 列表
}

func NewSubscriber(uuid string, strategy ISubscribeStrategy) *Subscriber {
	return &Subscriber{
		UUID:           uuid,
		Strategy:       strategy,
		Entities:       make(map[int64]enum.EntityType),
		EntitiesByType: make(map[enum.EntityType]map[int64]struct{}),
	}
}

// 获取已订阅的 Entities ID 列表
func (s *Subscriber) GetSubscribedEntities() []int64 {
	entities := make([]int64, 0, len(s.Entities))
	for id := range s.Entities {
		entities = append(entities, id)
	}
	return entities
}

// 无序遍历已订阅的 Entities ID 列表
func (s *Subscriber) RangeSubscribedEntities(callback func(entityId int64)) {
	for entityId := range s.Entities {
		callback(entityId)
	}
}

// 判断是否已订阅 Entity
func (s *Subscriber) HasSubscribedEntity(entityId int64) bool {
	_, ok := s.Entities[entityId]
	return ok
}

func (s *Subscriber) ClearSubscribedEntities() {
	s.Entities = make(map[int64]enum.EntityType)
	s.EntitiesByType = make(map[enum.EntityType]map[int64]struct{})
}

// 订阅 Entity
func (s *Subscriber) SubscribeEntity(entity IEntity) {
	id := entity.GetXXXId()
	if entity == nil {
		return
	}
	entityType := entity.GetEntityType()

	remain, ok := s.Entities[id]
	if ok {
		if remain == entityType {
			return
		}
		s.unsubscribeEntityById(id, remain, false)
	}

	s.Entities[id] = entityType
	if _, ok := s.EntitiesByType[entityType]; !ok {
		s.EntitiesByType[entityType] = make(map[int64]struct{})
	}
	s.EntitiesByType[entityType][id] = struct{}{}
}

func (s *Subscriber) UnsubscribeEntity(id int64) {
	entityType, ok := s.Entities[id]
	if !ok {
		return
	}
	s.unsubscribeEntityById(id, entityType, true)
}

func (s *Subscriber) unsubscribeEntityById(entityId int64, entityType enum.EntityType, clear bool) {
	delete(s.Entities, entityId)
	delete(s.EntitiesByType[entityType], entityId)
	if clear && len(s.EntitiesByType[entityType]) == 0 {
		delete(s.EntitiesByType, entityType)
	}
}
