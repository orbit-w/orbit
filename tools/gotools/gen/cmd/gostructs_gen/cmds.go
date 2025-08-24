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

var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate extension methods and utilities for protocol buffers",
	Long: `Generate extension methods for DeltaSyncMap and DeltaSyncList containers.
This command analyzes proto files and creates type-safe helper methods.
Optionally generates serialization utilities for skip_serialization fields.
Can also generate plain Go structs mirroring proto messages when --go-structs is set.`,
	Run: func(cmd *cobra.Command, args []string) {
		protoDir, _ := cmd.Flags().GetString("proto-dir")
		outputDir, _ := cmd.Flags().GetString("output-dir")
		protobufInclude, _ := cmd.Flags().GetString("protobuf-include")
		includeSerialize, _ := cmd.Flags().GetBool("serialize")
		genGoStructs, _ := cmd.Flags().GetBool("go-structs")
		structsPkg, _ := cmd.Flags().GetString("structs-package")

		if err := generateAll(protoDir, outputDir, protobufInclude, includeSerialize, genGoStructs, structsPkg); err != nil {
			fmt.Printf("Error generating code: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Successfully generated all code!")
	},
}
