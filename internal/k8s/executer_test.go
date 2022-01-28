package k8s_test

import (
	"encoding/json"
	"testing"

	"github.com/deviceinsight/kafkactl/internal"
	"github.com/deviceinsight/kafkactl/internal/k8s"
	"github.com/deviceinsight/kafkactl/testutil"
)

type TestRunner struct {
	binary string
	args   []string
}

func (runner *TestRunner) ExecuteAndReturn(binary string, args []string) ([]byte, error) {
	runner.binary = binary
	runner.args = args
	return nil, nil
}

func (runner *TestRunner) Execute(binary string, args []string) error {
	runner.binary = binary
	runner.args = args
	return nil
}

func TestExecWithImageAndImagePullSecretProvided(t *testing.T) {

	var context internal.ClientContext
	context.Kubernetes.Image = "private.registry.com/deviceinsight/kafkactl"
	context.Kubernetes.ImagePullSecret = "registry-secret"

	var testRunner = TestRunner{}
	var runner k8s.Runner = &testRunner

	exec := k8s.NewExecutor(context, &runner)

	err := exec.Run("scratch", "/kafkactl", []string{"version"}, []string{"ENV_A=1"})
	if err != nil {
		t.Fatal(err)
	}

	image := extractParam(t, testRunner.args, "--image")
	if image != context.Kubernetes.Image+":latest-scratch" {
		t.Fatalf("wrong image: %s", image)
	}

	overrides := extractParam(t, testRunner.args, "--overrides")
	var podOverrides k8s.PodOverrideType
	if err := json.Unmarshal([]byte(overrides), &podOverrides); err != nil {
		t.Fatalf("unable to unmarshall overrides: %v", err)
	}
	if len(podOverrides.Spec.ImagePullSecrets) != 1 || podOverrides.Spec.ImagePullSecrets[0].Name != context.Kubernetes.ImagePullSecret {
		t.Fatalf("wrong overrides: %s", overrides)
	}
}

func TestExecWithImageAndTagFails(t *testing.T) {

	var context internal.ClientContext
	context.Kubernetes.Image = "private.registry.com/deviceinsight/kafkactl:latest"

	var testRunner = TestRunner{}
	var runner k8s.Runner = &testRunner

	exec := k8s.NewExecutor(context, &runner)

	err := exec.Run("scratch", "/kafkactl", []string{"version"}, []string{"ENV_A=1"})
	testutil.AssertErrorContains(t, "image must not contain a tag", err)
}

func extractParam(t *testing.T, args []string, param string) string {

	var paramIdx int

	if paramIdx = indexOf(param, args); paramIdx < 0 {
		t.Fatalf("unable to find %s param: %v", param, args)
	}

	return args[paramIdx+1]
}

func indexOf(element string, data []string) int {
	for k, v := range data {
		if element == v {
			return k
		}
	}
	return -1
}
