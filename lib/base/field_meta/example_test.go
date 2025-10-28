package fieldmeta_test

import (
	"fmt"

	fieldmeta "gitee.com/orbit-w/orbit/lib/base/field_meta"
)

// 示例1：基础用法
func ExampleFieldMetas_basic() {
	fm := fieldmeta.NewFieldMetas()

	// 设置字段类型
	fm.SetFieldType(0, fieldmeta.FieldTypeSync)
	fm.SetFieldType(1, fieldmeta.FieldTypePersistent)

	// 检查字段类型
	fmt.Println(fm.HasFieldType(0, fieldmeta.FieldTypeSync))
	fmt.Println(fm.HasFieldType(1, fieldmeta.FieldTypePersistent))

	// Output:
	// true
	// true
}

// 示例2：多类型字段
func ExampleFieldMetas_multipleTypes() {
	fm := fieldmeta.NewFieldMetas()

	// 为一个字段设置多种类型
	fm.SetMultipleFieldTypes(5,
		fieldmeta.FieldTypeSync,
		fieldmeta.FieldTypePersistent,
		fieldmeta.FieldTypeValidated,
	)

	// 检查是否匹配所有类型
	fmt.Println(fm.MatchesAll(5,
		fieldmeta.FieldTypeSync,
		fieldmeta.FieldTypePersistent,
		fieldmeta.FieldTypeValidated))

	// Output:
	// true
}

// 示例3：查询具有指定类型的字段
func ExampleFieldMetas_getFields() {
	fm := fieldmeta.NewFieldMetas()

	// 设置多个字段
	fm.SetFieldType(0, fieldmeta.FieldTypeSync)
	fm.SetFieldType(5, fieldmeta.FieldTypeSync)
	fm.SetFieldType(10, fieldmeta.FieldTypeSync)

	// 获取所有可同步字段
	fields := fm.GetFieldsWithType(fieldmeta.FieldTypeSync)
	fmt.Println(fields)

	// 统计数量
	count := fm.CountFieldsWithType(fieldmeta.FieldTypeSync)
	fmt.Println(count)

	// Output:
	// [0 5 10]
	// 3
}

// 示例4：集合操作
func ExampleFieldMetas_setOperations() {
	fm := fieldmeta.NewFieldMetas()

	// 字段0,1,2 是可同步的
	fm.SetFieldType(0, fieldmeta.FieldTypeSync)
	fm.SetFieldType(1, fieldmeta.FieldTypeSync)
	fm.SetFieldType(2, fieldmeta.FieldTypeSync)

	// 字段1,2,3 是持久化的
	fm.SetFieldType(1, fieldmeta.FieldTypePersistent)
	fm.SetFieldType(2, fieldmeta.FieldTypePersistent)
	fm.SetFieldType(3, fieldmeta.FieldTypePersistent)

	// 交集：既可同步又持久化
	intersection := fm.GetIntersection(fieldmeta.FieldTypeSync, fieldmeta.FieldTypePersistent)
	fmt.Println(intersection)

	// 并集：可同步或持久化
	union := fm.GetUnion(fieldmeta.FieldTypeSync, fieldmeta.FieldTypePersistent)
	fmt.Println(union)

	// Output:
	// [1 2]
	// [0 1 2 3]
}

// 示例5：位掩码批量操作
func ExampleFieldMetas_maskOperations() {
	fm := fieldmeta.NewFieldMetas()

	// 使用位掩码批量设置字段
	// 0b1111 表示字段 0,1,2,3
	fm.SetMask(fieldmeta.FieldTypeSync, 0b1111)

	// 检查字段
	fmt.Println(fm.HasFieldType(0, fieldmeta.FieldTypeSync))
	fmt.Println(fm.HasFieldType(3, fieldmeta.FieldTypeSync))
	fmt.Println(fm.HasFieldType(4, fieldmeta.FieldTypeSync))

	// Output:
	// true
	// true
	// false
}

// 示例6：遍历字段
func ExampleFieldMetas_forEach() {
	fm := fieldmeta.NewFieldMetas()

	// 设置一些字段
	fm.SetFieldType(1, fieldmeta.FieldTypeSync)
	fm.SetFieldType(3, fieldmeta.FieldTypeSync)
	fm.SetFieldType(5, fieldmeta.FieldTypeSync)

	// 遍历所有可同步字段
	fm.ForEachFieldWithType(fieldmeta.FieldTypeSync, func(fieldID uint8) bool {
		fmt.Printf("Field %d is syncable\n", fieldID)
		return true
	})

	// Output:
	// Field 1 is syncable
	// Field 3 is syncable
	// Field 5 is syncable
}

// 示例7：游戏角色属性管理
func Example_gameCharacter() {
	// 定义角色属性字段ID
	const (
		FieldHP       = 0
		FieldMP       = 1
		FieldLevel    = 2
		FieldExp      = 3
		FieldGold     = 4
		FieldDiamond  = 5
		FieldName     = 6
		FieldVIPLevel = 7
	)

	fm := fieldmeta.NewFieldMetas()

	// 配置字段属性
	// HP/MP/Level 需要同步到客户端
	fm.SetFieldType(FieldHP, fieldmeta.FieldTypeSync)
	fm.SetFieldType(FieldMP, fieldmeta.FieldTypeSync)
	fm.SetFieldType(FieldLevel, fieldmeta.FieldTypeSync)

	// Gold/Diamond/Level 需要持久化
	fm.SetFieldType(FieldGold, fieldmeta.FieldTypePersistent)
	fm.SetFieldType(FieldDiamond, fieldmeta.FieldTypePersistent)
	fm.SetFieldType(FieldLevel, fieldmeta.FieldTypePersistent)

	// Name 可以被客户端修改，需要验证
	fm.SetMultipleFieldTypes(FieldName,
		fieldmeta.FieldTypeClientModifiable,
		fieldmeta.FieldTypeValidated,
	)

	// 查询需要同步的字段
	syncFields := fm.GetFieldsWithType(fieldmeta.FieldTypeSync)
	fmt.Printf("需要同步的字段: %v\n", syncFields)

	// 查询需要持久化的字段
	persistFields := fm.GetFieldsWithType(fieldmeta.FieldTypePersistent)
	fmt.Printf("需要持久化的字段: %v\n", persistFields)

	// 查询既需要同步又需要持久化的字段
	bothFields := fm.GetIntersection(fieldmeta.FieldTypeSync, fieldmeta.FieldTypePersistent)
	fmt.Printf("既需要同步又需要持久化: %v\n", bothFields)

	// Output:
	// 需要同步的字段: [0 1 2]
	// 需要持久化的字段: [2 4 5]
	// 既需要同步又需要持久化: [2]
}

// 示例8：使用位掩码进行高性能批量操作
func Example_highPerformance() {
	fm := fieldmeta.NewFieldMetas()

	// 假设有64个字段，快速配置前32个为可同步
	fm.SetMask(fieldmeta.FieldTypeSync, 0xFFFFFFFF) // 低32位全1

	// 快速配置字段16-31为持久化
	fm.SetMask(fieldmeta.FieldTypePersistent, 0xFFFF0000) // 第16-31位为1

	// 使用位运算快速判断
	syncMask := fm.GetMask(fieldmeta.FieldTypeSync)
	persistMask := fm.GetMask(fieldmeta.FieldTypePersistent)

	// 计算交集（既可同步又需持久化）
	intersectionMask := syncMask & persistMask
	fmt.Printf("交集掩码: 0x%X\n", intersectionMask)

	// 统计字段数量
	fmt.Printf("可同步字段数: %d\n", fm.CountFieldsWithType(fieldmeta.FieldTypeSync))
	fmt.Printf("需持久化字段数: %d\n", fm.CountFieldsWithType(fieldmeta.FieldTypePersistent))

	// Output:
	// 交集掩码: 0xFFFF0000
	// 可同步字段数: 32
	// 需持久化字段数: 16
}

// 示例9：克隆和独立修改
func ExampleFieldMetas_Clone() {
	fm := fieldmeta.NewFieldMetas()
	fm.SetFieldType(0, fieldmeta.FieldTypeSync)
	fm.SetFieldType(1, fieldmeta.FieldTypePersistent)

	// 克隆
	clone := fm.Clone()

	// 修改原始对象
	fm.SetFieldType(2, fieldmeta.FieldTypeReadOnly)

	// 克隆不受影响
	fmt.Println(clone.HasFieldType(0, fieldmeta.FieldTypeSync))
	fmt.Println(clone.HasFieldType(2, fieldmeta.FieldTypeReadOnly))

	// Output:
	// true
	// false
}

// 示例10：清除操作
func Example_clearOperations() {
	fm := fieldmeta.NewFieldMetas()

	// 设置多个字段和类型
	fm.SetMultipleFieldTypes(5,
		fieldmeta.FieldTypeSync,
		fieldmeta.FieldTypePersistent,
		fieldmeta.FieldTypeValidated,
	)
	fm.SetFieldType(10, fieldmeta.FieldTypeSync)

	// 清除单个字段的所有类型
	fm.ClearField(5)
	fmt.Println(fm.HasFieldType(5, fieldmeta.FieldTypeSync))
	fmt.Println(fm.HasFieldType(10, fieldmeta.FieldTypeSync))

	// 清除某种类型的所有字段
	fm.ClearMask(fieldmeta.FieldTypeSync)
	fmt.Println(fm.HasFieldType(10, fieldmeta.FieldTypeSync))

	// Output:
	// false
	// true
	// false
}
