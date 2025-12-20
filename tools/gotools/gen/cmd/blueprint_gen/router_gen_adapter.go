package blueprint_gen

import (
	"fmt"
	"path/filepath"
	"strings"

	"gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/pb_gen/net_message"
	blueprint_types "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types"
	router_gen "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/router_gen"
)

// generateRouterCode 从 BlueprintContext 生成 Router 代码
func generateRouterCode(ctx *BlueprintContext, controllerDir, outputFile, controllerPath string) error {
	// 将 BlueprintContext 转换为 router_gen 需要的格式
	requestInfos := convertNetWallsToRequestInfos(ctx.NetWalls)

	fmt.Printf("DEBUG: Found %d request infos for router generation\n", len(requestInfos))

	if len(requestInfos) == 0 {
		// 如果没有请求，跳过生成
		return nil
	}

	// 解析 Controller 文件
	controllerInfo, err := router_gen.DiscoverAndParseController(controllerDir)
	if err != nil {
		return fmt.Errorf("failed to parse controller: %w", err)
	}

	// 确定 Controller 文件路径
	if controllerPath == "" {
		controllerPath = filepath.Join(controllerDir, "controller.go")
	}

	// 创建生成上下文
	routerCtx := &router_gen.RouterGenContext{
		Requests:       requestInfos,
		Controller:     controllerInfo,
		OutputPath:     outputFile,
		ControllerPath: controllerPath,
	}

	// 生成代码
	if err := router_gen.GenerateRouterCode(routerCtx); err != nil {
		return fmt.Errorf("failed to generate router code: %w", err)
	}

	// 生成 Controller Handle 方法
	if err := router_gen.GenerateControllerMethods(routerCtx); err != nil {
		return fmt.Errorf("failed to generate controller methods: %w", err)
	}

	return nil
}

// convertNetWallsToRequestInfos 将 NetWallFile 列表转换为 RequestInfo 列表
func convertNetWallsToRequestInfos(netWalls []*NetWallFile) []*router_gen.RequestInfo {
	var requestInfos []*router_gen.RequestInfo

	for _, netwall := range netWalls {
		for _, req := range netwall.Requests {
			requestInfo := convertNetMessageToRequestInfo(req, netwall.PackageName)
			fmt.Println("DEBUG: requestInfo", netwall.PackageName, requestInfo)
			if requestInfo != nil {
				requestInfos = append(requestInfos, requestInfo)
			}
		}
	}

	return requestInfos
}

// convertNetMessageToRequestInfo 将 NetMessage 转换为 RequestInfo
func convertNetMessageToRequestInfo(req *net_message.NetMessage, packageName string) *router_gen.RequestInfo {
	// 包名处理：首字母大写用于类型引用，小写用于导入
	packageNameLower := strings.ToLower(packageName)
	packageNameUpper := strings.ToUpper(packageName[:1]) + packageName[1:]

	// 解析 EntityRef 字段
	entityRefs := extractEntityRefsFromFields(req.Fields)

	return &router_gen.RequestInfo{
		RequestName:      req.GetName(),
		PackageName:      packageNameLower,
		PackageNameUpper: packageNameUpper,
		EntityRefs:       entityRefs,
		HasRsp:           req.HasResponse(),
	}
}

// extractEntityRefsFromFields 从字段列表中提取 EntityRef 字段
func extractEntityRefsFromFields(fields []*blueprint_types.Field) []*router_gen.EntityRefInfo {
	var entityRefs []*router_gen.EntityRefInfo

	for _, field := range fields {
		// 检查是否是 EntityRef 类型
		if !isEntityRefField(field) {
			continue
		}

		// 提取实体名称
		entityName := extractEntityNameFromFieldName(field.Name)
		if entityName == "" {
			continue
		}

		// 生成参数名（小驼峰）
		paramName := toCamelCase(entityName) + "Entity"

		// 生成包装器类型
		wrapperType := fmt.Sprintf("*mme.%sEntityWrapper", entityName)
		wrapperTypeWithAlias := fmt.Sprintf("*mmeobj.%sEntityWrapper", entityName)

		entityRefs = append(entityRefs, &router_gen.EntityRefInfo{
			FieldName:            field.Name,
			EntityName:           entityName,
			ParamName:            paramName,
			WrapperType:          wrapperType,
			WrapperTypeWithAlias: wrapperTypeWithAlias,
		})
	}

	return entityRefs
}

// isEntityRefField 检查字段是否是 EntityRef 类型
func isEntityRefField(field *blueprint_types.Field) bool {
	// EntityRef 字段的命名格式：{EntityName}EntityRef
	// 例如：PlayerEntityRef
	if strings.HasSuffix(field.Name, "EntityRef") {
		return true
	}

	// 检查类型名（可能是 "EntityRef" 或 "MME.EntityRef"）
	typeName := field.Type.GetName()
	if typeName == "EntityRef" || strings.Contains(typeName, "EntityRef") {
		return true
	}

	// 检查完整类型名（带包名）
	if field.Type.TypeName != "" {
		if strings.Contains(field.Type.TypeName, "EntityRef") {
			return true
		}
	}

	return false
}

// extractEntityNameFromFieldName 从字段名提取实体名称
// 例如：PlayerEntityRef -> Player
func extractEntityNameFromFieldName(fieldName string) string {
	// 去掉 EntityRef 后缀
	if strings.HasSuffix(fieldName, "EntityRef") {
		return strings.TrimSuffix(fieldName, "EntityRef")
	}

	return ""
}

// toCamelCase 转换为小驼峰命名
func toCamelCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}
