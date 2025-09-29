package core

import (
	"fmt"
	"strings"
	"testing"
)

func TestParseCurlPrint(t *testing.T) {
	tests := []struct {
		name    string
		curlCmd string
	}{
		{
			name:    "Simple GET",
			curlCmd: `curl https://api.com`,
		},
		{
			name:    "POST with body",
			curlCmd: `curl -X POST https://api.com -d '{"user":"alice"}'`,
		},
		{
			name:    "With headers",
			curlCmd: `curl -X POST https://api.com -H 'Authorization: Bearer token' -H 'Content-Type: application/json' -d '{"name":"test"}'`,
		},
		{
			name:    "URL with query params",
			curlCmd: `curl 'https://api.com?foo=bar&baz=qux'`,
		},
		{
			name:    "Complex real-world example",
			curlCmd: `curl -X POST 'https://api.github.com/repos/owner/repo/issues?state=open' -H 'Authorization: Bearer ghp_token123' -H 'Accept: application/vnd.github+json' -d '{"title":"Bug report","body":"Description here"}'`,
		},
		{
			name: "Instantly.ai POST with multiline body",
			curlCmd: `curl -i -X POST \
  https://api.instantly.ai/api/v2/accounts/warmup-analytics \
  -H 'Authorization: Bearer <YOUR_TOKEN_HERE>' \
  -H 'Content-Type: application/json' \
  -d '{
    "emails": [
      "user@example.com"
    ]
  }'`,
		},
		{
			name:    "Instantly.ai GET with many query params",
			curlCmd: `curl -i -X GET 'https://api.instantly.ai/api/v2/campaigns/analytics?end_date=2024-01-01&exclude_total_leads_count=true&id=019981cb-fb99-705a-939e-7495cca31a4a&ids=019981cb-fb99-705a-939e-7496f101d2ed&start_date=2024-01-01' -H 'Authorization: Bearer <YOUR_TOKEN_HERE>'`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Println("\n" + strings.Repeat("=", 60))
			fmt.Printf("Test: %s\n", tt.name)
			fmt.Println(strings.Repeat("=", 60))
			fmt.Printf("Input: %s\n\n", tt.curlCmd)

			result, err := ParseCurl(tt.curlCmd)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			fmt.Printf("Method:   %s\n", result.Method)
			fmt.Printf("URL:      %s\n", result.URL)
			fmt.Printf("Base URL: %s\n", result.BaseURL)

			if len(result.Query) > 0 {
				fmt.Println("\nQuery Parameters:")
				for key, val := range result.Query {
					fmt.Printf("  %s = %s\n", key, val)
				}
			}

			if len(result.Headers) > 0 {
				fmt.Println("\nHeaders:")
				for key, val := range result.Headers {
					fmt.Printf("  %s: %s\n", key, val)
				}
			}

			if result.Body != "" {
				fmt.Printf("\nBody:\n  %s\n", result.Body)
			}

			fmt.Println(strings.Repeat("=", 60))
		})
	}
}