package k8s_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/deviceinsight/kafkactl/internal/output"

	"github.com/deviceinsight/kafkactl/internal"
	"github.com/deviceinsight/kafkactl/internal/k8s"
	"github.com/deviceinsight/kafkactl/internal/testutil"
)

type TestRunner struct {
	binary   string
	args     []string
	response []byte
}

func (runner *TestRunner) ExecuteAndReturn(binary string, args []string) ([]byte, error) {
	runner.binary = binary
	runner.args = args
	return runner.response, nil
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

	exec := k8s.NewExecutor(context, runner)

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
	if len(podOverrides.Spec.ImagePullSecrets) != 1 ||
		podOverrides.Spec.ImagePullSecrets[0].Name != context.Kubernetes.ImagePullSecret {
		t.Fatalf("wrong overrides: %s", overrides)
	}
}

func TestExecWithImageAndTagFails(t *testing.T) {

	var context internal.ClientContext
	context.Kubernetes.Image = "private.registry.com/deviceinsight/kafkactl:latest"

	var testRunner = TestRunner{}
	var runner k8s.Runner = &testRunner

	exec := k8s.NewExecutor(context, runner)

	err := exec.Run("scratch", "/kafkactl", []string{"version"}, []string{"ENV_A=1"})
	testutil.AssertErrorContains(t, "image must not contain a tag", err)
}

//nolint:gocognit
func TestParseKubectlVersion(t *testing.T) {

	var testRunner = TestRunner{}
	var runner k8s.Runner = &testRunner

	type tests struct {
		description   string
		kubectlOutput string
		wantErr       string
		wantVersion   k8s.Version
	}

	for _, test := range []tests{
		{
			description:   "parse_fails_for_unparsable_output",
			kubectlOutput: "unknown output",
			wantErr:       "unable to extract kubectl version",
		},
		{
			description: "parse_valid_kubectl_output_succeeds",
			kubectlOutput: `
				{
				  "ClientVersion": {
					"major": "1",
					"minor": "27",
					"gitVersion": "v1.27.1",
					"gitCommit": "4c9411232e10168d7b050c49a1b59f6df9d7ea4b",
					"gitTreeState": "clean",
					"buildDate": "2023-04-14T13:14:41Z",
					"goVersion": "go1.20.3",
					"compiler": "gc",
					"platform": "linux/amd64"
				  },
				  "kustomizeVersion": "v5.0.1"
				}`,
			wantVersion: k8s.Version{
				Major:      1,
				Minor:      27,
				GitVersion: "v1.27.1",
			},
		},
	} {
		t.Run(test.description, func(t *testing.T) {

			var err error

			output.Fail = func(failError error) {
				err = failError
			}

			testRunner.response = []byte(test.kubectlOutput)

			version := k8s.GetKubectlVersion("kubectl", runner)

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

			if test.wantVersion.GitVersion != version.GitVersion {
				t.Fatalf("different version expexted %s != %s", test.wantVersion.GitVersion, version.GitVersion)
			}

			if test.wantVersion.Major != version.Major {
				t.Fatalf("different major version expexted %d != %d", test.wantVersion.Major, version.Major)
			}

			if test.wantVersion.Minor != version.Minor {
				t.Fatalf("different minor version expexted %d != %d", test.wantVersion.Minor, version.Minor)
			}
		})
	}
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
