package protobuf

import (
	"os"
	"path/filepath"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/deviceinsight/kafkactl/v5/internal/output"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
)

type SearchContext struct {
	ProtosetFiles    []string
	ProtoFiles       []string
	ProtoImportPaths []string
}

func ResolveMessageType(context SearchContext, typeName string) *desc.MessageDescriptor {
	for _, descriptor := range makeDescriptors(context) {
		if msg := descriptor.FindMessage(typeName); msg != nil {
			return msg
		}
	}

	return nil
}

func makeDescriptors(context SearchContext) []*desc.FileDescriptor {
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
