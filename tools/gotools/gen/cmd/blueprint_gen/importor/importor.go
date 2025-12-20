package importor

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"sort"
	"strings"
)

// ImportInfo 保存单个 import 的信息
type ImportInfo struct {
	Path  string // 导入路径
	Alias string // 别名（如果有）
}

// ImportGroup 按类型分组的 imports
type ImportGroup struct {
	Standard      []ImportInfo // 标准库
	ThirdParty    []ImportInfo // 第三方库（github.com, gitee.com 等）
	Local         []ImportInfo // 本地项目包
	ProtoExternal []ImportInfo // 外部 proto 包（如 google.golang.org/protobuf/proto）
	ProtoLocal    []ImportInfo // 本地项目 proto 包
}

// SortStrategy 排序策略
type SortStrategy int

const (
	// SortByPath 按路径排序（默认）
	SortByPath SortStrategy = iota
	// SortByAlias 按别名优先排序
	SortByAlias

	ProtoPathPrefix string = "/pkg/proto"
)

// ImportManager import 管理器
type ImportManager struct {
	existingImports map[string]string // path -> alias
	protoPackages   map[string]bool   // proto 包名集合
	projectPrefix   string            // 项目前缀，用于识别本地包
	protoPathPrefix string            // proto 包路径前缀
	sortStrategy    SortStrategy      // 排序策略
}

// NewImportManager 创建 import 管理器
// protoPathPrefix: 项目proto包路径前缀
func NewImportManager(projectPrefix, protoPathPrefix string) *ImportManager {
	return &ImportManager{
		existingImports: make(map[string]string),
		protoPackages:   make(map[string]bool),
		projectPrefix:   projectPrefix,
		sortStrategy:    SortByPath, // 默认按路径排序
		protoPathPrefix: protoPathPrefix,
	}
}

func (im *ImportManager) GenerateProtoPath(packageName string) string {
	var prefix string
	if im.protoPathPrefix != "" {
		prefix = im.protoPathPrefix
	} else {
		prefix = ProtoPathPrefix
	}
	return fmt.Sprintf("%s%s/%s", im.projectPrefix, prefix, packageName)
}

// SetSortStrategy 设置排序策略
func (im *ImportManager) SetSortStrategy(strategy SortStrategy) {
	im.sortStrategy = strategy
}

// ExtractFromFile 从 Go 文件中提取现有 imports
func (im *ImportManager) ExtractFromFile(filePath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil // 文件不存在，不是错误
	}

	// 解析文件
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
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

		im.existingImports[path] = alias
	}

	return nil
}

// AddProtoPackage: 特殊处理 项目的proto包路径
func (im *ImportManager) AddProtoPackage(packageName string) {
	im.protoPackages[packageName] = true
}

// AddProtoPackages 批量添加 proto 包
func (im *ImportManager) AddProtoPackages(packages []string) {
	for _, pkg := range packages {
		im.protoPackages[pkg] = true
	}
}

// EnsureImport 确保某个 import 存在
func (im *ImportManager) EnsureImport(path, alias string) {
	if _, exists := im.existingImports[path]; !exists {
		im.existingImports[path] = alias
	}
}

// Generate 生成 import 代码块
func (im *ImportManager) Generate() string {
	group := im.groupImports()
	return im.formatImportGroup(group)
}

// groupImports 对 imports 进行分组
func (im *ImportManager) groupImports() *ImportGroup {
	group := &ImportGroup{
		Standard:      make([]ImportInfo, 0),
		ThirdParty:    make([]ImportInfo, 0),
		Local:         make([]ImportInfo, 0),
		ProtoExternal: make([]ImportInfo, 0),
		ProtoLocal:    make([]ImportInfo, 0),
	}

	// 从现有 imports 中分类
	for path, alias := range im.existingImports {
		info := ImportInfo{Path: path, Alias: alias}

		// Proto 相关的包特殊处理
		if strings.Contains(path, ProtoPathPrefix) {
			// 跳过项目内的 proto 包（将通过 protoPackages 重新生成）
			continue
		} else if strings.Contains(path, "google.golang.org/protobuf") ||
			strings.Contains(path, "github.com/golang/protobuf") {
			// 外部 proto 库
			group.ProtoExternal = append(group.ProtoExternal, info)
		} else if im.isStandardLib(path) {
			group.Standard = append(group.Standard, info)
		} else if strings.HasPrefix(path, im.projectPrefix) {
			// 本地项目包（不包括 proto）
			group.Local = append(group.Local, info)
		} else {
			// 第三方库
			group.ThirdParty = append(group.ThirdParty, info)
		}
	}

	// 添加 proto 包导入
	protoList := make([]string, 0, len(im.protoPackages))
	for pkg := range im.protoPackages {
		protoList = append(protoList, pkg)
	}
	sort.Strings(protoList)

	for _, pkg := range protoList {
		protoPath := im.GenerateProtoPath(pkg)
		group.ProtoLocal = append(group.ProtoLocal, ImportInfo{Path: protoPath})
	}

	// 检查是否已经有 google.golang.org/protobuf/proto
	hasProtoLib := false
	for _, info := range group.ProtoExternal {
		if info.Path == "google.golang.org/protobuf/proto" {
			hasProtoLib = true
			break
		}
	}

	// 如果有 proto 包但没有 protobuf 库，自动添加
	if !hasProtoLib && len(im.protoPackages) > 0 {
		group.ProtoExternal = append(group.ProtoExternal, ImportInfo{Path: "google.golang.org/protobuf/proto"})
	}

	// 根据排序策略自动排序所有分组
	sortFunc := im.sortImportInfos
	if im.sortStrategy == SortByAlias {
		sortFunc = im.sortByAlias
	}

	sortFunc(group.Standard)
	sortFunc(group.ThirdParty)
	sortFunc(group.Local)
	sortFunc(group.ProtoExternal)
	sortFunc(group.ProtoLocal)

	return group
}

// formatImportGroup 格式化 import 分组
func (im *ImportManager) formatImportGroup(group *ImportGroup) string {
	var result strings.Builder

	groups := [][]ImportInfo{
		group.Standard,
		group.ThirdParty,
		group.Local,
		group.ProtoExternal,
		group.ProtoLocal,
	}

	needsBlankLine := false
	for _, imports := range groups {
		if len(imports) == 0 {
			continue
		}

		if needsBlankLine {
			result.WriteString("\n")
		}

		for _, info := range imports {
			if info.Alias != "" {
				result.WriteString(fmt.Sprintf("\t%s \"%s\"\n", info.Alias, info.Path))
			} else {
				result.WriteString(fmt.Sprintf("\t\"%s\"\n", info.Path))
			}
		}

		needsBlankLine = true
	}

	return result.String()
}

// isStandardLib 判断是否是标准库
func (im *ImportManager) isStandardLib(path string) bool {
	// 标准库不包含域名
	return !strings.Contains(path, ".")
}

// sortImportInfos 对 ImportInfo 列表排序
// 排序规则：
// 1. 没有别名的包在前，有别名的包在后
// 2. 同类包按路径字母顺序排序
// 3. 有别名的包，如果别名相同则按路径排序
func (im *ImportManager) sortImportInfos(infos []ImportInfo) {
	sort.Slice(infos, func(i, j int) bool {
		// 如果一个有别名一个没有，没有别名的在前
		if (infos[i].Alias == "") != (infos[j].Alias == "") {
			return infos[i].Alias == ""
		}

		// 都有别名或都没有别名，按路径排序
		return infos[i].Path < infos[j].Path
	})
}

// SortByAlias 按别名优先排序
// 用于需要按别名排序的场景
func (im *ImportManager) sortByAlias(infos []ImportInfo) {
	sort.Slice(infos, func(i, j int) bool {
		// 如果都有别名，按别名排序
		if infos[i].Alias != "" && infos[j].Alias != "" {
			if infos[i].Alias != infos[j].Alias {
				return infos[i].Alias < infos[j].Alias
			}
			return infos[i].Path < infos[j].Path
		}

		// 如果一个有别名一个没有，没有别名的在前
		if (infos[i].Alias == "") != (infos[j].Alias == "") {
			return infos[i].Alias == ""
		}

		// 都没有别名，按路径排序
		return infos[i].Path < infos[j].Path
	})
}

// GetExistingImports 获取现有的 imports（用于调试）
func (im *ImportManager) GetExistingImports() map[string]string {
	return im.existingImports
}

// HasImport 检查是否已存在某个 import
func (im *ImportManager) HasImport(path string) bool {
	_, exists := im.existingImports[path]
	return exists
}

// GetImportAlias 获取 import 的别名
func (im *ImportManager) GetImportAlias(path string) (string, bool) {
	alias, exists := im.existingImports[path]
	return alias, exists
}

// RemoveImport 移除指定的 import
func (im *ImportManager) RemoveImport(path string) {
	delete(im.existingImports, path)
}

// ClearProtoPackages 清空所有 proto 包
func (im *ImportManager) ClearProtoPackages() {
	im.protoPackages = make(map[string]bool)
}

// GetProtoPackages 获取所有 proto 包名列表
func (im *ImportManager) GetProtoPackages() []string {
	packages := make([]string, 0, len(im.protoPackages))
	for pkg := range im.protoPackages {
		packages = append(packages, pkg)
	}
	sort.Strings(packages)
	return packages
}

// GetAllImports 获取所有 import 信息（按分组返回）
func (im *ImportManager) GetAllImports() *ImportGroup {
	return im.groupImports()
}
