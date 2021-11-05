package cmd_test

import (
	"os"
	"testing"

	"github.com/deviceinsight/kafkactl/testutil"
	"github.com/spf13/viper"
)

func TestEnvironmentVariableLoading(t *testing.T) {

	testutil.StartUnitTest(t)

	_ = os.Setenv("CONTEXTS_DEFAULT_BROKERS", "broker1:9092 broker2:9092")
	_ = os.Setenv("CONTEXTS_DEFAULT_TLS_ENABLED", "true")
	_ = os.Setenv("CONTEXTS_DEFAULT_TLS_CERT", "my-cert")
	_ = os.Setenv("CONTEXTS_DEFAULT_TLS_CERTKEY", "my-cert-key")
	_ = os.Setenv("CURRENT_CONTEXT", "non-existing-context")

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("version"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	if len(viper.GetStringSlice("contexts.default.brokers")) != 2 {
		t.Fatalf("expected two default brokers but got: %s", viper.GetString("contexts.default.brokers"))
	}

	testutil.AssertEquals(t, "broker1:9092", viper.GetStringSlice("contexts.default.brokers")[0])
	testutil.AssertEquals(t, "broker2:9092", viper.GetStringSlice("contexts.default.brokers")[1])
	testutil.AssertEquals(t, "true", viper.GetString("contexts.default.tls.enabled"))
	testutil.AssertEquals(t, "my-cert", viper.GetString("contexts.default.tls.cert"))
	testutil.AssertEquals(t, "my-cert-key", viper.GetString("contexts.default.tls.certKey"))
	testutil.AssertEquals(t, "non-existing-context", viper.GetString("current-context"))
}

func TestEnvironmentVariableLoadingAliases(t *testing.T) {

	testutil.StartUnitTest(t)

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

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("version"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	if len(viper.GetStringSlice("contexts.default.brokers")) != 2 {
		t.Fatalf("expected two default brokers but got: %s", viper.GetString("contexts.default.brokers"))
	}

	testutil.AssertEquals(t, "broker1:9092", viper.GetStringSlice("contexts.default.brokers")[0])
	testutil.AssertEquals(t, "broker2:9092", viper.GetStringSlice("contexts.default.brokers")[1])
	testutil.AssertEquals(t, "30", viper.GetString("contexts.default.requestTimeout"))
	testutil.AssertEquals(t, "true", viper.GetString("contexts.default.tls.enabled"))
	testutil.AssertEquals(t, "my-ca", viper.GetString("contexts.default.tls.ca"))
	testutil.AssertEquals(t, "my-cert", viper.GetString("contexts.default.tls.cert"))
	testutil.AssertEquals(t, "my-cert-key", viper.GetString("contexts.default.tls.certKey"))
	testutil.AssertEquals(t, "true", viper.GetString("contexts.default.tls.insecure"))
	testutil.AssertEquals(t, "true", viper.GetString("contexts.default.sasl.enabled"))
	testutil.AssertEquals(t, "user", viper.GetString("contexts.default.sasl.username"))
	testutil.AssertEquals(t, "pass", viper.GetString("contexts.default.sasl.password"))
	testutil.AssertEquals(t, "scram-sha512", viper.GetString("contexts.default.sasl.mechanism"))
	testutil.AssertEquals(t, "my-client", viper.GetString("contexts.default.clientID"))
	testutil.AssertEquals(t, "2.0.1", viper.GetString("contexts.default.kafkaVersion"))
	testutil.AssertEquals(t, "registry:8888", viper.GetString("contexts.default.avro.schemaRegistry"))
	testutil.AssertEquals(t, "hash", viper.GetString("contexts.default.defaultPartitioner"))
}
