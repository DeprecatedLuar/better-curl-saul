package parser

import (
	"fmt"
	"strings"
)

// parseFlags extracts flags from args and sets them in cmd, returning filtered args
func parseFlags(args []string, cmd *Command) ([]string, error) {
	var filteredArgs []string
	skip := 0

	for i, arg := range args {
		if skip > 0 {
			skip--
			continue
		}

		if strings.HasPrefix(arg, "--") {
			// Handle long flags
			switch arg {
			case "--raw":
				cmd.RawOutput = true
			case "--headers-only":
				cmd.ResponseFormat = "headers-only"
			case "--body-only":
				cmd.ResponseFormat = "body-only"
			case "--status-only":
				cmd.ResponseFormat = "status-only"
			case "--dry-run":
				cmd.DryRun = true
			case "--call":
				cmd.Call = true
			case "--create":
				cmd.Create = true
			default:
				return nil, fmt.Errorf("unknown flag: %s", arg)
			}
		} else if strings.HasPrefix(arg, "-") && len(arg) > 1 && !strings.HasPrefix(arg, "--") {
			// Handle short flags
			flagPart := arg[1:] // Remove leading -
			if flagPart == "v" {
				// -v flag: collect all following non-flag arguments as variable names
				varCount := 0
				for j := i + 1; j < len(args); j++ {
					if strings.HasPrefix(args[j], "-") {
						break // Stop at next flag
					}
					cmd.VariableFlags = append(cmd.VariableFlags, args[j])
					varCount++
				}
				skip = varCount // Skip these args in main loop
				// If no variables specified, empty slice signals "all variables"
				if len(cmd.VariableFlags) == 0 {
					cmd.VariableFlags = []string{}
				}
			} else {
				return nil, fmt.Errorf("unknown flag: %s", arg)
			}
		} else {
			// Not a flag, keep in filtered args
			filteredArgs = append(filteredArgs, arg)
		}
	}

	return filteredArgs, nil
}
