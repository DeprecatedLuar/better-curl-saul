// Package utils provides shared utility functions for Better-Curl-Saul
package utils

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/DeprecatedLuar/better-curl-saul/src/modules/display"
	"github.com/DeprecatedLuar/better-curl-saul/src/modules/errors"
)

// Version information - these variables are set at build time via ldflags
var (
	Version = "dev"
)

// GitHubRelease represents the GitHub API response for a release
type GitHubRelease struct {
	TagName string `json:"tag_name"`
}

// GetVersionInfo returns formatted version information
func GetVersionInfo() string {
	return "Better-Curl (Saul) " + "Beta " + Version
}

// CheckForUpdates checks GitHub for the latest release and compares with current version
func CheckForUpdates() (bool, string, error) {
	client := resty.New()
	client.SetTimeout(5 * time.Second)

	resp, err := client.R().Get("https://api.github.com/repos/DeprecatedLuar/better-curl-saul/releases/latest")
	if err != nil {
		return false, "", err
	}

	if resp.StatusCode() != 200 {
		return false, "", nil // Treat non-200 as "no update available"
	}

	var release GitHubRelease
	err = json.Unmarshal(resp.Body(), &release)
	if err != nil {
		return false, "", err
	}

	// Compare versions - if they're different, an update is available
	hasUpdate := release.TagName != Version && release.TagName != ""

	return hasUpdate, release.TagName, nil
}

// HandleUpdateCommand handles the update command logic and display
func HandleUpdateCommand() error {
	hasUpdate, latestVersion, err := CheckForUpdates()
	if err != nil {
		display.Warning(errors.WarnUpdateCheckFailed)
		return nil
	}

	if hasUpdate {
		display.Info(fmt.Sprintf(errors.InfoUpdateAvailable, latestVersion, Version, latestVersion))
	} else {
		display.Info(errors.InfoUpToDate)
	}
	return nil
}
