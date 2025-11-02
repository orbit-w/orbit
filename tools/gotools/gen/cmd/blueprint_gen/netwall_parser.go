package blueprint_gen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	
	"gopkg.in/yaml.v3"
)

// parseNetWallFiles 解析 NetWall YAML 文件
func (p *Parser) parseNetWallFiles() error {
	netwallDir := filepath.Join(p.blueprintDir, "netwall")
	
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
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	
	var yamlData map[string]interface{}
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		return err
	}
	
	netwallFile := NetWallFile{}
	
	// 解析 NetWall
	if netwallData, ok := yamlData["NetWall"].(map[string]interface{}); ok {
		netwall := NetWall{}
		
		if name, ok := netwallData["Name"].(string); ok {
			netwall.Name = name
		}
		
		// 解析 Requests
		if requests, ok := netwallData["Requests"].([]interface{}); ok {
			for _, reqItem := range requests {
				req := parseNetWallRequest(reqItem)
				if req.Name != "" {
					netwall.Requests = append(netwall.Requests, req)
				}
			}
		}
		
		// 解析 Notifies
		if notifies, ok := netwallData["Notifies"].([]interface{}); ok {
			for _, notifyItem := range notifies {
				notify := parseNetWallNotify(notifyItem)
				if notify.Name != "" {
					netwall.Notifies = append(netwall.Notifies, notify)
				}
			}
		}
		
		netwallFile.NetWall = netwall
	}
	
	// 解析 DataStructs
	if dataStructs, ok := yamlData["DataStructs"].([]interface{}); ok {
		for _, dsItem := range dataStructs {
			if dsMap, ok := dsItem.(map[string]interface{}); ok {
				for dsName, dsData := range dsMap {
					dataStruct := DataStruct{Name: dsName}
					
					if dsDataMap, ok := dsData.(map[string]interface{}); ok {
						// 解析字段
						for fieldName, fieldDef := range dsDataMap {
							if fieldStr, ok := fieldDef.(string); ok {
								field, err := ParseFieldDefinition(fieldName + " " + fieldStr)
								if err == nil {
									dataStruct.Fields = append(dataStruct.Fields, field)
								}
							}
						}
					} else if dsData == nil {
						// 空结构体
					}
					
					netwallFile.DataStructs = append(netwallFile.DataStructs, dataStruct)
				}
			} else if dsName, ok := dsItem.(string); ok {
				// 简单结构体，无字段
				netwallFile.DataStructs = append(netwallFile.DataStructs, DataStruct{
					Name: dsName,
				})
			}
		}
	}
	
	p.data.NetWalls = append(p.data.NetWalls, netwallFile)
	return nil
}

// parseNetWallRequest 解析 NetWall Request
func parseNetWallRequest(reqItem interface{}) NetWallMessage {
	req := NetWallMessage{}
	
	if reqMap, ok := reqItem.(map[string]interface{}); ok {
		for name, data := range reqMap {
			req.Name = name
			
			if dataMap, ok := data.(map[string]interface{}); ok {
				fields := make([]Field, 0)
				for fieldName, fieldDef := range dataMap {
					if fieldName == "Rsp" {
						// 解析响应
						if rspData, ok := fieldDef.(map[string]interface{}); ok {
							rsp := Response{}
							for rspFieldName, rspFieldDef := range rspData {
								if fieldStr, ok := rspFieldDef.(string); ok {
									field, err := ParseFieldDefinition(rspFieldName + " " + fieldStr)
									if err == nil {
										rsp.Fields = append(rsp.Fields, field)
									}
								}
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
	} else if reqName, ok := reqItem.(string); ok {
		// 简单请求，无参数
		req.Name = reqName
	}
	
	return req
}

// parseNetWallNotify 解析 NetWall Notify
func parseNetWallNotify(notifyItem interface{}) NetWallMessage {
	notify := NetWallMessage{}
	
	if notifyMap, ok := notifyItem.(map[string]interface{}); ok {
		for name, data := range notifyMap {
			notify.Name = name
			
			if dataMap, ok := data.(map[string]interface{}); ok {
				fields := make([]Field, 0)
				for fieldName, fieldDef := range dataMap {
					if fieldStr, ok := fieldDef.(string); ok {
						field, err := ParseFieldDefinition(fieldName + " " + fieldStr)
						if err == nil {
							fields = append(fields, field)
						}
					}
				}
				notify.Fields = fields
			}
		}
	} else if notifyName, ok := notifyItem.(string); ok {
		// 简单通知，无参数
		notify.Name = notifyName
	}
	
	return notify
}

// HasMMELocation 检查是否有 MMELocation 字段（编号 1000）
func (req *NetWallMessage) HasMMELocation() bool {
	for _, field := range req.Fields {
		if field.Number == 1000 {
			return true
		}
	}
	return false
}

// EnsureMMELocation 确保 MMELocation 字段存在（仅对 MME NetWall）
func (req *NetWallMessage) EnsureMMELocation() {
	if req.HasMMELocation() {
		return
	}
	
	// 添加 MMELocation 字段
	locField := Field{
		Name:   "Loc",
		Number: 1000,
		Type: FieldType{
			ValueType: "Core.MMELocation",
		},
	}
	
	req.Fields = append(req.Fields, locField)
}

