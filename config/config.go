//go:build linux

package config

import (
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/repo-scm/git/utils"
	"github.com/spf13/viper"
)

type Config struct {
	Overlay Overlay `yaml:"overlay"`
	Sshfs   Sshfs   `yaml:"sshfs"`
}

type Overlay struct {
	Mount string `yaml:"mount"`
}

type Sshfs struct {
	Mount   string   `yaml:"mount"`
	Options []string `yaml:"options"`
	Ports   []int    `yaml:"ports"`
}

const configData = `overlay:
  mount: "/mnt/repo-scm/git/overlay"
sshfs:
  mount: "/mnt/repo-scm/git/sshfs"
options:
  - "allow_other,default_permissions,follow_symlinks"
  - "cache=yes,kernel_cache,compression=no,big_writes,cache_timeout=115200"
  - "Cipher=aes128-ctr,StrictHostKeyChecking=no,UserKnownHostsFile=/dev/null"
ports:
  - 22
`

func LoadConfig(name string) (*Config, error) {
	var config Config

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	if name != "" {
		viper.SetConfigFile(name)
	} else {
		viper.AddConfigPath(path.Join(home, ".repo-scm"))
		viper.SetConfigName("git")
		viper.SetConfigType("yaml")
	}

	if err := viper.ReadInConfig(); err != nil {
		if name == "" {
			name = path.Join(home, ".repo-scm", "git.yaml")
		}
		if err := createConfig(name); err != nil {
			return nil, errors.Wrap(err, "failed to read or create config\n")
		}
	}

	if err := viper.Unmarshal(&config); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal config\n")
	}

	return &config, nil
}

func createConfig(name string) error {
	if err := os.MkdirAll(path.Dir(name), utils.PermDir); err != nil {
		return err
	}

	if err := os.WriteFile(name, []byte(configData), utils.PermFile); err != nil {
		return err
	}

	return nil
}
