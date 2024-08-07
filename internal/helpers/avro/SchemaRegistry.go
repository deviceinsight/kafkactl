package avro

import "github.com/riferrei/srclient"

func CreateSchemaRegistryClient(avroSchemaRegistry string) srclient.ISchemaRegistryClient {
	return srclient.CreateSchemaRegistryClient(avroSchemaRegistry)
}
