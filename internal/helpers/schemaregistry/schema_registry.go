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

	parsedUrl, err := url.Parse(baseURL)
	if err != nil {
		panic("Schema registry url invalid")
	}

	if injectSchema && parsedUrl.Port() == "443" {
		parsedUrl.Scheme = "https"
	}

	if parsedUrl.Port() == "" {
		if parsedUrl.Scheme == "https" {
			parsedUrl.Host = fmt.Sprintf("%s:%s", parsedUrl.Hostname(), "443")
		} else {
			parsedUrl.Host = fmt.Sprintf("%s:%s", parsedUrl.Hostname(), "80")
		}
	}

	return parsedUrl.String()
}
