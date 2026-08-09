package avro

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

const (
	maxSchemaResponseBytes = 1 << 20
	maxSchemaRedirects     = 10
	schemaRequestTimeout   = 5 * time.Second
)

type safeSchemaRequestError struct {
	message string
}

func (err *safeSchemaRequestError) Error() string {
	return err.message
}

// SchemaResolver loads and compiles Avro schemas from local files or HTTP(S)
// URLs. Successfully resolved URLs can be cached for the life of the resolver.
type SchemaResolver struct {
	jsonCodec    JSONCodec
	cacheEnabled bool
	cache        map[string]*MessageCodec
	cacheMutex   sync.Mutex
	httpClient   *http.Client
}

// NewSchemaResolver creates a resolver with an optional in-memory URL cache.
func NewSchemaResolver(jsonCodec JSONCodec, cacheEnabled bool) *SchemaResolver {
	return &SchemaResolver{
		jsonCodec:    jsonCodec,
		cacheEnabled: cacheEnabled,
		cache:        make(map[string]*MessageCodec),
		httpClient: &http.Client{
			Timeout:       schemaRequestTimeout,
			CheckRedirect: checkSchemaRedirect,
		},
	}
}

// ResolveSource treats HTTP(S) sources as URLs and all other sources as local
// paths. In particular, a Windows drive letter is not interpreted as a URL
// scheme.
func (resolver *SchemaResolver) ResolveSource(source string) (*MessageCodec, error) {
	trimmedSource := strings.TrimSpace(source)
	scheme, _, hasScheme := strings.Cut(trimmedSource, ":")
	if hasScheme && isHTTPScheme(scheme) {
		return resolver.ResolveURI(trimmedSource)
	}

	schema, err := os.ReadFile(source)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read avro schema file %q", source)
	}

	codec, err := NewMessageCodec(string(schema), resolver.jsonCodec)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load avro schema from file %q", source)
	}

	return codec, nil
}

// ResolveURI downloads and compiles an Avro schema from an HTTP(S) URL.
func (resolver *SchemaResolver) ResolveURI(rawURI string) (*MessageCodec, error) {
	trimmedURI := strings.TrimSpace(rawURI)
	parsedURI, err := parseSchemaURI(trimmedURI)
	if err != nil {
		return nil, err
	}

	if !resolver.cacheEnabled {
		return resolver.fetchSchema(parsedURI)
	}

	// URL fragments are not sent in HTTP requests and therefore must not create
	// separate entries for the same schema resource.
	cacheURI := *parsedURI
	cacheURI.Fragment = ""
	cacheURI.RawFragment = ""
	cacheKey := cacheURI.String()

	// Keep the lock while fetching so concurrent cache misses do not trigger
	// duplicate downloads. This also serializes misses for different URIs,
	// which keeps the command-scoped cache simple and deterministic.
	resolver.cacheMutex.Lock()
	defer resolver.cacheMutex.Unlock()

	if codec, found := resolver.cache[cacheKey]; found {
		return codec, nil
	}

	codec, err := resolver.fetchSchema(parsedURI)
	if err != nil {
		return nil, err
	}

	resolver.cache[cacheKey] = codec
	return codec, nil
}

func (resolver *SchemaResolver) fetchSchema(schemaURI *url.URL) (*MessageCodec, error) {
	displayURI := redactURI(schemaURI)

	request, err := http.NewRequest(http.MethodGet, schemaURI.String(), nil)
	if err != nil {
		return nil, errors.Errorf("failed to create avro schema request for %s", displayURI)
	}

	response, err := resolver.httpClient.Do(request)
	if err != nil {
		return nil, errors.Errorf("failed to fetch avro schema from %s: %s", displayURI, requestErrorDetail(err))
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, errors.Errorf("avro schema request to %s returned status %d", displayURI, response.StatusCode)
	}

	if response.ContentLength > maxSchemaResponseBytes {
		return nil, errors.Errorf("avro schema response from %s exceeds %d bytes", displayURI, maxSchemaResponseBytes)
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, maxSchemaResponseBytes+1))
	if err != nil {
		return nil, errors.Errorf("failed to read avro schema response from %s: %s", displayURI, requestErrorDetail(err))
	}
	if len(body) > maxSchemaResponseBytes {
		return nil, errors.Errorf("avro schema response from %s exceeds %d bytes", displayURI, maxSchemaResponseBytes)
	}

	codec, err := NewMessageCodec(string(body), resolver.jsonCodec)
	if err != nil {
		// The response may come from an address selected by a Kafka header. Do
		// not expose parser details that could quote the response body.
		return nil, errors.Errorf(
			"failed to load avro schema from %s: response is not a valid avro schema",
			displayURI,
		)
	}

	return codec, nil
}

func parseSchemaURI(rawURI string) (*url.URL, error) {
	if rawURI == "" {
		return nil, errors.New("avro schema URI must not be empty")
	}

	parsedURI, err := url.Parse(rawURI)
	if err != nil {
		return nil, errors.New("invalid avro schema URI")
	}

	if !isHTTPScheme(parsedURI.Scheme) {
		return nil, errors.Errorf(
			"unsupported avro schema URI scheme %q: only http and https are allowed",
			parsedURI.Scheme,
		)
	}
	if parsedURI.Hostname() == "" {
		return nil, errors.Errorf("avro schema URI %s must include a host", redactURI(parsedURI))
	}
	if parsedURI.User != nil {
		return nil, errors.Errorf("avro schema URI %s must not include user information", redactURI(parsedURI))
	}
	parsedURI.Scheme = strings.ToLower(parsedURI.Scheme)

	return parsedURI, nil
}

func isHTTPScheme(scheme string) bool {
	return strings.EqualFold(scheme, "http") || strings.EqualFold(scheme, "https")
}

func checkSchemaRedirect(request *http.Request, via []*http.Request) error {
	if len(via) >= maxSchemaRedirects {
		return &safeSchemaRequestError{message: fmt.Sprintf("stopped after %d redirects", maxSchemaRedirects)}
	}
	if !isHTTPScheme(request.URL.Scheme) {
		return &safeSchemaRequestError{message: "avro schema redirect target must use http or https"}
	}
	if request.URL.User != nil {
		return &safeSchemaRequestError{message: "avro schema redirect target must not include user information"}
	}
	if request.URL.Hostname() == "" {
		return &safeSchemaRequestError{message: "avro schema redirect target must include a host"}
	}
	request.URL.Scheme = strings.ToLower(request.URL.Scheme)

	// The resolver does not use authentication or cookies. Explicitly remove
	// sensitive headers that net/http may otherwise copy to a redirect request.
	request.Header.Del("Authorization")
	request.Header.Del("Proxy-Authorization")
	request.Header.Del("Cookie")
	request.Header.Del("Referer")

	return nil
}

func redactURI(schemaURI *url.URL) string {
	redactedURI := *schemaURI
	redactedURI.User = nil
	redactedURI.RawQuery = ""
	redactedURI.ForceQuery = false
	redactedURI.Fragment = ""
	return redactedURI.String()
}

func requestErrorDetail(err error) string {
	if errors.Is(err, context.DeadlineExceeded) {
		return "request timed out"
	}
	if errors.Is(err, context.Canceled) {
		return "request canceled"
	}

	var networkError net.Error
	if errors.As(err, &networkError) && networkError.Timeout() {
		return "request timed out"
	}

	var safeError *safeSchemaRequestError
	if errors.As(err, &safeError) {
		return safeError.Error()
	}

	// net/http errors can quote an untrusted Location header, including URL
	// credentials or signed query parameters. Keep unknown details out of user-
	// visible errors; the sanitized source URL is already included by the caller.
	return "request failed"
}
