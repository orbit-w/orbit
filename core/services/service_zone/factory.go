package servicezone

import (
	"fmt"

	mmeobj "gitee.com/orbit-w/orbit/app/mme"
	"gitee.com/orbit-w/orbit/app/proto/mme"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type EntityFactory func() IEntity

var entityFactories = make(map[mme.EntityType]EntityFactory)

func init() {
	RegisterEntityFactory(mme.EntityType_PlayerEntityType, func() IEntity {
		return mmeobj.NewPlayerEntityWrapper()
	})
}

func RegisterEntityFactory(entityType mme.EntityType, factory EntityFactory) {
	entityFactories[entityType] = factory
}

func LoadEntity(entityType mme.EntityType, raw bson.Raw) (IEntity, error) {
	factory, ok := entityFactories[entityType]
	if !ok {
		panic(fmt.Sprintf("entity factory not found for entity type %d", entityType))
	}
	entity := factory()
	if err := entity.Load(raw); err != nil {
		return nil, err
	}
	return entity, nil
}
