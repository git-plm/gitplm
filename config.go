package main

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Config struct {
	PMDir string `yaml:"pmDir"`
}

func loadConfig() (*Config, error) {
	config := &Config{}
	
	// Look for config file in current directory first, then home directory
	configPaths := []string{
		"gitplm.yaml",
		"gitplm.yml",
		".gitplm.yaml",
		".gitplm.yml",
	}
	
	// Also check home directory
	if homeDir, err := os.UserHomeDir(); err == nil {
		homePaths := []string{
			filepath.Join(homeDir, ".gitplm.yaml"),
			filepath.Join(homeDir, ".gitplm.yml"),
		}
		configPaths = append(configPaths, homePaths...)
	}
	
	var configData []byte
	var err error
	
	// Try to find and load a config file
	for _, path := range configPaths {
		if configData, err = os.ReadFile(path); err == nil {
			break
		}
	}
	
	// If no config file found, return empty config (not an error)
	if err != nil {
		return config, nil
	}
	
	err = yaml.Unmarshal(configData, config)
	if err != nil {
		return nil, err
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