package blueprint_gen

import (
	"fmt"
	"strings"

	blueprint_types "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types"
)

// TypeReferenceResolver 类型引用解析器，用于处理跨文件、跨命名空间的类型引用
type TypeReferenceResolver struct {
	ctx              *BlueprintContext
	typeConverter    *TypeConverter
	nameSpaces       map[string]string
	enumMap          map[string]*Enum
}

// NewTypeReferenceResolver 创建新的类型引用解析器
func NewTypeReferenceResolver(ctx *BlueprintContext, typeConverter *TypeConverter) *TypeReferenceResolver {
	resolver := &TypeReferenceResolver{
		ctx:           ctx,
		typeConverter: typeConverter,
		enumMap:       make(map[string]*Enum),
	}

	// 构建枚举映射，便于快速查找
	for _, enum := range ctx.Enums {
		resolver.enumMap[enum.Name] = enum
	}
	for _, netwallFile := range ctx.NetWalls {
		for _, enum := range netwallFile.Enums {
			resolver.enumMap[enum.Name] = enum
		}
	}

	return resolver
}

// ResolveTypeReference 解析类型引用，返回完整的类型名称（包含包名前缀，如果需要）
// fieldType: 字段类型
// currentPackageName: 当前 proto 文件的包名
// 返回: 解析后的类型名称（如 "MME.HeroMechanism" 或 "Core.Book"）
func (r *TypeReferenceResolver) ResolveTypeReference(fieldType *blueprint_types.FieldType, currentPackageName string) string {
	// 获取基础类型名称
	typeName := fieldType.Name
	if typeName == "" {
		typeName = fieldType.TypeName
	}
	if typeName == "" {
		return ""
	}

	// 获取命名空间映射（延迟初始化）
	if r.nameSpaces == nil {
		r.nameSpaces = r.ctx.GetNameSpace()
	}

	// 检查是否是枚举类型
	if r.isEnumType(typeName) {
		return r.resolveEnumTypeReference(typeName, currentPackageName)
	}

	// 检查是否是消息类型或 MME Object 类型
	// MME Object 类型（Entity, Manager, Module, Mechanism）也需要处理
	if fieldType.IsMessage() || fieldType.IsMMEObjectType() {
		return r.resolveMessageTypeReference(typeName, currentPackageName)
	}

	// 其他类型直接返回
	return r.typeConverter.ToProtoType(fieldType)
}

// resolveEnumTypeReference 解析枚举类型引用
func (r *TypeReferenceResolver) resolveEnumTypeReference(typeName, currentPackageName string) string {
	enum, exists := r.enumMap[typeName]
	if !exists {
		return typeName
	}

	// 如果枚举不在当前包中，需要添加包名前缀
	if enum.SourceProto != currentPackageName {
		packageName := r.getPackageNameForSource(enum.SourceProto)
		if packageName != "" && packageName != currentPackageName {
			// MME 包中的枚举不需要前缀（因为都在同一个包中）
			// 其他包的枚举需要添加包名前缀
			if packageName != "MME" {
				return fmt.Sprintf("%s.%s", packageName, typeName)
			}
		}
	}

	return typeName
}

// resolveMessageTypeReference 解析消息类型引用
func (r *TypeReferenceResolver) resolveMessageTypeReference(typeName, currentPackageName string) string {
	if r.nameSpaces == nil {
		r.nameSpaces = r.ctx.GetNameSpace()
	}

	packageName, ok := r.nameSpaces[typeName]
	if !ok {
		// 如果找不到命名空间，可能是基础类型或其他类型，直接返回
		return typeName
	}

	// 如果消息类型不在当前包中，需要添加包名前缀
	// 注意：MME 包内的类型都在同一个包中，不需要添加包名前缀
	if packageName != currentPackageName {
		// 对于 MME 包内的类型，即使在不同文件中，也不添加包名前缀
		// 因为它们都在同一个 MME 包中
		if packageName == "MME" && currentPackageName == "MME" {
			return typeName
		}
		return fmt.Sprintf("%s.%s", packageName, typeName)
	}

	return typeName
}

// isEnumType 检查类型名称是否是枚举类型
func (r *TypeReferenceResolver) isEnumType(typeName string) bool {
	_, exists := r.enumMap[typeName]
	return exists
}

// getPackageNameForSource 根据 SourceProto 返回对应的包名
func (r *TypeReferenceResolver) getPackageNameForSource(sourceProto string) string {
	switch sourceProto {
	case "entities", "managers", "modules", "mechanisms", "common":
		return "MME" // 这些都在 MME 包中
	default:
		// NetWall 包名，返回原值（首字母大写）
		if sourceProto != "" {
			return strings.ToUpper(sourceProto[:1]) + sourceProto[1:]
		}
		return sourceProto
	}
}

// CollectImportsFromField 从字段中收集需要的导入
// field: 字段定义
// currentPackageName: 当前 proto 文件的包名（如 "MME", "Core" 等）
// 返回: 需要导入的 proto 文件列表
func (r *TypeReferenceResolver) CollectImportsFromField(field *blueprint_types.Field, currentPackageName string) []string {
	// 将包名转换为 SourceProto
	currentSourceProto := r.getSourceProtoForPackageName(currentPackageName)
	return r.CollectImportsFromFieldWithSourceProto(field, currentSourceProto)
}

// CollectImportsFromFieldWithSourceProto 从字段中收集需要的导入（使用 SourceProto）
// field: 字段定义
// currentSourceProto: 当前 proto 文件的 SourceProto（如 "entities", "managers", "common" 等）
// 返回: 需要导入的 proto 文件列表
func (r *TypeReferenceResolver) CollectImportsFromFieldWithSourceProto(field *blueprint_types.Field, currentSourceProto string) []string {
	imports := make(map[string]bool)

	if r.nameSpaces == nil {
		r.nameSpaces = r.ctx.GetNameSpace()
	}

	// 检查字段类型本身
	r.collectImportsFromFieldType(&field.Type, currentSourceProto, imports)

	// 检查 map/xmap 的 value 类型
	if field.Type.ValueType != nil {
		r.collectImportsFromFieldType(field.Type.ValueType, currentSourceProto, imports)
	}

	// 检查 map/xmap 的 key 类型（如果 key 是消息类型）
	if field.Type.KeyType != nil {
		r.collectImportsFromFieldType(field.Type.KeyType, currentSourceProto, imports)
	}

	// 转换为列表并返回
	result := make([]string, 0, len(imports))
	for imp := range imports {
		result = append(result, imp)
	}
	return UniqueProtoImports(result)
}

// CollectImportsFromFieldsWithSourceProto 从字段列表中收集所有需要的导入（使用 SourceProto）
func (r *TypeReferenceResolver) CollectImportsFromFieldsWithSourceProto(fields []*blueprint_types.Field, currentSourceProto string) []string {
	imports := make(map[string]bool)

	for _, field := range fields {
		fieldImports := r.CollectImportsFromFieldWithSourceProto(field, currentSourceProto)
		for _, imp := range fieldImports {
			imports[imp] = true
		}
	}

	// 转换为列表并返回
	result := make([]string, 0, len(imports))
	for imp := range imports {
		result = append(result, imp)
	}
	return UniqueProtoImports(result)
}

// getSourceProtoForPackageName 根据包名返回对应的 SourceProto
func (r *TypeReferenceResolver) getSourceProtoForPackageName(packageName string) string {
	// 对于 MME 包，需要根据当前生成的文件来判断
	// 这里返回空字符串，表示需要调用方明确指定 SourceProto
	// 对于非 MME 包，返回空字符串
	return ""
}

// collectImportsFromFieldType 从字段类型中收集导入
// currentSourceProto: 当前 proto 文件的 SourceProto（如 "entities", "managers", "common" 等）
func (r *TypeReferenceResolver) collectImportsFromFieldType(fieldType *blueprint_types.FieldType, currentSourceProto string, imports map[string]bool) {
	typeName := fieldType.Name
	if typeName == "" {
		typeName = fieldType.TypeName
	}
	if typeName == "" {
		return
	}

	if r.nameSpaces == nil {
		r.nameSpaces = r.ctx.GetNameSpace()
	}

	// 检查是否是枚举类型
	if enum, exists := r.enumMap[typeName]; exists {
		// 如果枚举不在当前文件中，需要导入
		if enum.SourceProto != currentSourceProto {
			importFile := r.getProtoImportForSource(enum.SourceProto)
			if importFile != "" && importFile != r.getProtoImportForSource(currentSourceProto) {
				imports[importFile] = true
			}
		}
		return
	}

	// 检查是否是消息类型或 MME Object 类型
	// MME Object 类型（Entity, Manager, Module, Mechanism）也需要导入
	if fieldType.IsMessage() || fieldType.IsMMEObjectType() {
		// 获取消息类型的 SourceProto
		typeSourceProto := r.getSourceProtoForType(typeName)
		if typeSourceProto != "" && typeSourceProto != currentSourceProto {
			importFile := r.getProtoImportForSource(typeSourceProto)
			if importFile != "" && importFile != r.getProtoImportForSource(currentSourceProto) {
				imports[importFile] = true
			}
		}
	}
}

// getSourceProtoForType 根据类型名称返回对应的 SourceProto
func (r *TypeReferenceResolver) getSourceProtoForType(typeName string) string {
	// 检查是否是 Entity
	for _, entity := range r.ctx.Entities {
		if entity.Name == typeName {
			return "entities"
		}
	}
	// 检查是否是 Manager
	for _, manager := range r.ctx.Managers {
		if manager.Name == typeName {
			return "managers"
		}
	}
	// 检查是否是 Module
	for _, module := range r.ctx.Modules {
		if module.Name == typeName {
			return "modules"
		}
	}
	// 检查是否是 Mechanism
	for _, mechanism := range r.ctx.Mechanisms {
		if mechanism.Name == typeName {
			return "mechanisms"
		}
	}
	// 检查是否是 Common DataStruct
	if r.ctx.HeadFile != nil {
		for _, ds := range r.ctx.HeadFile.CommonDataStructs {
			if ds.Name == typeName {
				return "common"
			}
		}
	}
	return ""
}

// getProtoImportForSource 根据 SourceProto 返回对应的 proto 导入文件
func (r *TypeReferenceResolver) getProtoImportForSource(sourceProto string) string {
	switch sourceProto {
	case "entities":
		return "entities.proto"
	case "managers":
		return "managers.proto"
	case "modules":
		return "modules.proto"
	case "mechanisms":
		return "mechanisms.proto"
	case "common":
		return "common.proto"
	default:
		// NetWall 包名，转换为小写并添加 .proto 后缀
		if sourceProto != "" {
			return strings.ToLower(sourceProto) + ".proto"
		}
		return ""
	}
}

// getProtoImportForPackage 根据包名返回对应的 proto 导入文件
func (r *TypeReferenceResolver) getProtoImportForPackage(packageName string) string {
	switch packageName {
	case "MME":
		// MME 包可能来自多个文件，需要根据具体类型判断
		// 这里返回空，由调用方根据具体类型决定
		return ""
	case "Enum":
		// 枚举可能来自多个文件，需要根据具体类型判断
		return ""
	default:
		// NetWall 包名，转换为小写并添加 .proto 后缀
		if packageName != "" {
			return strings.ToLower(packageName) + ".proto"
		}
		return ""
	}
}

// CollectImportsFromFields 从字段列表中收集所有需要的导入
// currentPackageName: 当前 proto 文件的包名（如 "MME", "Core" 等）
// 注意：对于 MME 包，应该使用 CollectImportsFromFieldsWithSourceProto 并传入 SourceProto
func (r *TypeReferenceResolver) CollectImportsFromFields(fields []*blueprint_types.Field, currentPackageName string) []string {
	imports := make(map[string]bool)

	for _, field := range fields {
		fieldImports := r.CollectImportsFromField(field, currentPackageName)
		for _, imp := range fieldImports {
			imports[imp] = true
		}
	}

	// 转换为列表并返回
	result := make([]string, 0, len(imports))
	for imp := range imports {
		result = append(result, imp)
	}
	return UniqueProtoImports(result)
}

// CollectMessageImportsFromFields 从字段列表中收集消息类型的导入（不包括枚举）
func (r *TypeReferenceResolver) CollectMessageImportsFromFields(fields []*blueprint_types.Field, currentPackageName string) []string {
	imports := make(map[string]bool)

	if r.nameSpaces == nil {
		r.nameSpaces = r.ctx.GetNameSpace()
	}

	for _, field := range fields {
		typeName := field.Type.Name
		if typeName == "" {
			typeName = field.Type.TypeName
		}
		if typeName == "" {
			continue
		}

		// 只处理消息类型，跳过枚举
		if field.Type.IsMessage() && !r.isEnumType(typeName) {
			if packageName, ok := r.nameSpaces[typeName]; ok && packageName != currentPackageName {
				importFile := r.getProtoImportForType(typeName, packageName)
				if importFile != "" {
					imports[importFile] = true
				}
			}
		}

		// 检查 map/xmap 的 value 类型
		if field.Type.ValueType != nil {
			valueTypeName := field.Type.ValueType.Name
			if valueTypeName == "" {
				valueTypeName = field.Type.ValueType.TypeName
			}
			if valueTypeName != "" && field.Type.ValueType.IsMessage() && !r.isEnumType(valueTypeName) {
				if packageName, ok := r.nameSpaces[valueTypeName]; ok && packageName != currentPackageName {
					importFile := r.getProtoImportForType(valueTypeName, packageName)
					if importFile != "" {
						imports[importFile] = true
					}
				}
			}
		}
	}

	// 转换为列表并返回
	result := make([]string, 0, len(imports))
	for imp := range imports {
		result = append(result, imp)
	}
	return UniqueProtoImports(result)
}

// getProtoImportForType 根据类型名称和包名返回对应的 proto 导入文件
func (r *TypeReferenceResolver) getProtoImportForType(typeName, packageName string) string {
	if packageName == "MME" {
		// MME 包需要根据类型判断具体文件
		return r.getMMEProtoImportForType(typeName)
	}
	// 其他包直接使用包名
	return r.getProtoImportForPackage(packageName)
}

// getMMEProtoImportForType 根据 MME 类型名称返回对应的 proto 文件
func (r *TypeReferenceResolver) getMMEProtoImportForType(typeName string) string {
	// 检查是否是 Entity
	for _, entity := range r.ctx.Entities {
		if entity.Name == typeName {
			return "entities.proto"
		}
	}
	// 检查是否是 Manager
	for _, manager := range r.ctx.Managers {
		if manager.Name == typeName {
			return "managers.proto"
		}
	}
	// 检查是否是 Module
	for _, module := range r.ctx.Modules {
		if module.Name == typeName {
			return "modules.proto"
		}
	}
	// 检查是否是 Mechanism
	for _, mechanism := range r.ctx.Mechanisms {
		if mechanism.Name == typeName {
			return "mechanisms.proto"
		}
	}
	// 检查是否是 Common DataStruct
	if r.ctx.HeadFile != nil {
		for _, ds := range r.ctx.HeadFile.CommonDataStructs {
			if ds.Name == typeName {
				return "common.proto"
			}
		}
	}
	return ""
}

