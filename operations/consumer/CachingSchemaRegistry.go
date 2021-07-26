package consumer

import (
	schemaregistry "github.com/landoop/schema-registry"
	"github.com/pkg/errors"
)

type CachingSchemaRegistry struct {
	subjects []string
	schemas  map[int]string
	client   *schemaregistry.Client
}

func CreateCachingSchemaRegistry(avroSchemaRegistry string) (*CachingSchemaRegistry, error) {

	var err error

	registry := &CachingSchemaRegistry{}

	registry.schemas = make(map[int]string)
	registry.client, err = schemaregistry.NewClient(avroSchemaRegistry)

	if err != nil {
		return registry, errors.Wrap(err, "failed to create schema registry registry: ")
	}

	return registry, nil
}

func (registry *CachingSchemaRegistry) Subjects() ([]string, error) {
	var err error = nil

	if len(registry.subjects) == 0 {
		registry.subjects, err = registry.client.Subjects()
	}

	return registry.subjects, err
}

func (registry *CachingSchemaRegistry) GetSchemaByID(id int) (string, error) {
	var err error = nil

	if _, ok := registry.schemas[id]; !ok {
		var schema string
		schema, err = registry.client.GetSchemaByID(id)
		if err == nil && schema != "" {
			registry.schemas[id] = schema
		}
	}

	return registry.schemas[id], err
}
