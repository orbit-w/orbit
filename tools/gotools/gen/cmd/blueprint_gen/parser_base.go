package blueprint_gen

import (
	"fmt"
	"os"
	"path/filepath"

	types "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types"
	"gopkg.in/yaml.v3"
)

// BaseParser 通用解析器基类
type BaseParser struct {
	blueprintDir string
	ctx          *BlueprintData
}

// NewBaseParser 创建基础解析器
func NewBaseParser(blueprintDir string) *BaseParser {
	return &BaseParser{
		blueprintDir: blueprintDir,
		ctx: &BlueprintData{
			Entities:   make([]Entity, 0),
			Managers:   make([]Manager, 0),
			Modules:    make([]Module, 0),
			Mechanisms: make([]Mechanism, 0),
			NetWalls:   make([]NetWallFile, 0),
		},
	}
}

// GetContext 获取解析后的蓝图上下文
func (p *BaseParser) GetContext() *BlueprintData {
	return p.ctx
}

// ReadYAMLFile 读取并解析 YAML 文件
func (p *BaseParser) ReadYAMLFile(filePath string) (map[string]interface{}, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	var yamlData map[string]interface{}
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		return nil, fmt.Errorf("failed to parse YAML file %s: %w", filePath, err)
	}

	return yamlData, nil
}

// ReadYAMLFileList 读取并解析 YAML 列表文件
func (p *BaseParser) ReadYAMLFileList(filePath string) ([]interface{}, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	var yamlData []interface{}
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
	field, err := ParseFieldDefinition(fieldDef)
	if err != nil {
		return nil, err
	}

	typesField, err := ConvertFieldToTypesField(field)
	if err != nil {
		return nil, fmt.Errorf("failed to convert field: %w", err)
	}

	return typesField, nil
}

// ParseMessageFields 解析消息字段（通用逻辑）
func (p *BaseParser) ParseMessageFields(fieldsMap map[string]interface{}) ([]Field, error) {
	result := make([]Field, 0)

	for fieldName, fieldDef := range fieldsMap {
		if fieldStr, ok := fieldDef.(string); ok {
			field, err := ParseFieldDefinition(fieldName + " " + fieldStr)
			if err != nil {
				continue // 跳过解析失败的字段
			}
			result = append(result, field)
		}
	}

	return result, nil
}

// ParseMessageFieldsToTypes 解析消息字段为 types.Field 列表（通用逻辑）
func (p *BaseParser) ParseMessageFieldsToTypes(fieldsMap map[string]interface{}) ([]*types.Field, error) {
	result := make([]*types.Field, 0)

	for fieldName, fieldDef := range fieldsMap {
		if fieldStr, ok := fieldDef.(string); ok {
			field, err := ParseFieldDefinition(fieldName + " " + fieldStr)
			if err != nil {
				continue // 跳过解析失败的字段
			}

			typesField, err := ConvertFieldToTypesField(field)
			if err != nil {
				continue // 跳过转换失败的字段
			}

			result = append(result, typesField)
		}
	}

	return result, nil
}

// ParseRequestOrNotify 解析请求或通知（通用逻辑）
func (p *BaseParser) ParseRequestOrNotify(item interface{}, parseResponse bool) (Request, error) {
	req := Request{}

	if reqMap, ok := item.(map[string]interface{}); ok {
		for name, data := range reqMap {
			req.Name = name

			if dataMap, ok := data.(map[string]interface{}); ok {
				// 解析字段
				fields := make([]Field, 0)
				for fieldName, fieldDef := range dataMap {
					if fieldName == "Rsp" && parseResponse {
						// 解析响应
						if rspData, ok := fieldDef.(map[string]interface{}); ok {
							rsp := Response{}
							rspFields, err := p.ParseMessageFields(rspData)
							if err == nil {
								rsp.Fields = rspFields
							}
							req.Rsp = &rsp
						}
					} else {
						// 解析请求字段
						if fieldStr, ok := fieldDef.(string); ok {
							field, err := ParseFieldDefinition(fieldName + " " + fieldStr)
							if err == nil {
								fields = append(fields, field)
							}
						}
					}
				}
				req.Fields = fields
			}
		}
	} else if reqName, ok := item.(string); ok {
		// 简单请求，无参数
		req.Name = reqName
	}

	return req, nil
}

// ParseNotifyOnly 解析通知（不含响应）
func (p *BaseParser) ParseNotifyOnly(item interface{}) (Notify, error) {
	notify := Notify{}

	if notifyMap, ok := item.(map[string]interface{}); ok {
		for name, data := range notifyMap {
			notify.Name = name

			if dataMap, ok := data.(map[string]interface{}); ok {
				fields, err := p.ParseMessageFields(dataMap)
				if err == nil {
					notify.Fields = fields
				}
			}
		}
	}

	return notify, nil
}
