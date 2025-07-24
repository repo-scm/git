//go:build linux

package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/repo-scm/git/config"
	"github.com/repo-scm/git/utils"
)

const (
	runWelcome = `
🚀 Welcome to the Git Workspace!
📁 Current directory: %s
🧹 Use "git clean /path/to/name" to safely remove directories in overlayfs
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
	Short: "Run workspace",
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		config := GetConfig()

		var name string
		if len(args) == 0 {
			selectedName, err := selectWorkspaceInteractively(ctx, config)
			if err != nil {
				_, _ = fmt.Fprintln(os.Stderr, err.Error())
				os.Exit(1)
			}
			name = selectedName
		} else {
			name = args[0]
		}

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
		_, _ = fmt.Fprintf(cmd.OutOrStderr(), "  %s %s [workspace_name] [flags]\n\n", cmd.Root().Name(), cmd.Name())
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
		_, _ = fmt.Fprintf(cmd.OutOrStderr(), "  git run # Interactive workspace selection\n")
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

func selectWorkspaceInteractively(ctx context.Context, cfg *config.Config) (string, error) {
	// Get all available workspaces
	workspaces, err := QueryWorkspaces(ctx, cfg, false)
	if err != nil {
		return "", fmt.Errorf("failed to querying workspaces: %w", err)
	}

	if len(workspaces) == 0 {
		return "", errors.New("no workspaces found. Please create a workspace first using 'git create'")
	}

	// Extract workspace names for the prompt
	var workspaceNames []string
	for _, workspace := range workspaces {
		if workspace.Name != "" {
			workspaceNames = append(workspaceNames, workspace.Name)
		}
	}

	if len(workspaceNames) == 0 {
		return "", errors.New("no valid workspaces found")
	}

	// Create the interactive prompt
	prompt := promptui.Select{
		Label:        "Select a workspace to run",
		Items:        workspaceNames,
		Size:         10,
		HideSelected: false,
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}:",
			Active:   "▶ {{ . | cyan | bold }}",
			Inactive: "  {{ . | white }}",
			Selected: "✓ Selected workspace: {{ . | green | bold }}",
			Help:     "Use ↑/↓ arrow keys to navigate, Enter to select, Ctrl+C to cancel",
		},
	}

	index, result, err := prompt.Run()
	if err != nil {
		if errors.Is(err, promptui.ErrInterrupt) {
			return "", errors.New("operation cancelled by user")
		}
		return "", fmt.Errorf("failed to select workspace: %w", err)
	}

	_ = index

	return result, nil
}

func appendContentToBashrc(content string) error {
	file, err := os.OpenFile(os.ExpandEnv("$HOME/.bashrc"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, utils.PermFile)
	if err != nil {
		return fmt.Errorf("failed to open bashrc file: %w", err)
	}

	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	if _, err := file.WriteString(content); err != nil {
		return fmt.Errorf("failed to write bashrc file: %w", err)
	}

	return nil
}

func removeContentFromBashrc(content string) error {
	file := os.ExpandEnv("$HOME/.bashrc")

	old, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read bashrc file: %w", err)
	}

	_new := strings.ReplaceAll(string(old), content, "")
	if string(old) != _new {
		if err = os.WriteFile(file, []byte(_new), utils.PermFile); err != nil {
			return fmt.Errorf("failed to write bashrc file: %w", err)
		}
	}

	return nil
}
