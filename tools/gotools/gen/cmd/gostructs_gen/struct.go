package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/gostruct"

	"github.com/spf13/cobra"
)

var structCmd = &cobra.Command{
	Use:   "struct",
	Short: "Generate Go structs from proto messages",
	Long: `Parse proto files and generate corresponding Go structs with identical field information.
This command analyzes proto files and creates Go struct definitions that mirror the proto message structure.

Example usage:
  protogen struct -p pt -o generated --package structs`,
	Run: func(cmd *cobra.Command, args []string) {
		protoDir, _ := cmd.Flags().GetString("proto-dir")
		outputDir, _ := cmd.Flags().GetString("output-dir")
		protobufInclude, _ := cmd.Flags().GetString("protobuf-include")
		packageName, _ := cmd.Flags().GetString("package")
		emitForwarders, _ := cmd.Flags().GetBool("emit-entity-ops-forwarders")

		gostruct.SetGenerationOptions(gostruct.GenerationOptions{EmitEntityOpsForwarders: emitForwarders})

		if err := generateGoStructsFromProto(protoDir, outputDir, protobufInclude, packageName); err != nil {
			fmt.Printf("Error generating Go structs: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Successfully generated Go structs!")
	},
}

func init() {
	structCmd.Flags().String("package", "structs", "Go package name for generated structs")
	structCmd.Flags().Bool("emit-entity-ops-forwarders", false, "Emit Entity-level Ops forwarder methods (default: false)")
}

// generateGoStructsFromProto 从proto文件生成Go结构体
func generateGoStructsFromProto(protoDir, outputDir, protobufInclude, packageName string) error {
	// 创建上下文
	ctx := NewContext(protoDir, outputDir, protobufInclude)
	defer ctx.Clear()

	// 解析 proto 文件
	if err := parseProtoFiles(ctx); err != nil {
		return fmt.Errorf("failed to parse proto files: %v", err)
	}

	// 解析为Go结构体信息
	// 在生成前校验角色规则（Entity/Component/Data）
	if err := gostruct.ValidateMessageRoles(ctx); err != nil {
		return fmt.Errorf("role validation failed: %v", err)
	}

	structs, err := ParseProtoToGoStructs(ctx, packageName)
	if err != nil {
		return fmt.Errorf("failed to parse proto to Go structs: %v", err)
	}

	if len(structs) == 0 {
		fmt.Println("No proto messages found to generate structs from")
		return nil
	}

	// 将结构体按来源 proto 文件分组
	groups := gostruct.GroupStructsBySourceFile(structs)

	// 确保输出目录存在
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	total := 0
	for src, list := range groups {
		// 构建文件名：player.proto -> player.go
		base := filepath.Base(src)
		name := base
		if ext := filepath.Ext(base); ext != "" {
			name = base[:len(base)-len(ext)]
		}
		outputFile := filepath.Join(outputDir, name+".go")

		// 为每个分组单独生成代码与写入
		code := gostruct.GenerateGoStructCode(list)
		if err := writeToFile(outputFile, code); err != nil {
			return fmt.Errorf("failed to write output file %s: %v", outputFile, err)
		}
		fmt.Printf("Generated %d structs in %s (from %s)\n", len(list), outputFile, src)
		total += len(list)
	}

	fmt.Printf("Generated %d Go structs across %d files\n", total, len(groups))

	return nil
}
