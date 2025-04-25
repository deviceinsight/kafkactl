package protobuf_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/helpers/protobuf"
)

func TestParseMarshalOptions(t *testing.T) {

	type testCases struct {
		description     string
		input           []string
		existingOptions internal.ProtobufMarshalOptions
		wantOptions     internal.ProtobufMarshalOptions
		wantErr         string
	}

	for _, test := range []testCases{
		{
			description: "successful_parsing",
			input: []string{
				"allowPartial=true",
				"useprotonames",
				"useEnumNumbers=true",
				"EmitUnpopulated",
				"emitdefaultvalues=1",
			},
			wantOptions: internal.ProtobufMarshalOptions{
				AllowPartial:      true,
				UseProtoNames:     true,
				UseEnumNumbers:    true,
				EmitUnpopulated:   true,
				EmitDefaultValues: true,
			},
		},
		{
			description: "invalidate_existing_options",
			input: []string{
				"allowPartial=false",
				"useprotonames=false",
				"useEnumNumbers=0",
				"EmitUnpopulated=false",
				"emitdefaultvalues=0",
			},
			existingOptions: internal.ProtobufMarshalOptions{
				AllowPartial:      true,
				UseProtoNames:     true,
				UseEnumNumbers:    true,
				EmitUnpopulated:   true,
				EmitDefaultValues: true,
			},
			wantOptions: internal.ProtobufMarshalOptions{},
		},
		{
			description: "wrong_separator_fails",
			input:       []string{"allowpartial:true"},
			wantErr:     "unknown option \"allowpartial:true\"",
		},
		{
			description: "wrong_boolean_value_fails",
			input:       []string{"allowPartial=not_a_bool"},
			wantErr:     "parsing \"not_a_bool\": invalid syntax",
		},
	} {
		t.Run(test.description, func(t *testing.T) {

			options, err := protobuf.ParseMarshalOptions(test.input, test.existingOptions)

			if test.wantErr != "" {
				if err == nil {
					t.Errorf("want error %q but got nil", test.wantErr)
				}

				if !strings.Contains(err.Error(), test.wantErr) {
					t.Errorf("want error %q got %q", test.wantErr, err)
				}

				return
			}
			if err != nil {
				t.Errorf("doesn't want error but got %s", err)
			}

			if eq := reflect.DeepEqual(test.wantOptions, options); !eq {
				t.Errorf("want %v got %v", test.wantOptions, options)
			}
		})
	}
}
