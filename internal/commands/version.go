package commands

import (
	"github.com/DeprecatedLuar/better-curl-saul/pkg/display"
)

// VersionString holds the version information - set at build time via ldflags
var VersionString = "secretDevBuild"

// Version handles the version command - displays current version information
func Version() error {
	versionInfo := "Better-Curl (Saul) - " + "Beta " + VersionString
	display.Info(versionInfo)
	display.Plain("'When http gets complicated, Better Curl Saul'")
	return nil
}
