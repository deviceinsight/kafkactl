package k8s

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/deviceinsight/kafkactl/internal/env"

	"github.com/deviceinsight/kafkactl/internal"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var KafkaCtlVersion string

type Operation struct {
	context internal.ClientContext
}

func (operation *Operation) initialize() error {

	if !operation.context.Kubernetes.Enabled {
		return errors.Errorf("context is not a kubernetes context: %s", operation.context.Name)
	}

	if operation.context.Kubernetes.KubeContext == "" {
		return errors.Errorf("context has no kubernetes context set: contexts.%s.kubernetes.kubeContext", operation.context.Name)
	}

	if operation.context.Kubernetes.Namespace == "" {
		return errors.Errorf("context has no kubernetes namespace set: contexts.%s.kubernetes.namespace", operation.context.Name)
	}

	return nil
}

func (operation *Operation) Attach() error {

	var err error

	if operation.context, err = internal.CreateClientContext(); err != nil {
		return err
	}

	if err := operation.initialize(); err != nil {
		return err
	}

	exec := newExecutor(operation.context, &ShellRunner{})

	podEnvironment := parsePodEnvironment(operation.context)

	return exec.Run("ubuntu", "bash", nil, podEnvironment)
}

func (operation *Operation) TryRun(cmd *cobra.Command, args []string) bool {

	var err error

	if operation.context, err = internal.CreateClientContext(); err != nil {
		return false
	}

	if !operation.context.Kubernetes.Enabled {
		return false
	}

	if err := operation.Run(cmd, args); err != nil {
		output.Fail(err)
	}
	return true
}

func (operation *Operation) Run(cmd *cobra.Command, args []string) error {

	if err := operation.initialize(); err != nil {
		return err
	}

	exec := newExecutor(operation.context, &ShellRunner{})

	kafkaCtlCommand := parseCompleteCommand(cmd, []string{})
	kafkaCtlFlags, err := parseFlags(cmd)
	if err != nil {
		return err
	}

	podEnvironment := parsePodEnvironment(operation.context)

	kafkaCtlCommand = append(kafkaCtlCommand, args...)
	kafkaCtlCommand = append(kafkaCtlCommand, kafkaCtlFlags...)

	return exec.Run("scratch", "/kafkactl", kafkaCtlCommand, podEnvironment)
}

func parseFlags(cmd *cobra.Command) ([]string, error) {
	var flags []string
	var err error

	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if err == nil && flag.Changed {
			if flag.Value.Type() == "intSlice" || flag.Value.Type() == "int32Slice" {
				var intArray []int
				intArray, err = parseIntArray(flag.Value.String())
				if err == nil {
					for _, value := range intArray {
						flags = append(flags, fmt.Sprintf("--%s=%s", flag.Name, strconv.Itoa(value)))
					}
				}
			} else {
				flags = append(flags, fmt.Sprintf("--%s=%s", flag.Name, flag.Value.String()))
			}
		}
	})
	return flags, err
}

func parseIntArray(array string) ([]int, error) {
	var ints []int
	err := json.Unmarshal([]byte(array), &ints)
	return ints, err
}

func parseCompleteCommand(cmd *cobra.Command, found []string) []string {
	if cmd.Parent() == nil {
		return found
	}
	newCommand := []string{cmd.Name()}
	found = append(newCommand, found...)
	return parseCompleteCommand(cmd.Parent(), found)
}

func parsePodEnvironment(context internal.ClientContext) []string {

	var envVariables []string

	envVariables = appendStrings(envVariables, env.Brokers, context.Brokers)
	envVariables = appendBool(envVariables, env.TLSEnabled, context.TLS.Enabled)
	envVariables = appendStringIfDefined(envVariables, env.TLSCa, context.TLS.CA)
	envVariables = appendStringIfDefined(envVariables, env.TLSCert, context.TLS.Cert)
	envVariables = appendStringIfDefined(envVariables, env.TLSCertKey, context.TLS.CertKey)
	envVariables = appendBool(envVariables, env.TLSInsecure, context.TLS.Insecure)
	envVariables = appendBool(envVariables, env.SaslEnabled, context.Sasl.Enabled)
	envVariables = appendStringIfDefined(envVariables, env.SaslUsername, context.Sasl.Username)
	envVariables = appendStringIfDefined(envVariables, env.SaslPassword, context.Sasl.Password)
	envVariables = appendStringIfDefined(envVariables, env.SaslMechanism, context.Sasl.Mechanism)
	envVariables = appendStringIfDefined(envVariables, env.RequestTimeout, context.RequestTimeout.String())
	envVariables = appendStringIfDefined(envVariables, env.ClientID, context.ClientID)
	envVariables = appendStringIfDefined(envVariables, env.KafkaVersion, context.KafkaVersion.String())
	envVariables = appendStringIfDefined(envVariables, env.AvroSchemaRegistry, context.AvroSchemaRegistry)
	envVariables = appendStrings(envVariables, env.ProtobufProtoSetFiles, context.Protobuf.ProtosetFiles)
	envVariables = appendStrings(envVariables, env.ProtobufImportPaths, context.Protobuf.ProtoImportPaths)
	envVariables = appendStrings(envVariables, env.ProtobufProtoFiles, context.Protobuf.ProtoFiles)
	envVariables = appendStringIfDefined(envVariables, env.ProducerPartitioner, context.Producer.Partitioner)
	envVariables = appendStringIfDefined(envVariables, env.ProducerRequiredAcks, context.Producer.RequiredAcks)
	envVariables = appendIntIfGreaterZero(envVariables, env.ProducerMaxMessageBytes, context.Producer.MaxMessageBytes)

	return envVariables
}

func appendStrings(env []string, key string, value []string) []string {
	return append(env, fmt.Sprintf("%s=%s", key, strings.Join(value, " ")))
}

func appendBool(env []string, key string, value bool) []string {
	if value {
		return append(env, fmt.Sprintf("%s=%t", key, value))
	}
	return env
}

func appendStringIfDefined(env []string, key string, value string) []string {
	if value != "" {
		return append(env, fmt.Sprintf("%s=%s", key, value))
	}
	return env
}

func appendIntIfGreaterZero(env []string, key string, value int) []string {
	if value > 0 {
		return append(env, fmt.Sprintf("%s=%d", key, value))
	}
	return env
}
