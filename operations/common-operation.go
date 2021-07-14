package operations

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/Shopify/sarama"
	"github.com/deviceinsight/kafkactl/operations/helpers"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"io/ioutil"
	"os/user"
	"regexp"
	"strings"
	"time"
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

type TlsConfig struct {
	Enabled  bool
	CA       string
	Cert     string
	CertKey  string
	Insecure bool
}

type K8sConfig struct {
	Enabled     bool
	Binary      string
	KubeConfig  string
	KubeContext string
	Namespace   string
}

type ClientContext struct {
	Name               string
	Brokers            []string
	Tls                TlsConfig
	Sasl               SaslConfig
	Kubernetes         K8sConfig
	RequestTimeout     time.Duration
	ClientID           string
	KafkaVersion       sarama.KafkaVersion
	AvroSchemaRegistry string
	DefaultPartitioner string
	RequiredAcks       string
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
	context.Tls.Enabled = viper.GetBool("contexts." + context.Name + ".tls.enabled")
	context.Tls.CA = viper.GetString("contexts." + context.Name + ".tls.ca")
	context.Tls.Cert = viper.GetString("contexts." + context.Name + ".tls.cert")
	context.Tls.CertKey = viper.GetString("contexts." + context.Name + ".tls.certKey")
	context.Tls.Insecure = viper.GetBool("contexts." + context.Name + ".tls.insecure")
	context.ClientID = viper.GetString("contexts." + context.Name + ".clientID")

	context.RequestTimeout = viper.GetDuration("contexts." + context.Name + ".requestTimeout")

	if version, err := kafkaVersion(viper.GetString("contexts." + context.Name + ".kafkaVersion")); err == nil {
		context.KafkaVersion = version
	} else {
		return context, err
	}
	context.AvroSchemaRegistry = viper.GetString("contexts." + context.Name + ".avro.schemaRegistry")
	context.DefaultPartitioner = viper.GetString("contexts." + context.Name + ".defaultPartitioner")
	context.RequiredAcks = viper.GetString("contexts." + context.Name + ".requiredAcks")
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

	return context, nil
}

func CreateClient(context *ClientContext) (sarama.Client, error) {
	config, err := CreateClientConfig(context)
	if err == nil {
		return sarama.NewClient(context.Brokers, config)
	} else {
		return nil, err
	}
}

func CreateClusterAdmin(context *ClientContext) (sarama.ClusterAdmin, error) {
	config, err := CreateClientConfig(context)
	if err == nil {
		return sarama.NewClusterAdmin(context.Brokers, config)
	} else {
		return nil, err
	}
}

func CreateClientConfig(context *ClientContext) (*sarama.Config, error) {

	var config = sarama.NewConfig()
	config.Version = context.KafkaVersion
	config.ClientID = getClientID(context)

	if context.RequestTimeout > 0 {
		output.Debugf("using admin request timeout: %s", context.RequestTimeout.String())
		config.Admin.Timeout = context.RequestTimeout
	} else {
		output.Debugf("using default admin request timeout: 3s")
	}

	if context.Tls.Enabled {
		output.Debugf("TLS is enabled.")
		config.Net.TLS.Enable = true

		tlsConfig, err := setupTlsConfig(context.Tls)
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

func getClientID(context *ClientContext) string {

	var (
		err error
		usr *user.User
	)

	if context.ClientID != "" {
		return context.ClientID
	} else if usr, err = user.Current(); err != nil {
		output.Warnf("Failed to read current user: %v", err)
		return "kafkactl"
	} else {
		return "kafkactl-" + sanitizeUsername(usr.Username)
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
func setupTlsConfig(tlsConfig TlsConfig) (*tls.Config, error) {

	if !tlsConfig.Enabled {
		return nil, errors.Errorf("tls should be enabled at this point")
	}

	caPool, err := x509.SystemCertPool()
	if err != nil {
		output.Warnf("error reading system cert pool: %v", err)
		caPool = x509.NewCertPool()
	}

	if tlsConfig.CA != "" {
		caString, err := ioutil.ReadFile(tlsConfig.CA)
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
		output.Debugf("Assuming kafkaVersion: %s", sarama.V2_0_0_0)
		return sarama.V2_0_0_0, nil
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
