package mgo_builder

import (
	"fmt"
	"maps"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
)

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
	operations map[MongoUpdateOp]bson.M
}

// NewMongoUpdateBuilder 创建新的更新构建器
func NewMongoUpdateBuilder() *MongoUpdateBuilder {
	return &MongoUpdateBuilder{
		operations: make(map[MongoUpdateOp]bson.M),
	}
}

func (b *MongoUpdateBuilder) SetNestedPath(np *NestedPath, value any) {
	b.Set(np.Build(), value)
}

func (b *MongoUpdateBuilder) IncNestedPath(np *NestedPath, value any) {
	b.Inc(np.Build(), value)
}

func (b *MongoUpdateBuilder) UnsetNestedPath(np *NestedPath) {
	b.Unset(np.Build())
}

// AddOperation 添加操作
func (b *MongoUpdateBuilder) AddOperation(op MongoUpdateOp, path string, value any) {
	if b.operations[op] == nil {
		b.operations[op] = bson.M{}
	}
	b.operations[op][path] = value
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
			b.operations[op] = bson.M{}
		}
		maps.Copy(b.operations[op], opData)
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

	result := bson.M{}
	for op, opData := range b.operations {
		if len(opData) > 0 {
			result[string(op)] = opData
		}
	}
	return result
}

// Reset 重置构建器状态，用于对象池回收
func (b *MongoUpdateBuilder) Reset() {
	for op := range b.operations {
		for key := range b.operations[op] {
			delete(b.operations[op], key)
		}
		delete(b.operations, op)
	}
}

type NestedPath struct {
	parts []string
}

func (np *NestedPath) Field(name string) *NestedPath {
	np.parts = append(np.parts, name)
	return np
}

// Index 添加索引
func (np *NestedPath) Index(index int) *NestedPath {
	np.parts = append(np.parts, fmt.Sprintf("%d", index))
	return np
}

// ArrayAll 添加数组所有元素
func (np *NestedPath) ArrayAll() *NestedPath {
	np.parts = append(np.parts, "$[]")
	return np
}

func (np *NestedPath) Build() string {
	var pathBuilder strings.Builder
	for i, part := range np.parts {
		if i > 0 {
			pathBuilder.WriteByte('.')
		}
		pathBuilder.WriteString(part)
	}
	return pathBuilder.String()
}
