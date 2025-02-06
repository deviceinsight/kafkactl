package internal

import (
	"net/http"
	"time"

	"github.com/deviceinsight/kafkactl/v5/internal/helpers/avro"
	"github.com/deviceinsight/kafkactl/v5/internal/output"
	"github.com/pkg/errors"
	"github.com/riferrei/srclient"
)

type CachingSchemaRegistry struct {
	subjects      []string
	schemas       map[int]string
	latestSchemas map[string]*srclient.Schema
	client        srclient.ISchemaRegistryClient
}

func CreateCachingSchemaRegistry(context *ClientContext) (*CachingSchemaRegistry, error) {

	timeout := context.Avro.RequestTimeout

	if context.Avro.RequestTimeout <= 0 {
		timeout = 5 * time.Second
	}

	httpClient := &http.Client{Timeout: timeout}

	if context.Avro.TLS.Enabled {
		output.Debugf("avro TLS is enabled.")

		tlsConfig, err := setupTLSConfig(context.Avro.TLS)
		if err != nil {
			return nil, errors.Wrap(err, "failed to setup avro tls config")
		}

		httpClient.Transport = &http.Transport{
			TLSClientConfig: tlsConfig,
		}
	}

	baseURL := avro.FormatBaseURL(context.Avro.SchemaRegistry)
	client := srclient.CreateSchemaRegistryClientWithOptions(baseURL, httpClient, 16)

	if context.Avro.Username != "" {
		output.Debugf("avro BasicAuth is enabled.")
		client.SetCredentials(context.Avro.Username, context.Avro.Password)
	}
	return &CachingSchemaRegistry{
		client:        client,
		schemas:       make(map[int]string),
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

func (registry *CachingSchemaRegistry) GetSchemaByID(id int) (string, error) {
	var err error

	if _, ok := registry.schemas[id]; !ok {
		var schema *srclient.Schema
		schema, err = registry.client.GetSchema(id)
		if err == nil {
			registry.schemas[id] = schema.Schema()
		}
	}

	return registry.schemas[id], err
}
