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
	EnvFile     string
	ExampleFile string
}

type RequiredField struct {
	Type string
}

type ScanConfig struct {
	IgnorePaths []string
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
