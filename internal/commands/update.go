package commands

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/DeprecatedLuar/better-curl-saul/pkg/display"
	"github.com/go-resty/resty/v2"
)

// GitHubRelease represents the GitHub API response for a release
type GitHubRelease struct {
	TagName string `json:"tag_name"`
}

// Update handles the update command - checks for updates and displays status
func Update() error {
	hasUpdate, latestVersion, err := checkForUpdates()
	if err != nil {
		display.Warning(display.WarnUpdateCheckFailed)
		return nil
	}

	if hasUpdate {
		display.Info(fmt.Sprintf(display.InfoUpdateAvailable, latestVersion, VersionString, latestVersion))
	} else {
		display.Info(display.InfoUpToDate)
	}
	return nil
}

// checkForUpdates checks GitHub for the latest release and compares with current version
func checkForUpdates() (bool, string, error) {
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
	hasUpdate := release.TagName != VersionString && release.TagName != ""

	return hasUpdate, release.TagName, nil
}
