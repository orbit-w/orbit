package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// goStructsCmd 生成 Go 结构体命令
var goStructsCmd = &cobra.Command{
	Use:   "go_structs",
	Short: "Generate Go structs from proto messages",
	Long: `Generate Go structs that mirror proto message definitions.
This command parses proto files and creates corresponding Go struct definitions
with proper field types, tags, and documentation.

Example usage:
  protogen go_structs -p pt -o structs --package structs
  protogen go_structs --proto-dir=pt --output-dir=generated --package=models`,
	Run: func(cmd *cobra.Command, args []string) {
		protoDir, _ := cmd.Flags().GetString("proto-dir")
		outputDir, _ := cmd.Flags().GetString("output-dir")
		protobufInclude, _ := cmd.Flags().GetString("protobuf-include")
		packageName, _ := cmd.Flags().GetString("package")

		if err := generateGoStructsFromProto(protoDir, outputDir, protobufInclude, packageName); err != nil {
			fmt.Printf("Error generating Go structs: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Successfully generated Go structs!")
	},
}

func init() {
	// 这里可以添加全局标志
	rootCmd.PersistentFlags().StringP("proto-dir", "p", "pt", "Directory containing proto files")
	rootCmd.PersistentFlags().StringP("output-dir", "o", "pb", "Directory for generated Go files")
	rootCmd.PersistentFlags().String("protobuf-include", "", "Path to protobuf include directory (auto-detect if empty)")

	// 为 gen 命令添加 serialize 标志
	genCmd.Flags().BoolP("serialize", "s", false, "Generate serialization utilities for skip_serialization fields")
	// 为 gen 命令添加生成 Go structs 的标志
	genCmd.Flags().Bool("go-structs", false, "Also generate Go structs that mirror proto messages")
	genCmd.Flags().String("structs-package", "structs", "Go package name for generated structs (used with --go-structs)")

	// 为 go_structs 命令添加 package 标志
	goStructsCmd.Flags().String("package", "structs", "Go package name for generated structs")

	rootCmd.AddCommand(genCmd)
	rootCmd.AddCommand(structCmd)
	rootCmd.AddCommand(goStructsCmd)
}
