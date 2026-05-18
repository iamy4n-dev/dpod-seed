package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var scaffoldCmd = &cobra.Command{
	Use:   "scaffold",
	Short: "Generate directory layout for a new distro or package bundle",
}

var scaffoldDistroCmd = &cobra.Command{
	Use:   "distro <name>",
	Short: "Generate a new distro directory layout with a distro.yaml template",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(os.Stderr, "not implemented")
		os.Exit(1)
		return nil
	},
}

var scaffoldPackageCmd = &cobra.Command{
	Use:   "package <name>",
	Short: "Generate a new package bundle directory layout with a manifest.yaml template",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(os.Stderr, "not implemented")
		os.Exit(1)
		return nil
	},
}

func init() {
	scaffoldCmd.AddCommand(scaffoldDistroCmd, scaffoldPackageCmd)
}
