package blueprint_gen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
				req, err := p.ParseRequestOrNotify(reqItem, true)
				if err == nil && req.Name != "" {
					netwall.Requests = append(netwall.Requests, NetWallMessage{
						Name:    req.Name,
						Fields: req.Fields,
						Rsp:     req.Rsp,
						Comment: req.Comment,
					})
				}
			}
		}
		
		// 解析 Notifies
		if notifies, ok := netwallData["Notifies"].([]interface{}); ok {
			for _, notifyItem := range notifies {
				notify, err := p.ParseNotifyOnly(notifyItem)
				if err == nil && notify.Name != "" {
					netwall.Notifies = append(netwall.Notifies, NetWallMessage{
						Name:    notify.Name,
						Fields:  notify.Fields,
						Comment: notify.Comment,
					})
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
						fields, err := p.ParseMessageFields(dsDataMap)
						if err == nil {
							dataStruct.Fields = fields
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


