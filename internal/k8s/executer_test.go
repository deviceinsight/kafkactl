package k8s_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/k8s"
)

const sampleKubectlVersionOutput = `
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
				}`

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
	testRunner.response = []byte(sampleKubectlVersionOutput)
	var runner k8s.Runner = &testRunner

	exec, err := k8s.NewExecutor(context, runner)
	if err != nil {
		t.Fatal(err)
	}

	err = exec.Run("scratch", "/kafkactl", []string{"version"}, []string{"ENV_A=1"})
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

func TestExecWithoutPodOverridesProvided(t *testing.T) {

	var context internal.ClientContext
	context.Kubernetes.Image = "private.registry.com/deviceinsight/kafkactl"

	var testRunner = TestRunner{}
	testRunner.response = []byte(sampleKubectlVersionOutput)
	var runner k8s.Runner = &testRunner

	exec, err := k8s.NewExecutor(context, runner)
	if err != nil {
		t.Fatal(err)
	}

	err = exec.Run("scratch", "/kafkactl", []string{"version"}, []string{"ENV_A=1"})
	if err != nil {
		t.Fatal(err)
	}

	if indexOf("--overrides", testRunner.args) > -1 {
		t.Fatalf("unexpected arg: %s", testRunner.args)
	}
}

func TestExecWithImageAndTagAddsSuffix(t *testing.T) {

	var context internal.ClientContext
	context.Kubernetes.Image = "private.registry.com/deviceinsight/kafkactl:latest"

	var testRunner = TestRunner{}
	testRunner.response = []byte(sampleKubectlVersionOutput)
	var runner k8s.Runner = &testRunner

	exec, err := k8s.NewExecutor(context, runner)
	if err != nil {
		t.Fatal(err)
	}

	_ = exec.Run("scratch", "/kafkactl", []string{"version"}, []string{"ENV_A=1"})

	image := extractParam(t, testRunner.args, "--image")
	if image != context.Kubernetes.Image+"-scratch" {
		t.Fatalf("wrong image: %s", image)
	}
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
			description:   "parse_valid_kubectl_output_succeeds",
			kubectlOutput: sampleKubectlVersionOutput,
			wantVersion: k8s.Version{
				Major:      1,
				Minor:      27,
				GitVersion: "v1.27.1",
			},
		},
		{
			description: "parse_gcloud_kubectl_output_succeeds",
			kubectlOutput: `
				{
				  "ClientVersion": {
					"major": "1",
					"minor": "30+",
					"gitVersion": "v1.30.6-dispatcher",
					"gitCommit": "cab8b7130edcfb9d721071b8d9d606f6b4abe8d6",
					"gitTreeState": "clean",
					"buildDate": "2024-11-07T01:32:04Z",
					"goVersion": "go1.22.8",
					"compiler": "gc",
					"platform": "darwin/arm64"
				  },
				  "kustomizeVersion": "v5.0.4-0.20230601165947-6ce0bf390ce3"
				}`,
			wantVersion: k8s.Version{
				Major:      1,
				Minor:      30,
				GitVersion: "v1.30.6-dispatcher",
			},
		},
	} {
		t.Run(test.description, func(t *testing.T) {

			testRunner.response = []byte(test.kubectlOutput)

			version, err := k8s.GetKubectlVersion("kubectl", runner)

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
