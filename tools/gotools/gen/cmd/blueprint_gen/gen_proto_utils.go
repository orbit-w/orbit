package blueprint_gen

import (
	"fmt"
	"sort"
	"strings"
)

func GenerateProtoImport(packageName string) string {
	packageName = strings.ToLower(packageName)
	return fmt.Sprintf("%s.proto", packageName)
}

// UniqueProtoImports 排重&&排序 Proto 导入
// 返回排重&&排序后的 Proto 导入列表
func UniqueProtoImports(imports []string) []string {
	importMap := make(map[string]bool)
	uniqueImports := []string{}
	for _, imp := range imports {
		if !importMap[imp] {
			importMap[imp] = true
			uniqueImports = append(uniqueImports, imp)
		}
	}
	sort.Strings(uniqueImports)
	return uniqueImports
}
