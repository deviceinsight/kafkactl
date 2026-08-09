package consume

import (
	"bytes"
	"testing"
)

func TestDecodeAMQPValue32BitLength(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  []byte
	}{
		{
			name:  "str32 zero length",
			input: []byte{0xb1, 0x00, 0x00, 0x00, 0x00},
			want:  []byte{},
		},
		{
			name:  "vbin32 zero length",
			input: []byte{0xb0, 0x00, 0x00, 0x00, 0x00},
			want:  []byte{},
		},
		{
			name:  "str32 value",
			input: []byte{0xb1, 0x00, 0x00, 0x00, 0x03, 'a', 'b', 'c'},
			want:  []byte("abc"),
		},
		{
			name:  "vbin32 value",
			input: []byte{0xb0, 0x00, 0x00, 0x00, 0x03, 0x00, 0xff, 0x7f},
			want:  []byte{0x00, 0xff, 0x7f},
		},
		{
			name:  "str32 length exceeds int32",
			input: []byte{0xb1, 0xff, 0xff, 0xff, 0xff, 0x00},
			want:  []byte{0xb1, 0xff, 0xff, 0xff, 0xff, 0x00},
		},
		{
			name:  "vbin32 length exceeds int32",
			input: []byte{0xb0, 0xff, 0xff, 0xff, 0xff, 0x00},
			want:  []byte{0xb0, 0xff, 0xff, 0xff, 0xff, 0x00},
		},
		{
			name:  "truncated str32",
			input: []byte{0xb1, 0x00, 0x00, 0x00, 0x02, 'a'},
			want:  []byte{0xb1, 0x00, 0x00, 0x00, 0x02, 'a'},
		},
		{
			name:  "truncated vbin32",
			input: []byte{0xb0, 0x00, 0x00, 0x00, 0x02, 0x01},
			want:  []byte{0xb0, 0x00, 0x00, 0x00, 0x02, 0x01},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := decodeAMQPValue(test.input)
			if got == nil {
				t.Fatalf("decodeAMQPValue(% x) = nil, want a non-nil value", test.input)
			}
			if !bytes.Equal(got, test.want) {
				t.Fatalf("decodeAMQPValue(% x) = % x, want % x", test.input, got, test.want)
			}
		})
	}
}
