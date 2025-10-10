package config

import (
	"os"
	"path/filepath"
)

// config/paths.go (new file)
func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir() // Cross-platform
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ParentDirPath, AppDirName), nil
}

func GetPresetsPath() (string, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(configPath, PresetsDirName), nil
}

func GetAppPath() (string, error) {
	appPath, err := GetConfigPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(appPath, AppDirName), nil
}
