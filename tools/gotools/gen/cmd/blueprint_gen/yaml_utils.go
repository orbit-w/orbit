package blueprint_gen

import (
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

// YAMLDocumentHandler YAML 文档处理器接口
type YAMLDocumentHandler interface {
	// Handle 处理单个 YAML 文档，返回处理后的 map
	Handle(doc any) (map[string]any, error)
}

// MapDocumentHandler 处理 map 类型的 YAML 文档
type MapDocumentHandler struct{}

// Handle 处理 map 类型的文档
func (h *MapDocumentHandler) Handle(doc any) (map[string]any, error) {
	if docMap, ok := doc.(map[string]any); ok {
		return docMap, nil
	}
	return nil, fmt.Errorf("document is not a map, got %T", doc)
}

// ListDocumentHandler 处理列表类型的 YAML 文档
// 将列表转换为 map 的原因：
//  1. 统一访问接口：所有 YAML 文件（entities.yaml, managers.yaml 等）都是 map 格式
//     通过 yamlData["key"] 访问，保持一致性
//  2. 代码复用：避免为 headfile.yaml 写特殊的遍历逻辑
//  3. 语义匹配：headfile.yaml 的列表项本质是键值对（如 - DataStruct:, - Enums:）
//     转换为 map 更符合语义，便于通过 key 直接访问
type ListDocumentHandler struct {
	converter *ListToMapConverter
}

// NewListDocumentHandler 创建列表文档处理器
func NewListDocumentHandler() *ListDocumentHandler {
	return &ListDocumentHandler{
		converter: NewListToMapConverter(),
	}
}

// Handle 处理列表类型的文档
// 将列表转换为 map，以便统一访问接口
func (h *ListDocumentHandler) Handle(doc any) (map[string]any, error) {
	if docList, ok := doc.([]any); ok {
		return h.converter.Convert(docList), nil
	}
	return nil, fmt.Errorf("document is not a list, got %T", doc)
}

// ListToMapConverter 列表到 map 的转换器
type ListToMapConverter struct{}

// NewListToMapConverter 创建列表转换器
func NewListToMapConverter() *ListToMapConverter {
	return &ListToMapConverter{}
}

// Convert 将 YAML 列表转换为 map
// 列表项格式: - Key: value 或 - Key: {...}
// 每个列表项应该是一个 map，将其内容合并到结果 map 中
//
// 转换的原因：
// headfile.yaml 使用列表格式：
//   - DataStruct:
//   - Coord: ...
//   - Enums:
//   - ServiceZoneType: ...
//
// 转换为 map 后：
//
//	{
//	  "DataStruct": [...],
//	  "Enums": [...]
//	}
//
// 这样可以通过 yamlData["DataStruct"] 直接访问，与其他 YAML 文件保持一致
func (c *ListToMapConverter) Convert(list []any) map[string]any {
	result := make(map[string]any)
	for _, item := range list {
		if itemMap, ok := item.(map[string]any); ok {
			// 列表项是 map，将其内容合并到结果中
			for k, v := range itemMap {
				result[k] = v
			}
		}
	}
	return result
}

// YAMLDocumentProcessor YAML 文档处理器
// 负责处理不同类型的 YAML 文档（map 或 list）
type YAMLDocumentProcessor struct {
	handlers map[string]YAMLDocumentHandler
}

// NewYAMLDocumentProcessor 创建 YAML 文档处理器
func NewYAMLDocumentProcessor() *YAMLDocumentProcessor {
	return &YAMLDocumentProcessor{
		handlers: map[string]YAMLDocumentHandler{
			"map":  &MapDocumentHandler{},
			"list": NewListDocumentHandler(),
		},
	}
}

// ProcessDocument 处理单个 YAML 文档
// 根据文档类型选择合适的处理器
func (p *YAMLDocumentProcessor) ProcessDocument(doc any) (map[string]any, error) {
	// 处理 nil 文档（空文档，用 --- 分隔）
	if doc == nil {
		return make(map[string]any), nil
	}

	switch v := doc.(type) {
	case map[string]any:
		return p.handlers["map"].Handle(doc)
	case []any:
		return p.handlers["list"].Handle(doc)
	default:
		return nil, fmt.Errorf("unsupported YAML document type: %T", v)
	}
}

// ReadMultiDocumentYAML 读取多文档 YAML 文件
// 支持用 --- 分隔的多个文档，每个文档可以是 map 或 list
// 返回合并后的 map
func ReadMultiDocumentYAML(reader io.Reader) (map[string]any, error) {
	decoder := yaml.NewDecoder(reader)
	processor := NewYAMLDocumentProcessor()
	result := make(map[string]any)

	for {
		// 解析为 any 类型，以支持 map 和 list
		var doc any
		if err := decoder.Decode(&doc); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to decode YAML document: %w", err)
		}

		// 处理文档
		docMap, err := processor.ProcessDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to process YAML document: %w", err)
		}

		// 合并文档内容
		for k, v := range docMap {
			result[k] = v
		}
	}

	return result, nil
}
