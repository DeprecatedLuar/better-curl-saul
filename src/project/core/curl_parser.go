package core

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

type CurlRequest struct {
	Method  string
	URL     string
	BaseURL string
	Query   map[string]string
	Headers map[string]string
	Body    string
}

func ParseCurl(curlCmd string) (*CurlRequest, error) {
	curlCmd = strings.TrimSpace(curlCmd)

	if !strings.HasPrefix(curlCmd, "curl") {
		return nil, fmt.Errorf("command must start with 'curl'")
	}

	req := &CurlRequest{
		Method:  "GET",
		Headers: make(map[string]string),
		Query:   make(map[string]string),
	}

	// Extract method (-X or --request)
	methodRegex := regexp.MustCompile(`(?:-X|--request)\s+([A-Z]+)`)
	if match := methodRegex.FindStringSubmatch(curlCmd); len(match) > 1 {
		req.Method = match[1]
	}

	// Extract URL (look for http/https URLs with or without quotes)
	urlRegex := regexp.MustCompile(`(?:'(https?://[^']+)'|"(https?://[^"]+)"|(https?://\S+))`)
	if match := urlRegex.FindStringSubmatch(curlCmd); len(match) > 1 {
		if match[1] != "" {
			req.URL = match[1]
		} else if match[2] != "" {
			req.URL = match[2]
		} else if match[3] != "" {
			req.URL = match[3]
		}
	}

	// Parse URL for query parameters
	if req.URL != "" {
		parsedURL, err := url.Parse(req.URL)
		if err == nil {
			// Extract base URL (without query string)
			req.BaseURL = parsedURL.Scheme + "://" + parsedURL.Host + parsedURL.Path

			// Extract query parameters
			for key, values := range parsedURL.Query() {
				if len(values) > 0 {
					req.Query[key] = values[0]
				}
			}
		}
	}

	// Extract headers (-H or --header)
	headerRegex := regexp.MustCompile(`(?:-H|--header)\s+(?:'([^']+)'|"([^"]+)")`)
	for _, match := range headerRegex.FindAllStringSubmatch(curlCmd, -1) {
		headerStr := match[1]
		if headerStr == "" {
			headerStr = match[2]
		}

		parts := strings.SplitN(headerStr, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			req.Headers[key] = value
		}
	}

	// Extract body (-d or --data or --data-raw)
	bodyRegex := regexp.MustCompile(`(?:-d|--data|--data-raw)\s+(?:'([^']*)'|"([^"]*)"|(\S+))`)
	if match := bodyRegex.FindStringSubmatch(curlCmd); len(match) > 1 {
		if match[1] != "" {
			req.Body = match[1]
		} else if match[2] != "" {
			req.Body = match[2]
		} else if match[3] != "" {
			req.Body = match[3]
		}
	}

	return req, nil
}