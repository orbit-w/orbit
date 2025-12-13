package main

import (
	"fmt"
	"os"

	blueprint_gen "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen"
	router_gen "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/router_gen"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "gen",
	Short: "Orbit code generation tools",
	Long: `A suite of tools for generating code in the Orbit project.
This includes generating extension methods for protocol buffers (gostructs_gen) and router dispatch code (routergen).`,
}

func init() {
	router_gen.InitCmd(RootCmd)
	blueprint_gen.InitCmd(RootCmd)
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
