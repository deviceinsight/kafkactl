package protobuf

import (
	"slices"
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"
)

const schema = `syntax = "proto3";
package foo.bar;

message Outer {       // Level 0
  message MiddleAA {  // Level 1
    message Inner {   // Level 2
      int64 ival = 1;
      bool  booly = 2;
    }
  }
  message MiddleBB {  // Level 1
    message Inner {   // Level 2
      int32 ival = 1;
      bool  booly = 2;
    }
  }
}

message Outer2 {       // Level 0
  message MiddleAA {  // Level 1
    message Inner {   // Level 2
      int64 ival = 1;
      bool  booly = 2;
    }
  }
  message MiddleBB {  // Level 1
    message Inner {   // Level 2
      int32 ival = 1;
      bool  booly = 2;
    }
  }
}
`

func TestComputeIndexes(t *testing.T) {
	testCases := []struct {
		msgName  protoreflect.FullName
		expected []int64
	}{
		{"foo.bar.Outer", []int64{}},
		{"foo.bar.Outer2", []int64{1}},
		{"foo.bar.Outer.MiddleAA", []int64{0, 0}},
		{"foo.bar.Outer2.MiddleAA", []int64{1, 0}},
		{"foo.bar.Outer.MiddleAA.Inner", []int64{0, 0, 0}},
		{"foo.bar.Outer2.MiddleAA.Inner", []int64{1, 0, 0}},
	}
	fileDesc, err := ParseFileDescriptor(".", map[string]string{".": schema})
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testCases {
		t.Run(string(tc.msgName), func(t *testing.T) {
			result, err := ComputeIndexes(fileDesc, tc.msgName)
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(tc.expected, result) {
				t.Fatalf("expected: %+v, found: %+v", tc.expected, result)
			}
		})
	}
}

func TestComputeIndexesErrorsWhenCantCompute(t *testing.T) {
	fileDesc, err := ParseFileDescriptor(".", map[string]string{".": schema})
	if err != nil {
		t.Fatal(err)
	}
	msgName := protoreflect.FullName("foo.bar.MiddleAA")
	result, err := ComputeIndexes(fileDesc, msgName)
	if err == nil {
		t.Fatalf("expected: error, found: %+v", result)
	}
}
