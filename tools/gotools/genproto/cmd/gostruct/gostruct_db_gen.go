package gostruct

import (
	"fmt"
	"strings"
)

type MongoUpdateGenerator struct{}

// generateComponentMongoUpdateMethod emits a method that uses mgo_builder.MongoUpdateBuilder
// to build MongoDB updates for dirty fields, matching the current design pattern.
func (g *MongoUpdateGenerator) GenerateMongoUpdateMethod(s *GoStruct) string {
	if s == nil {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, "func (c *%s) BuildMongoUpdate(builder *mgo_builder.MongoUpdateBuilder, prefix string) {\n", s.Name)
	b.WriteString("\tif c == nil {\n\t\treturn\n\t}\n")

	for _, f := range s.Fields {
		if f.Name == "DirtyTracker" || strings.HasPrefix(f.Type, "dirty.") {
			continue
		}

		// compute path with prefix
		fmt.Fprintf(&b, "\t{\n\t\tpath := \"%s\"\n", f.ProtoName)
		b.WriteString("\t\tif prefix != \"\" {\n\t\t\tpath = prefix + \".\" + path\n\t\t}\n")

		// Check if field is dirty
		fmt.Fprintf(&b, "\t\tif c.IsDirty(%sDirty%sBit) {\n", s.Name, f.Name)

		// Handle nested component pointers
		switch {
		case f.IsPointerType():
			fmt.Fprintf(&b, "\t\t\tif c.%s == nil {\n", f.Name)
			b.WriteString("\t\t\t\tbuilder.Set(path, nil)\n")
			b.WriteString("\t\t\t} else {\n")
			fmt.Fprintf(&b, "\t\t\t\tc.%s.BuildMongoUpdate(builder, path)\n", f.Name)
			b.WriteString("\t\t\t}\n")
		case f.IsMapType():
			g.GenMongoUpdateMapFieldCode(f, &b)
		case f.IsRepeatedType():
			g.GenMongoUpdateRepeatedFieldCode(f, &b)
		default:
			fmt.Fprintf(&b, "\t\t\tbuilder.Set(path, c.%s)\n", f.Name)
		}

		b.WriteString("\t\t}\n\t}\n")
	}

	b.WriteString("}\n")
	return b.String()
}

func (g *MongoUpdateGenerator) GenMongoUpdateMapFieldCode(f *GoField, b *strings.Builder) {
	// Generate MongoDB update code for map fields with deep copy
	// Matches the pattern used in manually written BuildMongoUpdate methods
	g.genMapDeepCopyCode(f, b)
}

// genMapDeepCopyCode 生成 map 字段的深拷贝代码
// 参数说明：
//
//	f: GoField 字段信息
//	b: strings.Builder 用于构建代码
func (g *MongoUpdateGenerator) genMapDeepCopyCode(f *GoField, b *strings.Builder) {
	// 生成注释说明
	b.WriteString("\t\t\t// Deep copy map to prevent reference sharing\n")

	// 检查源 map 是否为 nil，在检查内部处理所有逻辑
	fmt.Fprintf(b, "\t\t\tif c.%s != nil {\n", f.Name)

	// 在 nil 检查内部直接创建目标 map
	fmt.Fprintf(b, "\t\t\t\tdst := make(%s, len(c.%s))\n", f.Type, f.Name)

	if f.MapInfo != nil && f.MapInfo.IsPointerValueType() {
		// 遍历源 map，根据值类型选择拷贝方式
		fmt.Fprintf(b, "\t\t\t\tfor k := range c.%s {\n", f.Name)
		fmt.Fprintf(b, "\t\t\t\t\tv := c.%s[k]\n", f.Name)
		// 指针值类型：需要深拷贝
		b.WriteString("\t\t\t\t\tif v != nil {\n")

		// 获取指针指向的类型名（去掉 * 前缀）
		fmt.Fprintf(b, "\t\t\t\t\t\tdst[k] = new(%s)\n", strings.TrimPrefix(f.MapInfo.ValueType, "*"))
		b.WriteString("\t\t\t\t\t\tv.DeepCopy(dst[k])\n")
		b.WriteString("\t\t\t\t\t}\n")
	} else {
		// 值类型：直接拷贝
		fmt.Fprintf(b, "\t\t\t\tfor k := range c.%s {\n", f.Name)
		fmt.Fprintf(b, "\t\t\t\t\tdst[k] = c.%s[k]\n", f.Name)
	}

	b.WriteString("\t\t\t\t}\n")

	// 设置到 MongoDB 更新构建器
	b.WriteString("\t\t\t\tbuilder.Set(path, dst)\n")
	b.WriteString("\t\t\t}\n")
}

func (g *MongoUpdateGenerator) GenMongoUpdateRepeatedFieldCode(f *GoField, b *strings.Builder) {
	g.genRepeatedDeepCopyCode(f, b)
}

// 生成 slice/array 类型字段深拷贝代码
// 参数说明：
//
//	f: GoField 字段信息
//	b: strings.Builder 用于构建代码
func (g *MongoUpdateGenerator) genRepeatedDeepCopyCode(f *GoField, b *strings.Builder) {
	// 生成注释说明
	b.WriteString("\t\t\t// Deep copy slice to prevent reference sharing\n")

	// 检查源 slice 是否为 nil，在检查内部处理所有逻辑
	fmt.Fprintf(b, "\t\t\tif c.%s != nil {\n", f.Name)

	// 在 nil 检查内部直接创建目标 slice
	fmt.Fprintf(b, "\t\t\t\tdst := make(%s, len(c.%s))\n", f.Type, f.Name)

	if f.RepeatedInfo != nil && f.RepeatedInfo.IsPointerValueType() {
		// 遍历源 slice，根据元素类型选择拷贝方式
		fmt.Fprintf(b, "\t\t\t\tfor i := range c.%s {\n", f.Name)
		fmt.Fprintf(b, "\t\t\t\t\tv := c.%s[i]\n", f.Name)
		// 指针类型元素：需要深拷贝
		b.WriteString("\t\t\t\t\tif v != nil {\n")

		// 获取指针指向的类型名（去掉 * 前缀）
		elementType := f.RepeatedInfo.ValueType
		baseType := strings.TrimPrefix(elementType, "*")
		fmt.Fprintf(b, "\t\t\t\t\t\tdst[i] = new(%s)\n", baseType)
		b.WriteString("\t\t\t\t\t\tv.DeepCopy(dst[i])\n")
		b.WriteString("\t\t\t\t\t}\n")
		b.WriteString("\t\t\t\t}\n")
	} else {
		// 值类型元素：直接拷贝
		fmt.Fprintf(b, "\t\t\t\tcopy(dst, c.%s)\n", f.Name)
	}

	// 设置到 MongoDB 更新构建器
	b.WriteString("\t\t\t\tbuilder.Set(path, dst)\n")
	b.WriteString("\t\t\t}\n")
}
