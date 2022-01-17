package cmd_test

import (
	"os"
	"testing"

	"github.com/deviceinsight/kafkactl/internal/env"

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

	_ = os.Setenv(env.RequestTimeout, "30")
	_ = os.Setenv(env.Brokers, "broker1:9092 broker2:9092")
	_ = os.Setenv(env.TLSEnabled, "true")
	_ = os.Setenv(env.TLSCa, "my-ca")
	_ = os.Setenv(env.TLSCert, "my-cert")
	_ = os.Setenv(env.TLSCertKey, "my-cert-key")
	_ = os.Setenv(env.TLSInsecure, "true")
	_ = os.Setenv(env.SaslEnabled, "true")
	_ = os.Setenv(env.SaslUsername, "user")
	_ = os.Setenv(env.SaslPassword, "pass")
	_ = os.Setenv(env.SaslMechanism, "scram-sha512")
	_ = os.Setenv(env.ClientID, "my-client")
	_ = os.Setenv(env.KafkaVersion, "2.0.1")
	_ = os.Setenv(env.AvroSchemaRegistry, "registry:8888")
	_ = os.Setenv(env.ProtobufProtoSetFiles, "/usr/include/protosets/ps1.protoset /usr/lib/ps2.protoset")
	_ = os.Setenv(env.ProtobufImportPaths, "/usr/include/protobuf /usr/lib/protobuf")
	_ = os.Setenv(env.ProtobufProtoFiles, "message.proto other.proto")
	_ = os.Setenv(env.ProducerPartitioner, "hash")
	_ = os.Setenv(env.ProducerRequiredAcks, "WaitForAll")
	_ = os.Setenv(env.ProducerMaxMessageBytes, "1234")

	for _, key := range env.Variables {
		if os.Getenv(key) == "" {
			t.Fatalf("missing test case for env variable: %s", key)
		}
	}

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
	testutil.AssertEquals(t, "/usr/include/protosets/ps1.protoset", viper.GetStringSlice("contexts.default.protobuf.protosetFiles")[0])
	testutil.AssertEquals(t, "/usr/include/protobuf", viper.GetStringSlice("contexts.default.protobuf.importPaths")[0])
	testutil.AssertEquals(t, "message.proto", viper.GetStringSlice("contexts.default.protobuf.protoFiles")[0])
	testutil.AssertEquals(t, "hash", viper.GetString("contexts.default.producer.partitioner"))
	testutil.AssertEquals(t, "WaitForAll", viper.GetString("contexts.default.producer.requiredAcks"))
	testutil.AssertEquals(t, "1234", viper.GetString("contexts.default.producer.maxMessageBytes"))
}
