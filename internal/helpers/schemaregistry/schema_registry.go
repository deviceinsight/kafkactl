package schemaregistry

import (
	"fmt"
	"net/url"
	"strings"
)

// FormatBaseURL will try to make sure that the schema:host:port:path pattern is followed on the `baseURL` field.
func FormatBaseURL(baseURL string) string {
	if baseURL == "" {
		return ""
	}

	// remove last slash, so the API can append the path with ease.
	if baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[0 : len(baseURL)-1]
	}

	schemaIdx := strings.Index(baseURL, "://")
	injectSchema := schemaIdx < 0

	if injectSchema {
		baseURL = fmt.Sprintf("http://%s", baseURL)
	}

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		panic("Schema registry url invalid")
	}

	if injectSchema && parsedURL.Port() == "443" {
		parsedURL.Scheme = "https"
	}

	if parsedURL.Port() == "" {
		if parsedURL.Scheme == "https" {
			parsedURL.Host = fmt.Sprintf("%s:%s", parsedURL.Hostname(), "443")
		} else {
			parsedURL.Host = fmt.Sprintf("%s:%s", parsedURL.Hostname(), "80")
		}
	}

	return parsedURL.String()
}
