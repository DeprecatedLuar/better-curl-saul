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
	fmt.Println("  saul [preset] [URL] [METHOD] [key=value] [key==value] [Key:value]  # HTTPie syntax")

	gohelp.PrintHeader("Global Commands")
	gohelp.Item("saul version", "Show version information")
	gohelp.Item("saul update", "Check for updates")
	gohelp.Item("saul ls [options]", "List presets directory (system ls command)")
	gohelp.Item("saul cp [source] [dest]", "Copy preset or variant")
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

	gohelp.PrintHeader("HTTPie-Style Syntax")
	fmt.Println("  Quick one-liners without explicit 'set' command:")
	fmt.Println()
	fmt.Println("  # URL + body fields")
	fmt.Println("  saul myapi https://api.example.com name=john email=john@example.com")
	fmt.Println()
	fmt.Println("  # Method + URL + headers")
	fmt.Println("  saul myapi POST https://api.example.com/users Authorization:Bearer-token")
	fmt.Println()
	fmt.Println("  # Query parameters (use ==)")
	fmt.Println("  saul myapi https://api.example.com/search q==searchterm limit==10")
	fmt.Println()
	fmt.Println("  # Mix everything in one command")
	fmt.Println("  saul myapi POST https://api.example.com/posts title=foo Auth:token tags==prod")
	fmt.Println()
	fmt.Println("  Syntax: = (body), == (query), : (headers), URLs and methods auto-detected")
	fmt.Println()

	gohelp.PrintHeader("Explicit Syntax (Classic)")
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

	gohelp.PrintHeader("Other Examples")
	fmt.Println("  # Copy presets and variants")
	fmt.Println("  saul cp myapi backup")
	fmt.Println("  saul cp myapi_submit myapi/submit")
	fmt.Println("  saul cp myapi/submit myapi/get")
	fmt.Println()
	fmt.Println("  # Check what's configured")
	fmt.Println("  saul pokeapi get url")
	fmt.Println("  saul pokeapi get body pokemon.name")
	fmt.Println()

	return nil
}
