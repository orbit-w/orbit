package blueprint_gen

import (
	"fmt"
	"strings"
	"text/template"
)

// Generator 代码生成器通用接口
type Generator interface {
	Generate(data *BlueprintData, outputDir string) error
	Name() string
}

// TemplateGenerator 基于模板的代码生成器
type TemplateGenerator struct {
	name         string
	templates    map[string]*template.Template
	templateFunc template.FuncMap
}

// NewTemplateGenerator 创建新的模板生成器
func NewTemplateGenerator(name string) *TemplateGenerator {
	return &TemplateGenerator{
		name:      name,
		templates: make(map[string]*template.Template),
		templateFunc: template.FuncMap{
			"toProtoType":      ToProtoTypeFromTypesFieldType,
			"toGoType":          ToGoTypeFromTypesFieldType,
			"toLower":           strings.ToLower,
			"toUpper":           strings.ToUpper,
			"camelToSnake":      CamelToSnake,
			"firstLower":         firstLower,
			"firstUpper":         firstUpper,
			"join":              strings.Join,
			"contains":          strings.Contains,
			"hasPrefix":          strings.HasPrefix,
			"hasSuffix":          strings.HasSuffix,
			"trimPrefix":         strings.TrimPrefix,
			"trimSuffix":         strings.TrimSuffix,
		},
	}
}

// AddTemplate 添加模板
func (g *TemplateGenerator) AddTemplate(name, tmpl string) error {
	t, err := template.New(name).Funcs(g.templateFunc).Parse(tmpl)
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", name, err)
	}
	g.templates[name] = t
	return nil
}

// RenderTemplate 渲染模板
func (g *TemplateGenerator) RenderTemplate(name string, data interface{}) (string, error) {
	t, ok := g.templates[name]
	if !ok {
		return "", fmt.Errorf("template %s not found", name)
	}

	var sb strings.Builder
	if err := t.Execute(&sb, data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", name, err)
	}

	return sb.String(), nil
}

// Name 返回生成器名称
func (g *TemplateGenerator) Name() string {
	return g.name
}

// CodeBuilder 代码构建器
type CodeBuilder struct {
	sb      strings.Builder
	indent  int
	indentStr string
}

// NewCodeBuilder 创建新的代码构建器
func NewCodeBuilder() *CodeBuilder {
	return &CodeBuilder{
		indent:     0,
		indentStr: "\t",
	}
}

// SetIndentStr 设置缩进字符串
func (b *CodeBuilder) SetIndentStr(str string) {
	b.indentStr = str
}

// WriteLine 写入一行代码
func (b *CodeBuilder) WriteLine(format string, args ...interface{}) {
	for i := 0; i < b.indent; i++ {
		b.sb.WriteString(b.indentStr)
	}
	if len(args) > 0 {
		b.sb.WriteString(fmt.Sprintf(format, args...))
	} else {
		b.sb.WriteString(format)
	}
	b.sb.WriteString("\n")
}

// Write 写入代码（不换行）
func (b *CodeBuilder) Write(format string, args ...interface{}) {
	if len(args) > 0 {
		b.sb.WriteString(fmt.Sprintf(format, args...))
	} else {
		b.sb.WriteString(format)
	}
}

// WriteEmptyLine 写入空行
func (b *CodeBuilder) WriteEmptyLine() {
	b.sb.WriteString("\n")
}

// Indent 增加缩进
func (b *CodeBuilder) Indent() {
	b.indent++
}

// Unindent 减少缩进
func (b *CodeBuilder) Unindent() {
	if b.indent > 0 {
		b.indent--
	}
}

// String 返回构建的代码
func (b *CodeBuilder) String() string {
	return b.sb.String()
}

// Reset 重置构建器
func (b *CodeBuilder) Reset() {
	b.sb.Reset()
	b.indent = 0
}

// 辅助函数
func firstLower(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}

func firstUpper(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

