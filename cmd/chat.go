//go:build linux

package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/repo-scm/git/config"
)

var (
	modelID   string
	quietMode bool
)

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Chat with workspace for git repo",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		config := GetConfig()
		name := args[0]
		prompt := args[1]
		if name == "" {
			_, _ = fmt.Fprintln(os.Stderr, "Please specify a workspace name")
			os.Exit(1)
		}
		if err := runChat(ctx, config, name, prompt); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	},
}

// nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(chatCmd)

	chatCmd.PersistentFlags().StringVarP(&modelID, "model", "m", "anthropic/claude-opus-4-20250514", "model id")
	chatCmd.PersistentFlags().BoolVarP(&quietMode, "quiet", "q", false, "quiet mode")

	chatCmd.SetUsageFunc(func(cmd *cobra.Command) error {
		_, _ = fmt.Fprintf(cmd.OutOrStderr(), "Usage:\n")
		_, _ = fmt.Fprintf(cmd.OutOrStderr(), "  %s %s <workspace_name> [prompt] [flags]\n\n", cmd.Root().Name(), cmd.Name())
		if cmd.HasLocalFlags() {
			_, _ = fmt.Fprintf(cmd.OutOrStderr(), "Flags:\n")
			cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
				_, _ = fmt.Fprintf(cmd.OutOrStderr(), "  -%s, --%s   %s", flag.Shorthand, flag.Name, flag.Usage)
				if flag.Name != "help" && flag.DefValue != "" {
					_, _ = fmt.Fprintf(cmd.OutOrStderr(), " (default %s)", flag.DefValue)
				}
				_, _ = fmt.Fprintf(cmd.OutOrStderr(), "\n")
			})
		}
		if cmd.HasInheritedFlags() {
			_, _ = fmt.Fprintf(cmd.OutOrStderr(), "\nGlobal Flags:\n")
			cmd.InheritedFlags().VisitAll(func(flag *pflag.Flag) {
				_, _ = fmt.Fprintf(cmd.OutOrStderr(), "  -%s, --%s   %s", flag.Shorthand, flag.Name, flag.Usage)
				if flag.DefValue != "" {
					_, _ = fmt.Fprintf(cmd.OutOrStderr(), " (default %s)", flag.DefValue)
				}
				_, _ = fmt.Fprintf(cmd.OutOrStderr(), "\n")
			})
		}
		_, _ = fmt.Fprintf(cmd.OutOrStderr(), "\nExample:\n")
		_, _ = fmt.Fprintf(cmd.OutOrStderr(), "  git chat your_workspace \"your_prompt\" --model \"anthropic/claude-opus-4-20250514\"\n")
		_, _ = fmt.Fprintf(cmd.OutOrStderr(), "  git chat your_workspace \"your_prompt\" --model \"anthropic/claude-opus-4-20250514\" --quiet\n")
		return nil
	})
}

func runChat(_ context.Context, cfg *config.Config, name, prompt string) error {
	return nil
}
