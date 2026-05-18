package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var ejectCmd = &cobra.Command{
	Use:   "eject",
	Short: "Remove dpod.lock, transferring file ownership back to the project",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runEject("dpod.lock", os.Stdin, os.Stdout)
	},
}

func runEject(lockPath string, in io.Reader, out io.Writer) error {
	if _, err := os.Stat(lockPath); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("no dpod.lock found — nothing to eject")
	}

	fmt.Fprint(out, "This will remove CLI ownership of all managed files. Files themselves are untouched. Continue? [y/N] ")

	scanner := bufio.NewScanner(in)
	scanner.Scan()
	answer := strings.TrimSpace(scanner.Text())

	if strings.ToLower(answer) != "y" {
		fmt.Fprintln(out, "Aborted.")
		return nil
	}

	if err := os.Remove(lockPath); err != nil {
		return fmt.Errorf("remove dpod.lock: %w", err)
	}
	fmt.Fprintln(out, "Ejected. dpod.lock removed. Files are yours.")
	return nil
}
