package config

import (
	"os"
	"path/filepath"
)

// LoadConfig loads configuration using hardcoded constants
// This is temporary until toml-vars-letsgooo library is ready
func LoadConfig() *Config {
	return &Config{
		ConfigDirPath:   ConfigDirPath,
		AppDirName:      AppDirName,
		PresetsDirName:  PresetsDirName,
		TimeoutSeconds:  DefaultTimeoutSeconds,
		MaxRetries:      DefaultMaxRetries,
		HTTPMethod:      DefaultHTTPMethod,
	}
}

type Config struct {
	ConfigDirPath   string
	AppDirName      string
	PresetsDirName  string
	TimeoutSeconds  int
	MaxRetries      int
	HTTPMethod      string
}

// GetConfigBase returns base config directory with environment validation
func GetConfigBase() (string, error) {
	home := os.Getenv("HOME")
	if home == "" {
		// Fallback mechanism for containerized environments
		return "/tmp/saul", nil
	}
	return home, nil
}

// GetPresetsPath returns full presets directory path
func (c *Config) GetPresetsPath() (string, error) {
	base, err := GetConfigBase()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, c.ConfigDirPath, c.AppDirName, c.PresetsDirName), nil
}

