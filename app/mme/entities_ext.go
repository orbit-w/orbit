package mme

import (
	"gitee.com/orbit-w/orbit/app/proto/mme"
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
	BuildMongoUpdate(builder *mgo_builder.MongoUpdateBuilder, path *mgo_builder.NestedPath)
	ToProto() proto.Message
	FromProto(msg proto.Message)
	ToIncrementalProtoWithContext(ctx mmemodel.SyncContext) proto.Message
}

type EntityFactory func() IEntity

var (
	mapEntityFactories = make(map[mme.EntityType]EntityFactory)
)

func init() {
	RegisterEntityFactory(mme.EntityType_PlayerEntityType, func() IEntity {
		return NewPlayerEntityWrapper()
	})
}

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
