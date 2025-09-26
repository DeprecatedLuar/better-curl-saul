// Package config provides centralized configuration management for Better-Curl-Saul.
// This package handles path resolution, environment validation, and configuration loading
// with fallback mechanisms for containerized and production environments.
package config


// LoadConfig loads configuration using hardcoded constants
// This is temporary until toml-vars-letsgooo library is ready
func LoadConfig() *Config {
	return &Config{
		ConfigDirPath:  ConfigDirPath,
		AppDirName:     AppDirName,
		PresetsDirName: PresetsDirName,
		TimeoutSeconds: DefaultTimeoutSeconds,
		MaxRetries:     DefaultMaxRetries,
		HTTPMethod:     DefaultHTTPMethod,
	}
}

type Config struct {
	ConfigDirPath  string
	AppDirName     string
	PresetsDirName string
	TimeoutSeconds int
	MaxRetries     int
	HTTPMethod     string
}


