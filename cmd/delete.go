//go:build linux

package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"syscall"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/repo-scm/git/config"
	"github.com/repo-scm/git/utils"
)

var (
	deleteAll bool
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete workspace",
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		var name string
		ctx := context.Background()
		config := GetConfig()
		if len(args) == 0 && !deleteAll {
			_, _ = fmt.Fprintln(os.Stderr, "Please specify a workspace name")
			os.Exit(1)
		}
		if len(args) == 1 {
			name = args[0]
		}
		if err := runDelete(ctx, config, name); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	},
}

// nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(deleteCmd)

	deleteCmd.PersistentFlags().BoolVarP(&deleteAll, "all", "a", false, "delete all workspaces")

	deleteCmd.SetUsageFunc(func(cmd *cobra.Command) error {
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
		_, _ = fmt.Fprintf(cmd.OutOrStderr(), "  git delete your_workspace\n")
		_, _ = fmt.Fprintf(cmd.OutOrStderr(), "  git delete --all\n")
		return nil
	})
}

func runDelete(ctx context.Context, cfg *config.Config, name string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		cancel()
	}()

	if name != "" {
		sshfsPath := path.Join(utils.ExpandTilde(cfg.Sshfs.Mount), name)
		overlayPath := path.Join(utils.ExpandTilde(cfg.Overlay.Mount), name)
		if err := UnmountOverlay(ctx, overlayPath); err != nil {
			if ctx.Err() != nil {
				fmt.Println("Operation cancelled")
				return ctx.Err()
			}
			fmt.Println(err.Error())
		}
		if err := UnmountSshfs(ctx, sshfsPath); err != nil {
			if ctx.Err() != nil {
				fmt.Println("Operation cancelled")
				return ctx.Err()
			}
			fmt.Println(err.Error())
		}
		return nil
	}

	workspaces, err := QueryWorkspaces(ctx, cfg, false)
	if err != nil {
		return err
	}

	for _, item := range workspaces {
		select {
		case <-ctx.Done():
			fmt.Println("Operation cancelled")
			return ctx.Err()
		default:
		}

		sshfsPath := path.Join(utils.ExpandTilde(cfg.Sshfs.Mount), item.Name)
		overlayPath := path.Join(utils.ExpandTilde(cfg.Overlay.Mount), item.Name)
		if err := UnmountOverlay(ctx, overlayPath); err != nil {
			if ctx.Err() != nil {
				fmt.Println("Operation cancelled")
				return ctx.Err()
			}
			fmt.Println(err.Error())
		}
		if err := UnmountSshfs(ctx, sshfsPath); err != nil {
			if ctx.Err() != nil {
				fmt.Println("Operation cancelled")
				return ctx.Err()
			}
			fmt.Println(err.Error())
		}
	}

	return nil
}

func UnmountOverlay(ctx context.Context, mount string) error {
	if mount == "" {
		return errors.New("mount are required\n")
	}

	mountDir := path.Dir(path.Clean(mount))
	mountName := path.Base(path.Clean(mount))

	upperPath := path.Join(mountDir, "upper-"+mountName)
	workPath := path.Join(mountDir, "work-"+mountName)

	defer func(mount, workPath, upperPath string) {
		_ = os.RemoveAll(mount)
		_ = os.RemoveAll(workPath)
		_ = os.RemoveAll(upperPath)
	}(mount, workPath, upperPath)

	cmd := exec.CommandContext(ctx, "fusermount", "-u", path.Clean(mount))

	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return errors.Wrap(err, "failed to unmount overlay\n")
	}

	fmt.Printf("successfully unmounted overlay\n")

	return nil
}

func UnmountSshfs(ctx context.Context, mount string) error {
	if mount == "" {
		return errors.New("mount is required\n")
	}

	defer func(path string) {
		_ = os.RemoveAll(path)
	}(mount)

	cmd := exec.CommandContext(ctx, "fusermount", "-u", path.Clean(mount))

	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		fmt.Println(err.Error())
		return nil
	}

	fmt.Printf("successfully unmounted sshfs\n")

	return nil
}
