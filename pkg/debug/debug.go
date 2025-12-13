package debug

import (
	"fmt"
	"os"
)

var enabled = false

// Enable turns on debug output
func Enable() {
	enabled = true
}

// IsEnabled returns whether debug mode is on
func IsEnabled() bool {
	return enabled
}

// Log prints a debug message if debug mode is enabled
func Log(format string, args ...interface{}) {
	if !enabled {
		return
	}
	fmt.Fprintf(os.Stderr, "[DEBUG] "+format+"\n", args...)
}
