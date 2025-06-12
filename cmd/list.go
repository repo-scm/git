//go:build linux

package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/repo-scm/git/config"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List workspaces for git repo",
	Args:  cobra.RangeArgs(1, 1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		config := GetConfig()
		if args[0] != "workspaces" {
			_, _ = fmt.Fprintln(os.Stderr, "invalid argument")
			os.Exit(1)
		}
		if err := runList(ctx, config); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	},
}

// nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(ctx context.Context, cfg *config.Config) error {
	return nil
}
