package schemaregistry

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/pkg/errors"
	"github.com/riferrei/srclient"
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

func SchemaToFileDescriptor(registry srclient.ISchemaRegistryClient, schema *srclient.Schema) (*desc.FileDescriptor, error) {
	dependencies, err := resolveDependencies(registry, schema.References())
	if err != nil {
		return nil, err
	}
	dependencies["."] = schema.Schema()

	return ParseFileDescriptor(".", dependencies)
}

func resolveDependencies(registry srclient.ISchemaRegistryClient, references []srclient.Reference) (map[string]string, error) {
	resolved := map[string]string{}
	for _, r := range references {
		latest, err := registry.GetLatestSchema(r.Subject)
		if err != nil {
			return map[string]string{}, errors.Wrap(err, fmt.Sprintf("couldn't fetch latest schema for subject %s", r.Subject))
		}
		resolved[r.Subject] = latest.Schema()
	}

	return resolved, nil
}

func ParseFileDescriptor(filename string, resolvedSchemas map[string]string) (*desc.FileDescriptor, error) {
	parser := protoparse.Parser{Accessor: protoparse.FileContentsFromMap(resolvedSchemas)}
	parsedFiles, err := parser.ParseFiles(filename)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't parse file descriptor")
	}
	return parsedFiles[0], nil
}
