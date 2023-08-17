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

	"github.com/deviceinsight/kafkactl/internal/helpers/avro"

	"github.com/IBM/sarama"
	"github.com/deviceinsight/kafkactl/internal/helpers"
	"github.com/deviceinsight/kafkactl/internal/helpers/protobuf"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var (
	invalidClientIDCharactersRegExp = regexp.MustCompile(`[^a-zA-Z0-9_-]`)
)

type SaslConfig struct {
	Enabled   bool
	Username  string
	Password  string
	Mechanism string
}

type TLSConfig struct {
	Enabled  bool
	CA       string
	Cert     string
	CertKey  string
	Insecure bool
}

type K8sConfig struct {
	Enabled         bool
	Binary          string
	Extra           []string
	KubeConfig      string
	KubeContext     string
	Namespace       string
	Image           string
	ImagePullSecret string
}

type ProducerConfig struct {
	Partitioner     string
	RequiredAcks    string
	MaxMessageBytes int
}

type ClientContext struct {
	Name               string
	Brokers            []string
	TLS                TLSConfig
	Sasl               SaslConfig
	Kubernetes         K8sConfig
	RequestTimeout     time.Duration
	ClientID           string
	KafkaVersion       sarama.KafkaVersion
	AvroSchemaRegistry string
	AvroJSONCodec      avro.JSONCodec
	Protobuf           protobuf.SearchContext
	Producer           ProducerConfig
}

type Config struct {
	Name  string
	Value string
}

func CreateClientContext() (ClientContext, error) {
	var context ClientContext

	context.Name = viper.GetString("current-context")

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
	context.TLS.CA = viper.GetString("contexts." + context.Name + ".tls.ca")
	context.TLS.Cert = viper.GetString("contexts." + context.Name + ".tls.cert")
	context.TLS.CertKey = viper.GetString("contexts." + context.Name + ".tls.certKey")
	context.TLS.Insecure = viper.GetBool("contexts." + context.Name + ".tls.insecure")
	context.ClientID = viper.GetString("contexts." + context.Name + ".clientID")

	context.RequestTimeout = viper.GetDuration("contexts." + context.Name + ".requestTimeout")

	if version, err := kafkaVersion(viper.GetString("contexts." + context.Name + ".kafkaVersion")); err == nil {
		context.KafkaVersion = version
	} else {
		return context, err
	}
	context.AvroSchemaRegistry = viper.GetString("contexts." + context.Name + ".avro.schemaRegistry")
	context.AvroJSONCodec = avro.ParseJSONCodec(viper.GetString("contexts." + context.Name + ".avro.jsonCodec"))
	context.Protobuf.ProtosetFiles = viper.GetStringSlice("contexts." + context.Name + ".protobuf.protosetFiles")
	context.Protobuf.ProtoImportPaths = viper.GetStringSlice("contexts." + context.Name + ".protobuf.importPaths")
	context.Protobuf.ProtoFiles = viper.GetStringSlice("contexts." + context.Name + ".protobuf.protoFiles")
	context.Producer.Partitioner = viper.GetString("contexts." + context.Name + ".producer.partitioner")
	context.Producer.RequiredAcks = viper.GetString("contexts." + context.Name + ".producer.requiredAcks")
	context.Producer.MaxMessageBytes = viper.GetInt("contexts." + context.Name + ".producer.maxMessageBytes")
	context.Sasl.Enabled = viper.GetBool("contexts." + context.Name + ".sasl.enabled")
	context.Sasl.Username = viper.GetString("contexts." + context.Name + ".sasl.username")
	context.Sasl.Password = viper.GetString("contexts." + context.Name + ".sasl.password")
	context.Sasl.Mechanism = viper.GetString("contexts." + context.Name + ".sasl.mechanism")

	viper.SetDefault("contexts."+context.Name+".kubernetes.binary", "kubectl")
	context.Kubernetes.Enabled = viper.GetBool("contexts." + context.Name + ".kubernetes.enabled")
	context.Kubernetes.Binary = viper.GetString("contexts." + context.Name + ".kubernetes.binary")
	context.Kubernetes.KubeConfig = viper.GetString("contexts." + context.Name + ".kubernetes.kubeConfig")
	context.Kubernetes.KubeContext = viper.GetString("contexts." + context.Name + ".kubernetes.kubeContext")
	context.Kubernetes.Namespace = viper.GetString("contexts." + context.Name + ".kubernetes.namespace")
	context.Kubernetes.Extra = viper.GetStringSlice("contexts." + context.Name + ".kubernetes.extra")
	context.Kubernetes.Image = viper.GetString("contexts." + context.Name + ".kubernetes.image")
	context.Kubernetes.ImagePullSecret = viper.GetString("contexts." + context.Name + ".kubernetes.imagePullSecret")

	return context, nil
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

	var config = sarama.NewConfig()
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
		output.Debugf("SASL is enabled (username = %s).", context.Sasl.Username)
		config.Net.SASL.Enable = true
		config.Net.SASL.User = context.Sasl.Username
		config.Net.SASL.Password = context.Sasl.Password
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
		case "plaintext":
			fallthrough
		case "":
			break
		default:
			return nil, errors.Errorf("Unknown sasl mechanism: %s", context.Sasl.Mechanism)
		}
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
	} else {
		return defaultPrefix + sanitizeUsername(usr.Username)
	}
}

func sanitizeUsername(u string) string {
	// Windows user may have format "DOMAIN|MACHINE\username", remove domain/machine if present
	s := strings.Split(u, "\\")
	u = s[len(s)-1]
	// Windows account can contain spaces or other special characters not supported
	// in client ID. Keep the bare minimum and ditch the rest.
	return invalidClientIDCharactersRegExp.ReplaceAllString(u, "")
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

func ListConfigs(admin *sarama.ClusterAdmin, resource sarama.ConfigResource) ([]Config, error) {

	var (
		configs       = make([]Config, 0)
		configEntries []sarama.ConfigEntry
		err           error
	)

	if configEntries, err = (*admin).DescribeConfig(resource); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to describe %v config", getResourceTypeName(resource.Type)))
	}

	for _, configEntry := range configEntries {

		if !configEntry.Default && configEntry.Source != sarama.SourceDefault {
			entry := Config{Name: configEntry.Name, Value: configEntry.Value}
			configs = append(configs, entry)
		}
	}

	return configs, nil
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
