package gostruct

import (
	"fmt"
	"strings"
)

// generateDeepCopyMethod generates a DeepCopy method for a struct
func generateDeepCopyMethod(s *GoStruct) string {
	if s == nil {
		return ""
	}

	// Skip OpsV2 types and other utility types
	if strings.Contains(s.Name, "OpsV2") || strings.Contains(s.Name, "Ops") {
		return ""
	}

	// Determine the method signature based on struct type
	if isEntity(s.Name) {
		return generateEntityDeepCopyMethod(s)
	} else if isComponent(s.Name) {
		return generateComponentDeepCopyMethod(s)
	} else {
		// For simple data structures, use DeepCopy with parameter
		return generateDataDeepCopyMethod(s)
	}
}

// generateEntityDeepCopyMethod generates DeepCopy method for Entity types
// Pattern: func (s *Entity) DeepCopy() *Entity
func generateEntityDeepCopyMethod(s *GoStruct) string {
	var b strings.Builder
	fmt.Fprintf(&b, "// DeepCopy creates a deep copy of %s\n", s.Name)
	fmt.Fprintf(&b, "func (s *%s) DeepCopy() *%s {\n", s.Name, s.Name)
	b.WriteString("\tif s == nil {\n\t\treturn nil\n\t}\n\n")

	fmt.Fprintf(&b, "\tco := &%s{}\n", s.Name)
	b.WriteString("\t*co = *s\n\n")

	for _, f := range s.Fields {
		if f.Name == "DirtyTracker" || strings.HasPrefix(f.Type, "dirty.") {
			continue
		}

		fieldCopyCode := generateEntityFieldDeepCopyCode(f)
		if fieldCopyCode != "" {
			b.WriteString(fieldCopyCode)
		}
	}

	b.WriteString("\treturn co\n")
	b.WriteString("}\n")

	return b.String()
}

// generateComponentDeepCopyMethod generates DeepCopy method for Component types
// Pattern: func (s *Component) DeepCopy(copy *Component)
func generateComponentDeepCopyMethod(s *GoStruct) string {
	var b strings.Builder
	fmt.Fprintf(&b, "// DeepCopy creates a deep copy of %s\n", s.Name)
	fmt.Fprintf(&b, "func (s *%s) DeepCopy(co *%s) {\n", s.Name, s.Name)
	b.WriteString("\tif s == nil {\n\t\treturn\n\t}\n\n")

	b.WriteString("\t*co = *s\n\n")

	for _, f := range s.Fields {
		if f.Name == "DirtyTracker" || strings.HasPrefix(f.Type, "dirty.") {
			continue
		}

		fieldCopyCode := generateComponentFieldDeepCopyCode(f)
		if fieldCopyCode != "" {
			b.WriteString(fieldCopyCode)
		}
	}

	b.WriteString("}\n")

	return b.String()
}

// generateDataDeepCopyMethod generates DeepCopy method for simple data types
// Pattern: func (s *Data) DeepCopy(copy *Data)
func generateDataDeepCopyMethod(s *GoStruct) string {
	var b strings.Builder
	fmt.Fprintf(&b, "func (s *%s) DeepCopy(co *%s) {\n", s.Name, s.Name)
	b.WriteString("\tif s == nil {\n\t\treturn\n\t}\n")
	b.WriteString("\t*co = *s\n")

	// For simple data types, check if there are any complex fields that need deep copying
	hasComplexFields := false
	for _, f := range s.Fields {
		if f.Name == "DirtyTracker" || strings.HasPrefix(f.Type, "dirty.") {
			continue
		}
		if strings.HasPrefix(f.Type, "map[") || strings.HasPrefix(f.Type, "[]") ||
			(strings.HasPrefix(f.Type, "*") && isStructType(f.Type)) {
			hasComplexFields = true
			break
		}
	}

	if hasComplexFields {
		b.WriteString("\n")
		for _, f := range s.Fields {
			if f.Name == "DirtyTracker" || strings.HasPrefix(f.Type, "dirty.") {
				continue
			}

			fieldCopyCode := generateDataFieldDeepCopyCode(f)
			if fieldCopyCode != "" {
				b.WriteString(fieldCopyCode)
			}
		}
	}

	b.WriteString("}\n")

	return b.String()
}

// generateEntityFieldDeepCopyCode generates deep copy code for Entity fields
func generateEntityFieldDeepCopyCode(f *GoField) string {
	if f == nil {
		return ""
	}

	var b strings.Builder

	if f.IsPointerType() && isComponentType(f.Type) {
		// Component pointers: create new instance and call its DeepCopy method
		typeName := extractTypeName(f.Type)
		fmt.Fprintf(&b, "\tif s.%s != nil {\n", f.Name)
		fmt.Fprintf(&b, "\t\tco.%s = &%s{}\n", f.Name, typeName)
		fmt.Fprintf(&b, "\t\ts.%s.DeepCopy(co.%s)\n", f.Name, f.Name)
		b.WriteString("\t}\n\n")
	} else if f.IsMapType() {
		genEntityMapFieldDeepCopyCode(f, &b)
	} else if f.IsRepeatedType() {
		// Handle slice types
		genEntityRepeatedFieldDeepCopyCode(f, &b)
	} else {
		// Basic types are already copied by *copy = *s
		return ""
	}

	return b.String()
}

// MapDeepCopyConfig 配置map深拷贝代码生成的参数
type MapDeepCopyConfig struct {
	TargetVarName   string // 目标变量名: "copy" 或 "co"
	AddExtraNewline bool   // 是否在最后添加额外换行
}

// genMapFieldDeepCopyCode 生成map字段深拷贝代码的统一函数
func genMapFieldDeepCopyCode(f *GoField, b *strings.Builder, cfg MapDeepCopyConfig) {
	fmt.Fprintf(b, "\tif s.%s != nil {\n", f.Name)

	_, valueType, ok := splitMapGoType(f.Type)
	if !ok {
		b.WriteString("\t}\n")
		if cfg.AddExtraNewline {
			b.WriteString("\n")
		}
		return
	}

	// 使用直接赋值策略 - 更简洁、高效
	fmt.Fprintf(b, "\t\t%s.%s = make(%s, len(s.%s))\n", cfg.TargetVarName, f.Name, f.Type, f.Name)
	fmt.Fprintf(b, "\t\tfor k, v := range s.%s {\n", f.Name)

	if f.MapInfo.IsPointerValueType() {
		// 深拷贝指针值
		b.WriteString("\t\t\tif v != nil {\n")
		fmt.Fprintf(b, "\t\t\t\t%s.%s[k] = new(%s)\n", cfg.TargetVarName, f.Name, strings.TrimPrefix(valueType, "*"))
		fmt.Fprintf(b, "\t\t\t\tv.DeepCopy(%s.%s[k])\n", cfg.TargetVarName, f.Name)
		b.WriteString("\t\t\t}\n")
	} else {
		// 简单值拷贝
		fmt.Fprintf(b, "\t\t\t%s.%s[k] = v\n", cfg.TargetVarName, f.Name)
	}
	b.WriteString("\t\t}\n")

	b.WriteString("\t}\n")
	if cfg.AddExtraNewline {
		b.WriteString("\n")
	}
}

func genEntityMapFieldDeepCopyCode(f *GoField, b *strings.Builder) {
	genMapFieldDeepCopyCode(f, b, MapDeepCopyConfig{
		TargetVarName:   "co",
		AddExtraNewline: true,
	})
}

// genEntityRepeatedFieldDeepCopyCode 为Entity类型的repeated字段(切片)生成深拷贝代码
// 参数说明:
//
//	f: 字段信息，包含字段名、类型等
//	b: 字符串构建器，用于输出生成的代码
//
// 生成的代码结构:
//  1. 检查源切片是否为nil
//  2. 为目标切片分配相同长度的内存
//  3. 根据元素类型选择拷贝策略:
//     - 指针结构体: 创建新实例并递归深拷贝
//     - 基础类型: 直接使用Go的copy()函数
func genEntityRepeatedFieldDeepCopyCode(f *GoField, b *strings.Builder) {
	// 生成nil检查: if s.FieldName != nil {
	fmt.Fprintf(b, "\tif s.%s != nil {\n", f.Name)

	// 提取切片元素类型，例如: []string -> string, []*User -> *User
	elemType := strings.TrimPrefix(f.Type, "[]")

	// 生成切片内存分配: copy.FieldName = make([]ElementType, len(s.FieldName))
	fmt.Fprintf(b, "\t\tco.%s = make(%s, len(s.%s))\n", f.Name, f.Type, f.Name)

	// 判断元素类型: 如果是指针结构体类型，需要深拷贝
	if strings.HasPrefix(elemType, "*") && isStructType(elemType) {
		// 生成遍历循环: for i, v := range s.FieldName {
		fmt.Fprintf(b, "\t\tfor i, v := range s.%s {\n", f.Name)

		// 生成空指针检查: if v != nil {
		b.WriteString("\t\t\tif v != nil {\n")

		// 为每个元素创建新实例: copy.FieldName[i] = new(ElementType)
		fmt.Fprintf(b, "\t\t\t\tco.%s[i] = new(%s)\n", f.Name, strings.TrimPrefix(elemType, "*"))

		// 递归调用深拷贝方法: v.DeepCopy(copy.FieldName[i])
		fmt.Fprintf(b, "\t\t\t\tv.DeepCopy(co.%s[i])\n", f.Name)

		b.WriteString("\t\t\t}\n") // 结束if v != nil
		b.WriteString("\t\t}\n")   // 结束for循环
	} else {
		// 基础类型或值类型: 使用Go内置的copy()函数进行浅拷贝
		// 生成: copy(copy.FieldName, s.FieldName)
		fmt.Fprintf(b, "\t\tcopy(co.%s, s.%s)\n", f.Name, f.Name)
	}

	// 结束if s.FieldName != nil 并添加空行分隔
	b.WriteString("\t}\n\n")
}

// generateComponentFieldDeepCopyCode generates deep copy code for Component fields
func generateComponentFieldDeepCopyCode(f *GoField) string {
	if f == nil {
		return ""
	}

	var b strings.Builder
	switch {
	case f.IsPointerType() && isComponentType(f.Type):
		// Nested component pointers
		typeName := extractTypeName(f.Type)
		fmt.Fprintf(&b, "\tif s.%s != nil {\n", f.Name)
		fmt.Fprintf(&b, "\t\tco.%s = &%s{}\n", f.Name, typeName)
		fmt.Fprintf(&b, "\t\ts.%s.DeepCopy(co.%s)\n", f.Name, f.Name)
		b.WriteString("\t}\n")
	case f.IsMapType():
		// Handle map types
		genComponentMapFieldDeepCopyCode(f, &b)
	case f.IsRepeatedType():
		// Handle slice types similar to Entity version but without extra newlines
		genComponentRepeatedFieldDeepCopyCode(f, &b)
	}

	return b.String()
}

func genComponentRepeatedFieldDeepCopyCode(f *GoField, b *strings.Builder) {
	// Handle slice types similar to Entity version but without extra newlines
	fmt.Fprintf(b, "\tif s.%s != nil {\n", f.Name)
	elemType := strings.TrimPrefix(f.Type, "[]")
	fmt.Fprintf(b, "\t\tco.%s = make(%s, len(s.%s))\n", f.Name, f.Type, f.Name)
	if strings.HasPrefix(elemType, "*") && isStructType(elemType) {
		fmt.Fprintf(b, "\t\tfor i, v := range s.%s {\n", f.Name)
		b.WriteString("\t\t\tif v != nil {\n")
		fmt.Fprintf(b, "\t\t\t\tco.%s[i] = new(%s)\n", f.Name, strings.TrimPrefix(elemType, "*"))
		fmt.Fprintf(b, "\t\t\t\tv.DeepCopy(co.%s[i])\n", f.Name)
		b.WriteString("\t\t\t}\n")
		b.WriteString("\t\t}\n")
	} else {
		fmt.Fprintf(b, "\t\tcopy(co.%s, s.%s)\n", f.Name, f.Name)
	}
	b.WriteString("\t}\n")
}

func genComponentMapFieldDeepCopyCode(f *GoField, b *strings.Builder) {
	genMapFieldDeepCopyCode(f, b, MapDeepCopyConfig{
		TargetVarName:   "co",
		AddExtraNewline: false,
	})
}

// generateDataFieldDeepCopyCode generates deep copy code for simple Data fields
func generateDataFieldDeepCopyCode(f *GoField) string {
	if f == nil {
		return ""
	}

	var b strings.Builder

	switch {
	case f.IsMapType():
		genDataMapFieldDeepCopyCode(f, &b)
	case f.IsRepeatedType():
		// Handle slice types
		fmt.Fprintf(&b, "\tif s.%s != nil {\n", f.Name)
		elemType := strings.TrimPrefix(f.Type, "[]")
		fmt.Fprintf(&b, "\t\tco.%s = make(%s, len(s.%s))\n", f.Name, f.Type, f.Name)
		if strings.HasPrefix(elemType, "*") && isStructType(elemType) {
			fmt.Fprintf(&b, "\t\tfor i, v := range s.%s {\n", f.Name)
			b.WriteString("\t\t\tif v != nil {\n")
			fmt.Fprintf(&b, "\t\t\t\tco.%s[i] = new(%s)\n", f.Name, strings.TrimPrefix(elemType, "*"))
			fmt.Fprintf(&b, "\t\t\t\tv.DeepCopy(co.%s[i])\n", f.Name)
			b.WriteString("\t\t\t}\n")
			b.WriteString("\t\t}\n")
		} else {
			fmt.Fprintf(&b, "\t\tcopy(co.%s, s.%s)\n", f.Name, f.Name)
		}
		b.WriteString("\t}\n")
	case f.IsPointerType() && isStructType(f.Type):
		// Handle pointer to struct
		fmt.Fprintf(&b, "\tif s.%s != nil {\n", f.Name)
		fmt.Fprintf(&b, "\t\tco.%s = new(%s)\n", f.Name, strings.TrimPrefix(f.Type, "*"))
		fmt.Fprintf(&b, "\t\ts.%s.DeepCopy(co.%s)\n", f.Name, f.Name)
		b.WriteString("\t}\n")
	default:
		// Basic types are already copied by *copy = *s
		return ""
	}
	return b.String()
}

func genDataMapFieldDeepCopyCode(f *GoField, b *strings.Builder) {
	genMapFieldDeepCopyCode(f, b, MapDeepCopyConfig{
		TargetVarName:   "co",
		AddExtraNewline: false,
	})
}
