package display

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

const (
	DefaultWidth      = 80
	MinWidth         = 40
	MaxWidth         = 120
	SeparatorChar    = "─"
	BulletSeparator  = " • "
)

// getTerminalWidth returns the current terminal width with intelligent fallback
func getTerminalWidth() int {
	if width, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && width > MinWidth {
		if width > MaxWidth {
			return MaxWidth
		}
		return width
	}
	return DefaultWidth
}

// calculateSeparatorWidth determines optimal separator width for readability
func calculateSeparatorWidth() int {
	termWidth := getTerminalWidth()
	if termWidth < 100 {
		// Use 80% of terminal width for smaller terminals
		return int(float64(termWidth) * 0.8)
	}
	// Use fixed 80 chars for larger terminals
	return DefaultWidth
}

// FormatSection creates clean formatting with simple header and footer separators
func FormatSection(title, content, metadata string) string {
	sepWidth := calculateSeparatorWidth()
	
	var result strings.Builder
	
	// Add initial line break for spacing from terminal
	result.WriteString("\n")
	
	// Simple header with title and metadata
	if metadata != "" {
		result.WriteString(fmt.Sprintf("%s%s%s\n", title, BulletSeparator, metadata))
	} else {
		result.WriteString(fmt.Sprintf("%s\n", title))
	}
	
	// Top separator
	result.WriteString(strings.Repeat(SeparatorChar, sepWidth))
	result.WriteString("\n")
	
	// Content
	if content != "" {
		result.WriteString("\n")
		result.WriteString(content)
		if !strings.HasSuffix(content, "\n") {
			result.WriteString("\n")
		}
	}
	
	return result.String()
}

// FormatResponse creates formatted response display with status metadata
func FormatResponse(status, contentType, timing, size string, content string) string {
	// Build clean metadata string
	metadata := fmt.Sprintf("%s%s%s%s%s", status, BulletSeparator, size, BulletSeparator, contentType)
	
	return FormatSection("Response:", content, metadata)
}

// FormatFileDisplay creates formatted file content display with file metadata
func FormatFileDisplay(fileType, size, entryCount string, content string) string {
	metadata := fmt.Sprintf("%s%s%s entries", size, BulletSeparator, entryCount)
	return FormatSection(fileType+":", content, metadata)
}

// FormatSimpleSection creates basic section formatting without metadata
func FormatSimpleSection(title, content string) string {
	return FormatSection(title+":", content, "")
}