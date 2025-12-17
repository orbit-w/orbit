package mme

import (
	"gitee.com/orbit-w/orbit/internal/game/proto/mme"
	"gitee.com/orbit-w/orbit/lib/module/db/mgo_builder"
	mmemodel "gitee.com/orbit-w/orbit/lib/module/mme_model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"google.golang.org/protobuf/proto"
)

type IEntity interface {
	Collection() string
	Load(raw bson.Raw) error
	Name() string
	GetXXXId() int64
	SetXXXId(id int64)
	GetEntityType() mme.EntityType
	InitFieldContext()
	ClearAllDirtyFlags()
	BuildMongoUpdate(builder *mgo_builder.MongoUpdateBuilder)
	ToProto() proto.Message
	FromProto(msg proto.Message)
	ToIncrementalProtoWithContext(ctx mmemodel.SyncContext) proto.Message
}

type (
	EntityFactory func() IEntity

	EntityDataFactory func() proto.Message
)

var (
	mapEntityFactories     = make(map[mme.EntityType]EntityFactory)
	mapEntityDataFactories = make(map[mme.EntityType]EntityDataFactory)
)

func RegisterEntityFactory(entityType mme.EntityType, factory EntityFactory) {
	mapEntityFactories[entityType] = factory
}

func GetEntityFactory(entityType mme.EntityType) EntityFactory {
	factory, ok := mapEntityFactories[entityType]
	if !ok {
		return nil
	}
	return factory
}

func RegisterEntityDataFactory(entityType mme.EntityType, factory EntityDataFactory) {
	mapEntityDataFactories[entityType] = factory
}

func GetEntityDataFactory(entityType mme.EntityType) EntityDataFactory {
	factory, ok := mapEntityDataFactories[entityType]
	if !ok {
		return nil
	}
	return factory
}
