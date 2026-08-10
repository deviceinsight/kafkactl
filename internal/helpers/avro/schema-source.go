package avro

import (
	"io"
	"net/http"
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
	if strings.HasPrefix(location, "http://") || strings.HasPrefix(location, "https://") {
		return source.fetchURL(location)
	}
	return source.fetchFile(location)
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
