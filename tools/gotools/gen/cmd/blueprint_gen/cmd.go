package blueprint_gen

import (
	"fmt"

	blueprint_types "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types"
	"github.com/spf13/cobra"
)

var (
	blueprintGenCmd = &cobra.Command{
		Use:   "blueprintgen",
		Short: "Generate proto and Go code from blueprint YAML files",
		Long: `Generate protobuf files and MME Go code from blueprint YAML files.
This tool parses YAML files in the blueprint directory and generates:
- Protocol buffer files in the protocol/ directory
- MME Go structure files in the app/mme/ directory`,
		Run: runBlueprintGen,
	}
)

func runBlueprintGen(cmd *cobra.Command, args []string) {
	blueprintDir, _ := cmd.Flags().GetString("blueprint-dir")
	protoOutput, _ := cmd.Flags().GetString("proto-output")
	goOutput, _ := cmd.Flags().GetString("go-output")
	protocolIDsOutput, _ := cmd.Flags().GetString("protocol-ids-output")
	debug, _ := cmd.Flags().GetBool("debug")

	if debug {
		println("Starting blueprint code generation...")
		println("Blueprint directory:", blueprintDir)
		println("Proto output directory:", protoOutput)
		println("Go output directory:", goOutput)
	}

	// 解析 YAML 文件
	parser := NewParser(blueprintDir)
	if err := parser.Parse(); err != nil {
		cmd.PrintErrln("Failed to parse YAML files:", err)
		return
	}

	data := parser.GetContext()
	if debug {
		fmt.Printf("Parsed %d entities\n", len(data.Entities))
		for _, e := range data.Entities {
			fmt.Printf("  Entity: %s, Fields: %d\n", e.Name, len(e.Fields))
		}
		fmt.Printf("Parsed %d managers\n", len(data.Managers))
		for _, m := range data.Managers {
			fmt.Printf("  Manager: %s, Fields: %d\n", m.Name, len(m.Fields))
			for _, f := range m.Fields {
				isXMap := f.Type.Kind == blueprint_types.FieldKindXMap
				keyType := "unknown"
				valueType := "unknown"
				if f.Type.KeyType != nil {
					if f.Type.KeyType.TypeName != "" {
						keyType = f.Type.KeyType.TypeName
					} else {
						keyType = f.Type.KeyType.String()
					}
				}
				if f.Type.ValueType != nil {
					if f.Type.ValueType.TypeName != "" {
						valueType = f.Type.ValueType.TypeName
					} else {
						valueType = f.Type.ValueType.String()
					}
				}
				fmt.Printf("    Field: %s (xmap=%v, key=%s, value=%s)\n", f.Name, isXMap, keyType, valueType)
			}
		}
		fmt.Printf("Parsed %d modules\n", len(data.Modules))
		for _, m := range data.Modules {
			fmt.Printf("  Module: %s, Mechanisms: %d\n", m.Name, len(m.Fields))
		}
		fmt.Printf("Parsed %d mechanisms\n", len(data.Mechanisms))
		for _, m := range data.Mechanisms {
			fmt.Printf("  Mechanism: %s, Fields: %d\n", m.Name, len(m.Fields))
		}
		fmt.Printf("Parsed %d netwalls\n", len(data.NetWalls))
	}

	// 生成 Proto 文件
	protoGen := NewProtoGenerator(data)
	if err := protoGen.Generate(protoOutput); err != nil {
		cmd.PrintErrln("Failed to generate proto files:", err)
		return
	}

	if debug {
		println("Proto files generated successfully")
	}

	// 生成 protocol_ids.pb.go 文件
	// protocol_ids.pb.go 应该输出到 app/proto/ 目录下，每个 NetWall 包生成一个文件
	// 使用 protocol-ids-output 参数，默认值为 app/proto
	if err := protoGen.GenerateProtocolIDs(protocolIDsOutput); err != nil {
		cmd.PrintErrln("Failed to generate protocol_ids.pb.go:", err)
		return
	}

	if debug {
		println("protocol_ids.pb.go generated successfully")
	}

	// 生成 Go 文件
	goGen := NewGoStructGenerator(data)
	if err := goGen.Generate(goOutput); err != nil {
		cmd.PrintErrln("Failed to generate Go files:", err)
		return
	}

	if debug {
		println("Go files generated successfully")
	}

	// 格式化生成的 Go 文件
	if err := FormatGoFiles(goOutput); err != nil {
		cmd.PrintErrln("Warning: Failed to format Go files:", err)
		// 不中断流程，仅打印警告
	} else if debug {
		println("Go files formatted successfully")
	}

	cmd.Println("Blueprint code generation completed successfully!")
}

// InitCmd 初始化命令
func InitCmd(father *cobra.Command) {
	blueprintGenCmd.Flags().String("blueprint-dir", "../protocol/blueprint", "Directory containing blueprint YAML files")
	blueprintGenCmd.Flags().String("proto-output", "../protocol/protocol", "Output directory for proto files")
	blueprintGenCmd.Flags().String("go-output", "app/mme", "Output directory for Go files")
	blueprintGenCmd.Flags().String("protocol-ids-output", "app/proto/pb", "Output directory for protocol_ids.pb.go file")
	blueprintGenCmd.Flags().Bool("debug", false, "Enable debug mode")

	father.AddCommand(blueprintGenCmd)
}
