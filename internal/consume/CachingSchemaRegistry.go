package consume

import (
	"github.com/riferrei/srclient"
)

type CachingSchemaRegistry struct {
	subjects []string
	schemas  map[int]string
	client   srclient.ISchemaRegistryClient
}

func CreateCachingSchemaRegistry(client srclient.ISchemaRegistryClient) *CachingSchemaRegistry {
	return &CachingSchemaRegistry{client: client, schemas: make(map[int]string)}
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
