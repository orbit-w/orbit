package blueprint_gen

import (
	"fmt"
	"maps"
	"sort"
	"strings"

	"gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/mmeobj"
	field_parser "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/parser"
	types "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types"
)

// parseMMEFiles 解析 MME YAML 文件
func (p *YamlParser) parseMMEFiles() error {
	// 解析 entities.yaml
	if err := p.ParseEntities(); err != nil {
		return fmt.Errorf("failed to parse entities: %w", err)
	}

	// 解析 manager.yaml
	if err := p.ParseManagers(); err != nil {
		return fmt.Errorf("failed to parse managers: %w", err)
	}

	// 解析 modules.yaml
	if err := p.ParseModules(); err != nil {
		return fmt.Errorf("failed to parse modules: %w", err)
	}

	// 解析 mechanisms.yaml
	if err := p.ParseMechanisms(); err != nil {
		return fmt.Errorf("failed to parse mechanisms: %w", err)
	}

	// 链接 Modules 和 Mechanisms
	p.ctx.LinkModules()
	// 链接 Managers 和 Modules
	p.ctx.LinkManagers()

	return nil
}

// ParseEntities 解析 entities.yaml
func (p *YamlParser) ParseEntities() error {
	filePath := p.ResolvePath("mme", "entities.yaml")
	if !FileExists(filePath) {
		return nil
	}

	yamlData, err := p.ReadYAMLFile(filePath)
	if err != nil {
		return err
	}

	entityList, ok := yamlData["Entity"].([]any)
	if !ok {
		return nil
	}

	for _, entityItem := range entityList {
		entityMap, ok := entityItem.(map[string]any)
		if !ok {
			continue
		}

		for entityName, entityData := range entityMap {
			entity := p.parseEntityFromMap(entityName, entityData)
			if len(entity.Fields) > 0 {
				p.ctx.AddEntity(entity)
			}
		}
	}

	// 解析枚举
	enumList, ok := yamlData["Enums"].([]any)
	if ok && enumList != nil {
		enumParser := NewEnumParser()
		for _, enumItem := range enumList {
			enum := enumParser.ParseEnumItem(enumItem)
			if enum == nil {
				panic(fmt.Sprintf("Enum not found for item %v", enumItem))
			}
			enum.SourceProto = "entities"
			p.ctx.AddEnum(enum)
		}
	}

	// 解析 Register，生成 Entity 类型枚举
	if err := p.parseEntityRegister(yamlData); err != nil {
		return fmt.Errorf("failed to parse Register: %w", err)
	}

	return nil
}

// parseEntityFromMap 从 map 数据中解析单个 Entity
// entityName: 实体名称
// entityData: 实体数据，通常是包含字段定义的 map
func (p *YamlParser) parseEntityFromMap(entityName string, entityData any) *mmeobj.Entity {
	entity := mmeobj.NewEntity()
	entity.Name = entityName

	fieldsMap, ok := entityData.(map[string]any)
	if !ok {
		return entity
	}

	// Entity 字段格式特殊（"ManagerName FieldName"），需要特殊处理
	for fieldKey, fieldValue := range fieldsMap {
		field := parseField(nil, fieldKey, fieldValue)
		if field == nil {
			panic(fmt.Sprintf("failed to parse field %s", fieldKey))
		}
		entity.Fields = append(entity.Fields, field)
	}

	// 按字段编号排序，保证字段顺序
	sortFieldsByNumber(entity.Fields)

	return entity
}

// extractNumberFromValue 从值中提取数字（支持多种数字类型）
func extractNumberFromValue(value any) int32 {
	switch v := value.(type) {
	case int:
		return int32(v)
	case int32:
		return v
	case int64:
		return int32(v)
	default:
		return 0
	}
}

// ParseManagers 解析 manager.yaml
func (p *YamlParser) ParseManagers() error {
	filePath := p.ResolvePath("mme", "manager.yaml")
	if !FileExists(filePath) {
		return nil
	}

	yamlData, err := p.ReadYAMLFile(filePath)
	if err != nil {
		return err
	}

	managerList, ok := yamlData["Managers"].([]any)
	if !ok {
		return nil
	}

	p.parseManagers(managerList)

	// 解析枚举
	enumList, ok := yamlData["Enums"].([]any)
	if ok && enumList != nil {
		enumParser := NewEnumParser()
		for _, enumItem := range enumList {
			enum := enumParser.ParseEnumItem(enumItem)
			if enum == nil {
				panic(fmt.Sprintf("Enum not found for item %v", enumItem))
			}
			enum.SourceProto = "managers"
			p.ctx.AddEnum(enum)
		}
	}

	return nil
}

func (p *YamlParser) parseManagers(items []any) {
	for _, item := range items {
		managerMap, ok := item.(map[string]any)
		if !ok {
			continue
		}

		// 遍历 map，区分 Manager 名称和字段定义
		// Manager 名称：key 不包含特殊字符（<, >, :），value 可能是 nil 或 map
		// 字段定义：key 包含特殊字符（<, >, :）
		for managerName, managerData := range managerMap {
			// 判断是否为 Manager 名称（不包含特殊字符的 key）
			if containsSpecialChars(managerName, []string{"<", ">", ":"}) {
				panic(fmt.Sprintf("managerName %s contains special chars", managerName))
			}

			// 这是 Manager 名称
			manager := p.parseManagerFromMap(managerName, managerData, managerMap)
			if len(manager.Fields) > 0 {
				p.ctx.AddManager(manager)
			}
		}
	}
}

// parseManagerFromMap 从 map 数据中解析单个 Manager 对象
// managerName: Manager 名称
// managerData: Manager 数据，可能是 nil 或包含字段定义的 map
// managerMap: 完整的 managerMap，用于从平铺结构中提取字段
// 返回: 解析后的 Manager
func (p *YamlParser) parseManagerFromMap(managerName string, managerData any, managerMap map[string]any) *mmeobj.Manager {
	manager := mmeobj.NewManager()
	manager.Name = managerName

	// 判断字段定义的位置
	var fieldsMap map[string]any
	if managerData != nil {
		// 如果 managerData 是 map，则字段定义在其中
		if dataMap, ok := managerData.(map[string]any); ok {
			fieldsMap = dataMap
		}
	}

	// 如果 managerData 是 nil 或不是 map，字段定义在平铺的 managerMap 中
	if fieldsMap == nil {
		fieldsMap = managerMap
	}

	fields := p.parseManagerFields(fieldsMap, managerName)
	manager.Fields = fields

	return manager
}

// parseManagerFields 解析 Manager 的字段
// fieldsMap: 包含字段定义的 map
// managerName: Manager 名称，用于验证
// 返回: 解析后的字段列表（已按编号排序）
func (p *YamlParser) parseManagerFields(fieldsMap map[string]any, managerName string) []*types.Field {
	// 如果 fieldsMap 中包含 managerName，说明 yaml 文件结构有问题
	if _, exists := fieldsMap[managerName]; exists {
		panic(fmt.Sprintf("fieldsMap contains managerName '%s', which indicates a malformed yaml structure", managerName))
	}

	return parseFieldsFromMap(fieldsMap, nil)
}

// parseModules 解析 modules.yaml
func (p *YamlParser) ParseModules() error {
	filePath := p.ResolvePath("mme", "modules.yaml")
	if !FileExists(filePath) {
		return nil
	}

	yamlData, err := p.ReadYAMLFile(filePath)
	if err != nil {
		return err
	}

	moduleList := yamlData["Modules"].([]any)
	for _, moduleItem := range moduleList {
		moduleMap := moduleItem.(map[string]any)
		module := mmeobj.NewModule()
		for moduleName, moduleData := range moduleMap {
			module.Name = moduleName
			p.parseModuleItem(module, moduleData.([]any))
		}
		if len(module.Fields) > 0 {
			p.ctx.AddModule(module)
		}
	}

	// 解析枚举
	enumList, ok := yamlData["Enums"].([]any)
	if ok && enumList != nil {
		enumParser := NewEnumParser()
		for _, enumItem := range enumList {
			enum := enumParser.ParseEnumItem(enumItem)
			if enum == nil {
				panic(fmt.Sprintf("Enum not found for item %v", enumItem))
			}
			enum.SourceProto = "modules"
			p.ctx.AddEnum(enum)
		}
	}

	return nil
}

// parseModuleItem 解析 Module 的单个 Item
// module: 模块
// moduleItem: 模块的 Item，通常是包含字段定义/Settings定义的 map
// 返回: 解析后的 Module
func (p *YamlParser) parseModuleItem(module *mmeobj.Module, moduleItem []any) {
	for i := range moduleItem {
		keyWordItem := moduleItem[i]
		keyWordItemMap := keyWordItem.(map[string]any)
		for keyWord, keyWordItem := range keyWordItemMap {
			switch keyWord {
			case ModuleKeyWordSettings:
				settings := keyWordItem.(map[string]any)
				if settings != nil {
					module.SetSettings(settings)
				}
			default:
				//如果不是其他特殊关键字，则认为是Field定义
				field := parseField(nil, keyWord, keyWordItem)
				if field == nil {
					panic(fmt.Sprintf("failed to parse field %s", keyWord))
				}

				module.AddField(field)
			}
		}
	}
}

// ParseMechanisms 解析 mechanisms.yaml
func (p *YamlParser) ParseMechanisms() error {
	filePath := p.ResolvePath("mme", "mechanisms.yaml")
	if !FileExists(filePath) {
		return nil
	}

	yamlData, err := p.ReadYAMLFile(filePath)
	if err != nil {
		return err
	}

	if mechanismList, ok := yamlData["Mechanisms"].([]any); ok {
		for _, mechItem := range mechanismList {
			m := p.parseMechanismItem(mechItem.(map[string]any))
			if m != nil {
				p.ctx.AddMechanism(m)
			}
		}
	}

	// 解析枚举
	enumList, ok := yamlData["Enums"].([]any)
	if ok && enumList != nil {
		enumParser := NewEnumParser()
		for _, enumItem := range enumList {
			enum := enumParser.ParseEnumItem(enumItem)
			if enum == nil {
				panic(fmt.Sprintf("Enum not found for item %v", enumItem))
			}
			enum.SourceProto = "mechanisms"
			p.ctx.AddEnum(enum)
		}
	}

	// 解析 Register，生成 Mechanism 类型枚举
	if err := p.parseMechanismRegister(yamlData); err != nil {
		return fmt.Errorf("failed to parse Mechanism Register: %w", err)
	}

	return nil
}

func (p *YamlParser) parseMechanismItem(mechItem map[string]any) *mmeobj.Mechanism {
	mechanism := mmeobj.NewMechanism()
	for mechKey, mechData := range mechItem {
		switch mechKey {
		case MechanismKeyWordSettings:
			settings := mechData.(map[string]any)
			if settings != nil {
				maps.Copy(mechanism.Settings, settings)
			}
		case MechanismKeyWordRequests:
		case MechanismKeyWordNotifies:
		default:
			// 解析结构体名称和Fields
			mechanism.Name = mechKey
			mechFieldsMap := mechData.(map[string]any)
			mechanism.Fields = parseFieldsFromMap(mechFieldsMap, nil)
		}
	}

	return mechanism
}

// ============================================================================
// 通用解析工具函数
// ============================================================================

// FieldParserConfig 字段解析配置
type FieldParserConfig struct {
	// ParseFieldFunc 用于解析字段，返回 *types.Field
	// 如果为 nil，使用默认的 parseAndConvertFieldDefinition
	ParseFieldFunc func(fieldDef string) *types.Field
}

// parseFieldsFromMap 从 map 中解析字段（通用函数）
// fieldsMap: 包含字段定义的 map
// config: 解析配置
// 返回: 解析后的字段列表（已按编号排序）
func parseFieldsFromMap(fieldsMap map[string]any, config *FieldParserConfig) []*types.Field {
	fields := make([]*types.Field, 0)

	for fieldKey, fieldValue := range fieldsMap {
		field := parseField(config, fieldKey, fieldValue)
		if field == nil {
			panic(fmt.Sprintf("failed to parse field %s", fieldKey))
		}
		fields = append(fields, field)
	}

	// 按字段编号排序，保证字段顺序
	sortFieldsByNumber(fields)

	return fields
}

func parseField(config *FieldParserConfig, fieldKey string, fieldValue any) *types.Field {
	// 构建字段定义
	fieldDef := buildFieldDefinition(fieldKey, fieldValue)
	if fieldDef == "" {
		return nil
	}

	// 解析字段
	var field *types.Field
	if config != nil && config.ParseFieldFunc != nil {
		field = config.ParseFieldFunc(fieldDef)
	} else {
		field = parseAndConvertFieldDefinition(fieldDef)
	}

	return field
}

// buildFieldDefinition 构建字段定义字符串（通用函数）
// fieldKey: 字段键，可能是完整的字段定义字符串，也可能是字段名
// fieldValue: 字段值，可能是编号、字符串或其他类型
// 返回: 完整的字段定义字符串
func buildFieldDefinition(fieldKey string, fieldValue any) string {
	if fieldValue == nil {
		return fieldKey
	}

	// 如果 fieldValue 是数字，添加到定义中
	if num := extractNumberFromValue(fieldValue); num > 0 {
		// 检查 fieldKey 是否已经包含编号
		if strings.Contains(fieldKey, ":") {
			return fieldKey
		}
		return fmt.Sprintf("%s: %d", fieldKey, num)
	}

	// 如果 fieldValue 是字符串，追加到定义中
	if valueStr, ok := fieldValue.(string); ok {
		// 检查 fieldKey 是否已经包含该字符串
		if strings.Contains(fieldKey, valueStr) {
			return fieldKey
		}
		return fieldKey + ": " + valueStr
	}

	return fieldKey
}

// parseAndConvertFieldDefinition 解析字段定义并转换为 *types.Field（通用函数）
// fieldDef: 字段定义字符串，格式: "int32 FieldName: 1 [blueprint:\"access=all\"]"
// 返回: 解析后的 *types.Field，如果解析失败返回 nil
func parseAndConvertFieldDefinition(fieldDef string) *types.Field {
	field := field_parser.ParseFieldDefinition(fieldDef)
	return field
}

// sortFieldsByNumber 按字段编号排序字段列表（通用函数）
func sortFieldsByNumber(fields []*types.Field) {
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Number < fields[j].Number
	})
}

// parseEntityRegister 解析 Register 部分，生成 Entity 类型枚举
// Register 格式:
//
//	Register:
//	  - PlayerEntity: 1
//	  - AnotherEntity: 2
func (p *YamlParser) parseEntityRegister(yamlData map[string]any) error {
	registerList, ok := yamlData["Register"].([]any)
	if !ok || registerList == nil {
		return nil // Register 是可选的
	}

	// 构建 Entity 名称映射，用于验证
	entityNameMap := make(map[string]bool)
	for _, entity := range p.ctx.Entities {
		entityNameMap[entity.Name] = true
	}

	// 创建枚举
	enum := NewEnum("EntityType", "Entity 类型枚举，由 entity register 自动生成", "common")

	// 添加默认枚举值 0（protobuf 要求第一个枚举值必须为 0）
	defaultEnumValue := NewEnumValue("EntityTypeUnknown", 0, "未知 Entity 类型")
	defaultEnumValue.Options[EnumValueOptionContent] = defaultEnumValue.Comment
	enum.Values = append(enum.Values, defaultEnumValue)

	// 解析每个 Register 项
	for _, registerItem := range registerList {
		registerMap, ok := registerItem.(map[string]any)
		if !ok {
			continue
		}

		for entityKey, entityValue := range registerMap {
			// 验证 Entity 是否存在
			if !entityNameMap[entityKey] {
				panic(fmt.Sprintf("Register 中定义的 Entity '%s' 不存在，请确保在 Entity 部分已定义", entityKey))
			}

			// 提取枚举值编号
			enumValueNumber := extractNumberFromValue(entityValue)
			if enumValueNumber == 0 {
				panic(fmt.Sprintf("Register 中 Entity '%s' 的编号不能为 0", entityKey))
			}

			// 生成枚举值名称（Entity 名称 + "Type"）
			enumValueName := GenEntityTypeEnumName(entityKey)

			// 创建枚举值
			enumValue := NewEnumValue(enumValueName, enumValueNumber, fmt.Sprintf("Entity 类型: %s", entityKey))
			enumValue.Options[EnumValueOptionContent] = enumValue.Comment

			enum.Values = append(enum.Values, enumValue)
		}
	}

	// 按编号排序枚举值
	sort.Slice(enum.Values, func(i, j int) bool {
		return enum.Values[i].Number < enum.Values[j].Number
	})

	// 检查是否有重复的编号
	numberMap := make(map[int32]string)
	for _, value := range enum.Values {
		if existingKey, exists := numberMap[value.Number]; exists {
			panic(fmt.Sprintf("Register 中存在重复的编号 %d: '%s' 和 '%s'", value.Number, existingKey, value.Name))
		}
		numberMap[value.Number] = value.Name
	}

	// 添加到上下文
	if len(enum.Values) > 0 {
		p.ctx.AddEnum(enum)
	}

	return nil
}

// parseMechanismRegister 解析 Mechanism Register 部分，生成 Mechanism 类型枚举
// Register 格式:
//
//	Register:
//	  - LevelUpMechanism: 1
//	  - HeroMechanism: 2
func (p *YamlParser) parseMechanismRegister(yamlData map[string]any) error {
	registerList, ok := yamlData["Register"].([]any)
	if !ok || registerList == nil {
		return nil // Register 是可选的
	}

	// 构建 Mechanism 名称映射，用于验证
	mechanismNameMap := make(map[string]bool)
	for _, mechanism := range p.ctx.Mechanisms {
		mechanismNameMap[mechanism.Name] = true
	}

	// 创建枚举
	enum := NewEnum("MechanismType", "Mechanism 类型枚举，由 mechanism register 自动生成", "common")

	// 添加默认枚举值 0（protobuf 要求第一个枚举值必须为 0）
	defaultEnumValue := NewEnumValue("Unknown", 0, "未知 Mechanism 类型")
	defaultEnumValue.Options[EnumValueOptionContent] = defaultEnumValue.Comment
	enum.Values = append(enum.Values, defaultEnumValue)

	// 解析每个 Register 项
	for _, registerItem := range registerList {
		registerMap, ok := registerItem.(map[string]any)
		if !ok {
			continue
		}

		for mechanismKey, mechanismValue := range registerMap {
			// 验证 Mechanism 是否存在
			if !mechanismNameMap[mechanismKey] {
				panic(fmt.Sprintf("Register 中定义的 Mechanism '%s' 不存在，请确保在 Mechanisms 部分已定义", mechanismKey))
			}

			// 提取枚举值编号
			enumValueNumber := extractNumberFromValue(mechanismValue)
			if enumValueNumber == 0 {
				panic(fmt.Sprintf("Register 中 Mechanism '%s' 的编号不能为 0", mechanismKey))
			}

			// 生成枚举值名称（Mechanism 名称 + "Type"）
			enumValueName := GenMechanismTypeEnumName(mechanismKey)

			// 创建枚举值
			enumValue := NewEnumValue(enumValueName, enumValueNumber, fmt.Sprintf("Mechanism 类型: %s", mechanismKey))
			enumValue.Options[EnumValueOptionContent] = enumValue.Comment

			enum.Values = append(enum.Values, enumValue)
		}
	}

	// 按编号排序枚举值
	sort.Slice(enum.Values, func(i, j int) bool {
		return enum.Values[i].Number < enum.Values[j].Number
	})

	// 检查是否有重复的编号
	numberMap := make(map[int32]string)
	for _, value := range enum.Values {
		if existingKey, exists := numberMap[value.Number]; exists {
			panic(fmt.Sprintf("Register 中存在重复的编号 %d: '%s' 和 '%s'", value.Number, existingKey, value.Name))
		}
		numberMap[value.Number] = value.Name
	}

	// 添加到上下文
	if len(enum.Values) > 0 {
		p.ctx.AddEnum(enum)
	}

	return nil
}
