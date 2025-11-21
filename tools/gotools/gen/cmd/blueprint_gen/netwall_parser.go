package blueprint_gen

import (
	"fmt"
	"os"
	"path/filepath"
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
func (p *YamlParser) parseNetWallFiles() error {
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
func (p *YamlParser) parseNetWallFile(filePath string) error {
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

func (p *YamlParser) parseRequestItems(ctx *NetWallFile, items []any) ([]*NetMessage, error) {
	requests := make([]*NetMessage, 0)
	for _, item := range items {
		req := p.parseRequestItem(ctx.PackageName, item)
		ctx.AddRequest(req)
	}
	return requests, nil
}

// 解析yaml的请求项，生成请求消息和响应消息
func (p *YamlParser) parseRequestItem(packageName string, item any) (req *NetMessage) {
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

func (p *YamlParser) parseNotifyItems(ctx *NetWallFile, items []any) ([]*NetMessage, error) {
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

func (p *YamlParser) parseNotifyItem(packageName string, item any) (notify *NetMessage) {
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

func (p *YamlParser) parseDataStructItems(ctx *NetWallFile, items []any) ([]*NetMessage, error) {
	dataStructs := make([]*NetMessage, 0)
	for _, item := range items {
		dataStruct := p.parseDataStructItem(ctx.PackageName, item)
		ctx.AddDataStruct(dataStruct)
	}
	return dataStructs, nil
}

func (p *YamlParser) parseDataStructItem(packageName string, item any) (dataStruct *NetMessage) {
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
func (p *YamlParser) parseEnumItems(ctx *NetWallFile, items []any) ([]*Enum, error) {
	enums := make([]*Enum, 0)
	enumParser := NewEnumParser()
	for _, item := range items {
		enum := enumParser.ParseEnumItem(item)
		if enum == nil {
			panic(fmt.Sprintf("Enum not found for item %v", item))
		}
		ctx.AddEnum(enum)
	}
	return enums, nil
}
