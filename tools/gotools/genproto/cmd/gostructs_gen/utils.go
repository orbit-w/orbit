package cmd

import (
	"os"
	"path/filepath"
	"strings"

	gogodesc "github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
)

// writeToFile 写入文件
func writeToFile(filePath, content string) error {
	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	return err
}

// placed here to avoid duplicate definitions across files
func buildMessageIndex(ctx *Context) {
	fds := ctx.GetFileDescriptorSet()
	for _, file := range fds.File {
		pkg := file.GetPackage()
		for _, m := range file.GetMessageType() {
			addMessageToIndex(ctx, m, pkg, "")
		}
	}
}

// addMessageToIndex 递归添加消息及其嵌套消息到索引中
func addMessageToIndex(ctx *Context, msg *gogodesc.DescriptorProto, pkg string, parentPrefix string) {
	msgName := msg.GetName()
	var fullName string

	if parentPrefix != "" {
		fullName = parentPrefix + "." + msgName
	} else {
		fullName = msgName
	}

	if pkg != "" {
		fullName = pkg + "." + fullName
	}

	ctx.AddMessageToIndex(fullName, msg)

	// 递归处理嵌套消息
	for _, nested := range msg.GetNestedType() {
		var newPrefix string
		if parentPrefix != "" {
			newPrefix = parentPrefix + "." + msgName
		} else {
			newPrefix = msgName
		}
		addMessageToIndex(ctx, nested, pkg, newPrefix)
	}
}

func trimLeadingDot(s string) string {
	if len(s) > 0 && s[0] == '.' {
		return s[1:]
	}
	return s
}

func shortTypeName(full string) string {
	if i := strings.LastIndex(full, "."); i >= 0 {
		return full[i+1:]
	}
	return full
}
