package fieldmeta

import (
	"testing"
)

func TestFieldMetas_Basic(t *testing.T) {
	fm := NewFieldMetas()

	// 测试设置单个字段类型
	fm.SetFieldType(0, FieldTypeSync)
	if !fm.HasFieldType(0, FieldTypeSync) {
		t.Errorf("Expected field 0 to have FieldTypeSync")
	}

	// 测试取消字段类型
	fm.UnsetFieldType(0, FieldTypeSync)
	if fm.HasFieldType(0, FieldTypeSync) {
		t.Errorf("Expected field 0 to not have FieldTypeSync after unset")
	}

	// 测试切换字段类型
	fm.ToggleFieldType(1, FieldTypeReadOnly)
	if !fm.HasFieldType(1, FieldTypeReadOnly) {
		t.Errorf("Expected field 1 to have FieldTypeReadOnly after toggle")
	}
	fm.ToggleFieldType(1, FieldTypeReadOnly)
	if fm.HasFieldType(1, FieldTypeReadOnly) {
		t.Errorf("Expected field 1 to not have FieldTypeReadOnly after second toggle")
	}
}

func TestFieldMetas_MultipleTypes(t *testing.T) {
	fm := NewFieldMetas()

	// 为同一个字段设置多种类型
	fm.SetMultipleFieldTypes(5, FieldTypeSync, FieldTypePersistent, FieldTypeValidated)

	// 验证所有类型都已设置
	if !fm.MatchesAll(5, FieldTypeSync, FieldTypePersistent, FieldTypeValidated) {
		t.Errorf("Expected field 5 to have all three types")
	}

	// 验证至少有一个类型
	if !fm.MatchesAny(5, FieldTypeReadOnly, FieldTypeSync) {
		t.Errorf("Expected field 5 to have at least one of the types")
	}

	// 获取字段的所有类型
	types := fm.GetFieldTypes(5)
	if len(types) != 3 {
		t.Errorf("Expected field 5 to have 3 types, got %d", len(types))
	}
}

func TestFieldMetas_GetFieldsWithType(t *testing.T) {
	fm := NewFieldMetas()

	// 设置多个字段为可同步类型
	fm.SetFieldType(0, FieldTypeSync)
	fm.SetFieldType(5, FieldTypeSync)
	fm.SetFieldType(10, FieldTypeSync)
	fm.SetFieldType(63, FieldTypeSync)

	// 获取所有可同步字段
	fields := fm.GetFieldsWithType(FieldTypeSync)
	if len(fields) != 4 {
		t.Errorf("Expected 4 sync fields, got %d", len(fields))
	}

	// 验证字段ID正确
	expectedFields := []uint8{0, 5, 10, 63}
	for i, expected := range expectedFields {
		if i >= len(fields) || fields[i] != expected {
			t.Errorf("Expected field %d to be %d, got %d", i, expected, fields[i])
		}
	}
}

func TestFieldMetas_CountFields(t *testing.T) {
	fm := NewFieldMetas()

	// 设置多个字段
	for i := uint8(0); i < 10; i++ {
		fm.SetFieldType(i, FieldTypePersistent)
	}

	// 统计数量
	count := fm.CountFieldsWithType(FieldTypePersistent)
	if count != 10 {
		t.Errorf("Expected 10 persistent fields, got %d", count)
	}

	// 检查是否有任何字段
	if !fm.HasAnyFieldWithType(FieldTypePersistent) {
		t.Errorf("Expected to have persistent fields")
	}

	// 清空并检查
	fm.ClearMask(FieldTypePersistent)
	if fm.HasAnyFieldWithType(FieldTypePersistent) {
		t.Errorf("Expected no persistent fields after clear")
	}
}

func TestFieldMetas_MaskOperations(t *testing.T) {
	fm := NewFieldMetas()

	// 直接设置掩码
	mask := uint64(0b1010101010) // 字段0,2,4,6,8
	fm.SetMask(FieldTypeSync, mask)

	// 验证掩码正确
	if fm.GetMask(FieldTypeSync) != mask {
		t.Errorf("Expected mask %b, got %b", mask, fm.GetMask(FieldTypeSync))
	}

	// 合并掩码
	additionalMask := uint64(0b0101010101) // 字段1,3,5,7,9
	fm.MergeMask(FieldTypeSync, additionalMask)

	expectedMerged := mask | additionalMask
	if fm.GetMask(FieldTypeSync) != expectedMerged {
		t.Errorf("Expected merged mask %b, got %b", expectedMerged, fm.GetMask(FieldTypeSync))
	}

	// 交集操作
	intersectMask := uint64(0b1111000000) // 字段6,7,8,9
	fm.IntersectMask(FieldTypeSync, intersectMask)

	expectedIntersect := expectedMerged & intersectMask
	if fm.GetMask(FieldTypeSync) != expectedIntersect {
		t.Errorf("Expected intersect mask %b, got %b", expectedIntersect, fm.GetMask(FieldTypeSync))
	}
}

func TestFieldMetas_IntersectionAndUnion(t *testing.T) {
	fm := NewFieldMetas()

	// 字段0,1,2 是可同步的
	fm.SetFieldType(0, FieldTypeSync)
	fm.SetFieldType(1, FieldTypeSync)
	fm.SetFieldType(2, FieldTypeSync)

	// 字段1,2,3 是持久化的
	fm.SetFieldType(1, FieldTypePersistent)
	fm.SetFieldType(2, FieldTypePersistent)
	fm.SetFieldType(3, FieldTypePersistent)

	// 交集：既可同步又持久化的字段（1,2）
	intersection := fm.GetIntersection(FieldTypeSync, FieldTypePersistent)
	if len(intersection) != 2 || intersection[0] != 1 || intersection[1] != 2 {
		t.Errorf("Expected intersection [1,2], got %v", intersection)
	}

	// 并集：可同步或持久化的字段（0,1,2,3）
	union := fm.GetUnion(FieldTypeSync, FieldTypePersistent)
	if len(union) != 4 {
		t.Errorf("Expected union length 4, got %d", len(union))
	}
	expectedUnion := []uint8{0, 1, 2, 3}
	for i, expected := range expectedUnion {
		if i >= len(union) || union[i] != expected {
			t.Errorf("Expected union field %d to be %d, got %d", i, expected, union[i])
		}
	}
}

func TestFieldMetas_ClearField(t *testing.T) {
	fm := NewFieldMetas()

	// 为字段5设置多种类型
	fm.SetMultipleFieldTypes(5, FieldTypeSync, FieldTypePersistent, FieldTypeReadOnly)

	// 清除字段5的所有类型
	fm.ClearField(5)

	// 验证所有类型都已清除
	if fm.MatchesAny(5, FieldTypeSync, FieldTypePersistent, FieldTypeReadOnly) {
		t.Errorf("Expected field 5 to have no types after clear")
	}
}

func TestFieldMetas_Clone(t *testing.T) {
	fm := NewFieldMetas()

	// 设置一些字段
	fm.SetFieldType(0, FieldTypeSync)
	fm.SetFieldType(1, FieldTypePersistent)

	// 克隆
	clone := fm.Clone()

	// 验证克隆包含相同的数据
	if !clone.HasFieldType(0, FieldTypeSync) {
		t.Errorf("Expected clone to have field 0 as sync")
	}
	if !clone.HasFieldType(1, FieldTypePersistent) {
		t.Errorf("Expected clone to have field 1 as persistent")
	}

	// 修改原始对象，验证克隆不受影响
	fm.SetFieldType(2, FieldTypeReadOnly)
	if clone.HasFieldType(2, FieldTypeReadOnly) {
		t.Errorf("Expected clone to be independent of original")
	}
}

func TestFieldMetas_ForEachFieldWithType(t *testing.T) {
	fm := NewFieldMetas()

	// 设置多个字段
	expectedFields := []uint8{1, 3, 5, 7, 9}
	for _, field := range expectedFields {
		fm.SetFieldType(field, FieldTypeSync)
	}

	// 遍历所有可同步字段
	visited := make([]uint8, 0, len(expectedFields))
	fm.ForEachFieldWithType(FieldTypeSync, func(fieldID uint8) bool {
		visited = append(visited, fieldID)
		return true
	})

	// 验证遍历结果
	if len(visited) != len(expectedFields) {
		t.Errorf("Expected to visit %d fields, visited %d", len(expectedFields), len(visited))
	}
	for i, expected := range expectedFields {
		if i >= len(visited) || visited[i] != expected {
			t.Errorf("Expected visited field %d to be %d, got %d", i, expected, visited[i])
		}
	}

	// 测试提前停止遍历
	count := 0
	fm.ForEachFieldWithType(FieldTypeSync, func(fieldID uint8) bool {
		count++
		return count < 3 // 只遍历前3个
	})
	if count != 3 {
		t.Errorf("Expected to visit 3 fields before stopping, visited %d", count)
	}
}

func TestFieldMetas_ClearAll(t *testing.T) {
	fm := NewFieldMetas()

	// 设置多个字段和类型
	fm.SetFieldType(0, FieldTypeSync)
	fm.SetFieldType(1, FieldTypePersistent)
	fm.SetFieldType(2, FieldTypeReadOnly)

	// 清空所有
	fm.ClearAll()

	// 验证所有掩码都已清空（测试前8种预定义类型）
	for ft := FieldType(0); ft < 8; ft++ {
		if fm.HasAnyFieldWithType(ft) {
			t.Errorf("Expected no fields with type %d after ClearAll", ft)
		}
	}
	// 额外测试几个高位类型
	for _, ft := range []FieldType{100, 120, 127} {
		if fm.HasAnyFieldWithType(ft) {
			t.Errorf("Expected no fields with type %d after ClearAll", ft)
		}
	}
}

func TestFieldMetas_BoundaryConditions(t *testing.T) {
	fm := NewFieldMetas()

	// 测试最大字段ID (63)
	fm.SetFieldType(63, FieldTypeSync)
	if !fm.HasFieldType(63, FieldTypeSync) {
		t.Errorf("Expected field 63 to work correctly")
	}

	// 测试超出范围的字段ID (应该被忽略)
	fm.SetFieldType(64, FieldTypeSync)
	fm.SetFieldType(255, FieldTypeSync)

	// 验证不会崩溃且不影响其他数据
	if !fm.HasFieldType(63, FieldTypeSync) {
		t.Errorf("Out of bounds operations should not affect valid data")
	}

	// 测试最大字段类型 (127)
	fm.SetFieldType(0, FieldType(127))
	if !fm.HasFieldType(0, FieldType(127)) {
		t.Errorf("Expected field type 127 to work correctly")
	}
}

// TestFieldMetas_Extended128Types 测试扩展的128种字段类型
func TestFieldMetas_Extended128Types(t *testing.T) {
	fm := NewFieldMetas()

	// 测试所有128种类型都能正常工作
	testTypes := []FieldType{
		0, 1, 7, // 预定义类型
		8, 50, 100, // 自定义类型
		120, 125, 127, // 高位类型
	}

	// 为字段0设置多种类型
	for _, ft := range testTypes {
		fm.SetFieldType(0, ft)
	}

	// 验证所有类型都已设置
	for _, ft := range testTypes {
		if !fm.HasFieldType(0, ft) {
			t.Errorf("Expected field 0 to have type %d", ft)
		}
	}

	// 测试清除高位类型
	fm.UnsetFieldType(0, FieldType(127))
	if fm.HasFieldType(0, FieldType(127)) {
		t.Errorf("Expected field 0 to not have type 127 after unset")
	}

	// 测试GetFieldTypes能正确返回多种类型
	types := fm.GetFieldTypes(0)
	if len(types) < len(testTypes)-1 { // -1 因为我们刚刚unset了一个
		t.Errorf("Expected at least %d types for field 0, got %d", len(testTypes)-1, len(types))
	}

	// 创建新的FieldMetas实例来测试不同字段使用不同的高位类型
	fm2 := NewFieldMetas()

	fm2.SetFieldType(10, FieldType(100))
	fm2.SetFieldType(20, FieldType(120))
	fm2.SetFieldType(30, FieldType(127))

	if !fm2.HasFieldType(10, FieldType(100)) {
		t.Errorf("Expected field 10 to have type 100")
	}
	if !fm2.HasFieldType(20, FieldType(120)) {
		t.Errorf("Expected field 20 to have type 120")
	}
	if !fm2.HasFieldType(30, FieldType(127)) {
		t.Errorf("Expected field 30 to have type 127")
	}

	// 测试批量操作对高位类型的支持
	fields := fm2.GetFieldsWithType(FieldType(100))
	if len(fields) != 1 || fields[0] != 10 {
		t.Errorf("Expected GetFieldsWithType(100) to return [10], got %v", fields)
	}

	// 测试统计
	count := fm2.CountFieldsWithType(FieldType(120))
	if count != 1 {
		t.Errorf("Expected count 1 for type 120, got %d", count)
	}
}

// TestFieldMetas_MemorySize 测试内存占用
func TestFieldMetas_MemorySize(t *testing.T) {
	// 验证结构体大小是预期的1KB
	fm := NewFieldMetas()

	// FieldMetas 应该是 128 * 8 = 1024 bytes
	expectedSize := MaxFieldTypes * 8

	// 虽然Go没有直接获取结构体大小的方法，但我们可以通过unsafe包验证
	// 这里只是一个说明性测试
	t.Logf("FieldMetas size should be %d bytes (1KB)", expectedSize)
	t.Logf("MaxFieldTypes: %d", MaxFieldTypes)
	t.Logf("MaxFieldID: %d", MaxFieldID)

	// 验证功能完整性 - 为每个字段设置对应的类型
	for i := uint8(0); i < MaxFieldID; i++ {
		// 使用字段ID作为类型ID（都在0-63范围内）
		fm.SetFieldType(i, FieldType(i))
	}

	// 验证所有字段都设置成功
	for i := uint8(0); i < MaxFieldID; i++ {
		if !fm.HasFieldType(i, FieldType(i)) {
			t.Errorf("Expected field %d to have type %d", i, i)
		}
	}
}

// 性能测试
func BenchmarkFieldMetas_SetFieldType(b *testing.B) {
	fm := NewFieldMetas()
	for i := 0; i < b.N; i++ {
		fm.SetFieldType(uint8(i%64), FieldType(i%8))
	}
}

func BenchmarkFieldMetas_HasFieldType(b *testing.B) {
	fm := NewFieldMetas()
	// 预设一些字段
	for i := uint8(0); i < 32; i++ {
		fm.SetFieldType(i, FieldTypeSync)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fm.HasFieldType(uint8(i%64), FieldTypeSync)
	}
}

func BenchmarkFieldMetas_GetFieldsWithType(b *testing.B) {
	fm := NewFieldMetas()
	// 预设一些字段
	for i := uint8(0); i < 32; i++ {
		fm.SetFieldType(i, FieldTypeSync)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fm.GetFieldsWithType(FieldTypeSync)
	}
}

func BenchmarkFieldMetas_CountFieldsWithType(b *testing.B) {
	fm := NewFieldMetas()
	// 预设一些字段
	for i := uint8(0); i < 32; i++ {
		fm.SetFieldType(i, FieldTypeSync)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fm.CountFieldsWithType(FieldTypeSync)
	}
}

func BenchmarkFieldMetas_GetIntersection(b *testing.B) {
	fm := NewFieldMetas()
	// 预设一些字段
	for i := uint8(0); i < 20; i++ {
		fm.SetFieldType(i, FieldTypeSync)
		fm.SetFieldType(i+10, FieldTypePersistent)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fm.GetIntersection(FieldTypeSync, FieldTypePersistent)
	}
}

func BenchmarkFieldMetas_Clone(b *testing.B) {
	fm := NewFieldMetas()
	// 预设一些字段
	for i := uint8(0); i < 32; i++ {
		fm.SetFieldType(i, FieldType(i%8))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fm.Clone()
	}
}
