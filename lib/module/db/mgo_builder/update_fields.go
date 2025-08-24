package mgo_builder

// UpdateFields 表示MongoDB更新操作中的字段映射
// 将 map[string]any 封装为专用类型，提供更好的类型安全性和可读性
type UpdateFields map[string]any

// NewUpdateFields 创建新的UpdateFields
func NewUpdateFields() UpdateFields {
	return make(UpdateFields)
}

// NewUpdateFieldsWithCapacity 创建指定容量的UpdateFields
func NewUpdateFieldsWithCapacity(capacity int) UpdateFields {
	return make(UpdateFields, capacity)
}

// Set 设置字段值
func (uf UpdateFields) Set(path string, value any) {
	uf[path] = value
}

// Get 获取字段值
func (uf UpdateFields) Get(path string) (any, bool) {
	value, exists := uf[path]
	return value, exists
}

// Delete 删除字段
func (uf UpdateFields) Delete(path string) {
	delete(uf, path)
}

// Has 检查字段是否存在
func (uf UpdateFields) Has(path string) bool {
	_, exists := uf[path]
	return exists
}

// Len 返回字段数量
func (uf UpdateFields) Len() int {
	return len(uf)
}

// IsEmpty 检查是否为空
func (uf UpdateFields) IsEmpty() bool {
	return len(uf) == 0
}

// Merge 合并另一个UpdateFields
func (uf UpdateFields) Merge(other UpdateFields) {
	for key, value := range other {
		uf[key] = value
	}
}

// Clone 创建副本
func (uf UpdateFields) Clone() UpdateFields {
	clone := make(UpdateFields, len(uf))
	for key, value := range uf {
		clone[key] = value
	}
	return clone
}

// Clear 清空所有字段
func (uf UpdateFields) Clear() {
	for key := range uf {
		delete(uf, key)
	}
}

// Keys 返回所有键
func (uf UpdateFields) Keys() []string {
	keys := make([]string, 0, len(uf))
	for key := range uf {
		keys = append(keys, key)
	}
	return keys
}

// Values 返回所有值
func (uf UpdateFields) Values() []any {
	values := make([]any, 0, len(uf))
	for _, value := range uf {
		values = append(values, value)
	}
	return values
}

// Range 遍历所有键值对
func (uf UpdateFields) Range(fn func(path string, value any) bool) {
	for path, value := range uf {
		if !fn(path, value) {
			break
		}
	}
}

// Filter 过滤字段，返回满足条件的新UpdateFields
func (uf UpdateFields) Filter(predicate func(path string, value any) bool) UpdateFields {
	result := NewUpdateFields()
	for path, value := range uf {
		if predicate(path, value) {
			result.Set(path, value)
		}
	}
	return result
}

// SetBatch 批量设置字段
func (uf UpdateFields) SetBatch(fields map[string]any) {
	for path, value := range fields {
		uf.Set(path, value)
	}
}
