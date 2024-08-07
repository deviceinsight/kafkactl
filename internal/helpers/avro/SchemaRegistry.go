package avro

import (
	"strings"

	"github.com/riferrei/srclient"
)

func CreateSchemaRegistryClient(baseURL string) srclient.ISchemaRegistryClient {
	baseURL = formatBaseURL(baseURL)
	return srclient.CreateSchemaRegistryClient(baseURL)
}

// formatBaseURL will try to make sure that the schema:host:port pattern is followed on the `baseURL` field.
func formatBaseURL(baseURL string) string {
	if baseURL == "" {
		return ""
	}

	// remove last slash, so the API can append the path with ease.
	if baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[0 : len(baseURL)-1]
	}

	portIdx := strings.LastIndexByte(baseURL, ':')

	schemaIdx := strings.Index(baseURL, "://")
	hasSchema := schemaIdx >= 0
	hasPort := portIdx > schemaIdx+1

	var port = "80"
	if hasPort {
		port = baseURL[portIdx+1:]
	}

	// find the schema based on the port.
	if !hasSchema {
		if port == "443" {
			baseURL = "https://" + baseURL
		} else {
			baseURL = "http://" + baseURL
		}
	} else if !hasPort {
		// has schema but not port.
		if strings.HasPrefix(baseURL, "https://") {
			port = "443"
		}
	}

	// finally, append the port part if it wasn't there.
	if !hasPort {
		baseURL += ":" + port
	}

	return baseURL
}
