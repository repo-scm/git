//go:build linux

package mount

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"syscall"

	"github.com/pkg/errors"
)

const (
	directoryPerm = 0755
)

func ParsePath(_ context.Context, name string) (remote, local string) {
	// Remote format: user@host:/remote/repo:/local/repo
	remotePattern := `^([^@]+)@([^:]+):([^:]+):([^:]+)$`
	remoteRegex := regexp.MustCompile(remotePattern)

	if matches := remoteRegex.FindStringSubmatch(name); matches != nil {
		_path := strings.Split(name, ":")
		return _path[0] + ":" + _path[1], _path[2]
	}

	return "", name
}

func MountSshfs(_ context.Context, key, remote, local string) error {
	if remote == "" || local == "" {
		return errors.New("remote and local are required\n")
	}

	if err := os.MkdirAll(local, directoryPerm); err != nil {
		return errors.Wrap(err, "failed to make directory\n")
	}

	config := "StrictHostKeyChecking=no,UserKnownHostsFile=/dev/null,port=22"
	if key != "" {
		config = fmt.Sprintf("%s,IdentityFile=%s", config, path.Clean(key))
	}

	cmd := exec.Command("sshfs",
		remote,
		path.Clean(local),
		"-o", "allow_other",
		"-o", "default_permissions",
		"-o", "follow_symlinks",
		"-o", config,
	)

	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "failed to mount sshfs\n")
	}

	fmt.Printf("Successfully mounted sshfs at %s\n", path.Clean(local))

	return nil
}

func UnmountSshfs(_ context.Context, local string) error {
	if local == "" {
		return errors.New("local is required\n")
	}

	defer func(path string) {
		_ = os.RemoveAll(path)
	}(local)

	cmd := exec.Command("fusermount", "-u", path.Clean(local))

	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "failed to unmount sshfs\n")
	}

	fmt.Printf("Successfully unmounted sshfs\n")

	return nil
}

func MountOverlay(_ context.Context, repo, mount string) error {
	if repo == "" || mount == "" {
		return errors.New("repo and mount are required\n")
	}

	mountDir := path.Dir(path.Clean(mount))
	mountName := path.Base(path.Clean(mount))

	upperPath := path.Join(mountDir, "cow-"+mountName)
	workPath := path.Join(mountDir, "work-"+mountName)

	dirs := []string{mount, upperPath, workPath}

	for _, item := range dirs {
		if err := os.MkdirAll(item, directoryPerm); err != nil {
			return errors.Wrap(err, "failed to make directory\n")
		}
	}

	opts := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s,index=off", repo, upperPath, workPath)

	if err := syscall.Mount("overlay", mount, "overlay", 0, opts); err != nil {
		return errors.Wrap(err, "failed to mount overlay\n")
	}

	fmt.Printf("Successfully mounted overlay at %s\n", mount)

	return nil
}

func UnmountOverlay(_ context.Context, repo, unmount string) error {
	if repo == "" || unmount == "" {
		return errors.New("repo and unmount are required\n")
	}

	unmountDir := path.Dir(path.Clean(unmount))
	unmountName := path.Base(path.Clean(unmount))

	upperPath := path.Join(unmountDir, "cow-"+unmountName)
	workPath := path.Join(unmountDir, "work-"+unmountName)

	defer func(unmount, workPath, upperPath string) {
		_ = os.RemoveAll(unmount)
		_ = os.RemoveAll(workPath)
		_ = os.RemoveAll(upperPath)
	}(unmount, workPath, upperPath)

	if err := syscall.Unmount(unmount, 0); err != nil {
		return errors.Wrap(err, "failed to unmount overlay\n")
	}

	fmt.Printf("Successfully unmounted overlay\n")

	return nil
}
