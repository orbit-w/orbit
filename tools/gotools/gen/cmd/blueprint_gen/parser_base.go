package blueprint_gen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gitee.com/orbit-w/meteor/bases/misc/utils"
	types "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types"
	"gopkg.in/yaml.v3"
)

// BaseParser 通用解析器基类
type BaseParser struct {
	blueprintDir string
	ctx          *BlueprintContext
}

// NewBaseParser 创建基础解析器
func NewBaseParser(blueprintDir string) *BaseParser {
	return &BaseParser{
		blueprintDir: blueprintDir,
		ctx:          NewBlueprintContext(),
	}
}

// GetContext 获取解析后的蓝图上下文
func (p *BaseParser) GetContext() *BlueprintContext {
	return p.ctx
}

// ReadYAMLFile 读取并解析 YAML 文件（支持多文档，用 --- 分隔）
// 支持 map 和 list 类型的文档，list 会自动转换为 map
func (p *BaseParser) ReadYAMLFile(filePath string) (map[string]any, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	reader := strings.NewReader(string(data))
	return ReadMultiDocumentYAML(reader)
}

// ReadYAMLFileList 读取并解析 YAML 列表文件
func (p *BaseParser) ReadYAMLFileList(filePath string) ([]any, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	var yamlData []any
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		return nil, fmt.Errorf("failed to parse YAML file %s: %w", filePath, err)
	}

	return yamlData, nil
}

// ResolvePath 解析文件路径
func (p *BaseParser) ResolvePath(paths ...string) string {
	return filepath.Join(append([]string{p.blueprintDir}, paths...)...)
}

// ParseFieldDefinitionToTypesField 将字段定义字符串解析为 types.Field
func (p *BaseParser) ParseFieldDefinitionToTypesField(fieldDef string) (*types.Field, error) {
	return ParseFieldDefinition(fieldDef)
}

// ParseMessageFields 解析消息字段（通用逻辑），统一使用 *types.Field
func (p *BaseParser) ParseMessageFields(fieldsMap map[string]any) ([]*types.Field, error) {
	result := make([]*types.Field, 0)

	for fieldKey, fieldDef := range fieldsMap {
		var field *types.Field
		var err error

		fieldStr := utils.ToString(fieldDef)
		// int32 FieldName: 1
		// fieldKey 包含类型和字段名，如 "string Query"
		field, err = ParseFieldDefinition(fieldKey + ": " + fieldStr)
		if err == nil && field != nil {
			result = append(result, field)
			continue
		}
	}

	return result, nil
}

// ParseMessageFieldsToTypes 解析消息字段为 types.Field 列表（通用逻辑）
// 现在直接使用 ParseMessageFields，因为已经统一使用 *types.Field
func (p *BaseParser) ParseMessageFieldsToTypes(fieldsMap map[string]any) ([]*types.Field, error) {
	return p.ParseMessageFields(fieldsMap)
}
