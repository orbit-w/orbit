package blueprint_gen

import (
	"fmt"
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
	if err := p.parseManagers(); err != nil {
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

	if entityList, ok := yamlData["Entity"].([]interface{}); ok {
		for _, entityItem := range entityList {
			if entityMap, ok := entityItem.(map[string]interface{}); ok {
				for entityName, entityData := range entityMap {
					entity := Entity{Name: entityName}

					// entityData 可能是 map，包含字段定义
					// 格式: { "HeroManager HeroManager": "1 [blueprint:\"access=all\"]" }
					if fieldsMap, ok := entityData.(map[string]interface{}); ok {
						for fieldKey, fieldValue := range fieldsMap {
							// fieldKey 格式: "HeroManager HeroManager" 或 "HeroManager HeroManager: 1"
							// fieldValue 格式: "1 [blueprint:\"access=all\"]" 或数字

							// fieldKey 格式: "HeroManager HeroManager"
							// fieldValue 格式: "1 [blueprint:\"access=all\"]" 或数字

							// 先提取类型名和字段名
							keyParts := strings.Fields(fieldKey)
							if len(keyParts) >= 2 {
								field := EntityField{}
								field.ManagerName = keyParts[0] // 类型名
								field.Name = keyParts[1]        // 字段名

								// 从 fieldValue 提取编号和选项
								if fieldValue != nil {
									if valueStr, ok := fieldValue.(string); ok {
										// valueStr 格式: "1 [blueprint:\"access=all\"]"
										valueParts := strings.Fields(valueStr)
										if len(valueParts) > 0 {
											var num int32
											numStr := valueParts[0]
											// 移除可能的选项部分
											if idx := strings.Index(numStr, "["); idx != -1 {
												field.Options = ParseFieldOptions(valueStr[idx:])
												numStr = numStr[:idx]
											} else if len(valueParts) > 1 {
												// 选项在下一个元素
												field.Options = ParseFieldOptions(valueParts[1])
											}
											if _, err := fmt.Sscanf(numStr, "%d", &num); err == nil {
												field.Number = num
											}
										}
									} else if num, ok := fieldValue.(int); ok {
										field.Number = int32(num)
									}
								}

								if field.Number > 0 {
									entity.Fields = append(entity.Fields, field)
								}
							}
						}
					}

					p.data.Entities = append(p.data.Entities, entity)
				}
			}
		}
	}

	return nil
}

// parseManagers 解析 manager.yaml
func (p *Parser) parseManagers() error {
	filePath := p.ResolvePath("mme", "manager.yaml")
	if !FileExists(filePath) {
		return nil
	}

	yamlData, err := p.ReadYAMLFile(filePath)
	if err != nil {
		return err
	}

	if managerList, ok := yamlData["Managers"].([]interface{}); ok {
		for _, managerItem := range managerList {
			if managerMap, ok := managerItem.(map[string]interface{}); ok {
				// 首先找出所有 Manager 名称（不包含 < 或 > 或 : 的 key）
				managersInMap := make(map[string]*Manager)
				for key := range managerMap {
					if !strings.Contains(key, "<") && !strings.Contains(key, ">") && !strings.Contains(key, ":") {
						managersInMap[key] = &Manager{Name: key}
					}
				}

				// 然后处理所有字段（包含 < 或 > 的 key）
				for fieldKey, fieldValue := range managerMap {
					// 跳过 Manager 名称
					if managersInMap[fieldKey] != nil {
						continue
					}

					// fieldKey 是字段定义字符串，格式: "xmap<int64, HeroModule> HeroMap: 1"
					// 或者: "xmap<int64, HeroModule> HeroMap"
					// fieldValue 是字段编号（可能是 int 或 nil）

					// 构建完整的字段定义
					fieldDef := fieldKey
					if fieldValue != nil {
						// 如果 fieldValue 是数字，添加到定义中
						if num, ok := fieldValue.(int); ok {
							fieldDef = fmt.Sprintf("%s: %d", fieldKey, num)
						} else if num, ok := fieldValue.(int32); ok {
							fieldDef = fmt.Sprintf("%s: %d", fieldKey, num)
						} else if num, ok := fieldValue.(int64); ok {
							fieldDef = fmt.Sprintf("%s: %d", fieldKey, num)
						}
					}

					// 解析字段定义
					field, err := ParseFieldDefinition(fieldDef)
					if err == nil {
						// 转换为 types.Field
						typesField, err := ConvertFieldToTypesField(field)
						if err == nil {
							// 找到对应的 Manager（如果有多个 Manager，可能需要更复杂的逻辑）
							// 目前假设只有一个 Manager
							for _, manager := range managersInMap {
								manager.Fields = append(manager.Fields, typesField)
								break // 只添加到第一个 Manager
							}
						}
					}
				}

				// 添加所有有字段的 Manager
				for _, manager := range managersInMap {
					if len(manager.Fields) > 0 {
						p.data.Managers = append(p.data.Managers, *manager)
					}
				}
			}
		}
	}

	return nil
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

	if moduleList, ok := yamlData["Modules"].([]interface{}); ok {
		for _, moduleItem := range moduleList {
			if moduleMap, ok := moduleItem.(map[string]interface{}); ok {
				for moduleName, moduleData := range moduleMap {
					module := Module{Name: moduleName}

					// moduleData 是列表，包含 Mechanism 引用
					// 格式: [{ "HeroMechanism Base": 1, Settings: {...} }, { "LevelUpMechanism LevelUp": 2, Settings: {...} }]
					if mechanismsList, ok := moduleData.([]interface{}); ok {
						for _, mechItem := range mechanismsList {
							if mechMap, ok := mechItem.(map[string]interface{}); ok {
								var fieldDef string
								var fieldNumber int32
								var settings map[string]interface{}

								// mechMap 的 key 是字段定义字符串，格式: "HeroMechanism Base"
								// value 是编号或 Settings
								for key, value := range mechMap {
									if key == "Settings" {
										// 处理 Settings
										if settingsMap, ok := value.(map[string]interface{}); ok {
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
													CustomOptions: make(map[string]interface{}),
												}
											}
											if field.Metadata.CustomOptions == nil {
												field.Metadata.CustomOptions = make(map[string]interface{})
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

					p.data.Modules = append(p.data.Modules, module)
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

	if mechanismList, ok := yamlData["Mechanisms"].([]interface{}); ok {
		for _, mechItem := range mechanismList {
			if mechMap, ok := mechItem.(map[string]interface{}); ok {
				for mechName, mechData := range mechMap {
					// 跳过 Settings, Requests, Notifies 这些不是 Mechanism 的 key
					if mechName == "Settings" || mechName == "Requests" || mechName == "Notifies" {
						continue
					}

					mechanism := Mechanism{Name: mechName}

					if mechDataMap, ok := mechData.(map[string]interface{}); ok {
						// 解析数据字段
						fields := make([]*types.Field, 0)

						// 解析 Settings
						if settings, ok := mechDataMap["Settings"].(map[string]interface{}); ok {
							mechanism.Settings = settings
						}

						// 解析 Requests
						if requests, ok := mechDataMap["Requests"].([]interface{}); ok {
							for _, reqItem := range requests {
								req, err := p.ParseRequestOrNotify(reqItem, true)
								if err == nil && req.Name != "" {
									mechanism.Requests = append(mechanism.Requests, req)
								}
							}
						}

						// 解析 Notifies
						if notifies, ok := mechDataMap["Notifies"].([]interface{}); ok {
							for _, notifyItem := range notifies {
								notify, err := p.ParseNotifyOnly(notifyItem)
								if err == nil && notify.Name != "" {
									mechanism.Notifies = append(mechanism.Notifies, notify)
								}
							}
						}

						// 解析数据字段（其他所有字段）
						// key 是字段定义字符串，格式: "int32 CurLevel: 1 [blueprint:\"access=all\"]"
						// value 可能是 nil、数字或其他元数据
						for key, value := range mechDataMap {
							if key != "Settings" && key != "Requests" && key != "Notifies" {
								// 这是一个数据字段定义
								// key 格式: "int32 CurLevel: 1 [blueprint:\"access=all\"]" 或 "int32 CurLevel"
								// value 可能是 nil 或数字

								// 构建完整的字段定义
								fieldDef := key
								if value != nil {
									// 如果 value 是数字，添加到定义中
									if num, ok := value.(int); ok {
										fieldDef = fmt.Sprintf("%s: %d", key, num)
									} else if num, ok := value.(int32); ok {
										fieldDef = fmt.Sprintf("%s: %d", key, num)
									} else if num, ok := value.(int64); ok {
										fieldDef = fmt.Sprintf("%s: %d", key, num)
									} else if valueStr, ok := value.(string); ok {
										// value 是字符串，格式: "1 [blueprint:\"access=all\"]"
										fieldDef = key + ": " + valueStr
									}
								}

								// 解析字段定义并转换为 *types.Field
								field, err := ParseFieldDefinition(fieldDef)
								if err == nil {
									typesField, err := ConvertFieldToTypesField(field)
									if err == nil {
										fields = append(fields, typesField)
									}
								}
							}
						}

						mechanism.Fields = fields
					} else {
						// mechData 不是 map，可能是其他类型（应该不会发生）
						// 但是需要处理这种情况，避免 panic
					}

					p.data.Mechanisms = append(p.data.Mechanisms, mechanism)
				}
			}
		}
	}

	return nil
}
