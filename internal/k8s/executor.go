package k8s

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/output"
	"golang.org/x/term"
)

type Version struct {
	Major      int
	Minor      int
	GitVersion string
}

type executor struct {
	kubectlBinary   string
	image           string
	imagePullSecret string
	version         Version
	runner          Runner
	clientID        string
	kubeConfig      string
	kubeContext     string
	serviceAccount  string
	keepPod         bool
	namespace       string
	labels          map[string]string
	annotations     map[string]string
	nodeSelector    map[string]string
	affinity        map[string]any
	tolerations     []map[string]any
}

const letterBytes = "abcdefghijklmnpqrstuvwxyz123456789"

func randomString(n int) string {

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[r.Intn(len(letterBytes))]
	}
	return string(b)
}

func getKubectlVersion(kubectlBinary string, runner Runner) Version {

	bytes, err := runner.ExecuteAndReturn(kubectlBinary, []string{"version", "--client", "-o", "json"})
	if err != nil {
		output.Fail(err)
		return Version{}
	}

	if len(bytes) == 0 {
		return Version{}
	}

	type versionOutput struct {
		ClientVersion struct {
			Major      string `json:"major"`
			Minor      string `json:"minor"`
			GitVersion string `json:"gitVersion"`
		} `json:"clientVersion"`
	}

	var jsonOutput versionOutput

	if err := json.Unmarshal(bytes, &jsonOutput); err != nil {
		output.Fail(fmt.Errorf("unable to extract kubectl version: %w", err))
		return Version{}
	}

	major, err := strconv.Atoi(jsonOutput.ClientVersion.Major)
	if err != nil {
		output.Fail(err)
	}

	minor, err := strconv.Atoi(jsonOutput.ClientVersion.Minor)
	if err != nil {
		output.Fail(err)
	}

	return Version{
		Major:      major,
		Minor:      minor,
		GitVersion: jsonOutput.ClientVersion.GitVersion,
	}
}

func newExecutor(context internal.ClientContext, runner Runner) *executor {
	return &executor{
		kubectlBinary:   context.Kubernetes.Binary,
		version:         getKubectlVersion(context.Kubernetes.Binary, runner),
		image:           context.Kubernetes.Image,
		imagePullSecret: context.Kubernetes.ImagePullSecret,
		clientID:        internal.GetClientID(&context, ""),
		kubeConfig:      context.Kubernetes.KubeConfig,
		kubeContext:     context.Kubernetes.KubeContext,
		namespace:       context.Kubernetes.Namespace,
		serviceAccount:  context.Kubernetes.ServiceAccount,
		keepPod:         context.Kubernetes.KeepPod,
		labels:          context.Kubernetes.Labels,
		annotations:     context.Kubernetes.Annotations,
		nodeSelector:    context.Kubernetes.NodeSelector,
		affinity:        context.Kubernetes.Affinity,
		tolerations:     context.Kubernetes.Tolerations,
		runner:          runner,
	}
}

func (kubectl *executor) SetKubectlBinary(bin string) {
	kubectl.kubectlBinary = bin
}

func (kubectl *executor) Run(dockerImageType, entryPoint string, kafkactlArgs []string, podEnvironment []string) error {

	dockerImage := getDockerImage(kubectl.image, dockerImageType)

	podName := fmt.Sprintf("kafkactl-%s-%s", strings.ToLower(kubectl.clientID), randomString(4))

	kubectlArgs := []string{
		"run", "-i", "--tty", "--restart=Never", podName,
		"--image", dockerImage,
	}

	if !kubectl.keepPod {
		kubectlArgs = slices.Insert(kubectlArgs, 1, "--rm")
	}

	if kubectl.kubeConfig != "" {
		kubectlArgs = append(kubectlArgs, "--kubeconfig", kubectl.kubeConfig)
	}

	podOverride := kubectl.createPodOverride()
	if !podOverride.IsEmpty() {
		podOverrideJSON, err := json.Marshal(podOverride)
		if err != nil {
			return errors.Wrap(err, "unable to create override")
		}

		kubectlArgs = append(kubectlArgs, "--overrides", string(podOverrideJSON))
	}

	kubectlArgs = append(kubectlArgs, "--context", kubectl.kubeContext)
	kubectlArgs = append(kubectlArgs, "--namespace", kubectl.namespace)

	for _, env := range podEnvironment {
		kubectlArgs = append(kubectlArgs, "--env", env)
	}

	kubectlArgs = addTerminalSizeEnv(kubectlArgs)

	kubectlArgs = append(kubectlArgs, "--command", "--", entryPoint)

	// Keep only kafkactl arguments that are relevant in k8s context
	allExceptConfigFileFilter := func(s string) bool {
		return !strings.HasPrefix(s, "-C=") && !strings.HasPrefix(s, "--config-file=")
	}
	kubectlArgs = append(kubectlArgs, filter(kafkactlArgs, allExceptConfigFileFilter)...)

	return kubectl.exec(kubectlArgs)
}

func addTerminalSizeEnv(args []string) []string {

	if !term.IsTerminal(0) {
		output.Debugf("no terminal detected")
		return args
	}

	width, height, err := term.GetSize(0)
	if err != nil {
		output.Debugf("unable to determine terminal size: %v", err)
		return args
	}

	args = append(args, "--env", fmt.Sprintf("TERM_WIDTH=%d", width))
	args = append(args, "--env", fmt.Sprintf("TERM_HEIGHT=%d", height))
	return args
}

func getDockerImage(image string, imageType string) string {

	if KafkaCtlVersion == "" {
		KafkaCtlVersion = "latest"
	}

	if image == "" {
		image = "deviceinsight/kafkactl"
	}

	if strings.Contains(image, ":") {
		return image + "-" + imageType
	}
	return image + ":" + KafkaCtlVersion + "-" + imageType
}

func filter(slice []string, predicate func(string) bool) (ret []string) {
	for _, s := range slice {
		if predicate(s) {
			ret = append(ret, s)
		}
	}
	return
}

func (kubectl *executor) exec(args []string) error {
	cmd := fmt.Sprintf("exec: %s %s", kubectl.kubectlBinary, join(args))
	output.Debugf("kubectl version: %s", kubectl.version.GitVersion)
	output.Debugf(cmd)
	err := kubectl.runner.Execute(kubectl.kubectlBinary, args)
	return err
}

func join(args []string) string {
	allArgs := ""
	for _, arg := range args {
		if strings.Contains(arg, " ") {
			arg = "\"" + arg + "\""
		}
		allArgs += " " + arg
	}
	return strings.TrimLeft(allArgs, " ")
}

func (kubectl *executor) GetVersion() Version {
	return kubectl.version
}

func (kubectl *executor) IsVersionAtLeast(major int, minor int) bool {
	return kubectl.version.Major >= major && kubectl.version.Minor >= minor
}
