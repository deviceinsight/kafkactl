package consume

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/output"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/riferrei/srclient"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
)

type RegistryProtobufMessageDeserializer struct {
	registry *internal.CachingSchemaRegistry
}

func (deserializer RegistryProtobufMessageDeserializer) canDeserialize(consumerMsg *sarama.ConsumerMessage, data []byte) bool {
	schemaID, err := deserializer.registry.ExtractSchemaID(data)
	if err == nil {
		schema, schemaErr := deserializer.registry.GetSchema(schemaID)
		if schemaErr != nil {
			output.Debugf("schema not found. id=%d partition=%d, offset=%d error=%v", schemaID,
				consumerMsg.Partition, consumerMsg.Offset, err)
			return false
		}
		isProtobufSchema := *schema.SchemaType() == srclient.Protobuf
		return isProtobufSchema
	}
	output.Debugf("failed to extract schema id. partition=%d, offset=%d error=%v", consumerMsg.Partition,
		consumerMsg.Offset, err)
	return false
}

func (deserializer RegistryProtobufMessageDeserializer) CanDeserializeValue(msg *sarama.ConsumerMessage, flags Flags) bool {
	return deserializer.canDeserialize(msg, msg.Value)
}

func (deserializer RegistryProtobufMessageDeserializer) CanDeserializeKey(msg *sarama.ConsumerMessage, flags Flags) bool {
	return deserializer.canDeserialize(msg, msg.Key)
}

func (deserializer RegistryProtobufMessageDeserializer) DeserializeValue(msg *sarama.ConsumerMessage) (*DeserializedData, error) {
	return deserializer.deserialize(msg.Value)
}

func (deserializer RegistryProtobufMessageDeserializer) DeserializeKey(msg *sarama.ConsumerMessage) (*DeserializedData, error) {
	return deserializer.deserialize(msg.Key)
}

func (deserializer RegistryProtobufMessageDeserializer) deserialize(rawData []byte) (*DeserializedData, error) {
	schemaID, err := deserializer.registry.ExtractSchemaID(rawData)
	schema, err := deserializer.registry.GetSchema(schemaID)
	if err != nil {
		return nil, err
	}
	output.Debugf("fetched schema %d", schemaID)
	resolvedSchemas, err := deserializer.resolveDependencies(schema.References())
	if err != nil {
		return nil, err
	}
	resolvedSchemas["."] = schema.Schema()
	fileDesc, err := deserializer.schemaToFileDescriptor(schema)
	if err != nil {
		return nil, err
	}
	indexes, bytes, err := readIndexes(rawData[5:])
	if err != nil {
		return nil, err
	}
	messageDescriptor, err := findMessageDescriptor(indexes, fileDesc.GetMessageTypes())
	if err != nil {
		return nil, err
	}
	dynmsg := dynamicpb.NewMessage(messageDescriptor.UnwrapMessage())
	err = proto.Unmarshal(rawData[5+bytes:], dynmsg)
	if err != nil {
		return nil, err
	}

	asJsonBytes, err := protojson.MarshalOptions{
		Multiline:       false,
		EmitUnpopulated: true,
		AllowPartial:    true,
		Resolver:        &ignoreUnrecognizedAny{protoregistry.GlobalTypes},
	}.Marshal(dynmsg)
	if err != nil {
		return nil, err
	}

	schemaString := schema.Schema()

	return &DeserializedData{
		data:     asJsonBytes,
		schema:   schemaString,
		schemaID: &schemaID,
	}, nil
}

func readIndexes(rawData []byte) ([]int64, int, error) {
	length, bytes := binary.Varint(rawData)
	indexes := []int64{}
	if bytes < 0 {
		return nil, 0, fmt.Errorf("can't read varint indexes length: %d", bytes)
	}
	if length == 0 {
		return []int64{0}, bytes, nil
	}
	for i := range int(length) {
		index, readBytes := binary.Varint(rawData[bytes:])
		if readBytes < 0 {
			return nil, 0, fmt.Errorf("can't read varint number %d", i)
		}
		indexes = append(indexes, index)
		bytes += readBytes
	}

	return indexes, bytes, nil
}

func findMessageDescriptor(indexes []int64, descriptors []*desc.MessageDescriptor) (*desc.MessageDescriptor, error) {
	if len(indexes) == 0 {
		return nil, errors.New("indexes can't be empty")
	} else if len(indexes) == 1 {
		return descriptors[indexes[0]], nil
	} else {
		return findMessageDescriptor(indexes[1:], descriptors)
	}
}

func (deserializer RegistryProtobufMessageDeserializer) schemaToFileDescriptor(schema *srclient.Schema) (*desc.FileDescriptor, error) {
	dependencies, err := deserializer.resolveDependencies(schema.References())
	if err != nil {
		return nil, err
	}
	dependencies["."] = schema.Schema()

	return parseFileDescriptor(".", dependencies)
}

func (deserializer RegistryProtobufMessageDeserializer) resolveDependencies(references []srclient.Reference) (map[string]string, error) {
	resolved := map[string]string{}
	for _, r := range references {
		latest, err := deserializer.registry.GetLatestSchema(r.Subject)
		if err != nil {
			return map[string]string{}, err
		}
		resolved[r.Subject] = latest.Schema()
	}

	return resolved, nil
}

func parseFileDescriptor(filename string, resolvedSchemas map[string]string) (*desc.FileDescriptor, error) {
	parser := protoparse.Parser{Accessor: protoparse.FileContentsFromMap(resolvedSchemas)}
	parsedFiles, err := parser.ParseFiles(filename)
	if err != nil {
		return nil, err
	}
	return parsedFiles[0], nil
}
