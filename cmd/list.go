//go:build linux

package cmd

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/repo-scm/git/config"
	"github.com/repo-scm/git/utils"
)

var (
	listAll     bool
	verboseMode bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List workspaces for git repo",
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		var name string
		ctx := context.Background()
		config := GetConfig()
		if len(args) == 0 && !listAll {
			_, _ = fmt.Fprintln(os.Stderr, "Please specify a workspace name")
			os.Exit(1)
		}
		if len(args) == 1 {
			name = args[0]
		}
		if err := runList(ctx, config, name); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	},
}

type Workspace struct {
	Name       string
	Mount      string
	Filesystem string
}

// nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.PersistentFlags().BoolVarP(&listAll, "all", "a", false, "list all workspaces")
	listCmd.PersistentFlags().BoolVarP(&verboseMode, "verbose", "v", false, "list in verbose mode")

	listCmd.SetUsageFunc(func(cmd *cobra.Command) error {
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
		_, _ = fmt.Fprintf(cmd.OutOrStderr(), "  git list your_workspace\n")
		_, _ = fmt.Fprintf(cmd.OutOrStderr(), "  git list your_workspace --verbose\n")
		_, _ = fmt.Fprintf(cmd.OutOrStderr(), "  git list --all\n")
		_, _ = fmt.Fprintf(cmd.OutOrStderr(), "  git list --all --verbose\n")
		return nil
	})
}

func runList(ctx context.Context, cfg *config.Config, name string) error {
	data := [][]string{
		{"NAME", "MOUNT", "FILESYSTEM"},
	}

	workspaces, err := QueryWorkspaces(ctx, cfg, verboseMode)
	if err != nil {
		return err
	}

	if name != "" {
		for _, item := range workspaces {
			if verboseMode {
				if strings.HasSuffix(path.Base(item.Mount), name) {
					data = append(data, []string{item.Name, item.Mount, item.Filesystem})
				}
			} else {
				if item.Name == name {
					data = append(data, []string{item.Name, item.Mount, item.Filesystem})
				}
			}
		}
		if err := utils.WriteTable(ctx, data); err != nil {
			return err
		}
		return nil
	}

	for _, item := range workspaces {
		data = append(data, []string{item.Name, item.Mount, item.Filesystem})
	}

	if err := utils.WriteTable(ctx, data); err != nil {
		return err
	}

	return nil
}

func QueryWorkspaces(_ context.Context, cfg *config.Config, verbose bool) ([]Workspace, error) {
	var workspaces []Workspace

	overlayPath := utils.ExpandTilde(cfg.Overlay.Mount)

	_ = filepath.WalkDir(overlayPath, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relPath, _ := filepath.Rel(overlayPath, p)
		if depth := strings.Count(relPath, string(filepath.Separator)); depth == 0 {
			if d.IsDir() {
				if p != overlayPath {
					var name string
					if verbose {
						if strings.HasPrefix(relPath, "upper-") || strings.HasPrefix(relPath, "work-") {
							name = ""
						} else {
							name = relPath
						}
						workspaces = append(workspaces, Workspace{name, p, "overlay"})
					} else {
						if !strings.HasPrefix(relPath, "upper-") && !strings.HasPrefix(relPath, "work-") {
							workspaces = append(workspaces, Workspace{relPath, p, "overlay"})
						}
					}
				}
			}
		}
		return nil
	})

	if verbose {
		sshfsPath := utils.ExpandTilde(cfg.Sshfs.Mount)
		_ = filepath.WalkDir(sshfsPath, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			relPath, _ := filepath.Rel(sshfsPath, p)
			if depth := strings.Count(relPath, string(filepath.Separator)); depth == 0 {
				if d.IsDir() {
					if p != sshfsPath {
						workspaces = append(workspaces, Workspace{"", p, "sshfs"})
					}
				}
			}
			return nil
		})
	}

	return workspaces, nil
}
