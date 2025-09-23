package config

const (
	// File permissions
	DirPermissions  = 0755
	FilePermissions = 0644

	// Directory configuration (hardcoded until library ready)
	ConfigDirPath   = ".config"
	AppDirName      = "saul"
	PresetsDirName  = "presets"

	// Default values
	DefaultTimeoutSeconds = 30
	DefaultMaxRetries     = 3
	DefaultHTTPMethod     = "GET"

)