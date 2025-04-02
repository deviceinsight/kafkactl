package internal

import (
	"encoding/binary"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/deviceinsight/kafkactl/v5/internal/helpers/schemaregistry"
	"github.com/deviceinsight/kafkactl/v5/internal/output"
	"github.com/pkg/errors"
	"github.com/riferrei/srclient"
)

const (
	WireFormatBytes = 5
	MagicByte       = 0
)

type CachingSchemaRegistry struct {
	srclient.ISchemaRegistryClient
	subjects []string
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

	baseURL := schemaregistry.FormatBaseURL(context.SchemaRegistry.URL)
	client := srclient.NewSchemaRegistryClient(baseURL, srclient.WithClient(httpClient),
		srclient.WithSemaphoreWeight(int64(16)))

	if context.SchemaRegistry.Username != "" {
		output.Debugf("schemaRegistry BasicAuth is enabled.")
		client.SetCredentials(context.SchemaRegistry.Username, context.SchemaRegistry.Password)
	}
	return &CachingSchemaRegistry{
		ISchemaRegistryClient: client,
	}, nil
}

func (registry *CachingSchemaRegistry) GetSubjects() ([]string, error) {
	var err error

	if len(registry.subjects) == 0 {
		registry.subjects, err = registry.ISchemaRegistryClient.GetSubjects()
	}

	return registry.subjects, err
}

func (registry *CachingSchemaRegistry) SubjectOfTypeExists(subject string, expectedSchemaType srclient.SchemaType) (bool, error) {
	subjects, err := registry.GetSubjects()
	if err != nil {
		return false, errors.Wrap(err, "failed to list available schemas")
	}

	if !slices.Contains(subjects, subject) {
		return false, nil
	}

	schema, err := registry.GetLatestSchema(subject)
	if err != nil {
		return false, errors.Wrap(err, fmt.Sprintf("failed to retrieve latest schema for subject %s", subject))
	}

	if schema.SchemaType() == nil && expectedSchemaType == srclient.Avro {
		return true, nil
	}

	if schema.SchemaType() == nil {
		return false, nil
	}

	if *schema.SchemaType() != expectedSchemaType {
		return false, nil
	}
	return true, nil
}

func (registry *CachingSchemaRegistry) ExtractSchemaID(data []byte) (int, error) {
	if len(data) < WireFormatBytes {
		return 0, fmt.Errorf("data too short. cannot extra schema id from message (len = %d)", len(data))
	}

	if data[0] != MagicByte {
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
