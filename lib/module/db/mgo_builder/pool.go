package mgo_builder

import "sync"

// MongoUpdateBuilderPool MongoUpdateBuilder 对象池
type MongoUpdateBuilderPool struct {
	pool sync.Pool
}

// NewMongoUpdateBuilderPool 创建新的 MongoUpdateBuilder 对象池
func NewMongoUpdateBuilderPool() *MongoUpdateBuilderPool {
	return &MongoUpdateBuilderPool{
		pool: sync.Pool{
			New: func() any {
				return NewMongoUpdateBuilder()
			},
		},
	}
}

// Get 从池中获取一个 MongoUpdateBuilder
func (p *MongoUpdateBuilderPool) Get() *MongoUpdateBuilder {
	return p.pool.Get().(*MongoUpdateBuilder)
}

// Put 将 MongoUpdateBuilder 放回池中
func (p *MongoUpdateBuilderPool) Put(builder *MongoUpdateBuilder) {
	if builder != nil {
		builder.Reset()
		p.pool.Put(builder)
	}
}

// 全局对象池实例
var (
	// DefaultBuilderPool 默认的 MongoUpdateBuilder 对象池
	DefaultBuilderPool = NewMongoUpdateBuilderPool()
)

// GetBuilder 从默认池中获取一个 MongoUpdateBuilder
func GetBuilder() *MongoUpdateBuilder {
	return DefaultBuilderPool.Get()
}

// PutBuilder 将 MongoUpdateBuilder 放回默认池中
func PutBuilder(builder *MongoUpdateBuilder) {
	DefaultBuilderPool.Put(builder)
}

// WithBuilder 使用 MongoUpdateBuilder，自动管理生命周期
func WithBuilder(fn func(*MongoUpdateBuilder)) {
	builder := GetBuilder()
	defer PutBuilder(builder)
	fn(builder)
}

// BuilderFunc 定义返回值的构建器函数类型
type BuilderFunc[T any] func(*MongoUpdateBuilder) T

// WithBuilderResult 使用 MongoUpdateBuilder 并返回结果，自动管理生命周期
func WithBuilderResult[T any](fn BuilderFunc[T]) T {
	builder := GetBuilder()
	defer PutBuilder(builder)
	return fn(builder)
}
