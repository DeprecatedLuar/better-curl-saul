package main

import (
	"fmt"
	httpModule "github.com/DeprecatedLuar/better-curl-saul/src/project/executor/http"
)

func main() {
	// Test with simple filtered JSON
	testJSON := `{"name":"pikachu","stats.0.base_stat":35,"types.0.type.name":"electric"}`
	result := httpModule.FormatResponseContent([]byte(testJSON), "pokeapi", false)
	fmt.Printf("Result: %s\n", result)
}
