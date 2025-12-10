package field_parser

import (
	"fmt"
	"regexp"
	"strings"

	types "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types"
	mmeobject "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types/mme_obj"
)

// ParseFieldDefinition 解析字段定义行，直接返回 *types.Field
// 格式: int32 FieldName: 1 [blueprint:"access=all"]
// 或者: HeroManager HeroManager: 1 [blueprint:"access=all"]
func ParseFieldDefinition(line string) *types.Field {
	field := new(types.Field)
	if err := BuildFieldType(line, field); err != nil {
		panic(fmt.Sprintf("failed to build field type: %s", err))
	}

	// 构建 FieldMetadata
	fieldMetadata := types.BuildFieldMetadata(field)
	field.Metadata = fieldMetadata

	if field.IsMMEObjectType() {
		name := field.GetTypeName()
		mmeObjectType, err := mmeobject.ParseMMEObjectType(name)
		if err != nil {
			panic(fmt.Sprintf("failed to parse mme object type: %s", name))
		}
		fieldMetadata.SetMMEObjectKind(mmeObjectType)
	}

	return field
}

func BuildFieldType(line string, field *types.Field) error {

	// 提取注释
	parts := strings.Split(line, "#")
	if len(parts) > 1 {
		field.Comment = strings.TrimSpace(parts[1])
		line = strings.TrimSpace(parts[0])
	}

	// 提取选项 [blueprint:"access=all"]
	var optionStr string
	if idx := strings.Index(line, "["); idx != -1 {
		optionStr = line[idx:]
		line = line[:idx]
		line = strings.TrimSpace(line)
	}
	field.Options = ParseFieldOptions(optionStr)

	// 解析字段编号 : N
	re := regexp.MustCompile(`:\s*(\d+)\s*`)
	matches := re.FindStringSubmatch(line)
	if len(matches) < 2 {
		return fmt.Errorf("field number not found in: %s", line)
	}

	var num int32
	if _, err := fmt.Sscanf(matches[1], "%d", &num); err != nil {
		return fmt.Errorf("failed to scan field number: %w", err)
	}
	field.Number = num

	// 移除 :N 部分
	line = re.ReplaceAllString(line, "")
	line = strings.TrimSpace(line)

	// 分割类型和字段名
	parts = strings.Fields(line)
	if len(parts) < 2 {
		return fmt.Errorf("invalid field format: %s", line)
	}

	// 类型部分是前面的所有部分（可能包含空格，如 "Core.MMELocation"）
	// 字段名是最后一部分
	field.Name = parts[len(parts)-1]
	typeParts := parts[:len(parts)-1]
	typeStr := strings.Join(typeParts, " ")

	// 使用 types.ParseTypeString 解析类型
	fieldType, err := types.ParseTypeString(typeStr)
	if err != nil {
		return fmt.Errorf("failed to parse type: %w", err)
	}
	field.Type = *fieldType

	return nil
}

// ParseFieldOptions 解析字段选项字符串
// 格式: [blueprint:"access=all", orbit:"Access=all"]
func ParseFieldOptions(optionStr string) types.FieldOption {
	opt := types.FieldOption{}

	if optionStr == "" {
		return opt
	}

	// 使用正则表达式提取 access=xxx
	re := regexp.MustCompile(`access\s*=\s*([a-z]+)`)
	matches := re.FindStringSubmatch(optionStr)
	if len(matches) >= 2 {
		opt.Access = matches[1]
	} else {
		opt.Access = "all" // 默认值
	}

	return opt
}
