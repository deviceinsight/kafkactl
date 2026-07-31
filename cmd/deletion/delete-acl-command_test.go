package deletion

import "testing"

func TestDeleteACLResourceNameFlag(t *testing.T) {
	cmd := newDeleteACLCmd()
	flag := cmd.Flags().Lookup("resource-name")

	if flag == nil {
		t.Fatal("resource-name flag is not registered")
	}

	if flag.Shorthand != "r" {
		t.Fatalf("resource-name shorthand = %q, want %q", flag.Shorthand, "r")
	}

	if flag.DefValue != "" {
		t.Fatalf("resource-name default = %q, want empty", flag.DefValue)
	}
}
