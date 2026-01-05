package internal

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"os/user"
	"regexp"
	"strings"
	"time"

	"github.com/deviceinsight/kafkactl/v5/internal/helpers/avro"

	"github.com/deviceinsight/kafkactl/v5/internal/auth"

	"github.com/deviceinsight/kafkactl/v5/internal/global"

	"github.com/IBM/sarama"
	"github.com/deviceinsight/kafkactl/v5/internal/helpers"
	"github.com/deviceinsight/kafkactl/v5/internal/output"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var invalidClientIDCharactersRegExp = regexp.MustCompile(`[^a-zA-Z0-9-]+`)

type TokenProvider struct {
	PluginName string
	Options    map[string]any
}

type SaslConfig struct {
	Enabled       bool
	Username      string
	Password      string
	Mechanism     string
	TokenProvider TokenProvider
	Version       string
}

type SchemaRegistryConfig struct {
	URL            string
	RequestTimeout time.Duration
	TLS            TLSConfig
	Username       string
	Password       string
}

type ProtobufConfig struct {
	ProtosetFiles    []string
	ProtoFiles       []string
	ProtoImportPaths []string
	MarshalOptions   ProtobufMarshalOptions
}

type ProtobufMarshalOptions struct {
	AllowPartial      bool
	UseProtoNames     bool
	UseEnumNumbers    bool
	EmitUnpopulated   bool
	EmitDefaultValues bool
}

type AvroConfig struct {
	JSONCodec avro.JSONCodec
}

type TLSConfig struct {
	Enabled  bool
	CA       string
	Cert     string
	CertKey  string
	Insecure bool
}

type K8sToleration struct {
	Key      string `json:"key" yaml:"key"`
	Operator string `json:"operator" yaml:"operator"`
	Value    string `json:"value" yaml:"value"`
	Effect   string `json:"effect" yaml:"effect"`
}

type K8sConfig struct {
	Enabled         bool
	Binary          string
	KubeConfig      string
	KubeContext     string
	Namespace       string
	Image           string
	ImagePullSecret string
	TLSSecret       string
	ServiceAccount  string
	AsUser          string
	KeepPod         bool
	Labels          map[string]string
	Annotations     map[string]string
	NodeSelector    map[string]string
	Affinity        map[string]any
	Resources       map[string]any
	Tolerations     []K8sToleration
}

type ConsumerConfig struct {
	IsolationLevel string
}

type ProducerConfig struct {
	Partitioner     string
	RequiredAcks    string
	MaxMessageBytes int
	ValueSerializer string
	KeySerializer   string
}

type ClientContext struct {
	Name           string
	Brokers        []string
	TLS            TLSConfig
	Sasl           SaslConfig
	Kubernetes     K8sConfig
	RequestTimeout time.Duration
	ClientID       string
	KafkaVersion   sarama.KafkaVersion
	Avro           AvroConfig
	Protobuf       ProtobufConfig
	SchemaRegistry SchemaRegistryConfig
	Producer       ProducerConfig
	Consumer       ConsumerConfig
}

type Config struct {
	Name  string
	Value string
}

func CreateClientContext() (ClientContext, error) {
	var context ClientContext
	var err error

	context.Name, err = global.GetCurrentContext()
	if err != nil {
		return context, err
	}

	if viper.Get("contexts."+context.Name) == nil {
		return context, errors.Errorf("no context with name %s found", context.Name)
	}

	if viper.GetString("contexts."+context.Name+".tlsCA") != "" ||
		viper.GetString("contexts."+context.Name+".tlsCert") != "" ||
		viper.GetString("contexts."+context.Name+".tlsCertKey") != "" ||
		viper.GetString("contexts."+context.Name+".tlsInsecure") != "" {
		return context, errors.Errorf("Your tls config contains fields that are not longer supported. Please update your config")
	}

	context.Brokers = viper.GetStringSlice("contexts." + context.Name + ".brokers")
	context.TLS.Enabled = viper.GetBool("contexts." + context.Name + ".tls.enabled")
	if context.TLS.CA, err = resolvePath("contexts." + context.Name + ".tls.ca"); err != nil {
		return context, err
	}
	if context.TLS.Cert, err = resolvePath("contexts." + context.Name + ".tls.cert"); err != nil {
		return context, err
	}
	if context.TLS.CertKey, err = resolvePath("contexts." + context.Name + ".tls.certKey"); err != nil {
		return context, err
	}
	context.TLS.Insecure = viper.GetBool("contexts." + context.Name + ".tls.insecure")
	context.ClientID = viper.GetString("contexts." + context.Name + ".clientID")

	context.RequestTimeout = viper.GetDuration("contexts." + context.Name + ".requestTimeout")

	if version, err := kafkaVersion(viper.GetString("contexts." + context.Name + ".kafkaVersion")); err == nil {
		context.KafkaVersion = version
	} else {
		return context, err
	}
	context.Avro.JSONCodec = avro.ParseJSONCodec(viper.GetString("contexts." + context.Name + ".avro.jsonCodec"))
	context.SchemaRegistry.URL = viper.GetString("contexts." + context.Name + ".schemaRegistry.url")
	context.SchemaRegistry.RequestTimeout = viper.GetDuration("contexts." + context.Name + ".schemaRegistry.requestTimeout")
	context.SchemaRegistry.TLS.Enabled = viper.GetBool("contexts." + context.Name + ".schemaRegistry.tls.enabled")
	if context.SchemaRegistry.TLS.CA, err = resolvePath("contexts." + context.Name + ".schemaRegistry.tls.ca"); err != nil {
		return context, err
	}
	if context.SchemaRegistry.TLS.Cert, err = resolvePath("contexts." + context.Name + ".schemaRegistry.tls.cert"); err != nil {
		return context, err
	}
	if context.SchemaRegistry.TLS.CertKey, err = resolvePath("contexts." + context.Name + ".schemaRegistry.tls.certKey"); err != nil {
		return context, err
	}
	context.SchemaRegistry.TLS.Insecure = viper.GetBool("contexts." + context.Name + ".schemaRegistry.tls.insecure")
	context.SchemaRegistry.Username = viper.GetString("contexts." + context.Name + ".schemaRegistry.username")
	context.SchemaRegistry.Password = viper.GetString("contexts." + context.Name + ".schemaRegistry.password")
	if context.Protobuf.ProtosetFiles, err = resolvePaths("contexts." + context.Name + ".protobuf.protosetFiles"); err != nil {
		return context, err
	}
	if context.Protobuf.ProtoImportPaths, err = resolvePaths("contexts." + context.Name + ".protobuf.importPaths"); err != nil {
		return context, err
	}
	context.Protobuf.ProtoFiles = viper.GetStringSlice("contexts." + context.Name + ".protobuf.protoFiles")
	context.Producer.Partitioner = viper.GetString("contexts." + context.Name + ".producer.partitioner")
	context.Producer.RequiredAcks = viper.GetString("contexts." + context.Name + ".producer.requiredAcks")
	context.Producer.MaxMessageBytes = viper.GetInt("contexts." + context.Name + ".producer.maxMessageBytes")
	context.Producer.ValueSerializer = viper.GetString("contexts." + context.Name + ".producer.valueSerializer")
	context.Producer.KeySerializer = viper.GetString("contexts." + context.Name + ".producer.keySerializer")
	context.Consumer.IsolationLevel = viper.GetString("contexts." + context.Name + ".consumer.isolationLevel")
	context.Sasl.Enabled = viper.GetBool("contexts." + context.Name + ".sasl.enabled")
	context.Sasl.Username = viper.GetString("contexts." + context.Name + ".sasl.username")
	context.Sasl.Password = viper.GetString("contexts." + context.Name + ".sasl.password")
	context.Sasl.Mechanism = viper.GetString("contexts." + context.Name + ".sasl.mechanism")
	context.Sasl.Version = viper.GetString("contexts." + context.Name + ".sasl.version")
	context.Sasl.TokenProvider.PluginName = viper.GetString("contexts." + context.Name + ".sasl.tokenProvider.plugin")
	context.Sasl.TokenProvider.Options = viper.GetStringMap("contexts." + context.Name + ".sasl.tokenProvider.options")

	viper.SetDefault("contexts."+context.Name+".kubernetes.binary", "kubectl")
	context.Kubernetes.Enabled = IsKubernetesEnabled()

	if context.Kubernetes.Enabled {
		if context.Kubernetes.Binary, err = resolvePath("contexts." + context.Name + ".kubernetes.binary"); err != nil {
			return context, err
		}
		if context.Kubernetes.KubeConfig, err = resolvePath("contexts." + context.Name + ".kubernetes.kubeConfig"); err != nil {
			return context, err
		}
	}
	context.Kubernetes.KubeContext = viper.GetString("contexts." + context.Name + ".kubernetes.kubeContext")
	context.Kubernetes.Namespace = viper.GetString("contexts." + context.Name + ".kubernetes.namespace")
	context.Kubernetes.Image = viper.GetString("contexts." + context.Name + ".kubernetes.image")
	context.Kubernetes.ImagePullSecret = viper.GetString("contexts." + context.Name + ".kubernetes.imagePullSecret")
	context.Kubernetes.TLSSecret = viper.GetString("contexts." + context.Name + ".kubernetes.tlsSecret")
	context.Kubernetes.ServiceAccount = viper.GetString("contexts." + context.Name + ".kubernetes.serviceAccount")
	context.Kubernetes.AsUser = viper.GetString("contexts." + context.Name + ".kubernetes.asUser")
	context.Kubernetes.KeepPod = viper.GetBool("contexts." + context.Name + ".kubernetes.keepPod")
	context.Kubernetes.Labels = viper.GetStringMapString("contexts." + context.Name + ".kubernetes.labels")
	context.Kubernetes.Annotations = viper.GetStringMapString("contexts." + context.Name + ".kubernetes.annotations")
	context.Kubernetes.NodeSelector = viper.GetStringMapString("contexts." + context.Name + ".kubernetes.nodeSelector")
	context.Kubernetes.Affinity = viper.GetStringMap("contexts." + context.Name + ".kubernetes.affinity")
	context.Kubernetes.Resources = viper.GetStringMap("contexts." + context.Name + ".kubernetes.resources")

	if err := viper.UnmarshalKey("contexts."+context.Name+".kubernetes.tolerations", &context.Kubernetes.Tolerations); err != nil {
		return context, err
	}

	return context, nil
}

func IsKubernetesEnabled() bool {
	contextName, _ := global.GetCurrentContext()
	return viper.GetBool("contexts." + contextName + ".kubernetes.enabled")
}

func CreateClient(context *ClientContext) (sarama.Client, error) {
	config, err := CreateClientConfig(context)
	if err == nil {
		return sarama.NewClient(context.Brokers, config)
	}
	return nil, err
}

func CreateClusterAdmin(context *ClientContext) (sarama.ClusterAdmin, error) {
	config, err := CreateClientConfig(context)
	if err == nil {
		return sarama.NewClusterAdmin(context.Brokers, config)
	}
	return nil, err
}

func CreateClientConfig(context *ClientContext) (*sarama.Config, error) {
	config := sarama.NewConfig()
	config.Version = context.KafkaVersion
	config.ClientID = GetClientID(context, "kafkactl-")

	if context.RequestTimeout > 0 {
		output.Debugf("using admin request timeout: %s", context.RequestTimeout.String())
		config.Admin.Timeout = context.RequestTimeout
	} else {
		output.Debugf("using default admin request timeout: 3s")
	}

	if context.TLS.Enabled {
		output.Debugf("TLS is enabled.")
		config.Net.TLS.Enable = true

		tlsConfig, err := setupTLSConfig(context.TLS)
		if err != nil {
			return nil, errors.Wrap(err, "failed to setup tls config")
		}
		config.Net.TLS.Config = tlsConfig
	}

	if context.Sasl.Enabled {
		if context.Sasl.Username != "" {
			output.Debugf("SASL is enabled (username = %s)", context.Sasl.Username)
		} else {
			output.Debugf("SASL is enabled")
		}
		config.Net.SASL.Enable = true
		config.Net.SASL.User = context.Sasl.Username
		config.Net.SASL.Password = context.Sasl.Password
		if strings.EqualFold(context.Sasl.Version, "v0") {
			config.Net.SASL.Version = sarama.SASLHandshakeV0
		}
		if strings.EqualFold(context.Sasl.Version, "v1") {
			config.Net.SASL.Version = sarama.SASLHandshakeV1
		}
		switch context.Sasl.Mechanism {
		case "scram-sha512":
			config.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512
			config.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
				return &helpers.XDGSCRAMClient{HashGeneratorFcn: helpers.SHA512}
			}
		case "scram-sha256":
			config.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA256
			config.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
				return &helpers.XDGSCRAMClient{HashGeneratorFcn: helpers.SHA256}
			}
		case "oauth":
			config.Net.SASL.Mechanism = sarama.SASLTypeOAuth
		case "plaintext":
			fallthrough
		case "":
			break
		default:
			return nil, errors.Errorf("Unknown sasl mechanism: %s", context.Sasl.Mechanism)
		}
	}

	if config.Net.SASL.Mechanism == sarama.SASLTypeOAuth {
		tokenProvider, err := auth.LoadTokenProviderPlugin(context.Sasl.TokenProvider.PluginName, context.Sasl.TokenProvider.Options, context.Brokers)
		if err != nil {
			return nil, errors.Wrap(err, "failed to load tokenProvider")
		}
		config.Net.SASL.TokenProvider = tokenProvider
	}

	return config, nil
}

func GetClientID(context *ClientContext, defaultPrefix string) string {
	var (
		err error
		usr *user.User
	)

	if context.ClientID != "" {
		return context.ClientID
	} else if usr, err = user.Current(); err != nil {
		output.Warnf("Failed to read current user: %v", err)
		return strings.TrimSuffix(defaultPrefix, "-")
	}
	return defaultPrefix + sanitizeUsername(usr.Username)
}

func resolvePaths(key string) ([]string, error) {
	filenames := viper.GetStringSlice(key)
	resolved := make([]string, len(filenames))
	var err error

	for i, filename := range filenames {
		resolved[i], err = global.ResolvePath(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve path for key %q: %w", key, err)
		}
	}

	return resolved, nil
}

func resolvePath(key string) (string, error) {
	filename := viper.GetString(key)

	if filename == "" {
		return filename, nil
	}

	if IsKubernetesEnabled() {
		return filename, nil
	}

	resolved, err := global.ResolvePath(filename)
	if err != nil {
		return resolved, fmt.Errorf("failed to resolve path for key %q: %w", key, err)
	}
	return resolved, nil
}

func sanitizeUsername(u string) string {
	// Windows user may have format "DOMAIN|MACHINE\username", remove domain/machine if present
	s := strings.Split(u, "\\")
	u = s[len(s)-1]
	// Windows account can contain spaces or other special characters not supported
	// in client ID. Keep the bare minimum and ditch the rest.
	return invalidClientIDCharactersRegExp.ReplaceAllString(u, "-")
}

// setupTlsConfig takes the paths to a tls certificate, CA, and certificate key in
// a PEM format and returns a constructed tls.Config object.
func setupTLSConfig(tlsConfig TLSConfig) (*tls.Config, error) {
	if !tlsConfig.Enabled {
		return nil, errors.Errorf("tls should be enabled at this point")
	}

	caPool, err := x509.SystemCertPool()
	if err != nil {
		output.Warnf("error reading system cert pool: %v", err)
		caPool = x509.NewCertPool()
	}

	if tlsConfig.CA != "" {
		caString, err := os.ReadFile(tlsConfig.CA)
		if err != nil {
			return nil, err
		}

		ok := caPool.AppendCertsFromPEM(caString)
		if !ok {
			return nil, errors.Errorf("unable to add ca at %s to certificate pool", tlsConfig.CA)
		}
	}

	var clientCert tls.Certificate

	if tlsConfig.Cert != "" && tlsConfig.CertKey != "" {
		clientCert, err = tls.LoadX509KeyPair(tlsConfig.Cert, tlsConfig.CertKey)
		if err != nil {
			return nil, err
		}
	}

	bundle := &tls.Config{
		RootCAs:      caPool,
		Certificates: []tls.Certificate{clientCert},
	}

	if tlsConfig.Insecure {
		bundle.InsecureSkipVerify = true
	}

	return bundle, nil
}

func kafkaVersion(s string) (sarama.KafkaVersion, error) {
	if s == "" {
		output.Debugf("Assuming kafkaVersion: %s", sarama.V2_5_0_0)
		return sarama.V2_5_0_0, nil
	}

	v, err := sarama.ParseKafkaVersion(strings.TrimPrefix(s, "v"))
	if err != nil {
		return sarama.KafkaVersion{}, err
	}

	output.Debugf("Using kafkaVersion: %s", v)

	return v, nil
}

func TopicExists(client *sarama.Client, name string) (bool, error) {
	var (
		err    error
		topics []string
	)

	if topics, err = (*client).Topics(); err != nil {
		return false, errors.Wrap(err, "failed to read topics")
	}

	for _, topic := range topics {
		if topic == name {
			return true, nil
		}
	}

	return false, nil
}

func ListConfigs(admin *sarama.ClusterAdmin, resource sarama.ConfigResource, includeDefaults bool) ([]Config, error) {
	var (
		configEntries []sarama.ConfigEntry
		err           error
	)

	if configEntries, err = (*admin).DescribeConfig(resource); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to describe %v config", getResourceTypeName(resource.Type)))
	}

	return listConfigsFromEntries(configEntries, includeDefaults), nil
}

func listConfigsFromEntries(configEntries []sarama.ConfigEntry, includeDefaults bool) []Config {
	configs := make([]Config, 0)

	for _, configEntry := range configEntries {
		if includeDefaults || (!configEntry.Default && configEntry.Source != sarama.SourceDefault) {
			entry := Config{Name: configEntry.Name, Value: configEntry.Value}
			configs = append(configs, entry)
		}
	}

	return configs
}

func getResourceTypeName(resourceType sarama.ConfigResourceType) string {
	switch resourceType {
	case sarama.TopicResource:
		return "topic"
	case sarama.BrokerResource:
		return "broker"
	case sarama.BrokerLoggerResource:
		return "brokerLogger"
	default:
		return "unknown"
	}
}
