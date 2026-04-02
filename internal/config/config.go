package config

import (
	"os"

	"github.com/BurntSushi/toml"
	"github.com/lovelyJason/openskills/internal/paths"
	"github.com/lovelyJason/openskills/internal/resource"
)

type Config struct {
	DefaultTargets     []string             `toml:"default_targets"`
	DefaultInstallMode resource.InstallMode `toml:"default_install_mode"`
	MaxBackups         int                  `toml:"max_backups"`
}

func Default() *Config {
	return &Config{
		DefaultTargets:     []string{},
		DefaultInstallMode: resource.ModeSymlink,
		MaxBackups:         10,
	}
}

func Load() (*Config, error) {
	cfg := Default()
	data, err := os.ReadFile(paths.ConfigFile())
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}
	if _, err := toml.Decode(string(data), cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) Save() error {
	if err := paths.EnsureDir(paths.ConfigDir()); err != nil {
		return err
	}
	f, err := os.Create(paths.ConfigFile())
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(c)
}
