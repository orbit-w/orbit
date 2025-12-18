package servicezone

import (
	mmeobj "gitee.com/orbit-w/orbit/internal/game/mme"
	"gitee.com/orbit-w/orbit/pkg/proto/mme"
)

// ISubscribeStrategy 订阅策略接口
// 用于定义不同的 Entity 订阅策略
type ISubscribeStrategy interface {
	// ShouldSubscribe 判断是否应该订阅指定的 Entity
	// entity: 要判断的 Entity
	// 返回: true 表示应该订阅，false 表示不订阅
	ShouldSubscribe(entity mmeobj.IEntity) bool

	// GetStrategyName 返回策略名称，用于日志和调试
	GetStrategyName() string

	// GetStrategyType 返回策略类型，对应 pb 中的 SubscribeStrategyType 枚举
	GetStrategyType() mme.SubscribeStrategyType
}

// AllEntitiesStrategy 全量订阅策略
// 订阅所有 Entities
type AllEntitiesStrategy struct{}

func NewAllEntitiesStrategy() *AllEntitiesStrategy {
	return &AllEntitiesStrategy{}
}

func (s *AllEntitiesStrategy) ShouldSubscribe(entity mmeobj.IEntity) bool {
	return true
}

func (s *AllEntitiesStrategy) GetStrategyName() string {
	return "AllEntities"
}

func (s *AllEntitiesStrategy) GetStrategyType() mme.SubscribeStrategyType {
	return mme.SubscribeStrategyType_All
}

// OnlyStrategy 按类型订阅策略（对应 pb 中的 Only）
// 只订阅指定类型的 Entities
type OnlyStrategy struct {
	AllowedTypes map[string]bool // 允许的 Entity 类型名称集合
}

func NewOnlyStrategy(allowedTypes []string) *OnlyStrategy {
	typeMap := make(map[string]bool)
	for _, t := range allowedTypes {
		typeMap[t] = true
	}
	return &OnlyStrategy{
		AllowedTypes: typeMap,
	}
}

func (s *OnlyStrategy) ShouldSubscribe(entity mmeobj.IEntity) bool {
	if entity == nil {
		return false
	}
	entityType := entity.Name()
	return s.AllowedTypes[entityType]
}

func (s *OnlyStrategy) GetStrategyName() string {
	return "Only"
}

func (s *OnlyStrategy) GetStrategyType() mme.SubscribeStrategyType {
	return mme.SubscribeStrategyType_Only
}

// ByTypeStrategy 按类型订阅策略（已废弃，使用 OnlyStrategy）
// Deprecated: 使用 OnlyStrategy 代替
type ByTypeStrategy = OnlyStrategy

// NewByTypeStrategy 创建按类型订阅策略（已废弃，使用 NewOnlyStrategy）
// Deprecated: 使用 NewOnlyStrategy 代替
func NewByTypeStrategy(allowedTypes []string) *ByTypeStrategy {
	return NewOnlyStrategy(allowedTypes)
}

// ByIdsStrategy 按 ID 列表订阅策略
// 只订阅指定 ID 的 Entities
type ByIdsStrategy struct {
	AllowedIds map[int64]bool // 允许的 Entity ID 集合
}

func NewByIdsStrategy(allowedIds []int64) *ByIdsStrategy {
	idMap := make(map[int64]bool)
	for _, id := range allowedIds {
		idMap[id] = true
	}
	return &ByIdsStrategy{
		AllowedIds: idMap,
	}
}

func (s *ByIdsStrategy) ShouldSubscribe(entity mmeobj.IEntity) bool {
	if entity == nil {
		return false
	}
	entityId := entity.GetXXXId()
	return s.AllowedIds[entityId]
}

func (s *ByIdsStrategy) ShouldSubscribeById(entityId int64) bool {
	return s.AllowedIds[entityId]
}

func (s *ByIdsStrategy) GetStrategyName() string {
	return "ByIds"
}

func (s *ByIdsStrategy) GetStrategyType() mme.SubscribeStrategyType {
	return mme.SubscribeStrategyType_ById
}

// ByDistanceStrategy 按距离订阅策略
// 只订阅距离指定位置一定范围内的 Entities
// 注意：需要 Entity 支持位置信息，这里提供基础框架
type ByDistanceStrategy struct {
	CenterX     float64 // 中心点 X 坐标
	CenterY     float64 // 中心点 Y 坐标
	MaxDistance float64 // 最大距离
}

func NewByDistanceStrategy(centerX, centerY, maxDistance float64) *ByDistanceStrategy {
	return &ByDistanceStrategy{
		CenterX:     centerX,
		CenterY:     centerY,
		MaxDistance: maxDistance,
	}
}

func (s *ByDistanceStrategy) ShouldSubscribe(entity mmeobj.IEntity) bool {
	// 注意：需要 Entity 实现位置接口才能使用此策略
	// 这里提供框架，实际使用时需要扩展 IEntity 接口或使用类型断言
	// 例如：
	// if posEntity, ok := entity.(IPositionEntity); ok {
	//     distance := calculateDistance(posEntity.GetX(), posEntity.GetY(), s.CenterX, s.CenterY)
	//     return distance <= s.MaxDistance
	// }
	return false
}

func (s *ByDistanceStrategy) GetStrategyName() string {
	return "ByDistance"
}

func (s *ByDistanceStrategy) GetStrategyType() mme.SubscribeStrategyType {
	return mme.SubscribeStrategyType_ByDistance
}

// CompositeStrategy 组合策略
// 支持多个策略的组合（AND 或 OR 逻辑）
type CompositeStrategy struct {
	Strategies []ISubscribeStrategy
	Logic      CompositeLogic // 组合逻辑：AND 或 OR
}

type CompositeLogic int

const (
	CompositeLogicAND CompositeLogic = iota // 所有策略都必须满足
	CompositeLogicOR                        // 任一策略满足即可
)

func NewCompositeStrategy(strategies []ISubscribeStrategy, logic CompositeLogic) *CompositeStrategy {
	return &CompositeStrategy{
		Strategies: strategies,
		Logic:      logic,
	}
}

func (s *CompositeStrategy) ShouldSubscribe(entity mmeobj.IEntity) bool {
	if len(s.Strategies) == 0 {
		return false
	}

	if s.Logic == CompositeLogicAND {
		// AND 逻辑：所有策略都必须返回 true
		for _, strategy := range s.Strategies {
			if !strategy.ShouldSubscribe(entity) {
				return false
			}
		}
		return true
	} else {
		// OR 逻辑：任一策略返回 true 即可
		for _, strategy := range s.Strategies {
			if strategy.ShouldSubscribe(entity) {
				return true
			}
		}
		return false
	}
}

func (s *CompositeStrategy) GetStrategyName() string {
	return "Composite"
}

func (s *CompositeStrategy) GetStrategyType() mme.SubscribeStrategyType {
	return mme.SubscribeStrategyType_Composite
}

// NewStrategyFromType 根据 pb 枚举类型创建策略实例
// 注意：某些策略类型（如 Only、ById、ByDistance、Composite）需要额外的参数，
// 这些参数需要通过具体的构造函数提供，此函数仅用于创建基础策略
func NewStrategyFromType(strategyType mme.SubscribeStrategyType) ISubscribeStrategy {
	switch strategyType {
	case mme.SubscribeStrategyType_All:
		return NewAllEntitiesStrategy()
	case mme.SubscribeStrategyType_Only:
		// Only 策略需要指定类型列表，这里返回空列表的策略
		// 实际使用时应该调用 NewOnlyStrategy(allowedTypes)
		return NewOnlyStrategy(nil)
	case mme.SubscribeStrategyType_ById:
		// ById 策略需要指定 ID 列表，这里返回空列表的策略
		// 实际使用时应该调用 NewByIdsStrategy(allowedIds)
		return NewByIdsStrategy(nil)
	case mme.SubscribeStrategyType_ByDistance:
		// ByDistance 策略需要指定位置和距离，这里返回默认值
		// 实际使用时应该调用 NewByDistanceStrategy(centerX, centerY, maxDistance)
		return NewByDistanceStrategy(0, 0, 0)
	case mme.SubscribeStrategyType_Composite:
		// Composite 策略需要指定子策略列表，这里返回空列表的策略
		// 实际使用时应该调用 NewCompositeStrategy(strategies, logic)
		return NewCompositeStrategy(nil, CompositeLogicAND)
	default:
		// 未知类型，返回全量订阅策略作为默认值
		return NewAllEntitiesStrategy()
	}
}
