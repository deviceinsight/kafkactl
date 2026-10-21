package avro

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/linkedin/goavro/v2"
	"github.com/pkg/errors"
)

const (
	httpRequestTimeout = 5 * time.Second

	// safety-guard: read at most 1MB
	maxSchemaBytes   = 1 << 20
	maxCachedSchemas = 1024
)

// cachedSchema memoizes a resolution attempt. Failures are cached too: without
// --ignore-errors they abort anyway, and with it an unresolvable location would
// otherwise be retried, and time out, once per message.
type cachedSchema struct {
	codec *goavro.Codec
	err   error
}

// SchemaSource resolves an avro schema from a local file path or an http(s) URL into a
// compiled codec, caching the result by location for the process lifetime.
type SchemaSource struct {
	jsonCodec  JSONCodec
	httpClient *http.Client
	cache      *lru.Cache[string, cachedSchema]
}

func NewSchemaSource(jsonCodec JSONCodec) *SchemaSource {
	cache, _ := lru.New[string, cachedSchema](maxCachedSchemas)

	return &SchemaSource{
		jsonCodec:  jsonCodec,
		httpClient: &http.Client{Timeout: httpRequestTimeout},
		cache:      cache,
	}
}

func (source *SchemaSource) Resolve(location string) (*goavro.Codec, error) {
	if cached, ok := source.cache.Get(location); ok {
		return cached.codec, cached.err
	}

	codec, err := source.resolve(location)
	source.cache.Add(location, cachedSchema{codec: codec, err: err})
	return codec, err
}

func (source *SchemaSource) resolve(location string) (*goavro.Codec, error) {
	schema, err := source.fetch(location)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to resolve avro schema from %s", location)
	}

	codec, err := CompileCodec(schema, source.jsonCodec)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse avro schema from %s", location)
	}

	return codec, nil
}

func (source *SchemaSource) fetch(location string) (string, error) {
	switch {
	case strings.HasPrefix(location, "http://") || strings.HasPrefix(location, "https://"):
		return source.fetchURL(location)
	case strings.HasPrefix(location, "file://"):
		path, err := fileURIPath(location)
		if err != nil {
			return "", err
		}
		return source.fetchFile(path)
	default:
		return source.fetchFile(location)
	}
}

func fileURIPath(location string) (string, error) {
	fileURL, err := url.Parse(location)
	if err != nil {
		return "", err
	}

	if fileURL.Host != "" && fileURL.Host != "localhost" {
		return "", errors.Errorf("unsupported file uri host %q", fileURL.Host)
	}

	path := fileURL.Path
	if len(path) >= 3 && path[0] == '/' && path[2] == ':' {
		// windows drive letter, e.g. file:///C:/schemas/x.avsc
		path = path[1:]
	}

	return path, nil
}

// IsSchemaURI reports whether location is an absolute uri supported by Resolve (http, https or file),
// as opposed to a plain file path.
func IsSchemaURI(location string) bool {
	for _, scheme := range []string{"http://", "https://", "file://"} {
		if strings.HasPrefix(location, scheme) {
			return true
		}
	}
	return false
}

func (source *SchemaSource) fetchURL(location string) (string, error) {
	resp, err := source.httpClient.Get(location) //nolint:gosec,noctx
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.Errorf("unexpected status code %d", resp.StatusCode)
	}

	return readLimited(resp.Body)
}

func (source *SchemaSource) fetchFile(location string) (string, error) {
	file, err := os.Open(location)
	if err != nil {
		return "", err
	}
	defer file.Close()

	return readLimited(file)
}

// readLimited reads a schema, refusing anything larger than maxSchemaBytes rather than
// buffering whatever the location happens to serve.
func readLimited(reader io.Reader) (string, error) {
	body, err := io.ReadAll(io.LimitReader(reader, maxSchemaBytes+1))
	if err != nil {
		return "", err
	}

	if len(body) > maxSchemaBytes {
		return "", errors.Errorf("schema is larger than the maximum of %d bytes", maxSchemaBytes)
	}

	return string(body), nil
}
