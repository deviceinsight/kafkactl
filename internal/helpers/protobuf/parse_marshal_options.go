package protobuf

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/deviceinsight/kafkactl/v5/internal"
)

const separator = "="

const optionAllowPartial = "allowpartial"
const optionUseProtoNames = "useprotonames"
const optionUseEnumNumbers = "useenumnumbers"
const optionEmitUnpopulated = "emitunpopulated"
const optionEmitDefaultValues = "emitdefaultvalues"

var AllMarshalOptions = []string{optionAllowPartial, optionUseProtoNames,
	optionUseEnumNumbers, optionEmitDefaultValues}

func ParseMarshalOptions(rawOptions []string, options internal.ProtobufMarshalOptions) (internal.ProtobufMarshalOptions, error) {

	var err error

	for _, optionFlag := range rawOptions {
		optionParts := strings.SplitN(optionFlag, separator, 2)
		optionKey := strings.ReplaceAll(strings.ReplaceAll(strings.ToLower(optionParts[0]), "-", ""), "_", "")

		switch optionKey {
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

func parseBooleanValue(parts []string) (bool, error) {

	if len(parts) < 2 {
		return true, nil
	}

	return strconv.ParseBool(parts[1])
}
