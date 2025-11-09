package commands

import (
	"fmt"

	"github.com/DeprecatedLuar/gohelp"
)

func Help() error {
	fmt.Println()
	fmt.Println("Better-Curl (Saul) - Workspace-based HTTP Client")

	gohelp.PrintHeader("Usage")
	fmt.Println("  saul [preset] [command] [target] [key=value]")

	gohelp.PrintHeader("Global Commands")
	gohelp.Item("saul version", "Show version information")
	gohelp.Item("saul update", "Check for updates")
	gohelp.Item("saul ls [options]", "List presets directory (system ls command)")
	gohelp.Item("saul rm [preset...]", "Delete one or more presets")
	gohelp.Item("saul help", "Show this help")

	gohelp.PrintHeader("Preset Commands")
	gohelp.Item("saul [preset]", "Create or switch to preset")
	gohelp.Item("saul [preset] set [target] [key=value]", "Set value in target file")
	gohelp.Item("saul [preset] get [target] [key]", "Get value from target file")
	gohelp.Item("saul [preset] call", "Execute HTTP request")
	gohelp.Item("saul call", "Execute HTTP request (current preset)")

	gohelp.PrintHeader("Targets")
	gohelp.Item("body", "HTTP request body (JSON)")
	gohelp.Item("headers", "HTTP headers")
	gohelp.Item("query", "Query/search payload data")
	gohelp.Item("request", "HTTP method, URL, and settings")
	gohelp.Item("variables", "Hard variables only (soft variables never stored)")

	gohelp.PrintHeader("Examples")
	fmt.Println("  # Special request syntax (no = sign)")
	fmt.Println("  saul pokeapi set url https://api.example.com")
	fmt.Println("  saul pokeapi set method POST")
	fmt.Println("  saul pokeapi set timeout 30")
	fmt.Println()
	fmt.Println("  # Regular TOML syntax (with = sign)")
	fmt.Println("  saul pokeapi set body pokemon.name=pikachu")
	fmt.Println("  saul pokeapi set header Content-Type=application/json")
	fmt.Println("  saul pokeapi set body pokemon.level=@level")
	fmt.Println()
	fmt.Println("  # Check what's configured")
	fmt.Println("  saul pokeapi get url")
	fmt.Println("  saul pokeapi get body pokemon.name")
	fmt.Println()

	return nil
}
