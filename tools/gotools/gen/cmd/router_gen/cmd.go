package router_gen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	blueprintDir   string
	controllerDir  string
	outputFile     string
	controllerPath string
)

// InitCmd 初始化命令
func InitCmd(rootCmd *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "routergen",
		Short: "Generate router handler functions from NetWall YAML and Controller files",
		Long: `Generate router handler functions that:
- Parse Request messages
- Load EntityRef entities
- Call Controller Handle methods

The generator reads NetWall YAML files to identify Request definitions and EntityRef fields,
then generates router handler functions that automatically handle message parsing, entity loading,
and method invocation.`,
		RunE: runRouterGen,
	}

	cmd.Flags().StringVar(&blueprintDir, "blueprint-dir", "protocol/blueprint/netwall", "Directory containing NetWall YAML files")
	cmd.Flags().StringVar(&controllerDir, "controller-dir", "orbit/app/controller_v2", "Directory containing Controller files")
	cmd.Flags().StringVar(&outputFile, "output", "orbit/app/routers/routers.go", "Output file path for generated router code")
	cmd.Flags().StringVar(&controllerPath, "controller-path", "", "Controller file path (default: inferred from controller-dir)")

	rootCmd.AddCommand(cmd)
}

// runRouterGen 运行代码生成器
func runRouterGen(cmd *cobra.Command, args []string) error {
	// 解析所有 NetWall YAML 文件
	requestInfos, err := parseAllNetWallFiles(blueprintDir)
	if err != nil {
		return fmt.Errorf("failed to parse NetWall files: %w", err)
	}

	if len(requestInfos) == 0 {
		return fmt.Errorf("no requests found in NetWall files")
	}

	// 解析 Controller 文件
	controllerInfo, err := discoverAndParseController(controllerDir)
	if err != nil {
		return fmt.Errorf("failed to parse controller: %w", err)
	}

	// 确定 Controller 文件路径
	if controllerPath == "" {
		// 从 controllerDir 推断
		controllerPath = filepath.Join(controllerDir, "controller.go")
	}

	// 创建生成上下文
	ctx := &RouterGenContext{
		Requests:       requestInfos,
		Controller:     controllerInfo,
		OutputPath:     outputFile,
		ControllerPath: controllerPath,
	}

	// 生成代码
	if err := generateRouterCode(ctx); err != nil {
		return fmt.Errorf("failed to generate router code: %w", err)
	}

	// 生成 Controller Handle 方法
	if err := GenerateControllerMethods(ctx); err != nil {
		return fmt.Errorf("failed to generate controller methods: %w", err)
	}

	fmt.Printf("Successfully generated router code: %s\n", outputFile)
	fmt.Printf("Generated %d request handlers\n", len(requestInfos))

	return nil
}

// parseAllNetWallFiles 解析所有 NetWall YAML 文件
func parseAllNetWallFiles(blueprintDir string) ([]*RequestInfo, error) {
	var allRequests []*RequestInfo

	// 检查目录是否存在
	if _, err := os.Stat(blueprintDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("blueprint directory does not exist: %s", blueprintDir)
	}

	// 读取所有 YAML 文件
	entries, err := os.ReadDir(blueprintDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read blueprint directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		yamlPath := filepath.Join(blueprintDir, entry.Name())
		requests, err := parseNetWallYAML(yamlPath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", entry.Name(), err)
		}

		allRequests = append(allRequests, requests...)
	}

	return allRequests, nil
}

// DiscoverAndParseController 发现并解析 Controller 文件（导出函数）
func DiscoverAndParseController(controllerDir string) (*ControllerInfo, error) {
	return discoverAndParseController(controllerDir)
}

// discoverAndParseController 发现并解析 Controller 文件（内部实现）
func discoverAndParseController(controllerDir string) (*ControllerInfo, error) {
	// 检查目录是否存在
	if _, err := os.Stat(controllerDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("controller directory does not exist: %s", controllerDir)
	}

	// 查找 Controller 文件
	files, err := discoverControllerFiles(controllerDir)
	if err != nil {
		return nil, fmt.Errorf("failed to discover controller files: %w", err)
	}

	// 尝试解析每个文件，找到包含 Controller 定义的文件
	for _, file := range files {
		controllerInfo, err := parseControllerFile(file)
		if err != nil {
			// 如果解析失败，继续尝试下一个文件
			continue
		}

		// 如果找到了 Controller 信息，返回
		if controllerInfo != nil {
			return controllerInfo, nil
		}
	}

	return nil, fmt.Errorf("controller info not found in any file in %s", controllerDir)
}

