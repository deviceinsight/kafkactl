package avro

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/linkedin/goavro/v2"
	"github.com/pkg/errors"
)

const httpRequestTimeout = 5 * time.Second

// SchemaSource resolves an avro schema from a local file path or an http(s) URL into a
// compiled codec, caching the compiled codec by location for the process lifetime.
type SchemaSource struct {
	jsonCodec  JSONCodec
	httpClient *http.Client
	mutex      sync.Mutex
	cache      map[string]*goavro.Codec
}

func NewSchemaSource(jsonCodec JSONCodec) *SchemaSource {
	return &SchemaSource{
		jsonCodec:  jsonCodec,
		httpClient: &http.Client{Timeout: httpRequestTimeout},
		cache:      make(map[string]*goavro.Codec),
	}
}

func (source *SchemaSource) Resolve(location string) (*goavro.Codec, error) {
	source.mutex.Lock()
	codec, ok := source.cache[location]
	source.mutex.Unlock()
	if ok {
		return codec, nil
	}

	schema, err := source.fetch(location)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to resolve avro schema from %s", location)
	}

	codec, err = CompileCodec(schema, source.jsonCodec)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse avro schema from %s", location)
	}

	source.mutex.Lock()
	source.cache[location] = codec
	source.mutex.Unlock()

	return codec, nil
}

func (source *SchemaSource) fetch(location string) (string, error) {
	switch {
	case strings.HasPrefix(location, "http://") || strings.HasPrefix(location, "https://"):
		return source.fetchURL(location)
	case strings.HasPrefix(location, "file://"):
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
		return source.fetchFile(path)
	default:
		return source.fetchFile(location)
	}
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (source *SchemaSource) fetchFile(location string) (string, error) {
	body, err := os.ReadFile(location)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
