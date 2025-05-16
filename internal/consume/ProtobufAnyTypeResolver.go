package consume

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/known/emptypb"
)

// An implementation of protojson's Resolver interface.
// Overriding the default resolver allows to ignore unknown protobuf.Any fields
type ignoreUnrecognizedAny struct {
	protoregistry.ExtensionTypeResolver
}

func (*ignoreUnrecognizedAny) FindMessageByName(name protoreflect.FullName) (protoreflect.MessageType, error) {
	if result, err := protoregistry.GlobalTypes.FindMessageByName(name); err == nil {
		return result, nil
	}
	return (&emptypb.Empty{}).ProtoReflect().Type(), nil
}

func (*ignoreUnrecognizedAny) FindMessageByURL(url string) (protoreflect.MessageType, error) {
	if result, err := protoregistry.GlobalTypes.FindMessageByURL(url); err == nil {
		return result, nil
	}
	return (&emptypb.Empty{}).ProtoReflect().Type(), nil
}
