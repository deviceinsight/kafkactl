package protobuf

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/bufbuild/protocompile"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/pkg/errors"
	"github.com/riferrei/srclient"

	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/output"
)

func SchemaToFileDescriptor(registry srclient.ISchemaRegistryClient, schema *srclient.Schema) (protoreflect.FileDescriptor, error) {
	dependencies, err := resolveSchemaDependencies(registry, schema)
	if err != nil {
		return nil, err
	}
	dependencies["."] = schema.Schema()

	return ParseFileDescriptor(".", dependencies)
}

func resolveSchemaDependencies(registry srclient.ISchemaRegistryClient, schema *srclient.Schema) (map[string]string, error) {
	dependencies := map[string]string{}
	err := resolveRefsDependencies(registry, dependencies, schema.References())
	if err != nil {
		return nil, err
	}
	return dependencies, nil
}

func resolveRefsDependencies(registry srclient.ISchemaRegistryClient, resolved map[string]string, references []srclient.Reference) error {
	for _, r := range references {
		if _, ok := resolved[r.Name]; ok {
			continue
		}
		latest, err := registry.GetSchemaByVersion(r.Subject, r.Version)
		if err != nil {
			return err
		}
		resolved[r.Name] = latest.Schema()
		err = resolveRefsDependencies(registry, resolved, latest.References())
		if err != nil {
			return err
		}
	}

	return nil
}

func ParseFileDescriptor(filename string, resolvedSchemas map[string]string) (protoreflect.FileDescriptor, error) {
	resolver := protocompile.SourceResolver{Accessor: protocompile.SourceAccessorFromMap(resolvedSchemas)}

	compiler := protocompile.Compiler{
		Resolver: protocompile.WithStandardImports(&resolver),
	}

	parsedFiles, err := compiler.Compile(context.Background(), filename)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't parse file descriptor")
	}

	return parsedFiles[0].ParentFile(), nil
}

func ComputeIndexes(fileDesc protoreflect.FileDescriptor, msgName protoreflect.FullName) ([]int64, error) {
	result := make([]int64, 0, 16)
	if fileDesc.Messages().Len() > 0 && fileDesc.Messages().Get(0).FullName() == msgName {
		return result, nil
	}

	messages := fileDesc.Messages()
	found := false

	for {
		found = false
		for i := range messages.Len() {
			message := messages.Get(i)
			if message.FullName() == msgName {
				return append(result, int64(i)), nil
			}

			if strings.Contains(string(msgName), string(message.FullName()+".")) {
				found = true
				result = append(result, int64(i))
				messages = message.Messages()
				break
			}
		}
		if !found {
			return nil, errors.Errorf("can't compute indexes for %s", msgName)
		}
	}
}

func FindMessageDescriptor(fileDesc protoreflect.FileDescriptor, msgName protoreflect.FullName) (protoreflect.MessageDescriptor, error) {
	return findMessageDescriptor(fileDesc.Messages(), msgName)
}

func findMessageDescriptor(messages protoreflect.MessageDescriptors, msgName protoreflect.FullName) (protoreflect.MessageDescriptor, error) {
	for i := range messages.Len() {
		message := messages.Get(i)

		if message.FullName() == msgName {
			return message, nil
		}

		if strings.HasPrefix(string(msgName), string(message.FullName()+".")) {
			return findMessageDescriptor(message.Messages(), msgName)
		}
	}

	return nil, errors.Errorf("can't find message descriptor for %s", msgName)
}

func ResolveMessageType(protobufConfig internal.ProtobufConfig, typeName protoreflect.FullName) protoreflect.MessageDescriptor {
	descSets := makeDescriptors(protobufConfig)
	return resolveMessageType(descSets, typeName)
}

func resolveMessageType(descSets []protoreflect.FileDescriptor, typeName protoreflect.FullName) protoreflect.MessageDescriptor {
	// make sure we use the shortname for using ByName:
	shortName := typeName.Name()
	var candidates []protoreflect.MessageDescriptor
	for _, descriptor := range descSets {
		if msg := descriptor.Messages().ByName(shortName); msg != nil {
			candidates = append(candidates, msg)
		}
	}

	// Prefer matching FullName first:
	for _, candidate := range candidates {
		if candidate.FullName() == typeName {
			return candidate
		}
	}

	// Fallback to matching short Name:
	for _, candidate := range candidates {
		if string(candidate.Name()) == string(typeName) {
			return candidate
		}
	}
	return nil
}

func makeDescriptors(protobufConfig internal.ProtobufConfig) []protoreflect.FileDescriptor {
	var ret []protoreflect.FileDescriptor

	ret = readFileDescriptors(protobufConfig.ProtosetFiles)
	importPaths := slices.Clone(protobufConfig.ProtoImportPaths)

	// extend import paths with existing files directories
	// this allows to specify only proto file path
	var protoFiles []string
	for _, protoFile := range protobufConfig.ProtoFiles {
		if existingFile := getExistingAbsoluteFile(protoFile); existingFile != "" {
			filename := filepath.Base(existingFile)
			path := filepath.Dir(existingFile)
			importPaths = append(importPaths, path)
			protoFiles = append(protoFiles, filename)
		} else {
			protoFiles = append(protoFiles, protoFile)
		}
	}

	resolver := protocompile.WithStandardImports(&protocompile.SourceResolver{ImportPaths: importPaths})

	compiler := protocompile.Compiler{
		Resolver: resolver,
	}

	parsedFiles, err := compiler.Compile(context.Background(), protoFiles...)
	if err != nil {
		output.Warnf("Proto compile error: %v", err)
		return ret
	}

	for _, protoFile := range parsedFiles {
		ret = append(ret, protoFile.ParentFile())
	}

	return ret
}

func readFileDescriptors(protoSetFiles []string) []protoreflect.FileDescriptor {
	var descriptors []protoreflect.FileDescriptor

	for _, protoSetFile := range protoSetFiles {
		var files descriptorpb.FileDescriptorSet

		b, err := os.ReadFile(protoSetFile)
		if err != nil {
			output.Warnf("Read protoset file %s failed: %s", protoSetFile, err)
			continue
		}

		if err = proto.Unmarshal(b, &files); err != nil {
			output.Warnf("Parse protoset file %s failed: %s", protoSetFile, err)
			continue
		}

		ff, err := protodesc.NewFiles(&files)
		if err != nil {
			output.Warnf("Process protoset file %s failed: %s", protoSetFile, err)
			continue
		}

		ff.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
			descriptors = append(descriptors, fd)
			return true
		})
	}

	return descriptors
}

func getExistingAbsoluteFile(protoFile string) string {
	_, err := os.Stat(protoFile)
	if err != nil {
		return ""
	}

	abs, err := filepath.Abs(protoFile)
	if err != nil {
		return ""
	}

	return abs
}
