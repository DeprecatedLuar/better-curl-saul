package gohelp

import (
	"fmt"
	"os"
	"strings"

	"github.com/muesli/reflow/ansi"
	"github.com/muesli/reflow/truncate"
	"golang.org/x/term"
)

const (
	Blue  = "\033[34m"
	Reset = "\033[0m"
)

func GetTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		return 80
	}
	return width
}

func Separator() {
	width := GetTerminalWidth() - 4
	fmt.Println(strings.Repeat("─", width))
}

func Header(title string) string {
	width := GetTerminalWidth() - 4
	prefix := "──["
	suffix := "]"
	displayWidth := 2 + 1 + len(title) + 1

	if displayWidth >= width {
		return prefix + title + suffix
	}

	remaining := width - displayWidth
	return prefix + title + suffix + strings.Repeat("─", remaining)
}

func TruncateLine(line string, maxWidth int) string {
	return truncate.StringWithTail(line, uint(maxWidth), ">")
}

func AlignDescriptions(text string, alignAt int) string {
	lines := strings.Split(text, "\n")
	var result strings.Builder

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			result.WriteString(line + "\n")
			continue
		}

		blueIdx := strings.Index(line, Blue)
		if blueIdx == -1 {
			result.WriteString(line + "\n")
			continue
		}

		commandPart := line[:blueIdx]
		descriptionPart := line[blueIdx:]

		paddedCommand := commandPart
		visibleLen := ansi.PrintableRuneWidth(commandPart)
		if visibleLen < alignAt {
			paddedCommand += strings.Repeat(" ", alignAt-visibleLen)
		}

		result.WriteString(paddedCommand + descriptionPart + "\n")
	}

	return strings.TrimSuffix(result.String(), "\n")
}

const defaultAlignment = 24

func Item(command, description string) {
	width := GetTerminalWidth()
	mainPart := "  " + command
	visibleLen := ansi.PrintableRuneWidth(mainPart)
	if visibleLen < defaultAlignment {
		mainPart += strings.Repeat(" ", defaultAlignment-visibleLen)
	}
	line := mainPart + Blue + description + Reset
	fmt.Println(TruncateLine(line, width))
}

func Paragraph(text string) {
	fmt.Println()
	fmt.Println(text)
	fmt.Println()
}

func PrintHeader(title string) {
	fmt.Println()
	fmt.Println(Header(title))
	fmt.Println()
}
