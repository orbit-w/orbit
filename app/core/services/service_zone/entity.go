package servicezone

import "gitee.com/orbit-w/orbit/lib/module/db/mgo_builder"

type IEntity interface {
	Name() string
	InitFieldContext()
	ClearAllDirtyFlags()
	BuildMongoUpdate(builder *mgo_builder.MongoUpdateBuilder, path *mgo_builder.NestedPath)
}
