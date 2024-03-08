package util_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/deviceinsight/kafkactl/internal/util"
)

func TestParseOffsets(t *testing.T) {

	type testCases struct {
		description string
		input       []string
		wantOffsets map[int32]int64
		wantErr     string
	}

	for _, test := range []testCases{
		{
			description: "successful_parsing",
			input:       []string{"1=222", "2=333", "5=444"},
			wantOffsets: map[int32]int64{1: 222, 2: 333, 5: 444},
		},
		{
			description: "wrong_separator_fails",
			input:       []string{"1:222"},
			wantErr:     "offset parameter has wrong format: 1:222 [1:222]",
		},
		{
			description: "partition_not_an_int_fails",
			input:       []string{"abc=222"},
			wantErr:     "parsing \"abc\": invalid syntax",
		},
		{
			description: "offset_not_an_int_fails",
			input:       []string{"1=nope"},
			wantErr:     "parsing \"nope\": invalid syntax",
		},
	} {
		t.Run(test.description, func(t *testing.T) {

			offsets, err := util.ParseOffsets(test.input)

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

			if eq := reflect.DeepEqual(test.wantOffsets, offsets); !eq {
				t.Errorf("want %q got %q", test.wantOffsets, offsets)
			}
		})
	}
}
