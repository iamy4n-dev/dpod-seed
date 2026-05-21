package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:     "dpod-seed",
	Short:   "Manage DevPod environments from verified upstream distros",
	Version: version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(
		initCmd,
		syncCmd,
		listCmd,
		validateCmd,
		ejectCmd,
		scaffoldCmd,
		versionCmd,
	)
}
