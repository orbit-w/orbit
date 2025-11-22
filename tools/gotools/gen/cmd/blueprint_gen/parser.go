package blueprint_gen

import (
	"fmt"
	"regexp"
	"strings"

	types "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types"
)

// YamlParser YAML 解析器
type YamlParser struct {
	*BaseParser
}

// NewParser 创建新的解析器
func NewParser(blueprintDir string) *YamlParser {
	return &YamlParser{
		BaseParser: NewBaseParser(blueprintDir),
	}
}

// Parse 解析所有 YAML 文件
func (p *YamlParser) Parse() error {
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

	// 检查 Entity 字段类型是否符合要求
	p.ctx.CheckEntityFields()

	return nil
}

// parseHeadFile 解析 headfile.yaml
func (p *YamlParser) parseHeadFile() error {
	filePath := p.ResolvePath("mme", "headfile.yaml")
	if !FileExists(filePath) {
		return nil
	}

	yamlData, err := p.ReadYAMLFile(filePath)
	if err != nil {
		return err
	}

	// 创建 HeadFileConfig
	headFileConfig := &HeadFileConfig{
		EntityFields:              make(map[string]any),
		ModuleFields:              make(map[string]any),
		ModuleStorageOption:       make([]string, 0),
		MechanismFields:           make(map[string]any),
		MechanismDataFieldOptions: make(map[string]FieldOptionDefinition),
		CommonDataStructs:         make([]DataStruct, 0),
	}

	// 解析 DataStruct
	if dataStructList, ok := yamlData["DataStruct"].([]any); ok {
		for _, dsItem := range dataStructList {
			dsMap, ok := dsItem.(map[string]any)
			if !ok {
				continue
			}

			for dsName, dsData := range dsMap {
				dataStruct := DataStruct{
					Name:   dsName,
					Fields: make([]*types.Field, 0),
				}

				if dsData != nil {
					fieldsMap, ok := dsData.(map[string]any)
					if ok {
						// 解析字段
						for fieldKey, fieldValue := range fieldsMap {
							field := parseField(nil, fieldKey, fieldValue)
							if field == nil {
								panic(fmt.Sprintf("failed to parse field %s in DataStruct %s", fieldKey, dsName))
							}
							dataStruct.Fields = append(dataStruct.Fields, field)
						}
						// 按字段编号排序
						sortFieldsByNumber(dataStruct.Fields)
					}
				}

				headFileConfig.CommonDataStructs = append(headFileConfig.CommonDataStructs, dataStruct)
			}
		}
	}

	// 解析 Enums
	if enumList, ok := yamlData["Enums"].([]any); ok && enumList != nil {
		enumParser := NewEnumParser()
		for _, enumItem := range enumList {
			enum := enumParser.ParseEnumItem(enumItem)
			if enum == nil {
				panic(fmt.Sprintf("Enum not found for item %v", enumItem))
			}
			// headfile.yaml 中的枚举来源标记为 "common"
			enum.SourceProto = "common"
			p.ctx.AddEnum(enum)
		}
	}

	// 设置 HeadFile 配置
	p.ctx.SetHeadFile(headFileConfig)

	return nil
}

type EnumParser struct {
}

func NewEnumParser() *EnumParser {
	return &EnumParser{}
}

// parseEnumItem 解析单个枚举项
// 格式:
//   - ServiceZoneType:
//     ServiceZoneTypePlay: 0 [Content:"逻辑服管理区域类型"]
//     ServiceZoneTypeUnion: 1 [Content:"联盟服管理区域类型"]
func (p *EnumParser) ParseEnumItem(item any) *Enum {
	itemMap := item.(map[string]any)
	for enumName, enumData := range itemMap {
		enum := &Enum{
			Name:    enumName,
			Values:  make([]*EnumValue, 0),
			Comment: "",
		}

		if enumData == nil {
			return enum
		}

		// enumData 应该是一个 map，包含枚举值
		valuesMap, ok := enumData.(map[string]any)
		if !ok {
			panic(fmt.Sprintf("Invalid enum values format for %s", enumName))
		}

		// 解析每个枚举值
		for valueName, valueDef := range valuesMap {
			enumValue := p.parseEnumValue(valueName, valueDef)
			if enumValue != nil {
				enum.Values = append(enum.Values, enumValue)
			}
		}

		// 按编号排序枚举值
		for i := 0; i < len(enum.Values)-1; i++ {
			for j := i + 1; j < len(enum.Values); j++ {
				if enum.Values[i].Number > enum.Values[j].Number {
					enum.Values[i], enum.Values[j] = enum.Values[j], enum.Values[i]
				}
			}
		}

		return enum
	}
	return nil
}

// parseEnumValue 解析单个枚举值
// 格式: ServiceZoneTypePlay: 0 [Content:"逻辑服管理区域类型"]
// 或者: ServiceZoneTypePlay: 0 [Content:"逻辑服管理区域类型", OtherKey:"OtherValue"]
// 或者: ServiceZoneTypePlay: 0
func (p *EnumParser) parseEnumValue(valueName string, valueDef any) *EnumValue {
	ev := &EnumValue{
		Name: valueName,
	}

	// 尝试不同的解析方式
	switch v := valueDef.(type) {
	case int:
		ev.Number = int32(v)
	case int32:
		ev.Number = v
	case int64:
		ev.Number = int32(v)
	case string:
		// 解析字符串格式: "0 [Content:\"注释\"]" 或 "0 [Content:\"注释\", OtherKey:\"OtherValue\"]" 或 "0"
		number, kvPairs := p.parseEnumValueString(v)
		ev.Number = number
		ev.Options = kvPairs
		ev.ParseComment()
	default:
		panic(fmt.Sprintf("Invalid enum value type: %T", valueDef))
	}

	return ev
}

// parseEnumValueString 解析枚举值字符串
// 格式: "0 [Content:\"注释\"]" 或 "0 [Content:\"注释\", OtherKey:\"OtherValue\"]" 或 "0"
func (p *EnumParser) parseEnumValueString(valueStr string) (int32, map[string]string) {
	var (
		number  int32
		kvPairs map[string]string
	)
	// 提取数字部分和中括号内容
	re := regexp.MustCompile(`(\d+)\s*(?:\[([^\]]+)\])?`)
	matches := re.FindStringSubmatch(valueStr)
	if len(matches) >= 2 {
		// 解析数字
		if _, err := fmt.Sscanf(matches[1], "%d", &number); err != nil {
			panic(fmt.Sprintf("Failed to parse enum value number: %s", valueStr))
		}

		// 解析中括号内的内容（如果存在）
		if len(matches) >= 3 && matches[2] != "" {
			bracketContent := matches[2]
			// 解析 Key-Value 对
			kvPairs = p.parseBracketKeyValuePairs(bracketContent)
		}
	} else {
		// 尝试直接解析为数字
		if _, err := fmt.Sscanf(valueStr, "%d", &number); err != nil {
			panic(fmt.Sprintf("Failed to parse enum value: %s", valueStr))
		}
	}

	return number, kvPairs
}

// parseBracketKeyValuePairs 解析中括号内的 Key-Value 对
// 格式: Content:"逻辑服管理区域类型", OtherKey:"OtherValue"
// 或者: Content : "逻辑服管理区域类型" , OtherKey : "OtherValue" (支持空格)
// 按逗号分割，每个部分按冒号分割，左边是 Key，右边是 Value
func (p *EnumParser) parseBracketKeyValuePairs(bracketContent string) map[string]string {
	kvPairs := make(map[string]string)

	// 使用正则表达式匹配 Key:"Value" 或 Key : "Value" 格式
	// 允许 Key 和冒号之间、冒号和引号之间、逗号前后有空格
	re := regexp.MustCompile(`(\w+)\s*:\s*"([^"]+)"`)
	matches := re.FindAllStringSubmatch(bracketContent, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			key := strings.TrimSpace(match[1])
			value := strings.TrimSpace(match[2])
			kvPairs[key] = value
		}
	}

	return kvPairs
}
