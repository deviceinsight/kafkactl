package internal

import (
	"encoding/binary"
	"fmt"
	"net/http"
	"time"

	"github.com/deviceinsight/kafkactl/v5/internal/helpers/avro"
	"github.com/deviceinsight/kafkactl/v5/internal/output"
	"github.com/pkg/errors"
	"github.com/riferrei/srclient"
)

const WireFormatBytes = 5

type CachingSchemaRegistry struct {
	subjects      []string
	schemas       map[int]*srclient.Schema
	latestSchemas map[string]*srclient.Schema
	client        srclient.ISchemaRegistryClient
}

func CreateCachingSchemaRegistry(context *ClientContext) (*CachingSchemaRegistry, error) {

	timeout := context.SchemaRegistry.RequestTimeout

	if context.SchemaRegistry.RequestTimeout <= 0 {
		timeout = 5 * time.Second
	}

	httpClient := &http.Client{Timeout: timeout}

	if context.SchemaRegistry.TLS.Enabled {
		output.Debugf("schemaRegistry TLS is enabled.")

		tlsConfig, err := setupTLSConfig(context.SchemaRegistry.TLS)
		if err != nil {
			return nil, errors.Wrap(err, "failed to setup schemaRegistry tls config")
		}

		httpClient.Transport = &http.Transport{
			TLSClientConfig: tlsConfig,
		}
	}

	baseURL := avro.FormatBaseURL(context.SchemaRegistry.URL)
	client := srclient.NewSchemaRegistryClient(baseURL, srclient.WithClient(httpClient),
		srclient.WithSemaphoreWeight(int64(16)))

	if context.SchemaRegistry.Username != "" {
		output.Debugf("schemaRegistry BasicAuth is enabled.")
		client.SetCredentials(context.SchemaRegistry.Username, context.SchemaRegistry.Password)
	}
	return &CachingSchemaRegistry{
		client:        client,
		schemas:       make(map[int]*srclient.Schema),
		latestSchemas: make(map[string]*srclient.Schema),
	}, nil

}

func (registry *CachingSchemaRegistry) GetSchemaByVersion(subject string, schemaVersion int) (*srclient.Schema, error) {
	return registry.client.GetSchemaByVersion(subject, schemaVersion)
}

func (registry *CachingSchemaRegistry) GetLatestSchema(subject string) (*srclient.Schema, error) {
	var err error

	if _, ok := registry.latestSchemas[subject]; !ok {
		var schema *srclient.Schema
		schema, err = registry.client.GetLatestSchema(subject)
		if err == nil {
			registry.latestSchemas[subject] = schema
		}
	}

	return registry.latestSchemas[subject], err
}

func (registry *CachingSchemaRegistry) Subjects() ([]string, error) {
	var err error

	if len(registry.subjects) == 0 {
		registry.subjects, err = registry.client.GetSubjects()
	}

	return registry.subjects, err
}

func (registry *CachingSchemaRegistry) GetSchemaByID(id int) (*srclient.Schema, error) {
	var err error

	if _, ok := registry.schemas[id]; !ok {
		var schema *srclient.Schema
		schema, err = registry.client.GetSchema(id)
		if err == nil {
			registry.schemas[id] = schema
		}
	}

	return registry.schemas[id], err
}

func (registry *CachingSchemaRegistry) ExtractSchemaID(data []byte) (int, error) {
	if len(data) < WireFormatBytes {
		return 0, fmt.Errorf("data too short. cannot extra schema id from message (len = %d)", len(data))
	}

	if data[0] != 0 {
		return 0, fmt.Errorf("confluent serialization format version number was %d != 0", data[0])
	}

	// https://docs.confluent.io/cloud/current/sr/fundamentals/serdes-develop/index.html#messages-wire-format
	id := int(binary.BigEndian.Uint32(data[1:WireFormatBytes]))

	if id == 0 {
		return 0, fmt.Errorf("schema id is 0")
	}
	return id, nil
}

func (registry *CachingSchemaRegistry) ExtractPayload(data []byte) []byte {
	return data[WireFormatBytes:]
}
