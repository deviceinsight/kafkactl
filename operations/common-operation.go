package operations

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/spf13/viper"
	"io/ioutil"
	"os/user"
	"regexp"
	"strings"
)

var (
	invalidClientIDCharactersRegExp = regexp.MustCompile(`[^a-zA-Z0-9_-]`)
)

type SaslConfig struct {
	Enabled  bool
	Username string
	Password string
}

type ClientContext struct {
	Name               string
	Brokers            []string
	TlsCA              string
	TlsCert            string
	TlsCertKey         string
	TlsInsecure        bool
	Sasl               SaslConfig
	ClientID           string
	KafkaVersion       sarama.KafkaVersion
	AvroSchemaRegistry string
	DefaultPartitioner string
}

func CreateClientContext() ClientContext {
	var context ClientContext

	context.Name = viper.GetString("current-context")
	context.Brokers = viper.GetStringSlice("contexts." + context.Name + ".brokers")
	context.TlsCA = viper.GetString("contexts." + context.Name + ".tlsCA")
	context.TlsCert = viper.GetString("contexts." + context.Name + ".tlsCert")
	context.TlsCertKey = viper.GetString("contexts." + context.Name + ".tlsCertKey")
	context.TlsInsecure = viper.GetBool("contexts." + context.Name + ".tlsInsecure")
	context.ClientID = viper.GetString("contexts." + context.Name + ".clientID")
	context.KafkaVersion = kafkaVersion(viper.GetString("contexts." + context.Name + ".kafkaVersion"))
	context.AvroSchemaRegistry = viper.GetString("contexts." + context.Name + ".avro.schemaRegistry")
	context.DefaultPartitioner = viper.GetString("contexts." + context.Name + ".defaultPartitioner")
	context.Sasl.Enabled = viper.GetBool("contexts." + context.Name + ".sasl.enabled")
	context.Sasl.Username = viper.GetString("contexts." + context.Name + ".sasl.username")
	context.Sasl.Password = viper.GetString("contexts." + context.Name + ".sasl.password")

	return context
}

func CreateClient(context *ClientContext) (sarama.Client, error) {
	return sarama.NewClient(context.Brokers, CreateClientConfig(context))
}

func CreateClientConfig(context *ClientContext) *sarama.Config {
	var (
		err    error
		config = sarama.NewConfig()
	)

	config.Version = context.KafkaVersion
	config.ClientID = getClientID(context)

	tlsConfig, err := setupCerts(context.TlsInsecure, context.TlsCert, context.TlsCA, context.TlsCertKey)
	if err != nil {
		output.Failf("failed to setup certificates err=%v", err)
	}
	if tlsConfig != nil {
		config.Net.TLS.Enable = true
		config.Net.TLS.Config = tlsConfig
	}

	if context.Sasl.Enabled {
		output.Debugf("SASL is enabled.")
		config.Net.SASL.Enable = true
		config.Net.SASL.User = context.Sasl.Username
		config.Net.SASL.Password = context.Sasl.Password
	}

	return config
}

func CreateClusterAdmin(context *ClientContext) (sarama.ClusterAdmin, error) {
	var (
		err    error
		config = sarama.NewConfig()
	)

	config.Version = context.KafkaVersion
	config.ClientID = getClientID(context)

	tlsConfig, err := setupCerts(context.TlsInsecure, context.TlsCert, context.TlsCA, context.TlsCertKey)
	if err != nil {
		output.Failf("failed to setup certificates err=%v", err)
	}
	if tlsConfig != nil {
		config.Net.TLS.Enable = true
		config.Net.TLS.Config = tlsConfig
	}

	if context.Sasl.Enabled {
		output.Debugf("SASL is enabled.")
		config.Net.SASL.Enable = true
		config.Net.SASL.User = context.Sasl.Username
		config.Net.SASL.Password = context.Sasl.Password
	}

	return sarama.NewClusterAdmin(context.Brokers, config)
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

// setupCerts takes the paths to a tls certificate, CA, and certificate key in
// a PEM format and returns a constructed tls.Config object.
func setupCerts(insecure bool, certPath, caPath, keyPath string) (*tls.Config, error) {
	if insecure {
		return &tls.Config{InsecureSkipVerify: true}, nil
	}
	if certPath == "" && caPath == "" && keyPath == "" {
		return nil, nil
	}

	if certPath == "" || caPath == "" || keyPath == "" {
		err := fmt.Errorf("certificate, CA and key path are required - got cert=%#v ca=%#v key=%#v", certPath, caPath, keyPath)
		return nil, err
	}

	caString, err := ioutil.ReadFile(caPath)
	if err != nil {
		return nil, err
	}

	caPool := x509.NewCertPool()
	ok := caPool.AppendCertsFromPEM(caString)
	if !ok {
		output.Failf("unable to add ca at %s to certificate pool", caPath)
	}

	clientCert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, err
	}

	bundle := &tls.Config{
		RootCAs:      caPool,
		Certificates: []tls.Certificate{clientCert},
	}
	bundle.BuildNameToCertificate()
	return bundle, nil
}

func kafkaVersion(s string) sarama.KafkaVersion {
	if s == "" {
		output.Debugf("Assuming kafkaVersion: %s", sarama.V2_0_0_0)
		return sarama.V2_0_0_0
	}

	v, err := sarama.ParseKafkaVersion(strings.TrimPrefix(s, "v"))
	if err != nil {
		output.Failf(err.Error())
	}

	output.Debugf("Using kafkaVersion: %s", v)

	return v
}

func TopicExists(client *sarama.Client, name string) (bool, error) {

	var (
		err    error
		topics []string
	)

	if topics, err = (*client).Topics(); err != nil {
		output.Failf("failed to read topics err=%v", err)
		return false, err
	}

	for _, topic := range topics {
		if topic == name {
			return true, nil
		}
	}

	return false, nil
}
