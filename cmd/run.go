//go:build linux

package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/repo-scm/git/config"
	"github.com/repo-scm/git/utils"
)

const (
	runWelcome = `
🚀 Welcome to the Git Workspace!
📁 Current directory: %s
💡 Custom PS1 is active
👋 Type exit when done
`

	runBye = `
👋 Thanks for using Git Workspace!
🏁 Done!
`

	runPS1 = `\[\033[0;32m\]git@repo-scm ➜ \[\033[01;34m\]%s \[\033[00m\]\$ `
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

	runCmd.SetUsageFunc(func(cmd *cobra.Command) error {
		_, _ = fmt.Fprintf(cmd.OutOrStderr(), "Usage:\n")
		_, _ = fmt.Fprintf(cmd.OutOrStderr(), "  %s %s <workspace_name> [flags]\n\n", cmd.Root().Name(), cmd.Name())
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
		_, _ = fmt.Fprintf(cmd.OutOrStderr(), "  git run your_workspace\n")
		return nil
	})
}

func runRun(_ context.Context, cfg *config.Config, name string) error {
	content := fmt.Sprintf(`export PS1="%s"`, fmt.Sprintf(runPS1, name))
	if err := appendContentToBashrc(content); err != nil {
		return err
	}

	defer func(content string) {
		_ = removeContentFromBashrc(content)
	}(content)

	mount := path.Join(utils.ExpandTilde(cfg.Overlay.Mount), name)
	script := fmt.Sprintf(`
echo '%s'
exec bash
`, fmt.Sprintf(runWelcome, mount))

	cmd := exec.Command("/bin/bash", "-c", script)
	cmd.Dir = mount
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	fmt.Print(runBye)

	return nil
}

func appendContentToBashrc(content string) error {
	file, err := os.OpenFile(os.ExpandEnv("$HOME/.bashrc"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, utils.PermFile)
	if err != nil {
		return errors.Wrap(err, "failed to open bashrc\n")
	}

	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	if _, err := file.WriteString(content); err != nil {
		return errors.Wrap(err, "failed to write bashrc\n")
	}

	return nil
}

func removeContentFromBashrc(content string) error {
	file := os.ExpandEnv("$HOME/.bashrc")

	old, err := os.ReadFile(file)
	if err != nil {
		return errors.Wrap(err, "failed to read bashrc\n")
	}

	_new := strings.ReplaceAll(string(old), content, "")
	if string(old) != _new {
		if err = os.WriteFile(file, []byte(_new), utils.PermFile); err != nil {
			return errors.Wrap(err, "failed to write bashrc\n")
		}
	}

	return nil
}
