package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Re-materialise DevPod config from dpod.yaml, overwriting owned files",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(os.Stderr, "not implemented")
		os.Exit(1)
		return nil
	},
}
