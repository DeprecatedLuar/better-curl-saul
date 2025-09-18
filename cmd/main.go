package main

import (
	"fmt"
	"os"

	"main/src/project/parser"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Println("Ok so let me break it down to you buddy: saul [preset] [set/rm/edit...] [url/body...] [key=value]")
		return
	}

	cmd, err := parser.ParseCommand(args)
	if err != nil {
		fmt.Printf("Oopsies: %v\n", err)
		return
	}

	fmt.Printf("\n  Global: %s\n", cmd.Global)
	fmt.Printf("  Preset: %s\n", cmd.Preset)
	fmt.Printf("  Command: %s\n", cmd.Command)
	fmt.Printf("  Target: %s\n", cmd.Target)
	fmt.Printf("  Key: %s\n", cmd.Key)
	fmt.Printf("  Value: %s\n\n", cmd.Value)
}
