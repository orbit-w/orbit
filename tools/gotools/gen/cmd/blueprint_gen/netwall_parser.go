package blueprint_gen

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	types "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types"
)

const (
	ResponseSuffix  = "_Rsp"
	RequestPrefix   = "Request_"
	CorePackageName = "Core"
	MMEPackageName  = "MME"

	CommonNetWallFileName = "common.yaml"
	CommonPackageName     = "Common"
)

// parseNetWallFiles 解析 NetWall YAML 文件
func (p *Parser) parseNetWallFiles() error {
	netwallDir := p.ResolvePath("netwall")

	// 检查目录是否存在
	if _, err := os.Stat(netwallDir); os.IsNotExist(err) {
		return nil // NetWall 目录可选
	}

	// 读取所有 YAML 文件
	files, err := os.ReadDir(netwallDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if !strings.HasSuffix(file.Name(), ".yaml") {
			continue
		}

		filePath := filepath.Join(netwallDir, file.Name())
		if err := p.parseNetWallFile(filePath); err != nil {
			return fmt.Errorf("failed to parse %s: %w", file.Name(), err)
		}
	}

	return nil
}

// parseNetWallFile 解析单个 NetWall YAML 文件
func (p *Parser) parseNetWallFile(filePath string) error {
	yamlData, err := p.ReadYAMLFile(filePath)
	if err != nil {
		return err
	}

	var netwallFile *NetWallFile

	// 解析 NetWall
	if netwallData, ok := yamlData["NetWall"].(map[string]any); ok {
		netwallFile = NewNetWallFile()
		if name, ok := netwallData["Name"].(string); ok {
			netwallFile.PackageName = name
		}

		// 解析 Requests
		if requests, ok := netwallData["Requests"].([]any); ok {
			p.parseRequestItems(netwallFile, requests)
		}

		// 解析 Notifies
		if notifies, ok := netwallData["Notifies"].([]any); ok {
			p.parseNotifyItems(netwallFile, notifies)
		}
	}

	if netwallFile == nil {
		if strings.HasSuffix(filePath, CommonNetWallFileName) {
			netwallFile = NewNetWallFile()
			netwallFile.PackageName = CommonPackageName
		} else {
			panic(fmt.Sprintf("NetWallFile not found for file %s", filePath))
		}
	}

	// 解析 DataStructs
	if dataStructs, ok := yamlData["DataStructs"].([]any); ok {
		p.parseDataStructItems(netwallFile, dataStructs)
	}

	// 解析 Enums
	if enums, ok := yamlData["Enums"].([]any); ok {
		p.parseEnumItems(netwallFile, enums)
	}

	p.ctx.AddNetWallFile(netwallFile)
	return nil
}

func (p *Parser) parseRequestItems(ctx *NetWallFile, items []any) ([]*NetMessage, error) {
	requests := make([]*NetMessage, 0)
	for _, item := range items {
		req := p.parseRequestItem(ctx.PackageName, item)
		ctx.AddRequest(req)
	}
	return requests, nil
}

// 解析yaml的请求项，生成请求消息和响应消息
func (p *Parser) parseRequestItem(packageName string, item any) (req *NetMessage) {
	itemMap := item.(map[string]any)
	fmt.Println("itemMap", itemMap)
	for key, data := range itemMap {
		switch key {
		case "Rsp":
		default:
			req = NewNetMessage(NetWallMessageTypeRequest)
			req.PackageName = packageName
			req.Name = key
			if data == nil {
				continue
			}
			fieldsMap := data.(map[string]any)
			fields, err := p.ParseMessageFields(fieldsMap)
			if err != nil {
				panic(fmt.Sprintf("Failed to parse fields for request %s: %v", key, err))
			}
			req.Fields = fields
		}
	}

	if req == nil {
		panic(fmt.Sprintf("Request not found for item %v", item))
	}

	rspData, ok := itemMap["Rsp"]
	if ok {
		rspDataMap := rspData.(map[string]any)
		rsp := NewNetMessage(NetWallMessageTypeResponse)
		rsp.PackageName = packageName
		rsp.Name = RequestPrefix + req.Name + ResponseSuffix
		fields, err := p.ParseMessageFields(rspDataMap)
		if err != nil {
			panic(fmt.Sprintf("Failed to parse fields for response %s: %v", rsp.Name, err))
		}
		rsp.Fields = fields
		req.Rsp = rsp
	}

	return req
}

func (p *Parser) parseNotifyItems(ctx *NetWallFile, items []any) ([]*NetMessage, error) {
	notifies := make([]*NetMessage, 0)
	for _, item := range items {
		notify := p.parseNotifyItem(ctx.PackageName, item)
		if notify == nil {
			panic(fmt.Sprintf("Notify not found for item %v", item))
		}
		ctx.AddNotify(notify)
	}
	return notifies, nil
}

func (p *Parser) parseNotifyItem(packageName string, item any) (notify *NetMessage) {
	itemMap := item.(map[string]any)
	for key, data := range itemMap {
		notify = NewNetMessage(NetWallMessageTypeNotify)
		notify.PackageName = packageName
		notify.Name = key
		if data == nil {
			continue
		}
		fieldsMap := data.(map[string]any)
		fields, err := p.ParseMessageFields(fieldsMap)
		if err != nil {
			panic(fmt.Sprintf("Failed to parse fields for notify %s: %v", key, err))
		}
		notify.Fields = fields
	}
	return notify
}

func (p *Parser) parseDataStructItems(ctx *NetWallFile, items []any) ([]*NetMessage, error) {
	dataStructs := make([]*NetMessage, 0)
	for _, item := range items {
		dataStruct := p.parseDataStructItem(ctx.PackageName, item)
		ctx.AddDataStruct(dataStruct)
	}
	return dataStructs, nil
}

func (p *Parser) parseDataStructItem(packageName string, item any) (dataStruct *NetMessage) {
	fmt.Println("parseDataStructItem", item)
	itemMap := item.(map[string]any)
	for key, data := range itemMap {
		dataStruct = NewNetMessage(NetWallMessageTypeDataStruct)
		dataStruct.PackageName = packageName
		dataStruct.Name = key
		if data == nil {
			// 允许没有字段的 DataStruct（如 OK）
			dataStruct.Fields = make([]*types.Field, 0)
			return dataStruct
		}
		fieldsMap := data.(map[string]any)
		fields, err := p.ParseMessageFields(fieldsMap)
		if err != nil {
			panic(fmt.Sprintf("Failed to parse fields for data struct %s: %v", key, err))
		}
		dataStruct.Fields = fields
		return dataStruct
	}
	// 如果 itemMap 为空，说明没有找到 DataStruct
	panic(fmt.Sprintf("Data struct not found for item %v", item))
}

// parseEnumItems 解析枚举列表
func (p *Parser) parseEnumItems(ctx *NetWallFile, items []any) ([]*Enum, error) {
	enums := make([]*Enum, 0)
	for _, item := range items {
		enum := p.parseEnumItem(item)
		if enum == nil {
			panic(fmt.Sprintf("Enum not found for item %v", item))
		}
		ctx.AddEnum(enum)
	}
	return enums, nil
}

// parseEnumItem 解析单个枚举项
// 格式:
//   - ServiceZoneType:
//     ServiceZoneTypePlay: 0 [Content:"逻辑服管理区域类型"]
//     ServiceZoneTypeUnion: 1 [Content:"联盟服管理区域类型"]
func (p *Parser) parseEnumItem(item any) *Enum {
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
func (p *Parser) parseEnumValue(valueName string, valueDef any) *EnumValue {
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
func (p *Parser) parseEnumValueString(valueStr string) (int32, map[string]string) {
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
func (p *Parser) parseBracketKeyValuePairs(bracketContent string) map[string]string {
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
