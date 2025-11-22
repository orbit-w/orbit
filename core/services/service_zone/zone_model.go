package servicezone

import (
	"fmt"

	mmeobj "gitee.com/orbit-w/orbit/app/mme"
	"gitee.com/orbit-w/orbit/app/proto/mme"
	"gitee.com/orbit-w/orbit/lib/module/persistence"
	"github.com/asynkron/protoactor-go/actor"
	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	// ZonePattern ServiceZone Actor 的 pattern
	ZonePattern = "zone-pattern"
)

var (
// zoneRegistry ServiceZone 注册表，用于 Actor Factory 根据 actorName 查找对应的 ServiceZone
// 在 behavior.go 中定义，这里只是声明
)

type ZoneType int32

const (
	ZoneTypeDungeon ZoneType = iota
	ZoneTypeCity
	ZoneTypeWild
	ZoneTypeArena
)

type ServiceZone struct {
	ID            string
	Type          ZoneType
	EntityTypeMap map[int64]mme.EntityType
	Entities      map[int64]mmeobj.IEntity
	// 订阅管理
	subscribers map[string]*Subscriber // 订阅者集合，key 为订阅者 ID

	actorPID    *actor.PID
	actorSystem *actor.ActorSystem
}

// NewServiceZone 创建新的服务区
func NewServiceZone(id string, zoneType ZoneType) *ServiceZone {
	return &ServiceZone{
		ID:            id,
		Type:          zoneType,
		EntityTypeMap: make(map[int64]mme.EntityType),
		Entities:      make(map[int64]mmeobj.IEntity),
		subscribers:   make(map[string]*Subscriber),
	}
}

// Start 启动 ServiceZone 的 Actor
// 需要在 Actor 系统启动后调用
func (zone *ServiceZone) Start(system *actor.ActorSystem) error {
	zone.actorSystem = system
	// 直接创建持久化Actor，使用supervision策略
	// 这里不使用supervision系统，而是直接创建，但使用supervision策略来保证容错
	decider := func(reason any) actor.Directive {
		// 使用Resume策略，出错后继续运行
		return actor.ResumeDirective
	}
	supervisor := actor.NewOneForOneStrategy(10, 1000, decider)

	props := actor.PropsFromProducer(func() actor.Actor {
		return NewZoneActorBehavior(zone)
	}, actor.WithSupervisor(supervisor))

	// 直接在Root下创建Actor
	ctx := system.Root
	pid, err := ctx.SpawnNamed(props, GenActorId(zone.ID))
	if err != nil {
		return err
	}

	zone.actorPID = pid
	return nil
}

// Stop 停止 ServiceZone 的 Actor
func (zone *ServiceZone) Stop() error {
	if zone.actorPID != nil {
		// 直接停止Actor
		ctx := zone.actorSystem.Root
		future := ctx.PoisonFuture(zone.actorPID)
		zone.actorPID = nil
		if err := future.Wait(); err != nil {
			return err
		}
	}
	return nil
}

func (zone *ServiceZone) GetActorPID() *actor.PID {
	return zone.actorPID
}

func (zone *ServiceZone) Load(id int64, entityType mme.EntityType) (mmeobj.IEntity, error) {
	factory := mmeobj.GetEntityFactory(entityType)
	if factory == nil {
		return nil, fmt.Errorf("entity factory not found for entity type %d", entityType)
	}
	entity := factory()
	future, err := persistence.Load(zone.ID, entity.Collection(), id)
	if err != nil {
		return nil, err
	}
	if err := future.Wait(); err != nil {
		return nil, err
	}
	resp, err := future.Result()
	if err != nil {
		return nil, err
	}
	loadResp, ok := resp.(ILoadResponse)
	if !ok {
		return nil, fmt.Errorf("invalid load response type: %T", resp)
	}

	if err := loadResp.GetError(); err != nil {
		return nil, err
	}

	if !loadResp.IsExists() {
		// 如果不存在，则新建一个实体
		entity.Load(bson.Raw{})
		entity.SetXXXId(id)
		return entity, nil
	}

	// 如果存在，则加载数据
	if err := entity.Load(loadResp.GetData()); err != nil {
		return nil, err
	}
	return entity, nil
}

// AddEntity 添加或更新 Entity（实现 IServiceZone 接口）
func (zone *ServiceZone) AddEntity(entity mmeobj.IEntity) {
	zone.SetEntity(entity)
}

// SetEntity 设置 Entity（添加或更新）
func (zone *ServiceZone) SetEntity(entity mmeobj.IEntity) {
	if entity == nil {
		return
	}

	// 更新实体对象引用（即使已存在也要更新，因为可能是新的对象实例）
	zone.addEntity(entity)

	// 更新订阅
	zone.UpdateSubscriptions(entity)
}

func (zone *ServiceZone) RemoveEntity(targetId int64) {
	entity := zone.getEntityById(targetId)
	if entity == nil {
		return
	}
	zone.removeEntity(targetId)
	// 从所有订阅者中移除
	zone.RemoveFromSubscriptions(entity)
}

func (zone *ServiceZone) getEntityById(targetId int64) mmeobj.IEntity {
	// 直接使用 map 查找，O(1) 复杂度
	entity, ok := zone.Entities[targetId]
	if !ok {
		return nil
	}
	// 验证实体 ID 是否匹配（防止 map key 和实体 ID 不一致的情况）
	if entity != nil && entity.GetXXXId() == targetId {
		return entity
	}
	return nil
}

func (zone *ServiceZone) addEntity(entity mmeobj.IEntity) {
	id := entity.GetXXXId()
	zone.Entities[id] = entity

	zone.EntityTypeMap[id] = entity.GetEntityType()
}

func (zone *ServiceZone) removeEntity(targetId int64) {
	_, ok := zone.EntityTypeMap[targetId]
	if !ok {
		return
	}
	delete(zone.EntityTypeMap, targetId)
	delete(zone.Entities, targetId)
}

// Subscribe 使用指定策略订阅 Entities
// subscriberId: 订阅者唯一标识
// strategy: 订阅策略
// 返回: 订阅到的 Entity 列表
func (zone *ServiceZone) Subscribe(subscriberId string, strategy ISubscribeStrategy) []mmeobj.IEntity {
	// 创建或获取订阅者
	subscriber, exists := zone.subscribers[subscriberId]
	if !exists {
		subscriber = NewSubscriber(subscriberId, strategy)
		zone.subscribers[subscriberId] = subscriber
	} else {
		// 更新策略
		subscriber.Strategy = strategy
		// 清空之前的订阅结果
		subscriber.ClearSubscribedEntities()
	}

	// 根据策略筛选 Entities
	subscribedEntities := make([]mmeobj.IEntity, 0)
	for _, entity := range zone.Entities {
		if strategy.ShouldSubscribe(entity) {
			subscriber.SubscribeEntity(entity)
			subscribedEntities = append(subscribedEntities, entity)
		}
	}

	return subscribedEntities
}

// SubscribeByIds 使用 ID 列表订阅（便捷方法）
func (zone *ServiceZone) SubscribeByIds(subscriberId string, entityIds []int64) []mmeobj.IEntity {
	strategy := NewByIdsStrategy(entityIds)
	return zone.Subscribe(subscriberId, strategy)
}

// SubscribeByType 使用类型订阅（便捷方法）
func (zone *ServiceZone) SubscribeByType(subscriberId string, entityTypes []string) []mmeobj.IEntity {
	strategy := NewByTypeStrategy(entityTypes)
	return zone.Subscribe(subscriberId, strategy)
}

// SubscribeAll 订阅所有 Entities（便捷方法）
func (zone *ServiceZone) SubscribeAll(subscriberId string) []mmeobj.IEntity {
	strategy := NewAllEntitiesStrategy()
	return zone.Subscribe(subscriberId, strategy)
}

// Unsubscribe 取消订阅
func (zone *ServiceZone) Unsubscribe(subscriberId string) {
	subscriber, exists := zone.subscribers[subscriberId]
	if !exists {
		return
	}
	subscriber.ClearSubscribedEntities()
	delete(zone.subscribers, subscriberId)
}

// GetSubscribedEntities 获取订阅者已订阅的 Entities
func (zone *ServiceZone) GetSubscribedEntities(subscriberId string) []mmeobj.IEntity {
	subscriber, exists := zone.subscribers[subscriberId]
	if !exists {
		return nil
	}

	entities := make([]mmeobj.IEntity, 0, len(subscriber.Entities))
	subscriber.RangeSubscribedEntities(func(entityId int64) {
		entity := zone.Entities[entityId]
		// 过滤掉已删除的实体（nil 元素）
		if entity != nil {
			entities = append(entities, entity)
		}
	})
	return entities
}

// GetSubscriberStrategyType 获取订阅者的策略类型（用于网络协议序列化）
func (zone *ServiceZone) GetSubscriberStrategyType(subscriberId string) (mme.SubscribeStrategyType, bool) {
	subscriber, exists := zone.subscribers[subscriberId]
	if !exists {
		return mme.SubscribeStrategyType_All, false
	}
	return subscriber.Strategy.GetStrategyType(), true
}

// SubscribeByStrategyType 根据 pb 枚举类型和参数创建策略并订阅
// 这是一个便捷方法，用于从网络协议中接收到的枚举类型创建订阅
func (zone *ServiceZone) SubscribeByStrategyType(
	subscriberId string,
	strategyType mme.SubscribeStrategyType,
	params any,
) []mmeobj.IEntity {
	var strategy ISubscribeStrategy

	switch strategyType {
	case mme.SubscribeStrategyType_All:
		strategy = NewAllEntitiesStrategy()
	case mme.SubscribeStrategyType_Only:
		if entityTypes, ok := params.([]string); ok {
			strategy = NewOnlyStrategy(entityTypes)
		} else {
			strategy = NewOnlyStrategy(nil)
		}
	case mme.SubscribeStrategyType_ById:
		if entityIds, ok := params.([]int64); ok {
			strategy = NewByIdsStrategy(entityIds)
		} else {
			strategy = NewByIdsStrategy(nil)
		}
	case mme.SubscribeStrategyType_Composite:
		if compositeParams, ok := params.(struct {
			Strategies []ISubscribeStrategy
			Logic      CompositeLogic
		}); ok {
			strategy = NewCompositeStrategy(compositeParams.Strategies, compositeParams.Logic)
		} else {
			strategy = NewCompositeStrategy(nil, CompositeLogicAND)
		}
	case mme.SubscribeStrategyType_ByDistance:
		// TODO: 实现按距离订阅策略
		panic("ByDistance strategy is not supported")
	default:
		// 未知类型，使用全量订阅作为默认值
		strategy = NewAllEntitiesStrategy()
	}

	return zone.Subscribe(subscriberId, strategy)
}

// UpdateSubscriptions 当 Entity 添加或更新时，更新所有订阅者
// 新添加的 Entity 会根据各订阅者的策略决定是否订阅
func (zone *ServiceZone) UpdateSubscriptions(entity mmeobj.IEntity) {
	if entity == nil {
		return
	}
	for _, subscriber := range zone.subscribers {
		if subscriber.Strategy.ShouldSubscribe(entity) {
			// 添加到订阅列表
			subscriber.SubscribeEntity(entity)
		} else {
			// 如果不再符合订阅条件，从订阅列表中移除
			subscriber.UnsubscribeEntity(entity.GetXXXId())
		}
	}
}

// RemoveFromSubscriptions 当 Entity 移除时，从所有订阅者中移除
func (zone *ServiceZone) RemoveFromSubscriptions(entity mmeobj.IEntity) {
	id := entity.GetXXXId()
	for _, subscriber := range zone.subscribers {
		subscriber.UnsubscribeEntity(id)
	}
}
