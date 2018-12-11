package operations

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"os/user"
	"regexp"
	"strings"
)

var (
	invalidClientIDCharactersRegExp = regexp.MustCompile(`[^a-zA-Z0-9_-]`)
)

type ClientContext struct {
	name         string
	brokers      []string
	tlsCA        string
	tlsCert      string
	tlsCertKey   string
	kafkaVersion sarama.KafkaVersion
}

func createClientContext() ClientContext {

	var context ClientContext

	context.name = viper.GetString("current-context")
	context.brokers = viper.GetStringSlice("contexts." + context.name + ".brokers")
	context.tlsCA = viper.GetString("contexts." + context.name + ".tlsCA")
	context.tlsCert = viper.GetString("contexts." + context.name + ".tlsCert")
	context.tlsCertKey = viper.GetString("contexts." + context.name + ".tlsCertKey")
	context.kafkaVersion = kafkaVersion(viper.GetString("contexts." + context.name + ".kafkaVersion"))

	return context
}

func createClient(context *ClientContext) (sarama.Client, error) {
	return sarama.NewClient(context.brokers, createClientConfig(context))
}

func createClientConfig(context *ClientContext) *sarama.Config {
	var (
		err    error
		usr    *user.User
		config = sarama.NewConfig()
	)

	config.Version = context.kafkaVersion

	if usr, err = user.Current(); err != nil {
		output.Warnf("Failed to read current user: %v", err)
	}
	config.ClientID = "kafkactl-" + sanitizeUsername(usr.Username)

	tlsConfig, err := setupCerts(context.tlsCert, context.tlsCA, context.tlsCertKey)
	if err != nil {
		output.Failf("failed to setup certificates err=%v", err)
	}
	if tlsConfig != nil {
		config.Net.TLS.Enable = true
		config.Net.TLS.Config = tlsConfig
	}

	return config
}

func createClusterAdmin(context *ClientContext) (sarama.ClusterAdmin, error) {
	var (
		err    error
		usr    *user.User
		config = sarama.NewConfig()
	)

	config.Version = context.kafkaVersion

	if usr, err = user.Current(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read current user err=%v", err)
	}
	config.ClientID = "kafkactl-" + sanitizeUsername(usr.Username)

	tlsConfig, err := setupCerts(context.tlsCert, context.tlsCA, context.tlsCertKey)
	if err != nil {
		output.Failf("failed to setup certificates err=%v", err)
	}
	if tlsConfig != nil {
		config.Net.TLS.Enable = true
		config.Net.TLS.Config = tlsConfig
	}

	return sarama.NewClusterAdmin(context.brokers, config)
}

func createOffsetManager(client sarama.Client, group string) (sarama.OffsetManager, error) {

	if group == "" {
		return nil, nil
	}

	return sarama.NewOffsetManagerFromClient(group, client)
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
func setupCerts(certPath, caPath, keyPath string) (*tls.Config, error) {
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
		return sarama.V2_0_0_0
	}

	v, err := sarama.ParseKafkaVersion(strings.TrimPrefix(s, "v"))
	if err != nil {
		output.Failf(err.Error())
	}

	return v
}
