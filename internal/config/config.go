package config

import (
	"os"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Env      EnvConfig
	Required map[string]RequiredField
	Scan     ScanConfig
}

type EnvConfig struct {
	EnvFile     string `toml:"env_file"`
	ExampleFile string `toml:"example_file"`
}

type RequiredField struct {
	Type string
}

type ScanConfig struct {
	IgnorePaths []string `toml:"ignore_paths"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return nil, err
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func Default() *Config {
	return &Config{
		Env: EnvConfig{
			EnvFile:     ".env",
			ExampleFile: ".env.example",
		},
		Required: map[string]RequiredField{},
		Scan: ScanConfig{
			IgnorePaths: []string{".git", "vendor", "node_modules", "dist"},
		},
	}
}
