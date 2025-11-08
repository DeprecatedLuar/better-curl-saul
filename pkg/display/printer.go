// Package display provides centralized console output functions with consistent formatting.
// All console output in Better-Curl-Saul should use these functions instead of direct fmt.Print* calls
// to maintain consistent styling, proper stderr/stdout separation, and Unix philosophy compliance.
package display

import (
	"fmt"
	"os"
)

// Error prints error messages to stderr with consistent formatting
func Error(msg string) {
	fmt.Fprintf(os.Stderr, "✗ %s\n", msg)
}

// Success prints success messages to stdout
func Success(msg string) {
	fmt.Printf("✓ %s\n", msg)
}

// Warning prints warning messages to stdout
func Warning(msg string) {
	fmt.Printf("%s\n", msg)
}

// Info prints informational messages to stdout
func Info(msg string) {
	fmt.Printf("» %s\n", msg)
}

// Tip prints helpful tips/hints to stdout
func Tip(msg string) {
	fmt.Printf("→ %s\n", msg)
}

// Plain prints messages without any formatting or prefixes
func Plain(msg string) {
	fmt.Printf("%s\n", msg)
}
