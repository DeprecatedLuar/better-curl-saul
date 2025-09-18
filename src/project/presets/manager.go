package presets

import (
	"fmt"
	"os"
	"path/filepath"

	"main/src/modules/config"
)

func CreatePreset(presetName string) error {
	presetsDir, err := config.GetPresetsDir()
	if err != nil {
		return err
	}

	presetPath := filepath.Join(presetsDir, presetName)

	err = os.MkdirAll(presetPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create preset directory %s: %w", presetPath, err)
	}

	return nil
}

func WriteTomlField(presetName, fileType, field, value string) error {
	presetsDir, err := config.GetPresetsDir()
	if err != nil {
		return err
	}

	presetPath := filepath.Join(presetsDir, presetName)
	tomlPath := filepath.Join(presetPath, fileType+".toml")

	err = CreatePreset(presetName)
	if err != nil {
		return err
	}

	content := fmt.Sprintf("%s = \"%s\"\n", field, value)

	file, err := os.OpenFile(tomlPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open/create TOML file %s: %w", tomlPath, err)
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return fmt.Errorf("failed to write to TOML file %s: %w", tomlPath, err)
	}

	return nil
}