package router_gen

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// GenerateControllerMethods 生成或更新 Controller Handle 方法（增量模式）
func GenerateControllerMethods(ctx *RouterGenContext) error {
	// 确定 Controller 文件路径
	controllerPath := ctx.ControllerPath
	if controllerPath == "" {
		// 尝试从 router output 推断
		controllerPath = inferControllerPath(ctx)
	}
	if controllerPath == "" {
		// 如果无法推断，使用默认路径
		if ctx.Controller != nil {
			controllerPath = findControllerFile(ctx.Controller)
		}
	}
	if controllerPath == "" {
		return fmt.Errorf("cannot determine controller file path, please specify --controller-path")
	}

	// 检查文件是否存在
	fileExists := false
	if _, err := os.Stat(controllerPath); err == nil {
		fileExists = true
	}

	// 解析现有的 Controller 文件，获取所有已存在的 Handle 方法（不管是否有实现）
	existingMethods, err := parseExistingControllerMethods(ctx.Controller, controllerPath)
	if err != nil {
		// 如果文件不存在或解析失败，existingMethods 为空，继续生成
		existingMethods = make(map[string]*MethodInfo)
	}

	// 筛选出需要生成的新方法（不存在的方法）
	newRequests := make([]*RequestInfo, 0)
	for _, req := range ctx.Requests {
		methodName := fmt.Sprintf("Handle%s", req.RequestName)
		if _, exists := existingMethods[methodName]; !exists {
			newRequests = append(newRequests, req)
		}
	}

	fmt.Printf("DEBUG: Controller generation - Existing methods: %d, New requests to generate: %d\n", len(existingMethods), len(newRequests))

	// 如果没有新方法需要生成，直接返回
	if len(newRequests) == 0 {
		return nil
	}

	// 如果文件已存在，使用增量模式追加新方法
	if fileExists {
		return appendNewMethods(ctx, controllerPath, newRequests)
	}

	// 如果文件不存在，生成完整的 Controller 文件
	controllerCode, err := generateControllerFile(ctx, existingMethods)
	if err != nil {
		return fmt.Errorf("failed to generate controller code: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(controllerPath, []byte(controllerCode), 0644); err != nil {
		return fmt.Errorf("failed to write controller file: %w", err)
	}

	return nil
}

// MethodInfo 方法信息
type MethodInfo struct {
	Name       string   // 方法名
	Params     []string // 参数列表（字符串形式）
	ReturnType string   // 返回类型
	Body       string   // 方法体（如果有实现）
	HasBody    bool     // 是否有方法体（不是空实现）
}

// parseExistingControllerMethods 解析现有的 Controller 方法
func parseExistingControllerMethods(controller *ControllerInfo, routerOutputPath string) (map[string]*MethodInfo, error) {
	methods := make(map[string]*MethodInfo)

	// 从 router output 路径推断 controller 路径
	controllerPath := inferControllerPathFromRouter(routerOutputPath)
	if controllerPath == "" {
		// 如果无法推断，尝试使用 controller 信息
		if controller != nil {
			// 尝试查找 controller 文件
			controllerPath = findControllerFile(controller)
		}
	}

	if controllerPath == "" {
		return methods, nil // 文件不存在，返回空 map
	}

	// 检查文件是否存在
	if _, err := os.Stat(controllerPath); os.IsNotExist(err) {
		return methods, nil
	}

	// 解析 Go 文件
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, controllerPath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse controller file: %w", err)
	}

	// 确定 Controller 类型名
	typeName := "Controller"
	if controller != nil && controller.TypeName != "" {
		typeName = controller.TypeName
	}

	// 查找 Controller 类型的方法
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			// 检查是否是 Controller 的方法
			if x.Recv != nil && len(x.Recv.List) > 0 {
				recv := x.Recv.List[0]
				if starExpr, ok := recv.Type.(*ast.StarExpr); ok {
					if ident, ok := starExpr.X.(*ast.Ident); ok {
						if ident.Name == typeName {
							// 这是 Controller 的方法
							methodInfo := extractMethodInfo(x, fset, controllerPath)
							if methodInfo != nil && strings.HasPrefix(methodInfo.Name, "Handle") {
								// 增量模式：记录所有 Handle 方法（不管是否有实现）
								// 只要方法存在就记录，用于判断是否需要生成
								methods[methodInfo.Name] = methodInfo
							}
						}
					}
				}
			}
		}
		return true
	})

	return methods, nil
}

// extractMethodInfo 提取方法信息
func extractMethodInfo(fn *ast.FuncDecl, fset *token.FileSet, filePath string) *MethodInfo {
	methodInfo := &MethodInfo{
		Name:    fn.Name.Name,
		Params:  make([]string, 0),
		HasBody: fn.Body != nil && len(fn.Body.List) > 0,
	}

	// 提取参数
	for _, param := range fn.Type.Params.List {
		paramType := formatNode(param.Type, fset)
		for _, name := range param.Names {
			paramName := name.Name
			methodInfo.Params = append(methodInfo.Params, fmt.Sprintf("%s %s", paramName, paramType))
		}
		if len(param.Names) == 0 {
			// 匿名参数
			methodInfo.Params = append(methodInfo.Params, paramType)
		}
	}

	// 提取返回类型
	if fn.Type.Results != nil && len(fn.Type.Results.List) > 0 {
		methodInfo.ReturnType = formatNode(fn.Type.Results.List[0].Type, fset)
	}

	// 提取方法体（如果有实现且不是简单的 return nil）
	if fn.Body != nil && methodInfo.HasBody {
		// 读取原始文件内容来提取方法体
		methodInfo.Body = extractMethodBody(fn, fset, filePath)

		// 如果提取的 body 为空，但方法有保护标记，仍然保留
		if methodInfo.Body == "" && hasPreserveMarker(fn) {
			// 强制保留（即使看起来是空实现）
			body := extractMethodBodyRaw(fn, fset, filePath)
			if body != "" {
				methodInfo.Body = body
			}
		}
	}

	return methodInfo
}

// formatNode 格式化 AST 节点为字符串
func formatNode(node ast.Node, fset *token.FileSet) string {
	var buf strings.Builder
	if err := format.Node(&buf, fset, node); err != nil {
		return ""
	}
	return strings.TrimSpace(buf.String())
}

// extractMethodBody 提取方法体
func extractMethodBody(fn *ast.FuncDecl, fset *token.FileSet, filePath string) string {
	if fn.Body == nil {
		return ""
	}

	// 读取文件内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		return ""
	}

	// 提取方法体的文本
	start := fset.Position(fn.Body.Pos()).Offset
	end := fset.Position(fn.Body.End()).Offset
	if start < 0 || end < 0 || start >= len(data) || end > len(data) {
		return ""
	}

	body := string(data[start+1 : end-1]) // 去掉 { 和 }
	bodyTrimmed := strings.TrimSpace(body)

	// 检查是否是简单的 return nil（空实现，不保留）
	if bodyTrimmed == "return nil" || bodyTrimmed == "return" {
		return ""
	}

	// 检查是否包含业务代码
	if !isBusinessCode(body) {
		return "" // 不是业务代码，不保留
	}

	return body
}

// isBusinessCode 判断方法体是否包含业务代码
// 业务代码的特征：
// 1. 不是简单的 return nil
// 2. 不是只有 TODO 注释
// 3. 有实际的业务逻辑（如函数调用、变量赋值等）
func isBusinessCode(body string) bool {
	if body == "" {
		return false
	}

	bodyTrimmed := strings.TrimSpace(body)

	// 检查是否是简单的 return nil
	if bodyTrimmed == "return nil" || bodyTrimmed == "return" {
		return false
	}

	// 检查是否只有 TODO 注释
	lines := strings.Split(bodyTrimmed, "\n")
	hasNonCommentCode := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// 跳过空行和注释
		if trimmed == "" || strings.HasPrefix(trimmed, "//") {
			continue
		}
		// 如果包含实际的代码（不是只有 return nil）
		if trimmed != "return nil" && trimmed != "return" {
			hasNonCommentCode = true
			break
		}
	}

	return hasNonCommentCode
}

// extractMethodBodyRaw 提取方法体（原始版本，不进行业务代码判断）
func extractMethodBodyRaw(fn *ast.FuncDecl, fset *token.FileSet, filePath string) string {
	if fn.Body == nil {
		return ""
	}

	// 读取文件内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		return ""
	}

	// 提取方法体的文本
	start := fset.Position(fn.Body.Pos()).Offset
	end := fset.Position(fn.Body.End()).Offset
	if start < 0 || end < 0 || start >= len(data) || end > len(data) {
		return ""
	}

	body := string(data[start+1 : end-1]) // 去掉 { 和 }
	return body
}

// hasPreserveMarker 检查方法是否有保护标记
// 支持注释：// +routergen:preserve 或 // routergen:preserve
func hasPreserveMarker(fn *ast.FuncDecl) bool {
	if fn.Doc == nil {
		return false
	}

	for _, comment := range fn.Doc.List {
		if strings.Contains(comment.Text, "+routergen:preserve") ||
			strings.Contains(comment.Text, "routergen:preserve") {
			return true
		}
	}

	return false
}

// generateControllerFile 生成 Controller 文件
func generateControllerFile(ctx *RouterGenContext, existingMethods map[string]*MethodInfo) (string, error) {
	var code strings.Builder

	// 生成文件头
	code.WriteString("// Code generated by routergen. DO NOT EDIT.\n")
	code.WriteString("// This file is automatically generated from NetWall YAML files.\n")
	code.WriteString("// To regenerate, run: go generate or blueprintgen.\n\n")

	// 生成包声明
	if ctx.Controller != nil {
		code.WriteString(fmt.Sprintf("package %s\n\n", ctx.Controller.PackageName))
	} else {
		code.WriteString("package controllerv2\n\n")
	}

	// 从现有文件中提取导入（如果文件存在）
	existingImports := extractExistingImports(ctx)

	// 生成导入语句
	code.WriteString("import (\n")

	// 收集所有需要的 proto 包导入
	protoPackages := make(map[string]bool)
	for _, req := range ctx.Requests {
		protoPackages[req.PackageName] = true
	}

	// 合并现有导入和新需要的导入，按标准分组
	allImports := mergeAndGroupImports(existingImports, protoPackages, ctx)
	code.WriteString(allImports)

	code.WriteString(")\n\n")

	// 确定 Controller 类型名和变量名
	typeName := "Controller"
	varName := "GControllerV2"
	if ctx.Controller != nil {
		if ctx.Controller.TypeName != "" {
			typeName = ctx.Controller.TypeName
		}
		if ctx.Controller.VarName != "" {
			varName = ctx.Controller.VarName
		}
	}

	// 生成 Controller 变量
	code.WriteString("var (\n")
	code.WriteString(fmt.Sprintf("\t%s = &%s{}\n", varName, typeName))
	code.WriteString(")\n\n")

	// 生成 Controller 类型
	code.WriteString(fmt.Sprintf("type %s struct{}\n\n", typeName))

	// 生成 Handle 方法
	for _, req := range ctx.Requests {
		methodName := fmt.Sprintf("Handle%s", req.RequestName)
		methodInfo, exists := existingMethods[methodName]

		// 生成方法签名
		code.WriteString(generateHandleMethod(req, ctx.Controller, methodInfo, exists))
		code.WriteString("\n")
	}

	// 格式化代码
	formatted, err := format.Source([]byte(code.String()))
	if err != nil {
		// 如果格式化失败，返回原始代码
		return code.String(), nil
	}

	return string(formatted), nil
}

// generateHandleMethod 生成单个 Handle 方法
func generateHandleMethod(req *RequestInfo, controller *ControllerInfo, existingMethod *MethodInfo, exists bool) string {
	var code strings.Builder

	methodName := fmt.Sprintf("Handle%s", req.RequestName)

	// 确定 Controller 类型名
	typeName := "Controller"
	if controller != nil && controller.TypeName != "" {
		typeName = controller.TypeName
	}

	// 生成方法签名
	code.WriteString(fmt.Sprintf("func (c *%s) %s(", typeName, methodName))

	// 第一个参数：Request
	requestType := fmt.Sprintf("%s.Request_%s", req.PackageName, req.RequestName)
	code.WriteString(fmt.Sprintf("req *%s", requestType))

	// 后续参数：EntityRef 实体（使用 EntityAgentTypeWithAlias）
	// Controller 层应该接收业务逻辑层的实体类型（PlayerEntityImpl），而不是数据层的包装器
	for _, entityRef := range req.EntityRefs {
		// 使用 EntityAgentTypeWithAlias (例如 *agent.PlayerEntityImpl)
		code.WriteString(fmt.Sprintf(", %s %s", entityRef.ParamName, entityRef.EntityAgentTypeWithAlias))
	}

	code.WriteString(") proto.Message {\n")

	// 生成方法体
	// 优先保留已有的业务代码实现
	if exists && existingMethod != nil && existingMethod.HasBody && existingMethod.Body != "" {
		// 保留已有的实现
		body := existingMethod.Body

		// 确保每行都有正确的缩进
		lines := strings.Split(body, "\n")
		for i, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" {
				// 如果行没有缩进，添加缩进
				if !strings.HasPrefix(line, "\t") && !strings.HasPrefix(line, " ") {
					code.WriteString("\t" + line)
				} else {
					code.WriteString(line)
				}
				// 如果不是最后一行，添加换行
				if i < len(lines)-1 {
					code.WriteString("\n")
				}
			} else if i < len(lines)-1 {
				// 空行也要保留（除了最后一行）
				code.WriteString("\n")
			}
		}
		code.WriteString("\n")
	} else {
		// 生成默认实现
		if len(req.EntityRefs) > 0 {
			code.WriteString("\t// TODO: 实现业务逻辑\n")
			for _, entityRef := range req.EntityRefs {
				code.WriteString(fmt.Sprintf("\t// %s 已经由 Router 层加载完成，可以直接使用\n", entityRef.ParamName))
			}
		} else {
			code.WriteString("\t// TODO: 实现业务逻辑\n")
		}
		code.WriteString("\treturn nil\n")
	}

	code.WriteString("}\n")

	return code.String()
}

// inferControllerPath 从 RouterGenContext 推断 Controller 文件路径
func inferControllerPath(ctx *RouterGenContext) string {
	// 从 router output 路径推断
	if ctx.OutputPath != "" {
		return inferControllerPathFromRouter(ctx.OutputPath)
	}

	// 如果无法推断，使用默认路径
	if ctx.Controller != nil {
		return findControllerFile(ctx.Controller)
	}

	return ""
}

// inferControllerPathFromRouter 从 router 输出路径推断 controller 路径
func inferControllerPathFromRouter(routerPath string) string {
	// router 路径通常是 app/routers/routers.go
	// controller 路径通常是 app/controller_v2/controller.go
	dir := filepath.Dir(routerPath)

	// 尝试多个可能的路径
	possiblePaths := []string{
		filepath.Join(filepath.Dir(dir), "controller_v2", "controller.go"),
		filepath.Join(filepath.Dir(dir), "controller", "controller.go"),
		filepath.Join(dir, "..", "controller_v2", "controller.go"),
		filepath.Join(dir, "..", "controller", "controller.go"),
	}

	for _, path := range possiblePaths {
		if absPath, err := filepath.Abs(path); err == nil {
			if _, err := os.Stat(absPath); err == nil {
				return absPath
			}
		}
	}

	return ""
}

// findControllerFile 查找 Controller 文件
func findControllerFile(controller *ControllerInfo) string {
	// 尝试多个可能的路径
	possibleDirs := []string{
		"internal/game/controller_v2",
		"internal/game/controller",
		"app/controller_v2",
		"app/controller",
		"orbit/internal/game/controller_v2",
		"orbit/internal/game/controller",
		"orbit/app/controller_v2",
		"orbit/app/controller",
	}

	for _, dir := range possibleDirs {
		controllerPath := filepath.Join(dir, "controller.go")
		if absPath, err := filepath.Abs(controllerPath); err == nil {
			if _, err := os.Stat(absPath); err == nil {
				return absPath
			}
		}
	}

	return ""
}

// appendNewMethods 增量模式：在现有文件中追加新方法
func appendNewMethods(ctx *RouterGenContext, controllerPath string, newRequests []*RequestInfo) error {
	// 读取现有文件内容
	existingContent, err := os.ReadFile(controllerPath)
	if err != nil {
		return fmt.Errorf("failed to read existing controller file: %w", err)
	}

	// 解析现有文件，找到最后一个方法的位置
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, controllerPath, existingContent, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse existing controller file: %w", err)
	}

	// 检查是否需要额外的导入
	needsMMEObjImport := false
	needsAgentImport := false
	for _, req := range newRequests {
		if len(req.EntityRefs) > 0 {
			needsMMEObjImport = true
			needsAgentImport = true
			break
		}
	}

	// 检查是否已经导入了 mmeobj 和 agent
	hasMMEObjImport := false
	hasAgentImport := false
	for _, imp := range node.Imports {
		if imp.Path != nil {
			importPath := strings.Trim(imp.Path.Value, "\"")
			if importPath == "gitee.com/orbit-w/orbit/internal/game/mme" {
				// 检查是否有别名 mmeobj
				if imp.Name != nil && imp.Name.Name == "mmeobj" {
					hasMMEObjImport = true
				}
			}
			if importPath == "gitee.com/orbit-w/orbit/internal/game/mme_agent/entities/player" {
				// 检查是否有别名 agent
				if imp.Name != nil && imp.Name.Name == "agent" {
					hasAgentImport = true
				}
			}
		}
	}

	// 如果需要导入但没有，添加导入
	content := string(existingContent)
	importsToAdd := []string{}
	if needsMMEObjImport && !hasMMEObjImport {
		importsToAdd = append(importsToAdd, "\tmmeobj \"gitee.com/orbit-w/orbit/internal/game/mme\"")
	}
	if needsAgentImport && !hasAgentImport {
		importsToAdd = append(importsToAdd, "\tagent \"gitee.com/orbit-w/orbit/internal/game/mme_agent/entities/player\"")
	}

	if len(importsToAdd) > 0 {
		// 使用 AST 找到 import 块的位置
		var importEndPos token.Pos
		if len(node.Imports) > 0 {
			// 找到最后一个 import 的位置
			lastImport := node.Imports[len(node.Imports)-1]
			importEndPos = lastImport.End()
		} else {
			// 如果没有导入，查找 import 关键字后的位置
			// 查找 "import (" 或 "import"
			importKeyword := strings.Index(content, "import (")
			if importKeyword != -1 {
				// 找到 import 块的开始
				importStart := importKeyword + len("import (")
				// 找到 import 块的结束
				importEnd := strings.Index(content[importStart:], ")")
				if importEnd != -1 {
					importEndPos = token.Pos(importStart + importEnd)
				}
			}
		}

		// 将位置转换为字节偏移
		if importEndPos > 0 {
			importEndOffset := fset.Position(importEndPos).Offset
			if importEndOffset > 0 && importEndOffset < len(content) {
				// 找到这一行的结束位置（换行符）
				lineEnd := strings.Index(content[importEndOffset:], "\n")
				if lineEnd == -1 {
					// 如果没有换行符，在当前位置后添加
					lineEnd = 0
				}
				insertPos := importEndOffset + lineEnd

				// 在最后一个导入后添加新导入
				before := content[:insertPos]
				after := content[insertPos:]
				// 确保有正确的缩进和格式
				newImport := "\n" + strings.Join(importsToAdd, "\n")
				content = before + newImport + after

				// 更新 existingContent
				existingContent = []byte(content)
				// 重新解析
				node, err = parser.ParseFile(fset, controllerPath, existingContent, parser.ParseComments)
				if err != nil {
					return fmt.Errorf("failed to re-parse controller file: %w", err)
				}
			}
		}
	}

	// 确定 Controller 类型名
	typeName := "Controller"
	if ctx.Controller != nil && ctx.Controller.TypeName != "" {
		typeName = ctx.Controller.TypeName
	}

	// 找到最后一个 Handle 方法的位置
	var lastMethodEnd token.Pos
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			// 检查是否是 Controller 的方法
			if x.Recv != nil && len(x.Recv.List) > 0 {
				recv := x.Recv.List[0]
				if starExpr, ok := recv.Type.(*ast.StarExpr); ok {
					if ident, ok := starExpr.X.(*ast.Ident); ok {
						if ident.Name == typeName && strings.HasPrefix(x.Name.Name, "Handle") {
							// 更新最后一个方法的位置
							if x.End() > lastMethodEnd {
								lastMethodEnd = x.End()
							}
						}
					}
				}
			}
		}
		return true
	})

	// 生成新方法的代码
	var newMethodsCode strings.Builder
	for _, req := range newRequests {
		methodCode := generateHandleMethod(req, ctx.Controller, nil, false)
		newMethodsCode.WriteString(methodCode)
		newMethodsCode.WriteString("\n")
	}

	// content 已经在前面声明过了，这里直接使用

	// 如果找到了最后一个方法，在它之后插入新方法
	if lastMethodEnd > 0 {
		insertPos := fset.Position(lastMethodEnd).Offset
		if insertPos > 0 && insertPos < len(content) {
			// 在最后一个方法之后插入新方法
			before := content[:insertPos]
			after := content[insertPos:]

			// 确保在最后一个方法后有换行
			if !strings.HasSuffix(before, "\n") {
				before += "\n"
			}
			if !strings.HasSuffix(before, "\n\n") {
				before += "\n"
			}

			content = before + newMethodsCode.String() + after
		} else {
			// 如果位置无效，直接追加到文件末尾
			if !strings.HasSuffix(content, "\n") {
				content += "\n"
			}
			content += newMethodsCode.String()
		}
	} else {
		// 如果没有找到任何方法，追加到文件末尾
		if !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		content += "\n" + newMethodsCode.String()
	}

	// 格式化代码
	formatted, err := format.Source([]byte(content))
	if err != nil {
		// 如果格式化失败，使用原始内容
		formatted = []byte(content)
	}

	// 写入文件
	if err := os.WriteFile(controllerPath, formatted, 0644); err != nil {
		return fmt.Errorf("failed to write controller file: %w", err)
	}

	return nil
}

// extractExistingImports 从现有 Controller 文件中提取导入
func extractExistingImports(ctx *RouterGenContext) map[string]string {
	imports := make(map[string]string) // path -> alias (or "" for no alias)

	// 确定 Controller 文件路径
	controllerPath := ctx.ControllerPath
	if controllerPath == "" {
		controllerPath = inferControllerPath(ctx)
	}
	if controllerPath == "" {
		return imports
	}

	// 检查文件是否存在
	if _, err := os.Stat(controllerPath); os.IsNotExist(err) {
		return imports
	}

	// 解析文件
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, controllerPath, nil, parser.ParseComments)
	if err != nil {
		return imports
	}

	// 提取所有导入
	for _, imp := range node.Imports {
		if imp.Path == nil {
			continue
		}

		path := strings.Trim(imp.Path.Value, "\"")
		alias := ""
		if imp.Name != nil {
			alias = imp.Name.Name
		}

		imports[path] = alias
	}

	return imports
}

// mergeAndGroupImports 合并并分组导入
func mergeAndGroupImports(existingImports map[string]string, protoPackages map[string]bool, ctx *RouterGenContext) string {
	var result strings.Builder

	// 分组：第三方库和本地项目包
	thirdParty := make([]string, 0) // meteor, zap 等外部依赖
	localPkg := make([]string, 0)   // 本地项目包（mmeobj, proto）
	localProto := make([]string, 0) // pkg/proto/*
	googlePb := ""                  // google.golang.org/protobuf/proto

	// 从现有导入中提取（除了要重新生成的 proto 包）
	for path, alias := range existingImports {
		// 跳过将被重新生成的 proto 包导入
		if strings.Contains(path, "/pkg/proto/") {
			continue
		}
		if path == "google.golang.org/protobuf/proto" {
			googlePb = path
			continue
		}

		// 区分第三方库和本地包
		if strings.Contains(path, "gitee.com/orbit-w/orbit/") {
			// 本地项目包
			if alias != "" {
				localPkg = append(localPkg, fmt.Sprintf("\t%s \"%s\"\n", alias, path))
			} else {
				localPkg = append(localPkg, fmt.Sprintf("\t\"%s\"\n", path))
			}
		} else if strings.HasPrefix(path, "gitee.com") ||
			strings.HasPrefix(path, "github.com") ||
			strings.HasPrefix(path, "go.uber.org") {
			// 第三方库（meteor, zap 等）
			if alias != "" {
				thirdParty = append(thirdParty, fmt.Sprintf("\t%s \"%s\"\n", alias, path))
			} else {
				thirdParty = append(thirdParty, fmt.Sprintf("\t\"%s\"\n", path))
			}
		}
	}

	// 确保有 mmeobj 导入
	if _, exists := existingImports["gitee.com/orbit-w/orbit/internal/game/mme"]; !exists {
		localPkg = append(localPkg, "\tmmeobj \"gitee.com/orbit-w/orbit/internal/game/mme\"\n")
	}

	// 检查是否需要 agent 导入（如果有 EntityRef）
	needsAgentImport := false
	for _, req := range ctx.Requests {
		if len(req.EntityRefs) > 0 {
			needsAgentImport = true
			break
		}
	}
	if needsAgentImport {
		if _, exists := existingImports["gitee.com/orbit-w/orbit/internal/game/mme_agent/entities/player"]; !exists {
			localPkg = append(localPkg, "\tagent \"gitee.com/orbit-w/orbit/internal/game/mme_agent/entities/player\"\n")
		}
	}

	// 添加所需的 proto 包导入
	sortedProtoPackages := make([]string, 0, len(protoPackages))
	for pkgName := range protoPackages {
		sortedProtoPackages = append(sortedProtoPackages, pkgName)
	}
	// 简单排序
	for i := 0; i < len(sortedProtoPackages)-1; i++ {
		for j := i + 1; j < len(sortedProtoPackages); j++ {
			if sortedProtoPackages[i] > sortedProtoPackages[j] {
				sortedProtoPackages[i], sortedProtoPackages[j] = sortedProtoPackages[j], sortedProtoPackages[i]
			}
		}
	}

	for _, pkgName := range sortedProtoPackages {
		localProto = append(localProto, fmt.Sprintf("\t\"gitee.com/orbit-w/orbit/pkg/proto/%s\"\n", pkgName))
	}

	// 确保有 google.golang.org/protobuf/proto
	if googlePb == "" {
		googlePb = "google.golang.org/protobuf/proto"
	}
	localProto = append(localProto, fmt.Sprintf("\t\"%s\"\n", googlePb))

	// 组装最终的导入
	// 1. 第三方库（meteor, zap）
	for _, imp := range thirdParty {
		result.WriteString(imp)
	}

	// 2. 本地项目包（空行分隔）
	if len(thirdParty) > 0 && (len(localPkg) > 0 || len(localProto) > 0) {
		result.WriteString("\n")
	}

	for _, imp := range localPkg {
		result.WriteString(imp)
	}

	for _, imp := range localProto {
		result.WriteString(imp)
	}

	return result.String()
}
