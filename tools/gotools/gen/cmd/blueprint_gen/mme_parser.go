package blueprint_gen

import (
	"fmt"
	"maps"
	"sort"
	"strings"

	types "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types"
)

// parseMMEFiles 解析 MME YAML 文件
func (p *Parser) parseMMEFiles() error {
	// 解析 entities.yaml
	if err := p.parseEntities(); err != nil {
		return fmt.Errorf("failed to parse entities: %w", err)
	}

	// 解析 manager.yaml
	if err := p.ParseManagers(); err != nil {
		return fmt.Errorf("failed to parse managers: %w", err)
	}

	// 解析 modules.yaml
	if err := p.parseModules(); err != nil {
		return fmt.Errorf("failed to parse modules: %w", err)
	}

	// 解析 mechanisms.yaml
	if err := p.parseMechanisms(); err != nil {
		return fmt.Errorf("failed to parse mechanisms: %w", err)
	}

	return nil
}

// parseEntities 解析 entities.yaml
func (p *Parser) parseEntities() error {
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

	return nil
}

// parseEntityFromMap 从 map 数据中解析单个 Entity
// entityName: 实体名称
// entityData: 实体数据，通常是包含字段定义的 map
func (p *Parser) parseEntityFromMap(entityName string, entityData any) Entity {
	entity := Entity{Name: entityName}

	fieldsMap, ok := entityData.(map[string]any)
	if !ok {
		return entity
	}

	// Entity 字段格式特殊（"ManagerName FieldName"），需要特殊处理
	for fieldKey, fieldValue := range fieldsMap {
		field := parseEntityField(fieldKey, fieldValue)
		if field != nil {
			entity.Fields = append(entity.Fields, field)
		}
	}

	// 按字段编号排序，保证字段顺序
	sortFieldsByNumber(entity.Fields)

	return entity
}

// parseEntityField 解析 EntityField，返回 *types.Field
// fieldKey 格式: "HeroManager HeroManager"
// fieldValue 格式: "1 [blueprint:\"access=all\"]" 或数字
func parseEntityField(fieldKey string, fieldValue any) *types.Field {
	// 使用通用工具函数构建字段定义字符串
	fieldDef := buildFieldDefinition(fieldKey, fieldValue)
	if fieldDef == "" {
		return nil
	}

	// 使用通用工具函数解析字段定义
	field := parseAndConvertFieldDefinition(fieldDef)
	if field == nil {
		return nil
	}

	// Entity 字段的特殊处理：类型是 MMEObject，TypeName 是 ManagerName
	// 从 fieldKey 提取 ManagerName（第一个词）
	keyParts := strings.Fields(fieldKey)
	if len(keyParts) < 2 {
		return nil
	}
	managerName := keyParts[0]

	// 调整类型信息为 MMEObject
	field.Type.Kind = types.FieldKindMMEObject
	field.Type.Label = types.FieldLabelOptional
	field.Type.TypeName = managerName

	return field
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
func (p *Parser) ParseManagers() error {
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

	return nil
}

func (p *Parser) parseManagers(items []any) {
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
func (p *Parser) parseManagerFromMap(managerName string, managerData any, managerMap map[string]any) Manager {
	manager := Manager{Name: managerName}

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
func (p *Parser) parseManagerFields(fieldsMap map[string]any, managerName string) []*types.Field {
	// 如果 fieldsMap 中包含 managerName，说明 yaml 文件结构有问题
	if _, exists := fieldsMap[managerName]; exists {
		panic(fmt.Sprintf("fieldsMap contains managerName '%s', which indicates a malformed yaml structure", managerName))
	}

	return parseFieldsFromMap(fieldsMap, nil)
}

// parseModules 解析 modules.yaml
func (p *Parser) parseModules() error {
	filePath := p.ResolvePath("mme", "modules.yaml")
	if !FileExists(filePath) {
		return nil
	}

	yamlData, err := p.ReadYAMLFile(filePath)
	if err != nil {
		return err
	}

	if moduleList, ok := yamlData["Modules"].([]any); ok {
		for _, moduleItem := range moduleList {
			if moduleMap, ok := moduleItem.(map[string]any); ok {
				for moduleName, moduleData := range moduleMap {
					module := Module{Name: moduleName}

					// moduleData 是列表，包含 Mechanism 引用
					// 格式: [{ "HeroMechanism Base": 1, Settings: {...} }, { "LevelUpMechanism LevelUp": 2, Settings: {...} }]
					if mechanismsList, ok := moduleData.([]any); ok {
						for _, mechItem := range mechanismsList {
							if mechMap, ok := mechItem.(map[string]any); ok {
								var fieldDef string
								var fieldNumber int32
								var settings map[string]any

								// mechMap 的 key 是字段定义字符串，格式: "HeroMechanism Base"
								// value 是编号或 Settings
								for key, value := range mechMap {
									if key == "Settings" {
										// 处理 Settings
										if settingsMap, ok := value.(map[string]any); ok {
											settings = settingsMap
										}
									} else {
										// key 是字段定义字符串，格式: "HeroMechanism Base"
										// value 是编号（int）或 nil

										// 先尝试从 value 获取编号
										if value != nil {
											if num, ok := value.(int); ok {
												fieldNumber = int32(num)
											} else if num, ok := value.(int32); ok {
												fieldNumber = num
											} else if num, ok := value.(int64); ok {
												fieldNumber = int32(num)
											}
										}

										// 构建字段定义字符串，格式: "HeroMechanism FieldName: Number"
										if fieldNumber != 0 {
											fieldDef = fmt.Sprintf("%s: %d", key, fieldNumber)
										} else {
											fieldDef = key
										}

										// 如果从 value 没有获取到编号，尝试从 key 中解析（key 可能包含编号）
										if fieldNumber == 0 {
											// 解析字段定义: "HeroMechanism Base: 1"
											parts2 := strings.Fields(fieldDef)
											if len(parts2) >= 3 {
												for i, part := range parts2 {
													if strings.HasPrefix(part, ":") && i+1 < len(parts2) {
														var num int32
														if _, err := fmt.Sscanf(parts2[i+1], "%d", &num); err == nil {
															fieldNumber = num
															fieldDef = fmt.Sprintf("%s: %d", key, num)
														}
														break
													}
												}
											}
										}
									}
								}

								// 解析字段定义，创建 types.Field
								if fieldDef != "" {
									// 解析字段定义：格式 "HeroMechanism FieldName: Number"
									parts := strings.Fields(fieldDef)
									if len(parts) >= 2 {
										mechanismName := parts[0]
										fieldName := strings.TrimSuffix(parts[1], ":")

										// 如果编号为 0，尝试从字段定义中解析
										if fieldNumber == 0 && len(parts) >= 3 {
											if _, err := fmt.Sscanf(parts[2], "%d", &fieldNumber); err != nil {
												fieldNumber = 0
											}
										}

										// 创建 types.Field
										field := &types.Field{
											Name:   fieldName,
											Number: fieldNumber,
											Type: types.FieldType{
												Kind:     types.FieldKindMMEObject, // Mechanism 是 MME Object 类型
												Label:    types.FieldLabelOptional,
												TypeName: mechanismName,
											},
											Options: types.FieldOption{},
										}

										// 如果有 Settings，将其存储到 Metadata 的 CustomOptions 中
										if settings != nil {
											if field.Metadata == nil {
												field.Metadata = &types.FieldMetadata{
													CustomOptions: make(map[string]any),
												}
											}
											if field.Metadata.CustomOptions == nil {
												field.Metadata.CustomOptions = make(map[string]any)
											}
											field.Metadata.CustomOptions["Settings"] = settings
										}

										if fieldNumber > 0 {
											module.Mechanisms = append(module.Mechanisms, field)
										}
									}
								}
							}
						}
					}

					p.ctx.AddModule(module)
				}
			}
		}
	}

	return nil
}

// parseMechanisms 解析 mechanisms.yaml
func (p *Parser) parseMechanisms() error {
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

	return nil
}

func (p *Parser) parseMechanismItem(mechItem map[string]any) *Mechanism {
	mechanism := &Mechanism{}
	for mechKey, mechData := range mechItem {
		switch mechKey {
		case MechanismKeyWordSettings:
			settings := mechData.(map[string]any)
			maps.Copy(mechanism.Settings, settings)
		case MechanismKeyWordRequests:
			requests := mechData.([]any)
			for _, reqItem := range requests {
				req, err := p.ParseRequestOrNotify(reqItem, true)
				if err == nil && req.Name != "" {
					mechanism.Requests = append(mechanism.Requests, req)
				}
			}
		case MechanismKeyWordNotifies:
			notifies := mechData.([]any)
			for _, notifyItem := range notifies {
				notify, err := p.ParseNotifyOnly(notifyItem)
				if err == nil && notify.Name != "" {
					mechanism.Notifies = append(mechanism.Notifies, notify)
				}
			}
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
		// 构建字段定义
		fieldDef := buildFieldDefinition(fieldKey, fieldValue)
		if fieldDef == "" {
			continue
		}

		// 解析字段
		var field *types.Field
		if config != nil && config.ParseFieldFunc != nil {
			field = config.ParseFieldFunc(fieldDef)
		} else {
			field = parseAndConvertFieldDefinition(fieldDef)
		}

		if field != nil {
			fields = append(fields, field)
		}
	}

	// 按字段编号排序，保证字段顺序
	sortFieldsByNumber(fields)

	return fields
}

// buildFieldDefinition 构建字段定义字符串（通用函数）
// fieldKey: 字段键，可能是完整的字段定义字符串，也可能是字段名
// fieldValue: 字段值，可能是编号、字符串或其他类型
// 返回: 完整的字段定义字符串
func buildFieldDefinition(fieldKey string, fieldValue any) string {
	fmt.Println("buildFieldDefinition fieldKey", fieldKey, "fieldValue", fieldValue)
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
	field, err := ParseFieldDefinition(fieldDef)
	if err != nil {
		return nil
	}
	return field
}

// sortFieldsByNumber 按字段编号排序字段列表（通用函数）
func sortFieldsByNumber(fields []*types.Field) {
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Number < fields[j].Number
	})
}
