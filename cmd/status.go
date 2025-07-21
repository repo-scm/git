//go:build linux

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show toolchains status",
	RunE: func(cmd *cobra.Command, args []string) error {
		_ = checkOverlayfs()
		_ = checkSshfs()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func checkOverlayfs() error {
	targetPath := "/usr/local/bin/fuse-overlayfs"

	fmt.Printf("fuse-overlayfs:\n")

	if info, err := os.Stat(targetPath); err == nil {
		fmt.Printf("  Installed: ✓ %s (%d bytes)\n", targetPath, info.Size())
		if info.Mode()&0111 != 0 {
			fmt.Printf("  Executable: ✓\n")
		} else {
			fmt.Printf("  Executable: ✗ (not executable)\n")
		}
	} else {
		fmt.Printf("  Installed: ✗ (not found at %s)\n", targetPath)
		fmt.Printf("  Run 'sudo %s install' to install\n", rootCmd.Use)
	}

	return nil
}

func checkSshfs() error {
	sshfsPath := "/usr/bin/sshfs"

	fmt.Printf("\nsshfs:\n")

	if info, err := os.Stat(sshfsPath); err == nil {
		fmt.Printf("  Installed: ✓ %s (%d bytes)\n", sshfsPath, info.Size())
		if info.Mode()&0111 != 0 {
			fmt.Printf("  Executable: ✓\n")
		} else {
			fmt.Printf("  Executable: ✗ (not executable)\n")
		}
	} else {
		fmt.Printf("  Installed: ✗ (not found at %s)\n", sshfsPath)
		fmt.Printf("  Run 'sudo apt install sshfs' to install\n")
	}

	return nil
}
