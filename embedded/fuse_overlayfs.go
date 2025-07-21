//go:build linux

package embedded

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/repo-scm/git/utils"
)

const (
	targetPathFuseOverlayFs = "/usr/local/bin/fuse-overlayfs"
)

//go:embed fuse-overlayfs
var fuseOverlayfsData []byte

func InstallFuseOverlayfs() error {
	if err := os.MkdirAll("/usr/local/bin", utils.PermDir); err != nil {
		return fmt.Errorf("failed to create /usr/local/bin directory: %v", err)
	}

	if err := os.WriteFile(targetPathFuseOverlayFs, fuseOverlayfsData, utils.PermDir); err != nil {
		return fmt.Errorf("failed to write fuse-overlayfs to %s: %v", targetPathFuseOverlayFs, err)
	}

	fmt.Printf("Successfully installed fuse-overlayfs to %s\n", targetPathFuseOverlayFs)

	return nil
}
