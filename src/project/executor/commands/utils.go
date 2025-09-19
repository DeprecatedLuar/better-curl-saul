package commands

import (
	"strings"
)


// NormalizeTarget converts target aliases to canonical names
func NormalizeTarget(target string) string {
	switch strings.ToLower(target) {
	case "body":
		return "body"
	case "headers", "header":
		return "headers"
	case "query", "queries":
		return "query"
	case "request", "req", "url":
		return "request"
	case "variables", "vars", "var":
		return "variables"
	default:
		return ""
	}
}