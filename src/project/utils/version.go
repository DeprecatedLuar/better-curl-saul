// Package utils provides shared utility functions for Better-Curl-Saul
package utils

// Version information - these variables are set at build time via ldflags
var (
	Version = "dev"
)

// GetVersionInfo returns formatted version information
func GetVersionInfo() string {
	return "Better-Curl (Saul) " + "Beta " + Version
}
