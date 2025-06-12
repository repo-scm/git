//go:build linux

package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/repo-scm/git/config"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run workspace for git repo",
	Args:  cobra.RangeArgs(1, 1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		config := GetConfig()
		name := args[0]
		if err := runRun(ctx, config, name); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	},
}

// nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(runCmd)
}

func runRun(ctx context.Context, cfg *config.Config, name string) error {
	return nil
}
