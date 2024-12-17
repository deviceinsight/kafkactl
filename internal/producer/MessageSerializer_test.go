package producer

import (
	"reflect"
	"testing"
)

func TestDecodeBytesBase64WhenInputIsNil(t *testing.T) {
	got, err := decodeBytes(nil, "base64")
	if got != nil {
		t.Errorf("Expected nil, got %v", got)
	}
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
}

func TestDecodeBytesBase64WhenInputIsInvalid(t *testing.T) {
	got, err := decodeBytes([]byte("..this..is..not..base64.."), "base64")
	if got != nil {
		t.Errorf("Expected nil, got %v", got)
	}
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestDecodeBytesBase64WithRegularInput(t *testing.T) {
	// length 4
	got, err := decodeBytes([]byte("dGVzdA=="), "base64")
	if string(got) != "test" {
		t.Errorf("Expected hello, got %v", string(got))
	}
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
	// length 5
	got, err = decodeBytes([]byte("aGVsbG8="), "base64")
	if string(got) != "hello" {
		t.Errorf("Expected hello, got %v", string(got))
	}
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}

	// length 6
	got, err = decodeBytes([]byte("aGVsbG8h"), "base64")
	if string(got) != "hello!" {
		t.Errorf("Expected hello, got %v", string(got))
	}
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
}

func TestDecodeBytesBase64WithZeroPaddedInput(t *testing.T) {
	got, err := decodeBytes([]byte("AAAAAAAD"), "base64")
	expected := []byte{0, 0, 0, 0, 0, 3}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Expected hello, got %v", got)
	}
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
}
