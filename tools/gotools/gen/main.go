package main

import (
	"fmt"
	"os"

	cmd "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/gostructs_gen"
	"gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/routergen"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "gen",
	Short: "Orbit code generation tools",
	Long: `A suite of tools for generating code in the Orbit project.
This includes generating extension methods for protocol buffers (gostructs_gen) and router dispatch code (routergen).`,
}

func init() {
	cmd.InitCmd(RootCmd)
	routergen.InitCmd(RootCmd)
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
