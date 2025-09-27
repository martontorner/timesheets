package config

import (
	"os"

	"gopkg.in/yaml.v3"

	"github.com/tornermarton/timesheets/internal/entries"
)

type Config struct {
	Source entries.TimeEntrySourceConfig `yaml:"source"`
	Target entries.TimeEntryTargetConfig `yaml:"target"`
}

func Read(path string) (*Config, error) {
	var cfg Config

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func Print(config *Config) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	os.Stdout.Write(data)

	return nil
}
