package util

import (
	"testing"
)

func TestTimestamp(t *testing.T) {

	var validExamples = []string{
		"2014-04-26T17:24:37.123Z",
		"2014-04-26T17:24:37.123",
		"2009-08-12T22:15:09Z",
		"2017-07-19T03:21:51",
		"2013-04-01T22:43",
		"2014-04-26",
		"1384216367189",
	}

	for _, example := range validExamples {

		if val, err := ParseTimestamp(example); err != nil {
			t.Fatalf("Error converting %s - err %v", example, err)
		} else {
			t.Logf("converted %s to: %v", example, val)
		}
	}

	var invalidExample = "14249B99055"
	if val, err := ParseTimestamp(invalidExample); err != nil {
		t.Logf("Invalid timestamp %s rejected - err %v", invalidExample, err)
	} else {
		t.Fatalf("Invalid timestamp %s not rejected and converted to: %v", invalidExample, val)
	}
}

func TestStringArraysEqual(t *testing.T) {
	var reference = []string{"a", "b", "c"}
	var differents = [][]string{
		{"c", "a", "b"},
		{"a", "a", "a"},
		{"a"},
	}

	for _, different := range differents {
		if StringArraysEqual(reference, different) {
			t.Fatalf("Only arrays with the same elements in the same order are considered equal. %s : %s", reference, different)
		}
	}

	var equal = []string{"a", "b", "c"}
	if !StringArraysEqual(reference, equal) {
		t.Fatalf("Arrays with the same elements in the same order shall be considered equal. %s : %s", reference, equal)
	}
}
