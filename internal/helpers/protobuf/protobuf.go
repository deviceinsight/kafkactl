package protobuf

import (
	"io/ioutil"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/deviceinsight/kafkactl/output"
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

	absFiles, err := protoparse.ResolveFilenames(context.ProtoImportPaths, context.ProtoFiles...)
	if err != nil {
		output.Warnf("Resolve proto files failed: %s", err)
		return ret
	}

	protoFiles, err := (&protoparse.Parser{
		ImportPaths:      append([]string{"."}, context.ProtoImportPaths...),
		InferImportPaths: true,
		ErrorReporter: func(err protoparse.ErrorWithPos) error {
			output.Warnf("Proto parser error [%s]: %s", err.GetPosition(), err)
			return nil
		},
		WarningReporter: func(pos protoparse.ErrorWithPos) {
			output.Warnf("Proto parse warning: %s", err)
		},
	}).ParseFiles(absFiles...)
	if err != nil {
		output.Warnf("Proto files parse error: %s", err)
	}

	ret = append(ret, protoFiles...)

	return ret
}

func appendProtosets(descs []*desc.FileDescriptor, protosetFiles []string) []*desc.FileDescriptor {
	for _, protosetFile := range protosetFiles {
		var files descriptorpb.FileDescriptorSet

		b, err := ioutil.ReadFile(protosetFile)
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
