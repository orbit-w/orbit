package cmd

import (
	"gitee.com/orbit-w/orbit/tools/gotools/genproto/cmd/gostruct"
)

// ParseProtoToGoStructs 解析proto文件，生成Go结构体信息
func ParseProtoToGoStructs(ctx *Context, packageName string) ([]*gostruct.GoStruct, error) {
	config := gostruct.DefaultGenerationConfig(packageName)
	// adapt Context to gostruct.ProtoContext
	return gostruct.ParseProtoToGoStructsWithConfig(ctx, config)
}
