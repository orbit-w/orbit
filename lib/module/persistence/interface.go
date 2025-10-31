package persistence

import (
	"gitee.com/orbit-w/orbit/lib/module/db/mgo_builder"
)

// Wrapper 定义可以构建MongoDB更新的包装器接口
// 所有MME Wrapper类都应该实现此接口
type Wrapper interface {
	Collection() string
	// BuildMongoUpdate 构建MongoDB更新操作
	// builder: MongoDB更新构建器
	// path: 嵌套字段路径
	BuildMongoUpdate(builder *mgo_builder.MongoUpdateBuilder, path *mgo_builder.NestedPath)
}
