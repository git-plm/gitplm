package main

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

// FieldConfig says how the CSV columns of one IPN category are presented to
// KiCad. Every column is served hidden under its own name, so a category only
// states its exceptions:
//
//	Value   the column that populates KiCad's built-in Value field
//	Visible the columns KiCad displays on the schematic; all others are hidden
//	Rename  columns served under a different KiCad field name
//
// Visible and Rename are keyed by CSV column name, not by KiCad field name.
type FieldConfig struct {
	Value   string            `yaml:"value"`
	Visible []string          `yaml:"visible"`
	Rename  map[string]string `yaml:"rename"`
}

type HTTPConfig struct {
	Enabled bool   `yaml:"enabled"`
	Port    int    `yaml:"port"`
	Token   string `yaml:"token"`
	// Fields configures the fields served for each IPN category (RES, CAP,
	// ...). The "default" key applies to every category, and a category's own
	// settings are applied on top of it.
	Fields map[string]FieldConfig `yaml:"fields"`
}

// FieldsForCategory returns the field configuration for a category: the
// "default" settings with the category's own applied on top. A category
// replaces the default's value column and visible list outright, and adds to
// its renames.
func (h HTTPConfig) FieldsForCategory(category string) FieldConfig {
	merged := h.Fields["default"]

	fields, ok := h.Fields[strings.ToUpper(category)]
	if !ok {
		fields, ok = h.Fields[strings.ToLower(category)]
	}
	if !ok {
		return merged
	}

	if fields.Value != "" {
		merged.Value = fields.Value
	}

	// A category that lists no visible columns of its own inherits the
	// default's. `visible: []` is a category that displays nothing.
	if fields.Visible != nil {
		merged.Visible = fields.Visible
	}

	if len(fields.Rename) > 0 {
		renames := make(map[string]string, len(merged.Rename)+len(fields.Rename))
		for column, name := range merged.Rename {
			renames[column] = name
		}
		for column, name := range fields.Rename {
			renames[column] = name
		}
		merged.Rename = renames
	}

	return merged
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
