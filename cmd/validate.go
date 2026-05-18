package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Resolve dpod.yaml without writing files — suitable for CI",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(os.Stderr, "not implemented")
		os.Exit(1)
		return nil
	},
}
