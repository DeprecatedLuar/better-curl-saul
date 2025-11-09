package parser

import (
	"strings"
)

// IsHTTPMethod checks if a string is a valid HTTP method
func IsHTTPMethod(s string) bool {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
	upper := strings.ToUpper(s)
	for _, method := range methods {
		if upper == method {
			return true
		}
	}
	return false
}

// LooksLikeURL checks if a string resembles a URL
func LooksLikeURL(s string) bool {
	// URLs typically contain dots or colons (for ports)
	// Exclude strings with "://" followed by nothing or just ":" for headers
	if strings.Contains(s, "://") {
		return true
	}
	// Must have dots (domain) or start with localhost/IP patterns
	return strings.Contains(s, ".") ||
	       strings.HasPrefix(s, "localhost") ||
	       strings.HasPrefix(s, "127.0.0.1") ||
	       strings.HasPrefix(s, "192.168.")
}

// IsHTTPieArg checks if an argument is HTTPie-style (contains separators)
func IsHTTPieArg(s string) bool {
	// Contains = or : (body/query or header)
	// Exclude URLs with ://
	if strings.Contains(s, "://") {
		return false
	}
	return strings.Contains(s, "=") || strings.Contains(s, ":")
}

// ParseHTTPieSyntax parses HTTPie-style command syntax into a Command
// Example: POST api.com Authorization:token name=value q==search
func ParseHTTPieSyntax(args []string, preset string) (Command, error) {
	cmd := Command{
		Preset: preset,
		Command: "httpie", // Mark as HTTPie syntax for routing
	}

	var kvPairs []KeyValuePair

	for _, arg := range args {
		if IsHTTPMethod(arg) {
			// HTTP method -> request target
			kvPairs = append(kvPairs, KeyValuePair{
				Target: "request",
				Key:    "method",
				Value:  strings.ToUpper(arg),
			})
		} else if strings.Contains(arg, "==") {
			// Query param: key==value (check before single =)
			parts := strings.SplitN(arg, "==", 2)
			if len(parts) == 2 {
				kvPairs = append(kvPairs, KeyValuePair{
					Target: "query",
					Key:    parts[0],
					Value:  parts[1],
				})
			}
		} else if strings.Contains(arg, "=") {
			// Body param: key=value
			parts := strings.SplitN(arg, "=", 2)
			if len(parts) == 2 {
				kvPairs = append(kvPairs, KeyValuePair{
					Target: "body",
					Key:    parts[0],
					Value:  parts[1],
				})
			}
		} else if strings.Contains(arg, ":") && !strings.Contains(arg, "://") {
			// Header: Key:value
			parts := strings.SplitN(arg, ":", 2)
			if len(parts) == 2 {
				kvPairs = append(kvPairs, KeyValuePair{
					Target: "headers",
					Key:    parts[0],
					Value:  parts[1],
				})
			}
		} else if LooksLikeURL(arg) {
			// URL -> request target (check last to avoid catching key=value.com)
			kvPairs = append(kvPairs, KeyValuePair{
				Target: "request",
				Key:    "url",
				Value:  arg,
			})
		}
	}

	cmd.KeyValuePairs = kvPairs

	return cmd, nil
}
