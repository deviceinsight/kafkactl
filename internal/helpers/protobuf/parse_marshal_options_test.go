package protobuf_test

import (
	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/helpers/protobuf"
	"reflect"
	"strings"
	"testing"
)

func TestParseMarshalOptions(t *testing.T) {

	type testCases struct {
		description string
		input       []string
		wantOptions internal.ProtobufMarshalOptions
		wantErr     string
	}

	for _, test := range []testCases{
		{
			description: "successful_parsing",
			input: []string{
				"multiline=true",
				"indent=    ",
				"allowPartial=true",
				"useprotonames",
				"useEnumNumbers=true",
				"EmitUnpopulated",
				"emitdefaultvalues=1",
			},
			wantOptions: internal.ProtobufMarshalOptions{
				Multiline:         true,
				Indent:            "    ",
				AllowPartial:      true,
				UseProtoNames:     true,
				UseEnumNumbers:    true,
				EmitUnpopulated:   true,
				EmitDefaultValues: true,
			},
		},
		{
			description: "wrong_separator_fails",
			input:       []string{"multiline:true"},
			wantErr:     "unknown option \"multiline:true\"",
		},
		{
			description: "indent_without_value_fails",
			input:       []string{"indent"},
			wantErr:     "no value provided",
		},
		{
			description: "wrong_boolean_value_fails",
			input:       []string{"multiline=not_a_bool"},
			wantErr:     "parsing \"not_a_bool\": invalid syntax",
		},
	} {
		t.Run(test.description, func(t *testing.T) {

			options, err := protobuf.ParseMarshalOptions(test.input)

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
