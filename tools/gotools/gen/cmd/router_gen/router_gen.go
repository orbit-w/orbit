package routergen

import "github.com/spf13/cobra"

var (
	genRouterCmd = &cobra.Command{
		Use:   "protocolgen",
		Short: "Generate router code from proto files",
		Run:   runRouterGluegen,
	}
)

func runRouterGluegen(cmd *cobra.Command, args []string) {

}

func InitCmd(father *cobra.Command) {
	genRouterCmd.Flags().String("proto-dir", "app/proto/pb", "Directory containing proto files")
	genRouterCmd.Flags().String("output-dir", "app/proto/pb", "Output directory for generated files")
	genRouterCmd.Flags().Bool("debug", false, "Enable debug mode")
	genRouterCmd.Flags().Bool("quiet", false, "Enable quiet mode")

	father.AddCommand(genRouterCmd)
}
