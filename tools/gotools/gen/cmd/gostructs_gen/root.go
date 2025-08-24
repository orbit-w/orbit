package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "protogen",
	Short: "Protocol buffer extension generator",
	Long: `A tool to automatically generate extension methods for protocol buffer messages.
This tool analyzes proto3 files and generates helper methods for DeltaSyncMap and other containers.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
