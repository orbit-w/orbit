package router_gen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// parseNetWallYAML 解析 NetWall YAML 文件（支持多文档格式）
func parseNetWallYAML(yamlPath string) ([]*RequestInfo, error) {
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %w", err)
	}

	// 使用多文档解析器（支持 --- 分隔的多个文档）
	reader := strings.NewReader(string(data))
	yamlData, err := readMultiDocumentYAML(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	netwallData, ok := yamlData["NetWall"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("netWall section not found")
	}

	packageName, _ := netwallData["Name"].(string)
	if packageName == "" {
		return nil, fmt.Errorf("netWall name not found")
	}

	requests, ok := netwallData["Requests"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("requests section not found or invalid")
	}

	var requestInfos []*RequestInfo
	for _, reqItem := range requests {
		reqInfo := parseRequestItem(reqItem, packageName)
		if reqInfo != nil {
			requestInfos = append(requestInfos, reqInfo)
		}
	}

	return requestInfos, nil
}

// parseRequestItem 解析单个 Request 项
func parseRequestItem(item interface{}, packageName string) *RequestInfo {
	itemMap, ok := item.(map[string]interface{})
	if !ok {
		return nil
	}

	var requestName string
	var fields map[string]interface{}
	var hasRsp bool

	for key, value := range itemMap {
		if key == "Rsp" {
			hasRsp = true
			continue
		}
		requestName = key
		if value != nil {
			fields, _ = value.(map[string]interface{})
		}
	}

	if requestName == "" {
		return nil
	}

	// 解析 EntityRef 字段
	entityRefs := parseEntityRefs(fields)

	// 包名处理：首字母大写用于类型引用，小写用于导入
	packageNameLower := strings.ToLower(packageName)
	packageNameUpper := strings.ToUpper(packageName[:1]) + packageName[1:]

	return &RequestInfo{
		RequestName:      requestName,
		PackageName:      packageNameLower,
		PackageNameUpper: packageNameUpper,
		EntityRefs:       entityRefs,
		HasRsp:           hasRsp,
	}
}

// parseEntityRefs 解析 EntityRef 字段
func parseEntityRefs(fields map[string]interface{}) []*EntityRefInfo {
	var entityRefs []*EntityRefInfo

	for fieldKey := range fields {
		// YAML 格式：EntityRef PlayerEntityRef: 1
		// 解析后 fieldKey 是 "EntityRef PlayerEntityRef"
		// fieldValue 是字段编号（数字），这里不需要使用
		
		// 检查字段键是否包含 "EntityRef"
		if !strings.Contains(fieldKey, "EntityRef") {
			continue
		}

		// 提取字段名（用于访问 Request 字段）
		// 从 "EntityRef PlayerEntityRef" 提取 "PlayerEntityRef"
		fieldName := extractFieldName(fieldKey)
		if fieldName == "" {
			continue
		}

		// 从字段名提取实体名称
		// PlayerEntityRef -> Player
		entityName := extractEntityName(fieldName)
		if entityName == "" {
			continue
		}

		// 生成参数名（小驼峰）
		paramName := toCamelCase(entityName) + "Entity"

		// 生成包装器类型
		wrapperType := fmt.Sprintf("*mme.%sEntityWrapper", entityName)
		wrapperTypeWithAlias := fmt.Sprintf("*mmeobj.%sEntityWrapper", entityName)

		entityRefs = append(entityRefs, &EntityRefInfo{
			FieldName:           fieldName,
			EntityName:          entityName,
			ParamName:           paramName,
			WrapperType:         wrapperType,
			WrapperTypeWithAlias: wrapperTypeWithAlias,
		})
	}

	return entityRefs
}

// extractFieldName 从字段键中提取字段名
// 例如：EntityRef PlayerEntityRef -> PlayerEntityRef
func extractFieldName(fieldKey string) string {
	// 如果包含空格，取最后一部分
	parts := strings.Fields(fieldKey)
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	// 如果没有空格，直接返回（可能是 PlayerEntityRef）
	return fieldKey
}

// extractEntityName 从 EntityRef 字段名提取实体名称
// 例如：PlayerEntityRef -> Player, HeroEntityRef -> Hero
func extractEntityName(fieldName string) string {
	// 处理 "EntityRef PlayerEntityRef" 格式
	fieldName = strings.TrimPrefix(fieldName, "EntityRef ")

	// 去掉 EntityRef 后缀
	if strings.HasSuffix(fieldName, "EntityRef") {
		return strings.TrimSuffix(fieldName, "EntityRef")
	}

	return ""
}

// toCamelCase 转换为小驼峰命名
// 例如：Player -> player, Hero -> hero
func toCamelCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}

// parseControllerFile 解析 Controller 文件，提取 Controller 信息
func parseControllerFile(controllerPath string) (*ControllerInfo, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, controllerPath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse controller file: %w", err)
	}

	var controllerInfo *ControllerInfo

	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.GenDecl:
			if x.Tok == token.VAR {
				for _, spec := range x.Specs {
					if valueSpec, ok := spec.(*ast.ValueSpec); ok {
						for i, name := range valueSpec.Names {
							if strings.HasPrefix(name.Name, "G") && strings.Contains(name.Name, "Controller") {
								// 找到 Controller 变量
								varName := name.Name
								var typeName string

								// 获取类型
								if valueSpec.Type != nil {
									if ident, ok := valueSpec.Type.(*ast.Ident); ok {
										typeName = ident.Name
									}
								}

								// 获取包名
								packageName := node.Name.Name

								controllerInfo = &ControllerInfo{
									PackageName: packageName,
									VarName:     varName,
									TypeName:    typeName,
								}
								return false
							}

							// 如果没有显式类型，尝试从值推断
							if valueSpec.Values != nil && i < len(valueSpec.Values) {
								if compositeLit, ok := valueSpec.Values[i].(*ast.CompositeLit); ok {
									if ident, ok := compositeLit.Type.(*ast.Ident); ok {
										typeName := ident.Name
										varName := name.Name
										packageName := node.Name.Name

										if controllerInfo == nil {
											controllerInfo = &ControllerInfo{
												PackageName: packageName,
												VarName:     varName,
												TypeName:    typeName,
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
		return true
	})

	if controllerInfo == nil {
		return nil, fmt.Errorf("controller info not found in file")
	}

	return controllerInfo, nil
}

// discoverControllerFiles 发现 Controller 文件
func discoverControllerFiles(controllerDir string) ([]string, error) {
	var files []string

	err := filepath.Walk(controllerDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			// 跳过测试文件和生成的文件
			if strings.HasSuffix(path, "_test.go") || strings.Contains(path, "_gen.go") {
				return nil
			}
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// readMultiDocumentYAML 读取多文档 YAML 文件（支持 --- 分隔）
func readMultiDocumentYAML(reader io.Reader) (map[string]interface{}, error) {
	decoder := yaml.NewDecoder(reader)
	result := make(map[string]interface{})

	for {
		var doc interface{}
		if err := decoder.Decode(&doc); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to decode YAML document: %w", err)
		}

		// 处理文档
		var docMap map[string]interface{}
		switch v := doc.(type) {
		case map[string]interface{}:
			docMap = v
		case nil:
			// 空文档，跳过
			continue
		default:
			return nil, fmt.Errorf("unsupported YAML document type: %T", v)
		}

		// 合并文档内容
		for k, v := range docMap {
			result[k] = v
		}
	}

	return result, nil
}

