package display

import "fmt"

// SectionHeader creates a visual header for content sections
func SectionHeader(title string) string {
	return fmt.Sprintf("┌─ %s ─┐", title)
}

// SectionFooter creates a separator line to close sections
func SectionFooter() string {
	return "──────────────────────"
}

// SectionWrap wraps content with header and footer for complete section formatting
func SectionWrap(title, content string) string {
	return fmt.Sprintf("%s\n%s\n%s\n%s",
		SectionHeader(title), SectionFooter(), content, SectionFooter())
}