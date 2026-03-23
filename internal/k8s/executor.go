package k8s

import (
	"context"
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
	kubectlBinary    string
	image            string
	imagePullSecret  string
	tlsSecret        string
	saslSecretName   string
	createSaslSecret bool
	saslSecret       string
	saslConfig       internal.SaslConfig
	version          Version
	runner           Runner
	clientID         string
	kubeConfig       string
	kubeContext      string
	serviceAccount   string
	asUser           string
	keepPod          bool
	namespace        string
	labels           map[string]string
	annotations      map[string]string
	nodeSelector     map[string]string
	affinity         map[string]any
	resources        map[string]any
	tolerations      []internal.K8sToleration
	ctx              context.Context
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

func getKubectlVersion(kubectlBinary string, runner Runner) (Version, error) {
	bytes, err := runner.ExecuteAndReturn(kubectlBinary, []string{"version", "--client", "-o", "json"})
	if err != nil {
		return Version{}, err
	}

	if len(bytes) == 0 {
		return Version{}, fmt.Errorf("version response empty")
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
		return Version{}, fmt.Errorf("unable to extract kubectl version: %w", err)
	}

	major, err := strconv.Atoi(jsonOutput.ClientVersion.Major)
	if err != nil {
		return Version{}, err
	}

	minor, err := strconv.Atoi(strings.ReplaceAll(jsonOutput.ClientVersion.Minor, "+", ""))
	if err != nil {
		return Version{}, err
	}

	return Version{
		Major:      major,
		Minor:      minor,
		GitVersion: jsonOutput.ClientVersion.GitVersion,
	}, nil
}

func newExecutor(ctx context.Context, clientContext internal.ClientContext, runner Runner) (*executor, error) {
	version, err := getKubectlVersion(clientContext.Kubernetes.Binary, runner)
	if err != nil {
		return nil, err
	}

	return &executor{
		kubectlBinary:    clientContext.Kubernetes.Binary,
		version:          version,
		image:            clientContext.Kubernetes.Image,
		imagePullSecret:  clientContext.Kubernetes.ImagePullSecret,
		tlsSecret:        clientContext.Kubernetes.TLSSecret,
		saslSecretName:   clientContext.Kubernetes.SaslSecret.Name,
		createSaslSecret: clientContext.Kubernetes.SaslSecret.Create,
		saslConfig:       clientContext.Sasl,
		clientID:         internal.GetClientID(&clientContext, ""),
		kubeConfig:       clientContext.Kubernetes.KubeConfig,
		kubeContext:      clientContext.Kubernetes.KubeContext,
		namespace:        clientContext.Kubernetes.Namespace,
		serviceAccount:   clientContext.Kubernetes.ServiceAccount,
		asUser:           clientContext.Kubernetes.AsUser,
		keepPod:          clientContext.Kubernetes.KeepPod,
		labels:           clientContext.Kubernetes.Labels,
		annotations:      clientContext.Kubernetes.Annotations,
		nodeSelector:     clientContext.Kubernetes.NodeSelector,
		affinity:         clientContext.Kubernetes.Affinity,
		resources:        clientContext.Kubernetes.Resources,
		tolerations:      clientContext.Kubernetes.Tolerations,
		runner:           runner,
		ctx:              ctx,
	}, nil
}

func (kubectl *executor) SetKubectlBinary(bin string) {
	kubectl.kubectlBinary = bin
}

func (kubectl *executor) Run(dockerImageType, entryPoint string, kafkactlArgs []string, podEnvironment []string, additionalKubectlArgs ...string) error {
	dockerImage := getDockerImage(kubectl.image, dockerImageType)

	suffix := randomString(4)
	podName := fmt.Sprintf("kafkactl-%s-%s", strings.ToLower(kubectl.clientID), suffix)

	if kubectl.saslSecretName != "" && kubectl.createSaslSecret {
		return errors.Errorf("kubernetes.saslSecret.name and kubernetes.saslSecret.create are mutually exclusive")
	}

	if kubectl.createSaslSecret {
		// RBAC check
		canIArgs := []string{"auth", "can-i", "create", "secrets"}
		canIArgs = kubectl.addGlobalArgs(canIArgs)
		if bytes, err := kubectl.runner.ExecuteAndReturn(kubectl.kubectlBinary, canIArgs); err != nil {
			return errors.Wrapf(err, "RBAC check failed in namespace %s", kubectl.namespace)
		} else if !strings.Contains(string(bytes), "yes") {
			return errors.Errorf("no permission to create secrets in namespace %s", kubectl.namespace)
		}

		kubectl.saslSecret = fmt.Sprintf("kafkactl-sasl-%s", suffix)

		// ensure we don't take over lifecycle of an existing secret
		checkArgs := kubectl.addGlobalArgs([]string{"get", "secret", kubectl.saslSecret})
		if _, err := kubectl.runner.ExecuteAndReturn(kubectl.kubectlBinary, checkArgs); err == nil {
			return errors.Errorf("secret %s already exists; kafkactl will not manage the lifecycle of existing secrets", kubectl.saslSecret)
		}

		createSecretArgs := []string{
			"create", "secret", "generic", kubectl.saslSecret,
			fmt.Sprintf("--from-literal=username=%s", kubectl.saslConfig.Username),
			fmt.Sprintf("--from-literal=password=%s", kubectl.saslConfig.Password),
		}
		createSecretArgs = kubectl.addGlobalArgs(createSecretArgs)

		if _, err := kubectl.runner.ExecuteAndReturn(kubectl.kubectlBinary, createSecretArgs); err != nil {
			return errors.Wrapf(err, "unable to create secret %s", kubectl.saslSecret)
		}

		defer func() {
			deleteSecretArgs := []string{"delete", "secret", kubectl.saslSecret, "--ignore-not-found=true"}
			deleteSecretArgs = kubectl.addGlobalArgs(deleteSecretArgs)
			if _, err := kubectl.runner.ExecuteAndReturn(kubectl.kubectlBinary, deleteSecretArgs); err != nil {
				output.Warnf("unable to delete secret %s: %v", kubectl.saslSecret, err)
			}
		}()
	} else if kubectl.saslSecretName != "" {
		kubectl.saslSecret = kubectl.saslSecretName
	}

	kubectlArgs := []string{
		"run", "-i", "--restart=Never", podName,
		"--image", dockerImage,
	}

	kubectlArgs = append(kubectlArgs, additionalKubectlArgs...)

	if !kubectl.keepPod {
		kubectlArgs = slices.Insert(kubectlArgs, 1, "--rm")
	}

	kubectlArgs = kubectl.addGlobalArgs(kubectlArgs)

	podOverrides := kubectl.createPodOverrides()
	if len(podOverrides) > 0 {
		podOverridesJSON, err := json.Marshal(podOverrides)
		if err != nil {
			return errors.Wrap(err, "unable to create overrides")
		}

		kubectlArgs = append(kubectlArgs, "--override-type", "json")
		kubectlArgs = append(kubectlArgs, "--overrides", string(podOverridesJSON))
	}

	for _, env := range podEnvironment {
		kubectlArgs = append(kubectlArgs, "--env", env)
	}

	kubectlArgs = addTerminalSizeEnv(kubectlArgs)

	kubectlArgs = append(kubectlArgs, "--command", "--", entryPoint)

	// Keep only kafkactl arguments that are relevant in k8s context
	allExceptConfigFileFilter := func(s string) bool {
		return !strings.HasPrefix(s, "-C=") && !strings.HasPrefix(s, "--config-file=") && !strings.HasPrefix(s, "--context=")
	}
	kubectlArgs = append(kubectlArgs, filter(kafkactlArgs, allExceptConfigFileFilter)...)
	errChan := make(chan error, 1)

	go func() {
		errChan <- kubectl.exec(kubectlArgs)
		close(errChan)
	}()

	select {
	case <-kubectl.ctx.Done():
		err := kubectl.exec([]string{"delete", "pod", podName, "-n", kubectl.namespace, "--wait=true"})
		if err != nil {
			output.Warnf("delete pod %s returned an error %w", podName, err)
			return err
		}
		return context.Canceled
	case err := <-errChan:
		return err
	}
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

func (kubectl *executor) addGlobalArgs(args []string) []string {
	if kubectl.kubeConfig != "" {
		args = append(args, "--kubeconfig", kubectl.kubeConfig)
	}

	if kubectl.asUser != "" {
		args = append(args, "--as", kubectl.asUser)
	}

	args = append(args, "--context", kubectl.kubeContext)
	args = append(args, "--namespace", kubectl.namespace)

	return args
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
