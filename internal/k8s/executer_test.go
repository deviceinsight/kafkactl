package k8s_test

import (
	"context"
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
	var clientContext internal.ClientContext
	clientContext.Kubernetes.Image = "private.registry.com/deviceinsight/kafkactl"
	clientContext.Kubernetes.ImagePullSecret = "registry-secret"

	testRunner := TestRunner{}
	testRunner.response = []byte(sampleKubectlVersionOutput)
	var runner k8s.Runner = &testRunner

	exec, err := k8s.NewExecutor(context.Background(), clientContext, runner)
	if err != nil {
		t.Fatal(err)
	}

	err = exec.Run("scratch", "/kafkactl", []string{"version"}, []string{"ENV_A=1"})
	if err != nil {
		t.Fatal(err)
	}

	image := extractParam(t, testRunner.args, "--image")
	if image != clientContext.Kubernetes.Image+":latest-scratch" {
		t.Fatalf("wrong image: %s", image)
	}

	overrideType := extractParam(t, testRunner.args, "--override-type")
	if overrideType != "json" {
		t.Fatalf("wrong override-type: %s", overrideType)
	}

	overrides := extractParam(t, testRunner.args, "--overrides")
	var podOverrides k8s.JSONPatchType
	if err := json.Unmarshal([]byte(overrides), &podOverrides); err != nil {
		t.Fatalf("unable to unmarshall overrides: %v", err)
	}

	// Find the imagePullSecrets patch operation
	found := false
	for _, patch := range podOverrides {
		if patch.Path == "/spec/imagePullSecrets" && patch.Op == "add" {
			found = true
			// Verify the value is correct
			valueBytes, _ := json.Marshal(patch.Value)
			var secrets []struct{ Name string }
			if err := json.Unmarshal(valueBytes, &secrets); err != nil {
				t.Fatalf("unable to parse imagePullSecrets value: %v", err)
			}
			if len(secrets) != 1 || secrets[0].Name != clientContext.Kubernetes.ImagePullSecret {
				t.Fatalf("wrong imagePullSecrets in patch: %s", overrides)
			}
			break
		}
	}
	if !found {
		t.Fatalf("imagePullSecrets patch operation not found in overrides: %s", overrides)
	}
}

func TestExecWithTLSSecretProvided(t *testing.T) {
	var clientContext internal.ClientContext
	clientContext.Kubernetes.Image = "private.registry.com/deviceinsight/kafkactl"
	clientContext.Kubernetes.TLSSecret = "my-tls-secret"

	testRunner := TestRunner{}
	testRunner.response = []byte(sampleKubectlVersionOutput)
	var runner k8s.Runner = &testRunner

	exec, err := k8s.NewExecutor(context.Background(), clientContext, runner)
	if err != nil {
		t.Fatal(err)
	}

	err = exec.Run("scratch", "/kafkactl", []string{"version"}, []string{"ENV_A=1"})
	if err != nil {
		t.Fatal(err)
	}

	overrideType := extractParam(t, testRunner.args, "--override-type")
	if overrideType != "json" {
		t.Fatalf("wrong override-type: %s", overrideType)
	}

	overrides := extractParam(t, testRunner.args, "--overrides")
	var podOverrides k8s.JSONPatchType
	if err := json.Unmarshal([]byte(overrides), &podOverrides); err != nil {
		t.Fatalf("unable to unmarshall overrides: %v", err)
	}

	// Find the volumes patch operation
	foundVolumes := false
	foundVolumeMounts := false

	for _, patch := range podOverrides {
		if patch.Path == "/spec/volumes" && patch.Op == "add" {
			foundVolumes = true
			// Verify the volume is correct
			valueBytes, _ := json.Marshal(patch.Value)
			var volumes []struct {
				Name   string `json:"name"`
				Secret struct {
					SecretName string `json:"secretName"`
				} `json:"secret"`
			}
			if err := json.Unmarshal(valueBytes, &volumes); err != nil {
				t.Fatalf("unable to parse volumes value: %v", err)
			}
			if len(volumes) != 1 || volumes[0].Name != "kafkactl-tls" || volumes[0].Secret.SecretName != clientContext.Kubernetes.TLSSecret {
				t.Fatalf("wrong volumes in patch: %s", overrides)
			}
		}

		if patch.Path == "/spec/containers/0/volumeMounts" && patch.Op == "add" {
			foundVolumeMounts = true
			// Verify the volumeMount is correct
			valueBytes, _ := json.Marshal(patch.Value)
			var volumeMounts []struct {
				Name      string `json:"name"`
				MountPath string `json:"mountPath"`
				ReadOnly  bool   `json:"readOnly"`
			}
			if err := json.Unmarshal(valueBytes, &volumeMounts); err != nil {
				t.Fatalf("unable to parse volumeMounts value: %v", err)
			}
			if len(volumeMounts) != 1 || volumeMounts[0].Name != "kafkactl-tls" ||
				volumeMounts[0].MountPath != "/etc/ssl/certs/kafkactl" || !volumeMounts[0].ReadOnly {
				t.Fatalf("wrong volumeMounts in patch: %s", overrides)
			}
		}
	}

	if !foundVolumes {
		t.Fatalf("volumes patch operation not found in overrides: %s", overrides)
	}
	if !foundVolumeMounts {
		t.Fatalf("volumeMounts patch operation not found in overrides: %s", overrides)
	}
}

func TestExecWithResourcesProvided(t *testing.T) {
	var clientContext internal.ClientContext
	clientContext.Kubernetes.Image = "private.registry.com/deviceinsight/kafkactl"
	clientContext.Kubernetes.Resources = map[string]any{
		"requests": map[string]any{
			"memory": "64Mi",
			"cpu":    "250m",
		},
		"limits": map[string]any{
			"memory": "128Mi",
			"cpu":    "500m",
		},
	}

	testRunner := TestRunner{}
	testRunner.response = []byte(sampleKubectlVersionOutput)
	var runner k8s.Runner = &testRunner

	exec, err := k8s.NewExecutor(context.Background(), clientContext, runner)
	if err != nil {
		t.Fatal(err)
	}

	err = exec.Run("scratch", "/kafkactl", []string{"version"}, []string{"ENV_A=1"})
	if err != nil {
		t.Fatal(err)
	}

	overrideType := extractParam(t, testRunner.args, "--override-type")
	if overrideType != "json" {
		t.Fatalf("wrong override-type: %s", overrideType)
	}

	overrides := extractParam(t, testRunner.args, "--overrides")
	var podOverrides k8s.JSONPatchType
	if err := json.Unmarshal([]byte(overrides), &podOverrides); err != nil {
		t.Fatalf("unable to unmarshall overrides: %v", err)
	}

	// Find the resources patch operation
	found := false
	for _, patch := range podOverrides {
		if patch.Path == "/spec/containers/0/resources" && patch.Op == "add" {
			found = true
			// Verify the value is correct
			valueBytes, _ := json.Marshal(patch.Value)
			var resources map[string]map[string]string
			if err := json.Unmarshal(valueBytes, &resources); err != nil {
				t.Fatalf("unable to parse resources value: %v", err)
			}
			if resources["requests"]["memory"] != "64Mi" ||
				resources["requests"]["cpu"] != "250m" ||
				resources["limits"]["memory"] != "128Mi" ||
				resources["limits"]["cpu"] != "500m" {
				t.Fatalf("wrong resources in patch: %s", overrides)
			}
			break
		}
	}
	if !found {
		t.Fatalf("resources patch operation not found in overrides: %s", overrides)
	}
}

func TestExecWithoutPodOverridesProvided(t *testing.T) {
	var clientContext internal.ClientContext
	clientContext.Kubernetes.Image = "private.registry.com/deviceinsight/kafkactl"

	testRunner := TestRunner{}
	testRunner.response = []byte(sampleKubectlVersionOutput)
	var runner k8s.Runner = &testRunner

	exec, err := k8s.NewExecutor(context.Background(), clientContext, runner)
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
	var clientContext internal.ClientContext
	clientContext.Kubernetes.Image = "private.registry.com/deviceinsight/kafkactl:latest"

	testRunner := TestRunner{}
	testRunner.response = []byte(sampleKubectlVersionOutput)
	var runner k8s.Runner = &testRunner

	exec, err := k8s.NewExecutor(context.Background(), clientContext, runner)
	if err != nil {
		t.Fatal(err)
	}

	_ = exec.Run("scratch", "/kafkactl", []string{"version"}, []string{"ENV_A=1"})

	image := extractParam(t, testRunner.args, "--image")
	if image != clientContext.Kubernetes.Image+"-scratch" {
		t.Fatalf("wrong image: %s", image)
	}
}

//nolint:gocognit
func TestParseKubectlVersion(t *testing.T) {
	testRunner := TestRunner{}
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
