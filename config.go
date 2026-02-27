package main

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type HTTPConfig struct {
	Enabled bool   `yaml:"enabled"`
	Port    int    `yaml:"port"`
	Token   string `yaml:"token"`
}

type Config struct {
	PMDir string     `yaml:"pmDir"`
	HTTP  HTTPConfig `yaml:"http"`
}

var configNames = []string{
	"gitplm.yaml",
	"gitplm.yml",
	".gitplm.yaml",
	".gitplm.yml",
}

// findConfigFile searches for a config file starting from the current directory,
// walking up parent directories to the filesystem root, then falling back to
// the home directory.
func findConfigFile() (string, bool) {
	dir, err := os.Getwd()
	if err == nil {
		dir, _ = filepath.Abs(dir)
		for {
			for _, name := range configNames {
				p := filepath.Join(dir, name)
				if _, err := os.Stat(p); err == nil {
					return p, true
				}
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}

	if homeDir, err := os.UserHomeDir(); err == nil {
		for _, name := range configNames {
			p := filepath.Join(homeDir, name)
			if _, err := os.Stat(p); err == nil {
				return p, true
			}
		}
	}

	return "", false
}

func loadConfig() (*Config, error) {
	config := &Config{}

	configPath, found := findConfigFile()
	if !found {
		return config, nil
	}

	configData, err := os.ReadFile(configPath)
	if err != nil {
		return config, nil
	}

	err = yaml.Unmarshal(configData, config)
	if err != nil {
		return nil, err
	}

	// Resolve relative pmDir against the config file's directory
	if config.PMDir != "" && !filepath.IsAbs(config.PMDir) {
		config.PMDir = filepath.Join(filepath.Dir(configPath), config.PMDir)
	}

	return config, nil
}

func saveConfig(pmDir string) error {
	config := Config{
		PMDir: pmDir,
	}

	data, err := yaml.Marshal(&config)
	if err != nil {
		return err
	}

	return os.WriteFile("gitplm.yml", data, 0644)
}
