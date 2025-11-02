package blueprint_gen

import (
	"fmt"
	"os"
	"path/filepath"
	
	"gopkg.in/yaml.v3"
)

// Parser YAML 解析器
type Parser struct {
	blueprintDir string
	data         *BlueprintData
}

// NewParser 创建新的解析器
func NewParser(blueprintDir string) *Parser {
	return &Parser{
		blueprintDir: blueprintDir,
		data: &BlueprintData{
			Entities:   make([]Entity, 0),
			Managers:   make([]Manager, 0),
			Modules:    make([]Module, 0),
			Mechanisms: make([]Mechanism, 0),
			NetWalls:   make([]NetWallFile, 0),
		},
	}
}

// Parse 解析所有 YAML 文件
func (p *Parser) Parse() error {
	// 解析 headfile.yaml
	if err := p.parseHeadFile(); err != nil {
		return fmt.Errorf("failed to parse headfile: %w", err)
	}
	
	// 解析 MME 文件
	if err := p.parseMMEFiles(); err != nil {
		return fmt.Errorf("failed to parse MME files: %w", err)
	}
	
	// 解析 NetWall 文件
	if err := p.parseNetWallFiles(); err != nil {
		return fmt.Errorf("failed to parse NetWall files: %w", err)
	}
	
	return nil
}

// GetData 获取解析后的数据
func (p *Parser) GetData() *BlueprintData {
	return p.data
}

// parseHeadFile 解析 headfile.yaml
func (p *Parser) parseHeadFile() error {
	filePath := filepath.Join(p.blueprintDir, "mme", "headfile.yaml")
	if !FileExists(filePath) {
		return nil // headfile 可选
	}
	
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	
	// headfile.yaml 的格式是列表，每个列表项是一个对象
	var yamlData []interface{}
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		return err
	}
	
	config := &HeadFileConfig{
		EntityFields:              make(map[string]interface{}),
		ModuleFields:              make(map[string]interface{}),
		MechanismFields:           make(map[string]interface{}),
		MechanismDataFieldOptions: make(map[string]FieldOptionDefinition),
		CommonDataStructs:         make([]DataStruct, 0),
	}
	
	// 遍历列表项
	for _, item := range yamlData {
		if itemMap, ok := item.(map[string]interface{}); ok {
			// 解析 EntityFields
			if entityFields, ok := itemMap["EntityFields"].(map[string]interface{}); ok {
				for k, v := range entityFields {
					config.EntityFields[k] = v
				}
			} else if _, ok := itemMap["EntityFields"]; ok {
				// EntityFields 存在但为空（可能是 nil 或空 map）
				// 保持空 map
			}
			
			// 解析 ModuleFields
			if moduleFields, ok := itemMap["ModuleFields"].(map[string]interface{}); ok {
				for k, v := range moduleFields {
					config.ModuleFields[k] = v
				}
			}
			
			// 解析 ModuleStorageOption
			if moduleStorageOption, ok := itemMap["ModuleStorageOption"].(map[string]interface{}); ok {
				if choices, ok := moduleStorageOption["Choices"].([]interface{}); ok {
					config.ModuleStorageOption = make([]string, 0, len(choices))
					for _, choice := range choices {
						if s, ok := choice.(string); ok {
							config.ModuleStorageOption = append(config.ModuleStorageOption, s)
						}
					}
				}
			}
			
			// 解析 MechanismFields
			if mechanismFields, ok := itemMap["MechanismFields"].(map[string]interface{}); ok {
				for k, v := range mechanismFields {
					config.MechanismFields[k] = v
				}
			}
			
			// 解析 MechanismDataFieldOptions
			if options, ok := itemMap["MechanismDataFieldOption"].([]interface{}); ok {
				for _, opt := range options {
					if optMap, ok := opt.(map[string]interface{}); ok {
						for name, def := range optMap {
							optionDef := FieldOptionDefinition{}
							if defMap, ok := def.(map[string]interface{}); ok {
								if defaultVal, ok := defMap["Default"].(string); ok {
									optionDef.Default = defaultVal
								}
								if enum, ok := defMap["Enum"].(string); ok {
									optionDef.Enum = enum
								}
								if desc, ok := defMap["Description"].(string); ok {
									optionDef.Description = desc
								}
								if choices, ok := defMap["Choices"].([]interface{}); ok {
									optionDef.Choices = make([]string, 0, len(choices))
									for _, choice := range choices {
										if s, ok := choice.(string); ok {
											optionDef.Choices = append(optionDef.Choices, s)
										}
									}
								}
							}
							config.MechanismDataFieldOptions[name] = optionDef
						}
					}
				}
			}
			
			// 解析 CommonDataStruct
			if commonStructs, ok := itemMap["CommonDataStruct"].([]interface{}); ok {
				for _, cs := range commonStructs {
					if csMap, ok := cs.(map[string]interface{}); ok {
						for name, fields := range csMap {
							structFields, err := parseDataStructFields(fields)
							if err != nil {
								continue
							}
							config.CommonDataStructs = append(config.CommonDataStructs, DataStruct{
								Name:   name,
								Fields: structFields,
							})
						}
					}
				}
			}
		}
	}
	
	p.data.HeadFile = config
	return nil
}

// parseDataStructFields 解析数据结构的字段
func parseDataStructFields(fields interface{}) ([]Field, error) {
	if fieldsMap, ok := fields.(map[string]interface{}); ok {
		result := make([]Field, 0)
		for fieldName, fieldDef := range fieldsMap {
			// fieldDef 可能是字符串 "int32: 1" 或者更复杂的结构
			if fieldStr, ok := fieldDef.(string); ok {
				// 格式: "int32: 1" 或 "map<int32, int32>: 2"
				field, err := ParseFieldDefinition(fieldName + " " + fieldStr)
				if err != nil {
					continue
				}
				result = append(result, field)
			}
		}
		return result, nil
	}
	return nil, fmt.Errorf("invalid fields format")
}

