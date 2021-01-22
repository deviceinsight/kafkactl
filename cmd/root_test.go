package cmd_test

import (
	"github.com/deviceinsight/kafkactl/test_util"
	"github.com/spf13/viper"
	"os"
	"testing"
)

func TestEnvironmentVariableLoading(t *testing.T) {

	test_util.StartUnitTest(t)

	_ = os.Setenv("CONTEXTS_DEFAULT_BROKERS", "broker1:9092 broker2:9092")
	_ = os.Setenv("CONTEXTS_DEFAULT_TLS_ENABLED", "true")
	_ = os.Setenv("CONTEXTS_DEFAULT_TLS_CERT", "my-cert")
	_ = os.Setenv("CONTEXTS_DEFAULT_TLS_CERTKEY", "my-cert-key")
	_ = os.Setenv("CURRENT_CONTEXT", "non-existing-context")

	kafkaCtl := test_util.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("version"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	if len(viper.GetStringSlice("contexts.default.brokers")) != 2 {
		t.Fatalf("expected two default brokers but got: %s", viper.GetString("contexts.default.brokers"))
	}

	test_util.AssertEquals(t, "broker1:9092", viper.GetStringSlice("contexts.default.brokers")[0])
	test_util.AssertEquals(t, "broker2:9092", viper.GetStringSlice("contexts.default.brokers")[1])
	test_util.AssertEquals(t, "true", viper.GetString("contexts.default.tls.enabled"))
	test_util.AssertEquals(t, "my-cert", viper.GetString("contexts.default.tls.cert"))
	test_util.AssertEquals(t, "my-cert-key", viper.GetString("contexts.default.tls.certKey"))
	test_util.AssertEquals(t, "non-existing-context", viper.GetString("current-context"))
}

func TestEnvironmentVariableLoadingAliases(t *testing.T) {

	test_util.StartUnitTest(t)

	_ = os.Setenv("REQUESTTIMEOUT", "30")
	_ = os.Setenv("BROKERS", "broker1:9092 broker2:9092")
	_ = os.Setenv("TLS_ENABLED", "true")
	_ = os.Setenv("TLS_CA", "my-ca")
	_ = os.Setenv("TLS_CERT", "my-cert")
	_ = os.Setenv("TLS_CERTKEY", "my-cert-key")
	_ = os.Setenv("TLS_INSECURE", "true")
	_ = os.Setenv("SASL_ENABLED", "true")
	_ = os.Setenv("SASL_USERNAME", "user")
	_ = os.Setenv("SASL_PASSWORD", "pass")
	_ = os.Setenv("SASL_MECHANISM", "scram-sha512")
	_ = os.Setenv("CLIENTID", "my-client")
	_ = os.Setenv("KAFKAVERSION", "2.0.1")
	_ = os.Setenv("AVRO_SCHEMAREGISTRY", "registry:8888")
	_ = os.Setenv("DEFAULTPARTITIONER", "hash")

	kafkaCtl := test_util.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("version"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	if len(viper.GetStringSlice("contexts.default.brokers")) != 2 {
		t.Fatalf("expected two default brokers but got: %s", viper.GetString("contexts.default.brokers"))
	}

	test_util.AssertEquals(t, "broker1:9092", viper.GetStringSlice("contexts.default.brokers")[0])
	test_util.AssertEquals(t, "broker2:9092", viper.GetStringSlice("contexts.default.brokers")[1])
	test_util.AssertEquals(t, "30", viper.GetString("contexts.default.requestTimeout"))
	test_util.AssertEquals(t, "true", viper.GetString("contexts.default.tls.enabled"))
	test_util.AssertEquals(t, "my-ca", viper.GetString("contexts.default.tls.ca"))
	test_util.AssertEquals(t, "my-cert", viper.GetString("contexts.default.tls.cert"))
	test_util.AssertEquals(t, "my-cert-key", viper.GetString("contexts.default.tls.certKey"))
	test_util.AssertEquals(t, "true", viper.GetString("contexts.default.tls.insecure"))
	test_util.AssertEquals(t, "true", viper.GetString("contexts.default.sasl.enabled"))
	test_util.AssertEquals(t, "user", viper.GetString("contexts.default.sasl.username"))
	test_util.AssertEquals(t, "pass", viper.GetString("contexts.default.sasl.password"))
	test_util.AssertEquals(t, "scram-sha512", viper.GetString("contexts.default.sasl.mechanism"))
	test_util.AssertEquals(t, "my-client", viper.GetString("contexts.default.clientID"))
	test_util.AssertEquals(t, "2.0.1", viper.GetString("contexts.default.kafkaVersion"))
	test_util.AssertEquals(t, "registry:8888", viper.GetString("contexts.default.avro.schemaRegistry"))
	test_util.AssertEquals(t, "hash", viper.GetString("contexts.default.defaultPartitioner"))
}
