# gohelp

Simple Go library for beautiful CLI help screens with automatic formatting.

## Quick Start

```go
import "github.com/DeprecatedLuar/gohelp"

gohelp.PrintHeader("Commands")
gohelp.Item("start [options]", "Start the service")
gohelp.Item("stop", "Stop the service")

gohelp.Paragraph("Additional information here")
```

## API

Four simple functions:

- **`PrintHeader("Title")`** - Section header with decorative lines
- **`Item("command", "description")`** - Aligned key-value pair (auto-colored blue)
- **`Paragraph("text")`** - Regular text with spacing
- **`Separator()`** - Full-width horizontal line

## Features

- Adapts to terminal width automatically
- Auto-aligns descriptions at column 24
- Handles long lines with smart truncation (adds `>` indicator)
- Preserves ANSI colors when truncating
- Automatic spacing between elements

## Usage in Your Project

**Local development (no GitHub required):**

```bash
# Add to your project's go.mod
go mod edit -replace github.com/DeprecatedLuar/gohelp=./lib/gohelp
```

**From GitHub (once published):**

```bash
go get github.com/DeprecatedLuar/gohelp
```

## Example Output

```
──[Commands]────────────────────────────────────
  start [options]       Start the service
  stop                  Stop the service
```

See `_examples/main.go` for a complete working example.
