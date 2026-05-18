package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var ejectCmd = &cobra.Command{
	Use:   "eject",
	Short: "Remove dpod.lock, transferring file ownership back to the project",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(os.Stderr, "not implemented")
		os.Exit(1)
		return nil
	},
}
