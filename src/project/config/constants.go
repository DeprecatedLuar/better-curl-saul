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

	// Command constants
	SaulVersion = "version"
	SaulSet     = "set"
	SaulRemove  = "remove"
	SaulEdit    = "edit"
)

var ShortAliases = map[string]string{
	"v":  "version",
	"s":  "set",
	"rm": "remove",
	"ed": "edit",
}