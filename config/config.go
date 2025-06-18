//go:build linux

package config

import (
	_ "embed"
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"github.com/repo-scm/git/utils"
)

//go:embed git.yaml
var configData string

type Config struct {
	Models  []Model `yaml:"models"`
	Overlay Overlay `yaml:"overlay"`
	Sshfs   Sshfs   `yaml:"sshfs"`
}

type Model struct {
	ProviderName string `yaml:"provider_name"`
	ApiBase      string `yaml:"api_base"`
	ApiKey       string `yaml:"api_key"`
	ModelId      string `yaml:"model_id"`
}

type Overlay struct {
	Mount string `yaml:"mount"`
}

type Sshfs struct {
	Mount   string   `yaml:"mount"`
	Options []string `yaml:"options"`
	Ports   []int    `yaml:"ports"`
}

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
		viper.SetConfigFile(name)
		if err := viper.ReadInConfig(); err != nil {
			return nil, errors.Wrap(err, "failed to read config after creation\n")
		}
	}

	buf, err := os.ReadFile(viper.ConfigFileUsed())
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(buf, &config); err != nil {
		return nil, err
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
