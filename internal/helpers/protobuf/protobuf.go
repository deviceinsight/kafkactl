package protobuf

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/output"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/pkg/errors"
	"github.com/riferrei/srclient"
)

func SchemaToFileDescriptor(registry srclient.ISchemaRegistryClient, schema *srclient.Schema) (*desc.FileDescriptor, error) {
	dependencies, err := resolveDependencies(registry, schema.References())
	if err != nil {
		return nil, err
	}
	dependencies["."] = schema.Schema()

	return ParseFileDescriptor(".", dependencies)
}

func resolveDependencies(registry srclient.ISchemaRegistryClient, references []srclient.Reference) (map[string]string, error) {
	resolved := map[string]string{}
	for _, r := range references {
		refSchema, err := registry.GetSchemaByVersion(r.Subject, r.Version)
		if err != nil {
			return map[string]string{}, errors.Wrap(err, fmt.Sprintf("couldn't fetch latest schema for subject %s", r.Subject))
		}
		resolved[r.Subject] = refSchema.Schema()
	}

	return resolved, nil
}

func ParseFileDescriptor(filename string, resolvedSchemas map[string]string) (*desc.FileDescriptor, error) {
	parser := protoparse.Parser{Accessor: protoparse.FileContentsFromMap(resolvedSchemas)}
	parsedFiles, err := parser.ParseFiles(filename)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't parse file descriptor")
	}
	return parsedFiles[0], nil
}

func ComputeIndexes(fileDesc *desc.FileDescriptor, msgName string) ([]int64, error) {
	result := make([]int64, 0, 16)
	if len(fileDesc.GetMessageTypes()) > 0 && fileDesc.GetMessageTypes()[0].GetFullyQualifiedName() == msgName {
		return result, nil
	}

	messages := fileDesc.GetMessageTypes()
	found := false

	for {
		found = false
		for i, message := range messages {
			if message.GetFullyQualifiedName() == msgName {
				return append(result, int64(i)), nil
			}

			if strings.Contains(msgName, message.GetFullyQualifiedName()+".") {
				found = true
				result = append(result, int64(i))
				messages = message.GetNestedMessageTypes()
				break
			}
		}
		if !found {
			return nil, errors.Errorf("can't compute indexes for %s", msgName)
		}
	}
}

func ResolveMessageType(context internal.ProtobufConfig, typeName string) *desc.MessageDescriptor {
	for _, descriptor := range makeDescriptors(context) {
		if msg := descriptor.FindMessage(typeName); msg != nil {
			return msg
		}
	}

	return nil
}

func makeDescriptors(context internal.ProtobufConfig) []*desc.FileDescriptor {
	var ret []*desc.FileDescriptor

	ret = appendProtosets(ret, context.ProtosetFiles)
	importPaths := append([]string{}, context.ProtoImportPaths...)

	// extend import paths with existing files directories
	// this allows to specify only proto file path
	for _, existingFile := range getExistingFiles(context.ProtoFiles) {
		importPaths = append(importPaths, filepath.Dir(existingFile))
	}

	resolvedFilenames, err := protoparse.ResolveFilenames(importPaths, context.ProtoFiles...)
	if err != nil {
		output.Warnf("Resolve proto files failed: %s", err)
		return ret
	}

	protoFiles, err := (&protoparse.Parser{
		ImportPaths:      importPaths,
		InferImportPaths: true,
		ErrorReporter: func(err protoparse.ErrorWithPos) error {
			output.Warnf("Proto parser error [%s]: %s", err.GetPosition(), err)
			return nil
		},
		WarningReporter: func(err protoparse.ErrorWithPos) {
			output.Warnf("Proto parse warning: %s", err)
		},
	}).ParseFiles(resolvedFilenames...)
	if err != nil {
		output.Warnf("Proto files parse error: %s", err)
	}

	ret = append(ret, protoFiles...)

	return ret
}

func appendProtosets(descs []*desc.FileDescriptor, protosetFiles []string) []*desc.FileDescriptor {
	for _, protosetFile := range protosetFiles {
		var files descriptorpb.FileDescriptorSet

		b, err := os.ReadFile(protosetFile)
		if err != nil {
			output.Warnf("Read protoset file %s failed: %s", protosetFile, err)
			continue
		}

		if err = proto.Unmarshal(b, &files); err != nil {
			output.Warnf("Parse protoset file %s failed: %s", protosetFile, err)
			continue
		}

		fds, err := desc.CreateFileDescriptorsFromSet(&files)
		if err != nil {
			output.Warnf("Convert file %s to descriptors failed: %s", protosetFile, err)
			continue
		}

		for _, fd := range fds {
			descs = append(descs, fd)
		}

	}

	return descs
}

func getExistingFiles(protoFiles []string) []string {
	var existing []string

	for _, protoFile := range protoFiles {
		_, err := os.Stat(protoFile)
		if err != nil {
			continue
		}

		abs, err := filepath.Abs(protoFile)
		if err != nil {
			continue
		}

		existing = append(existing, abs)
	}

	return existing
}
