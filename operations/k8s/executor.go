package k8s

import (
	"fmt"
	"github.com/deviceinsight/kafkactl/operations"
	"github.com/deviceinsight/kafkactl/output"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
)

type Version struct {
	Major int
	Minor int
	Patch int
}

type executor struct {
	kubectlBinary string
	version       Version
	runner        Runner
	kubeConfig    string
	kubeContext   string
	namespace     string
	extra         []string
}

const letterBytes = "abcdefghijklmnpqrstuvwxyz123456789"

func randomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func getKubectlVersion(kubectlBinary string, runner Runner) Version {

	bytes, err := runner.ExecuteAndReturn(kubectlBinary, []string{"version", "--client", "--short"})
	if err != nil {
		output.Fail(err)
	}

	if len(bytes) == 0 {
		return Version{}
	}

	re := regexp.MustCompile(`v(?P<major>\d+)\.(?P<minor>\d+)\.(?P<patch>\d+)`)
	matches := re.FindStringSubmatch(string(bytes))

	result := make(map[string]string)
	for i, name := range re.SubexpNames() {
		result[name] = matches[i]
	}

	major, err := strconv.Atoi(result["major"])
	if err != nil {
		output.Fail(err)
	}

	minor, err := strconv.Atoi(result["minor"])
	if err != nil {
		output.Fail(err)
	}

	patch, err := strconv.Atoi(result["patch"])
	if err != nil {
		output.Fail(err)
	}

	return Version{
		Major: major,
		Minor: minor,
		Patch: patch,
	}
}

func newExecutor(context operations.ClientContext, runner Runner) *executor {
	return &executor{
		kubectlBinary: context.Kubernetes.Binary,
		version:       getKubectlVersion(context.Kubernetes.Binary, runner),
		kubeConfig:    context.Kubernetes.KubeConfig,
		kubeContext:   context.Kubernetes.KubeContext,
		namespace:     context.Kubernetes.Namespace,
		runner:        runner,
	}
}

func (kubectl *executor) SetExtraArgs(args ...string) {
	kubectl.extra = args
}

func (kubectl *executor) SetKubectlBinary(bin string) {
	kubectl.kubectlBinary = bin
}

func (kubectl *executor) Run(dockerImageType string, kafkactlArgs []string, podEnvironment []string) error {

	if KafkaCtlVersion == "" {
		KafkaCtlVersion = "latest"
	}
	dockerImage := "deviceinsight/kafkactl:" + KafkaCtlVersion + "-" + dockerImageType

	podName := "kafkactl-" + randomString(10)

	kubectlArgs := []string{"run", "--generator=run-pod/v1", "--rm", "-i", "--tty", "--restart=Never", podName,
		"--image", dockerImage}

	if kubectl.kubeConfig != "" {
		kubectlArgs = append(kubectlArgs, "--kubeconfig", kubectl.kubeConfig)
	}

	kubectlArgs = append(kubectlArgs, "--context", kubectl.kubeContext)
	kubectlArgs = append(kubectlArgs, "--namespace", kubectl.namespace)

	for _, env := range podEnvironment {
		kubectlArgs = append(kubectlArgs, "--env", env)
	}

	if len(kubectl.extra) > 0 {
		kubectlArgs = append(kubectlArgs, kubectl.extra...)
	}

	kubectlArgs = append(kubectlArgs, "--")
	kubectlArgs = append(kubectlArgs, kafkactlArgs...)

	return kubectl.exec(kubectlArgs)
}

func (kubectl *executor) exec(args []string) error {

	cmd := fmt.Sprintf("exec: %s %s", kubectl.kubectlBinary, join(args))
	output.Debugf("kubectl version: %d.%d.%d", kubectl.version.Major, kubectl.version.Minor, kubectl.version.Patch)
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
