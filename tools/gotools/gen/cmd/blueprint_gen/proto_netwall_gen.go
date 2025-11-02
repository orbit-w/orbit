package blueprint_gen

import (
	"fmt"
	"strings"
)

// generateNetWallProto 生成 NetWall Proto 文件
func (g *ProtoGenerator) generateNetWallProto(outputDir string) error {
	// 收集所有 NetWall 的 Request 和 Notify
	allRequests := make(map[string][]NetWallMessage)
	allNotifies := make(map[string][]NetWallMessage)
	allDataStructs := make(map[string]DataStruct) // name -> DataStruct
	
	// 收集 DataStructs（去重）
	for _, netwallFile := range g.data.NetWalls {
		for _, ds := range netwallFile.DataStructs {
			allDataStructs[ds.Name] = ds
		}
		
		netwall := netwallFile.NetWall
		
		// 收集 Requests
		for _, req := range netwall.Requests {
			allRequests[netwall.Name] = append(allRequests[netwall.Name], req)
		}
		
		// 收集 Notifies
		for _, notify := range netwall.Notifies {
			allNotifies[netwall.Name] = append(allNotifies[netwall.Name], notify)
		}
	}
	
	// 生成 structs.proto
	if err := g.generateStructsProto(outputDir, allDataStructs); err != nil {
		return fmt.Errorf("failed to generate structs.proto: %w", err)
	}
	
	// 生成 request.proto
	if err := g.generateRequestProto(outputDir, allRequests); err != nil {
		return fmt.Errorf("failed to generate request.proto: %w", err)
	}
	
	// 生成 notify.proto
	if err := g.generateNotifyProto(outputDir, allNotifies); err != nil {
		return fmt.Errorf("failed to generate notify.proto: %w", err)
	}
	
	return nil
}

// generateStructsProto 生成 structs.proto
func (g *ProtoGenerator) generateStructsProto(outputDir string, dataStructs map[string]DataStruct) error {
	sb := strings.Builder{}
	
	// 文件头部
	imports := []string{}
	if g.data.HeadFile != nil && len(g.data.HeadFile.CommonDataStructs) > 0 {
		imports = append(imports, "protocol/common.proto")
	}
	sb.WriteString(generateProtoHeader("Core", imports))
	
	// 生成所有数据结构
	for _, ds := range dataStructs {
		sb.WriteString(fmt.Sprintf("// %s\n", ds.Name))
		sb.WriteString(fmt.Sprintf("message %s {\n", ds.Name))
		
		for _, field := range ds.Fields {
			sb.WriteString(generateFieldProto(field, field.Number))
		}
		
		sb.WriteString("}\n\n")
	}
	
	content := sb.String()
	return WriteFile(outputDir+"/structs.proto", content)
}

// generateRequestProto 生成 request.proto
func (g *ProtoGenerator) generateRequestProto(outputDir string, allRequests map[string][]NetWallMessage) error {
	sb := strings.Builder{}
	
	// 文件头部
	imports := []string{"protocol/structs.proto"}
	if g.data.HeadFile != nil && len(g.data.HeadFile.CommonDataStructs) > 0 {
		imports = append(imports, "protocol/common.proto")
	}
	sb.WriteString(generateProtoHeader("Core", imports))
	
	sb.WriteString("message Request {\n")
	
	// 生成所有 NetWall 的 Request
	for _, requests := range allRequests {
		for _, req := range requests {
			sb.WriteString(fmt.Sprintf("  message %s {\n", req.Name))
			
			// 生成字段
			for _, field := range req.Fields {
				protoType := ToProtoType(field.Type)
				sb.WriteString(fmt.Sprintf("    %s %s = %d;\n", protoType, field.Name, field.Number))
			}
			
			// 生成 Rsp
			if req.Rsp != nil {
				sb.WriteString("    message Rsp {\n")
				for _, field := range req.Rsp.Fields {
					sb.WriteString(generateFieldProto(field, field.Number))
				}
				sb.WriteString("    }\n")
			}
			
			sb.WriteString("  }\n\n")
		}
	}
	
	sb.WriteString("}\n")
	
	content := sb.String()
	return WriteFile(outputDir+"/request.proto", content)
}

// generateNotifyProto 生成 notify.proto
func (g *ProtoGenerator) generateNotifyProto(outputDir string, allNotifies map[string][]NetWallMessage) error {
	sb := strings.Builder{}
	
	// 文件头部
	imports := []string{"protocol/structs.proto"}
	if g.data.HeadFile != nil && len(g.data.HeadFile.CommonDataStructs) > 0 {
		imports = append(imports, "protocol/common.proto")
	}
	sb.WriteString(generateProtoHeader("Core", imports))
	
	sb.WriteString("message Notify {\n")
	
	// 生成所有 NetWall 的 Notify
	for _, notifies := range allNotifies {
		for _, notify := range notifies {
			sb.WriteString(fmt.Sprintf("  message %s {\n", notify.Name))
			
			// 生成字段
			for _, field := range notify.Fields {
				protoType := ToProtoType(field.Type)
				sb.WriteString(fmt.Sprintf("    %s %s = %d;\n", protoType, field.Name, field.Number))
			}
			
			sb.WriteString("  }\n\n")
		}
	}
	
	sb.WriteString("}\n")
	
	content := sb.String()
	return WriteFile(outputDir+"/notify.proto", content)
}

