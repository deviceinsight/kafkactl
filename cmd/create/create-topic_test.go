package create

import (
	"github.com/deviceinsight/kafkactl/output"
	"github.com/deviceinsight/kafkactl/util"
	"strings"
	"testing"
)

func TestServiceFunc(t *testing.T) {
	t.Parallel()
}

func TestInvalidServiceFunc3(t *testing.T) {
	t.Parallel()
}

func TestCreateTopicIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	_, out, err := util.ExecuteCommandC("create", "topic", "my-topic")

	if err != nil {
		t.Fatalf("failed to execute create topic command: %v", err)
	}

	expected := "created topic my-topic"

	if out != expected {
		t.Fatalf("unexpected output: %s", out)
	}
}
