package config

const (
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