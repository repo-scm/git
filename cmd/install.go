//go:build linux

package cmd

import (
	"fmt"
	"os"

	"github.com/repo-scm/git/embedded"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install toolchains",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Installing embedded fuse-overlayfs...\n")
		if err := embedded.InstallFuseOverlayfs(); err != nil {
			if os.IsPermission(err) {
				return fmt.Errorf("permission denied: try running with sudo")
			}
			return err
		}
		fmt.Println("fuse-overlayfs installed successfully!")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
