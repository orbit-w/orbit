package servicezone

import (
	"gitee.com/orbit-w/orbit/app/proto/enum"
	"gitee.com/orbit-w/orbit/lib/module/db/mgo_builder"
	mmemodel "gitee.com/orbit-w/orbit/lib/module/mme_model"
	"google.golang.org/protobuf/proto"
)

type IEntity interface {
	Name() string
	// 获取 Entity ID（XXXId）
	GetXXXId() int64
	// 获取 Entity 类型
	GetEntityType() enum.EntityType
	// 初始化字段上下文
	InitFieldContext()
	// 清除所有脏标记
	ClearAllDirtyFlags()
	// 构建 MongoDB 更新操作
	BuildMongoUpdate(builder *mgo_builder.MongoUpdateBuilder, path *mgo_builder.NestedPath)
	// 将 Entity 数据转换为完整的 protobuf 结构体
	ToProto() proto.Message
	// 从 protobuf 结构体加载数据到 Entity
	FromProto(msg proto.Message)
	// 根据脏标记位构建增量数据的 protoMessage，用于增量同步
	ToIncrementalProtoWithContext(ctx mmemodel.SyncContext) proto.Message
}
