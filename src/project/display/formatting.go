package display

import (
	"fmt"
	"math"
	"strings"
)

const sectionWidth = 55
const headerToFooterRatio = 1.078 // 55:51 ratio (accounting for ┘ character)

// SectionHeader creates a visual header for content sections
func SectionHeader(title string) string {
	// Build exactly sectionWidth characters
	start := "┌─ " + title + " "
	needed := sectionWidth - len(start) - 1 // -1 for closing ┐
	return start + strings.Repeat("─", needed) + "┐"
}

// SectionFooter creates a separator line to close sections
func SectionFooter() string {
	footerWidth := int(math.Round(float64(sectionWidth) / headerToFooterRatio))
	return strings.Repeat("─", footerWidth-1) + "┘"
}

// SectionStart creates a section header with proper spacing for content
func SectionStart(title string) string {
	return fmt.Sprintf("\n%s\n", SectionHeader(title))
}

// SectionWrap wraps content with header and footer for complete section formatting
func SectionWrap(title, content string) string {
	return fmt.Sprintf("%s\n%s\n%s\n%s",
		SectionHeader(title), SectionFooter(), content, SectionFooter())
}
