package servicezone

import (
	"gitee.com/orbit-w/orbit/lib/module/db/mgo_builder"
	mmemodel "gitee.com/orbit-w/orbit/lib/module/mme_model"
	"google.golang.org/protobuf/proto"
)

type IEntity interface {
	Name() string
	InitFieldContext()
	ClearAllDirtyFlags()
	BuildMongoUpdate(builder *mgo_builder.MongoUpdateBuilder, path *mgo_builder.NestedPath)
	ToProto() proto.Message
	FromProto(msg proto.Message)
	ToIncrementalProtoWithContext(ctx mmemodel.SyncContext) proto.Message
}
