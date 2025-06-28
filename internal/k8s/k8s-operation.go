package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/deviceinsight/kafkactl/v5/internal/global"

	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var KafkaCtlVersion string

type Operation interface {
	Attach() error
	Run(cmd *cobra.Command, args []string) error
}

type operation struct {
	runner Runner
}

func NewOperation() Operation {
	return &operation{
		runner: &ShellRunner{},
	}
}

func (op *operation) initialize(context internal.ClientContext) error {
	if !context.Kubernetes.Enabled {
		return errors.Errorf("context is not a kubernetes context: %s", context.Name)
	}

	if context.Kubernetes.KubeContext == "" {
		return errors.Errorf("context has no kubernetes context set: contexts.%s.kubernetes.kubeContext", context.Name)
	}

	if context.Kubernetes.Namespace == "" {
		return errors.Errorf("context has no kubernetes namespace set: contexts.%s.kubernetes.namespace", context.Name)
	}

	return nil
}

func (op *operation) Attach() error {
	clientContext, err := internal.CreateClientContext()
	if err != nil {
		return err
	}

	if err := op.initialize(clientContext); err != nil {
		return err
	}

	exec, err := newExecutor(context.Background(), clientContext, op.runner)
	if err != nil {
		return err
	}

	podEnvironment := parsePodEnvironment(clientContext)

	return exec.Run("ubuntu", "bash", nil, podEnvironment, "--tty")
}

func (op *operation) Run(cmd *cobra.Command, args []string) error {
	context, err := internal.CreateClientContext()
	if err != nil {
		return err
	}

	if !context.Kubernetes.Enabled {
		return fmt.Errorf("kubernetes not enabled")
	}

	return op.run(cmd.Context(), context, cmd, args)
}

func (op *operation) run(ctx context.Context, clientContext internal.ClientContext, cmd *cobra.Command, args []string) error {
	if err := op.initialize(clientContext); err != nil {
		return err
	}

	exec, err := newExecutor(ctx, clientContext, op.runner)
	if err != nil {
		return err
	}

	kafkaCtlCommand := parseCompleteCommand(cmd, []string{})
	kafkaCtlFlags, err := parseFlags(cmd)
	if err != nil {
		return err
	}

	podEnvironment := parsePodEnvironment(clientContext)

	kafkaCtlCommand = append(kafkaCtlCommand, args...)
	kafkaCtlCommand = append(kafkaCtlCommand, kafkaCtlFlags...)

	return exec.Run("scratch", "/kafkactl", kafkaCtlCommand, podEnvironment)
}

func parseFlags(cmd *cobra.Command) ([]string, error) {
	var flags []string
	var err error

	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if err == nil && flag.Changed {
			if parsedFlags, parseErr := parseFlag(flag, cmd.Flags()); err != nil {
				err = parseErr
			} else {
				flags = append(flags, parsedFlags...)
			}
		}
	})
	return flags, err
}

func parseFlag(flag *pflag.Flag, flagSet *pflag.FlagSet) (flags []string, err error) {
	switch flag.Value.Type() {
	case "intSlice":
		var intSlice []int
		intSlice, err = flagSet.GetIntSlice(flag.Name)
		if err == nil {
			for _, value := range intSlice {
				flags = append(flags, fmt.Sprintf("--%s=%d", flag.Name, value))
			}
		}
		return flags, err
	case "int32Slice":
		var int32Slice []int32
		int32Slice, err = flagSet.GetInt32Slice(flag.Name)
		if err == nil {
			for _, value := range int32Slice {
				flags = append(flags, fmt.Sprintf("--%s=%d", flag.Name, value))
			}
		}
	case "stringArray":
		var strArray []string
		strArray, err = flagSet.GetStringArray(flag.Name)
		if err == nil {
			for _, value := range strArray {
				flags = append(flags, fmt.Sprintf("--%s=%s", flag.Name, value))
			}
		}
	case "stringSlice":
		var strSlice []string
		strSlice, err = flagSet.GetStringSlice(flag.Name)
		if err == nil {
			for _, value := range strSlice {
				flags = append(flags, fmt.Sprintf("--%s=%s", flag.Name, value))
			}
		}
	default:
		flags = append(flags, fmt.Sprintf("--%s=%s", flag.Name, flag.Value.String()))
	}
	return flags, err
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

	envVariables = appendStrings(envVariables, global.Brokers, context.Brokers)
	envVariables = appendBool(envVariables, global.TLSEnabled, context.TLS.Enabled)
	envVariables = appendStringIfDefined(envVariables, global.TLSCa, context.TLS.CA)
	envVariables = appendStringIfDefined(envVariables, global.TLSCert, context.TLS.Cert)
	envVariables = appendStringIfDefined(envVariables, global.TLSCertKey, context.TLS.CertKey)
	envVariables = appendBool(envVariables, global.TLSInsecure, context.TLS.Insecure)
	envVariables = appendBool(envVariables, global.SaslEnabled, context.Sasl.Enabled)
	envVariables = appendStringIfDefined(envVariables, global.SaslUsername, context.Sasl.Username)
	envVariables = appendStringIfDefined(envVariables, global.SaslPassword, context.Sasl.Password)
	envVariables = appendStringIfDefined(envVariables, global.SaslMechanism, context.Sasl.Mechanism)
	envVariables = appendStringIfDefined(envVariables, global.SaslTokenProviderPlugin, context.Sasl.TokenProvider.PluginName)
	envVariables = appendMapIfDefined(envVariables, global.SaslTokenProviderOptions, context.Sasl.TokenProvider.Options)
	envVariables = appendStringIfDefined(envVariables, global.RequestTimeout, context.RequestTimeout.String())
	envVariables = appendStringIfDefined(envVariables, global.ClientID, context.ClientID)
	envVariables = appendStringIfDefined(envVariables, global.KafkaVersion, context.KafkaVersion.String())
	envVariables = appendStringIfDefined(envVariables, global.AvroJSONCodec, context.Avro.JSONCodec.String())
	envVariables = appendStringIfDefined(envVariables, global.SchemaRegistryURL, context.SchemaRegistry.URL)
	envVariables = appendStringIfDefined(envVariables, global.SchemaRegistryRequestTimeout, context.SchemaRegistry.RequestTimeout.String())
	envVariables = appendBool(envVariables, global.SchemaRegistryTLSEnabled, context.SchemaRegistry.TLS.Enabled)
	envVariables = appendStringIfDefined(envVariables, global.SchemaRegistryTLSCa, context.SchemaRegistry.TLS.CA)
	envVariables = appendStringIfDefined(envVariables, global.SchemaRegistryTLSCert, context.SchemaRegistry.TLS.Cert)
	envVariables = appendStringIfDefined(envVariables, global.SchemaRegistryTLSCertKey, context.SchemaRegistry.TLS.CertKey)
	envVariables = appendBool(envVariables, global.SchemaRegistryTLSInsecure, context.SchemaRegistry.TLS.Insecure)
	envVariables = appendStringIfDefined(envVariables, global.SchemaRegistryUsername, context.SchemaRegistry.Username)
	envVariables = appendStringIfDefined(envVariables, global.SchemaRegistryPassword, context.SchemaRegistry.Password)
	envVariables = appendStrings(envVariables, global.ProtobufProtoSetFiles, context.Protobuf.ProtosetFiles)
	envVariables = appendStrings(envVariables, global.ProtobufImportPaths, context.Protobuf.ProtoImportPaths)
	envVariables = appendStrings(envVariables, global.ProtobufProtoFiles, context.Protobuf.ProtoFiles)
	envVariables = appendStringIfDefined(envVariables, global.ProducerPartitioner, context.Producer.Partitioner)
	envVariables = appendStringIfDefined(envVariables, global.ProducerRequiredAcks, context.Producer.RequiredAcks)
	envVariables = appendIntIfGreaterZero(envVariables, global.ProducerMaxMessageBytes, context.Producer.MaxMessageBytes)

	return envVariables
}

func appendStrings(env []string, key string, value []string) []string {
	if len(value) > 0 {
		return append(env, fmt.Sprintf("%s=%s", key, strings.Join(value, " ")))
	}
	return env
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

func appendMapIfDefined(env []string, key string, value map[string]any) []string {
	if len(value) > 0 {
		jsonMap, err := json.Marshal(value)
		if err != nil {
			panic(err)
		}
		return append(env, fmt.Sprintf("%s=%s", key, jsonMap))
	}
	return env
}
