package cmd

import "github.com/spf13/cobra"

func InitCmd(rootCmd *cobra.Command) {
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
