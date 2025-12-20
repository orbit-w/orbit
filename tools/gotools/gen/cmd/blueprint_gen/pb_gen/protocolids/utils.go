package protocolids

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// codeBuilder 代码构建器实现
type codeBuilder struct {
	sb        strings.Builder
	indent    int
	indentStr string
}

// newCodeBuilder 创建新的代码构建器
func newCodeBuilder() *codeBuilder {
	return &codeBuilder{
		indent:    0,
		indentStr: "\t",
	}
}

// SetIndentStr 设置缩进字符串
func (b *codeBuilder) SetIndentStr(str string) {
	b.indentStr = str
}

// WriteLine 写入一行代码
func (b *codeBuilder) WriteLine(format string, args ...any) {
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

// WriteEmptyLine 写入空行
func (b *codeBuilder) WriteEmptyLine() {
	b.sb.WriteString("\n")
}

// Indent 增加缩进
func (b *codeBuilder) Indent() {
	b.indent++
}

// Unindent 减少缩进
func (b *codeBuilder) Unindent() {
	if b.indent > 0 {
		b.indent--
	}
}

// String 返回构建的代码
func (b *codeBuilder) String() string {
	return b.sb.String()
}

// defaultFileWriter 默认文件写入器实现
type defaultFileWriter struct{}

// EnsureDir 确保目录存在
func (w *defaultFileWriter) EnsureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}

// WriteFile 写入文件内容
func (w *defaultFileWriter) WriteFile(filePath, content string) error {
	if err := w.EnsureDir(filepath.Dir(filePath)); err != nil {
		return err
	}
	return os.WriteFile(filePath, []byte(content), 0644)
}

