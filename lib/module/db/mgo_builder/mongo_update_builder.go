package mgo_builder

// MongoUpdateOp 定义 MongoDB 更新操作类型
type MongoUpdateOp string

const (
	OpSet      MongoUpdateOp = "$set"
	OpUnset    MongoUpdateOp = "$unset"
	OpInc      MongoUpdateOp = "$inc"
	OpAddToSet MongoUpdateOp = "$addToSet"
)

// MongoUpdateBuilder 高效的 MongoDB 更新构建器
type MongoUpdateBuilder struct {
	operations map[MongoUpdateOp]UpdateFields
}

// NewMongoUpdateBuilder 创建新的更新构建器
func NewMongoUpdateBuilder() *MongoUpdateBuilder {
	return &MongoUpdateBuilder{
		operations: make(map[MongoUpdateOp]UpdateFields),
	}
}

// AddOperation 添加操作
func (b *MongoUpdateBuilder) AddOperation(op MongoUpdateOp, path string, value any) {
	if b.operations[op] == nil {
		b.operations[op] = NewUpdateFields()
	}
	b.operations[op].Set(path, value)
}

// Set 添加 $set 操作
func (b *MongoUpdateBuilder) Set(path string, value any) {
	b.AddOperation(OpSet, path, value)
}

// Unset 添加 $unset 操作
func (b *MongoUpdateBuilder) Unset(path string) {
	b.AddOperation(OpUnset, path, "")
}

// Inc 添加 $inc 操作
func (b *MongoUpdateBuilder) Inc(path string, value any) {
	b.AddOperation(OpInc, path, value)
}

// Merge 合并另一个构建器的操作
func (b *MongoUpdateBuilder) Merge(other *MongoUpdateBuilder) {
	if other == nil {
		return
	}

	for op, opData := range other.operations {
		if b.operations[op] == nil {
			b.operations[op] = NewUpdateFields()
		}
		b.operations[op].Merge(opData)
	}
}

// IsEmpty 检查是否为空
func (b *MongoUpdateBuilder) IsEmpty() bool {
	return len(b.operations) == 0
}

// Build 构建最终的 MongoDB 更新文档
func (b *MongoUpdateBuilder) Build() map[string]any {
	if b.IsEmpty() {
		return nil
	}

	result := make(map[string]any)
	for op, opData := range b.operations {
		if !opData.IsEmpty() {
			result[string(op)] = map[string]any(opData)
		}
	}
	return result
}

// Reset 重置构建器状态，用于对象池回收
func (b *MongoUpdateBuilder) Reset() {
	for op := range b.operations {
		b.operations[op].Clear()
		delete(b.operations, op)
	}
}

// MapChangeTracker 跟踪 Map 字段的细粒度变更
type MapChangeTracker struct {
	SetKeys   UpdateFields    // 设置的键值对
	UnsetKeys map[string]bool // 删除的键
	IncKeys   UpdateFields    // 递增的键值对
}

// NewMapChangeTracker 创建新的 Map 变更跟踪器
func NewMapChangeTracker() *MapChangeTracker {
	return &MapChangeTracker{
		SetKeys:   NewUpdateFields(),
		UnsetKeys: make(map[string]bool),
		IncKeys:   NewUpdateFields(),
	}
}

// TrackSet 跟踪设置操作
func (t *MapChangeTracker) TrackSet(key string, value any) {
	t.SetKeys.Set(key, value)
	delete(t.UnsetKeys, key) // 如果之前标记删除，现在取消
}

// TrackUnset 跟踪删除操作
func (t *MapChangeTracker) TrackUnset(key string) {
	t.UnsetKeys[key] = true
	t.SetKeys.Delete(key) // 如果之前标记设置，现在取消
}

// TrackInc 跟踪递增操作
func (t *MapChangeTracker) TrackInc(key string, value any) {
	t.IncKeys.Set(key, value)
}

// IsEmpty 检查是否为空
func (t *MapChangeTracker) IsEmpty() bool {
	return t.SetKeys.IsEmpty() && len(t.UnsetKeys) == 0 && t.IncKeys.IsEmpty()
}

// ApplyToBuilder 将跟踪的变更应用到构建器
func (t *MapChangeTracker) ApplyToBuilder(builder *MongoUpdateBuilder, pathPrefix string) {
	for key, value := range t.SetKeys {
		path := pathPrefix + "." + key
		builder.Set(path, value)
	}

	for key := range t.UnsetKeys {
		path := pathPrefix + "." + key
		builder.Unset(path)
	}

	for key, value := range t.IncKeys {
		path := pathPrefix + "." + key
		builder.Inc(path, value)
	}
}

// Reset 重置跟踪器状态，用于对象池回收
func (t *MapChangeTracker) Reset() {
	t.SetKeys.Clear()
	for key := range t.UnsetKeys {
		delete(t.UnsetKeys, key)
	}
	t.IncKeys.Clear()
}
