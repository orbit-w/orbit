package servicezone

type ZoneType int32

const (
	ZoneTypeDungeon ZoneType = iota
	ZoneTypeCity
	ZoneTypeWild
	ZoneTypeArena
)

type ServiceZone struct {
	ID        string
	Type      ZoneType
	EntityMap map[int64]int32
	Entities  []IEntity
	// 订阅管理
	subscribers map[string]*Subscriber // 订阅者集合，key 为订阅者 ID
}

// Subscriber 订阅者信息
type Subscriber struct {
	ID       string             // 订阅者唯一标识
	Strategy ISubscribeStrategy // 订阅策略
	Entities map[int64]IEntity  // 已订阅的 Entities，key 为 Entity ID
}

// NewServiceZone 创建新的服务区
func NewServiceZone(id string, zoneType ZoneType) *ServiceZone {
	return &ServiceZone{
		ID:          id,
		Type:        zoneType,
		EntityMap:   make(map[int64]int32),
		Entities:    make([]IEntity, 0),
		subscribers: make(map[string]*Subscriber),
	}
}

// Subscribe 使用指定策略订阅 Entities
// subscriberId: 订阅者唯一标识
// strategy: 订阅策略
// 返回: 订阅到的 Entity 列表
func (zone *ServiceZone) Subscribe(subscriberId string, strategy ISubscribeStrategy) []IEntity {
	// 创建或获取订阅者
	subscriber, exists := zone.subscribers[subscriberId]
	if !exists {
		subscriber = &Subscriber{
			ID:       subscriberId,
			Strategy: strategy,
			Entities: make(map[int64]IEntity),
		}
		zone.subscribers[subscriberId] = subscriber
	} else {
		// 更新策略
		subscriber.Strategy = strategy
		// 清空之前的订阅结果
		subscriber.Entities = make(map[int64]IEntity)
	}

	// 根据策略筛选 Entities
	subscribedEntities := make([]IEntity, 0)
	for _, entity := range zone.Entities {
		if strategy.ShouldSubscribe(entity) {
			entityId := entity.GetXXXId()
			subscriber.Entities[entityId] = entity
			subscribedEntities = append(subscribedEntities, entity)
		}
	}

	return subscribedEntities
}

// SubscribeByIds 使用 ID 列表订阅（便捷方法）
func (zone *ServiceZone) SubscribeByIds(subscriberId string, entityIds []int64) []IEntity {
	strategy := NewByIdsStrategy(entityIds)
	return zone.Subscribe(subscriberId, strategy)
}

// SubscribeByType 使用类型订阅（便捷方法）
func (zone *ServiceZone) SubscribeByType(subscriberId string, entityTypes []string) []IEntity {
	strategy := NewByTypeStrategy(entityTypes)
	return zone.Subscribe(subscriberId, strategy)
}

// SubscribeAll 订阅所有 Entities（便捷方法）
func (zone *ServiceZone) SubscribeAll(subscriberId string) []IEntity {
	strategy := NewAllEntitiesStrategy()
	return zone.Subscribe(subscriberId, strategy)
}

// Unsubscribe 取消订阅
func (zone *ServiceZone) Unsubscribe(subscriberId string) {
	delete(zone.subscribers, subscriberId)
}

// GetSubscribedEntities 获取订阅者已订阅的 Entities
func (zone *ServiceZone) GetSubscribedEntities(subscriberId string) []IEntity {
	subscriber, exists := zone.subscribers[subscriberId]
	if !exists {
		return nil
	}

	entities := make([]IEntity, 0, len(subscriber.Entities))
	for _, entity := range subscriber.Entities {
		entities = append(entities, entity)
	}
	return entities
}

// UpdateSubscriptions 当 Entity 添加或更新时，更新所有订阅者
// 新添加的 Entity 会根据各订阅者的策略决定是否订阅
func (zone *ServiceZone) UpdateSubscriptions(entity IEntity) {
	if entity == nil {
		return
	}
	entityId := entity.GetXXXId()
	for _, subscriber := range zone.subscribers {
		if subscriber.Strategy.ShouldSubscribe(entity) {
			// 添加到订阅列表
			subscriber.Entities[entityId] = entity
		} else {
			// 如果不再符合订阅条件，从订阅列表中移除
			delete(subscriber.Entities, entityId)
		}
	}
}

// RemoveFromSubscriptions 当 Entity 移除时，从所有订阅者中移除
func (zone *ServiceZone) RemoveFromSubscriptions(entityId int64) {
	for _, subscriber := range zone.subscribers {
		delete(subscriber.Entities, entityId)
	}
}
