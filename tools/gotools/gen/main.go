package main

import (
	"fmt"
	"os"

	cmd "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/gostructs_gen"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "protogen",
	Short: "Protocol buffer extension generator",
	Long: `A tool to automatically generate extension methods for protocol buffer messages.
This tool analyzes proto3 files and generates helper methods for DeltaSyncMap and other containers.`,
}

func init() {
	cmd.InitCmd(RootCmd)
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	Execute()
}
