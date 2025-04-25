package protobuf

import (
	"fmt"
	"github.com/deviceinsight/kafkactl/v5/internal"
	"strconv"
	"strings"
)

const separator = "="

const optionMultiline = "multiline"
const optionIndent = "indent"
const optionAllowPartial = "allowpartial"
const optionUseProtoNames = "useprotonames"
const optionUseEnumNumbers = "useenumnumbers"
const optionEmitUnpopulated = "emitunpopulated"
const optionEmitDefaultValues = "emitdefaultvalues"

var AllMarshalOptions = []string{optionMultiline, optionIndent, optionAllowPartial, optionUseProtoNames,
	optionUseEnumNumbers, optionEmitDefaultValues, optionEmitDefaultValues}

func ParseMarshalOptions(rawOptions []string) (internal.ProtobufMarshalOptions, error) {

	options := internal.ProtobufMarshalOptions{}
	var err error

	for _, optionFlag := range rawOptions {
		optionParts := strings.SplitN(optionFlag, separator, 2)
		optionKey := strings.ReplaceAll(strings.ReplaceAll(strings.ToLower(optionParts[0]), "-", ""), "_", "")

		switch optionKey {
		case optionMultiline:
			if options.Multiline, err = parseBooleanValue(optionParts); err != nil {
				return options, err
			}
		case optionIndent:
			if options.Indent, err = parseStringValue(optionParts); err != nil {
				return options, err
			}
		case optionAllowPartial:
			if options.AllowPartial, err = parseBooleanValue(optionParts); err != nil {
				return options, err
			}
		case optionUseProtoNames:
			if options.UseProtoNames, err = parseBooleanValue(optionParts); err != nil {
				return options, err
			}
		case optionUseEnumNumbers:
			if options.UseEnumNumbers, err = parseBooleanValue(optionParts); err != nil {
				return options, err
			}
		case optionEmitUnpopulated:
			if options.EmitUnpopulated, err = parseBooleanValue(optionParts); err != nil {
				return options, err
			}
		case optionEmitDefaultValues:
			if options.EmitDefaultValues, err = parseBooleanValue(optionParts); err != nil {
				return options, err
			}
		default:
			return options, fmt.Errorf("unknown option %q. known options are: %v", optionKey, AllMarshalOptions)
		}
	}

	return options, nil
}

func parseStringValue(parts []string) (string, error) {

	if len(parts) < 2 {
		return "", fmt.Errorf("no value provided")
	}

	return parts[1], nil
}

func parseBooleanValue(parts []string) (bool, error) {

	if len(parts) < 2 {
		return true, nil
	}

	return strconv.ParseBool(parts[1])
}
