package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/tornermarton/timesheets/internal/entries"
)

type Config struct {
	Source entries.TimeEntrySourceConfig `yaml:"source"`
	Target entries.TimeEntryTargetConfig `yaml:"target"`

	TimeZone *string `yaml:"timezone"`
}

func GetDefaultPath() string {
	if config, err := os.UserConfigDir(); err == nil {
		return filepath.Join(config, "timesheets", "config.yaml")
	}

	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".config", "timesheets", "config.yaml")
	}

	if cwd, err := os.Getwd(); err == nil {
		return filepath.Join(cwd, "config.yaml")
	}

	panic("Could not determine default config path")
}

func Read(path string) (*Config, error) {
	var cfg Config

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func Print(config *Config) error {
	content, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	os.Stdout.Write(content)

	return nil
}
