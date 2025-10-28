package fieldmeta

// FieldType 字段访问类型标记
// 支持0-127共128种不同的字段类型
type FieldType uint8

const (
	// 预定义的常用字段类型

	// FieldTypeSync 可同步字段（可以同步到客户端）
	FieldTypeSync FieldType = iota
	// FieldTypeReadOnly 只读字段（不可修改）
	FieldTypeReadOnly
	// FieldTypeClientModifiable 客户端可修改字段
	FieldTypeClientModifiable
	// FieldTypePersistent 持久化字段（需要保存到数据库）
	FieldTypePersistent
	// FieldTypePrivate 私有字段（仅服务器使用）
	FieldTypePrivate
	// FieldTypeComputed 计算字段（由其他字段计算得出）
	FieldTypeComputed
	// FieldTypeValidated 需要验证的字段
	FieldTypeValidated
	// FieldTypeEncrypted 加密字段
	FieldTypeEncrypted

	// 预留更多类型供业务使用 (8-127)
	// 用户可以根据需要定义自己的 FieldType
	// 例如: const FieldTypeCustom1 FieldType = 8
)

const (
	// MaxFieldTypes 最大支持的字段类型数量（128种）
	// FieldType 使用 uint8，范围是 0-127
	MaxFieldTypes = 128

	// MaxFieldID 最大支持的字段ID（64个字段，ID范围0-63）
	MaxFieldID = 64
)

// FieldMetas 字段元数据管理器
// 使用位图方案高效管理多个字段的多种状态
//
// 容量规格：
// - 支持最多 64 个字段（字段ID: 0-63）
// - 支持最多 128 种字段类型（类型ID: 0-127）
// - 内存占用: 128 * 8 = 1024 bytes (1KB)
//
// 性能特性：
// - 所有单字段操作都是 O(1) 时间复杂度
// - 零内存分配
// - CPU缓存友好
type FieldMetas struct {
	// masks 存储不同类型的位掩码
	// masks[i] 表示第i种FieldType的位掩码
	// 每个位对应一个字段ID（0-63）
	masks [MaxFieldTypes]uint64
}

// NewFieldMetas 创建一个新的字段元数据管理器
func NewFieldMetas() *FieldMetas {
	return &FieldMetas{}
}

// SetFieldType 为指定字段设置某种类型标记
// fieldID: 字段ID (0-63)
// fieldType: 字段类型 (0-127)
func (fm *FieldMetas) SetFieldType(fieldID uint8, fieldType FieldType) {
	if fieldID >= MaxFieldID {
		return
	}
	fm.masks[fieldType] |= (1 << fieldID)
}

// UnsetFieldType 为指定字段取消某种类型标记
// fieldID: 字段ID (0-63)
// fieldType: 字段类型 (0-127)
func (fm *FieldMetas) UnsetFieldType(fieldID uint8, fieldType FieldType) {
	if fieldID >= MaxFieldID {
		return
	}
	fm.masks[fieldType] &^= (1 << fieldID)
}

// HasFieldType 检查指定字段是否具有某种类型标记
// fieldID: 字段ID (0-63)
// fieldType: 字段类型 (0-127)
func (fm *FieldMetas) HasFieldType(fieldID uint8, fieldType FieldType) bool {
	if fieldID >= MaxFieldID {
		return false
	}
	return (fm.masks[fieldType] & (1 << fieldID)) != 0
}

// ToggleFieldType 切换指定字段的某种类型标记
// fieldID: 字段ID (0-63)
// fieldType: 字段类型 (0-127)
func (fm *FieldMetas) ToggleFieldType(fieldID uint8, fieldType FieldType) {
	if fieldID >= MaxFieldID {
		return
	}
	fm.masks[fieldType] ^= (1 << fieldID)
}

// SetMultipleFieldTypes 为指定字段同时设置多种类型标记
// fieldID: 字段ID (0-63)
// fieldTypes: 字段类型列表
func (fm *FieldMetas) SetMultipleFieldTypes(fieldID uint8, fieldTypes ...FieldType) {
	if fieldID >= MaxFieldID {
		return
	}
	for _, ft := range fieldTypes {
		fm.masks[ft] |= (1 << fieldID)
	}
}

// ClearField 清除指定字段的所有类型标记
// fieldID: 字段ID (0-63)
func (fm *FieldMetas) ClearField(fieldID uint8) {
	if fieldID >= MaxFieldID {
		return
	}
	mask := ^(uint64(1) << fieldID)
	for i := range fm.masks {
		fm.masks[i] &= mask
	}
}

// GetFieldTypes 获取指定字段的所有类型标记
// fieldID: 字段ID (0-63)
// 返回该字段具有的所有类型标记
func (fm *FieldMetas) GetFieldTypes(fieldID uint8) []FieldType {
	if fieldID >= MaxFieldID {
		return nil
	}
	types := make([]FieldType, 0, 8) // 预估大部分字段不会有太多类型
	// 遍历所有masks数组
	for i := range fm.masks {
		if (fm.masks[i] & (1 << fieldID)) != 0 {
			types = append(types, FieldType(i))
		}
	}
	return types
}

// GetFieldsWithType 获取具有指定类型标记的所有字段ID
// fieldType: 字段类型 (0-127)
// 返回所有具有该类型标记的字段ID列表
func (fm *FieldMetas) GetFieldsWithType(fieldType FieldType) []uint8 {
	fields := make([]uint8, 0, MaxFieldID)
	mask := fm.masks[fieldType]
	for i := uint8(0); i < MaxFieldID; i++ {
		if (mask & (1 << i)) != 0 {
			fields = append(fields, i)
		}
	}
	return fields
}

// CountFieldsWithType 统计具有指定类型标记的字段数量
// fieldType: 字段类型 (0-127)
func (fm *FieldMetas) CountFieldsWithType(fieldType FieldType) int {
	return popcount(fm.masks[fieldType])
}

// HasAnyFieldWithType 检查是否存在任何具有指定类型标记的字段
// fieldType: 字段类型 (0-127)
func (fm *FieldMetas) HasAnyFieldWithType(fieldType FieldType) bool {
	return fm.masks[fieldType] != 0
}

// GetMask 获取指定类型的位掩码（用于批量操作）
// fieldType: 字段类型 (0-127)
func (fm *FieldMetas) GetMask(fieldType FieldType) uint64 {
	return fm.masks[fieldType]
}

// SetMask 直接设置指定类型的位掩码（用于批量操作）
// fieldType: 字段类型 (0-127)
// mask: 位掩码
func (fm *FieldMetas) SetMask(fieldType FieldType, mask uint64) {
	fm.masks[fieldType] = mask
}

// MergeMask 合并指定类型的位掩码（用于批量操作）
// fieldType: 字段类型 (0-127)
// mask: 位掩码
func (fm *FieldMetas) MergeMask(fieldType FieldType, mask uint64) {
	fm.masks[fieldType] |= mask
}

// IntersectMask 与指定类型的位掩码做交集（用于批量操作）
// fieldType: 字段类型 (0-127)
// mask: 位掩码
func (fm *FieldMetas) IntersectMask(fieldType FieldType, mask uint64) {
	fm.masks[fieldType] &= mask
}

// ClearMask 清空指定类型的位掩码
// fieldType: 字段类型 (0-127)
func (fm *FieldMetas) ClearMask(fieldType FieldType) {
	fm.masks[fieldType] = 0
}

// ClearAll 清空所有位掩码
func (fm *FieldMetas) ClearAll() {
	for i := range fm.masks {
		fm.masks[i] = 0
	}
}

// Clone 克隆一个新的FieldMetas
func (fm *FieldMetas) Clone() *FieldMetas {
	clone := &FieldMetas{}
	copy(clone.masks[:], fm.masks[:])
	return clone
}

// MatchesAll 检查指定字段是否匹配所有给定的类型标记
// fieldID: 字段ID (0-63)
// fieldTypes: 字段类型列表
func (fm *FieldMetas) MatchesAll(fieldID uint8, fieldTypes ...FieldType) bool {
	if fieldID >= MaxFieldID {
		return false
	}
	for _, ft := range fieldTypes {
		if (fm.masks[ft] & (1 << fieldID)) == 0 {
			return false
		}
	}
	return true
}

// MatchesAny 检查指定字段是否匹配任一给定的类型标记
// fieldID: 字段ID (0-63)
// fieldTypes: 字段类型列表
func (fm *FieldMetas) MatchesAny(fieldID uint8, fieldTypes ...FieldType) bool {
	if fieldID >= MaxFieldID {
		return false
	}
	for _, ft := range fieldTypes {
		if (fm.masks[ft] & (1 << fieldID)) != 0 {
			return true
		}
	}
	return false
}

// GetIntersection 获取多种类型标记的交集字段
// fieldTypes: 字段类型列表
// 返回同时具有所有这些类型标记的字段ID列表
func (fm *FieldMetas) GetIntersection(fieldTypes ...FieldType) []uint8 {
	if len(fieldTypes) == 0 {
		return nil
	}

	// 计算所有类型掩码的交集
	var intersection uint64 = 0xFFFFFFFFFFFFFFFF
	for _, ft := range fieldTypes {
		intersection &= fm.masks[ft]
	}

	// 提取字段ID
	fields := make([]uint8, 0, MaxFieldID)
	for i := uint8(0); i < MaxFieldID; i++ {
		if (intersection & (1 << i)) != 0 {
			fields = append(fields, i)
		}
	}
	return fields
}

// GetUnion 获取多种类型标记的并集字段
// fieldTypes: 字段类型列表
// 返回至少具有其中一种类型标记的字段ID列表
func (fm *FieldMetas) GetUnion(fieldTypes ...FieldType) []uint8 {
	if len(fieldTypes) == 0 {
		return nil
	}

	// 计算所有类型掩码的并集
	var union uint64 = 0
	for _, ft := range fieldTypes {
		union |= fm.masks[ft]
	}

	// 提取字段ID
	fields := make([]uint8, 0, MaxFieldID)
	for i := uint8(0); i < MaxFieldID; i++ {
		if (union & (1 << i)) != 0 {
			fields = append(fields, i)
		}
	}
	return fields
}

// ForEachFieldWithType 遍历所有具有指定类型标记的字段
// fieldType: 字段类型 (0-127)
// fn: 回调函数，返回false时停止遍历
func (fm *FieldMetas) ForEachFieldWithType(fieldType FieldType, fn func(fieldID uint8) bool) {
	if fn == nil {
		return
	}
	mask := fm.masks[fieldType]
	for i := uint8(0); i < MaxFieldID && mask != 0; i++ {
		if (mask & 1) != 0 {
			if !fn(i) {
				return
			}
		}
		mask >>= 1
	}
}

// popcount 计算uint64中1的个数（汉明重量）
// 使用MIT HAKMEM算法的变体，高效计算
func popcount(x uint64) int {
	// 使用分治法计算
	x = x - ((x >> 1) & 0x5555555555555555)
	x = (x & 0x3333333333333333) + ((x >> 2) & 0x3333333333333333)
	x = (x + (x >> 4)) & 0x0f0f0f0f0f0f0f0f
	x = x + (x >> 8)
	x = x + (x >> 16)
	x = x + (x >> 32)
	return int(x & 0x7f)
}
