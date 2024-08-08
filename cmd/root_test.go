package cmd_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/deviceinsight/kafkactl/v5/internal/global"

	"github.com/deviceinsight/kafkactl/v5/internal/testutil"
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

	_ = os.Setenv(global.RequestTimeout, "30")
	_ = os.Setenv(global.Brokers, "broker1:9092 broker2:9092")
	_ = os.Setenv(global.TLSEnabled, "true")
	_ = os.Setenv(global.TLSCa, "my-ca")
	_ = os.Setenv(global.TLSCert, "my-cert")
	_ = os.Setenv(global.TLSCertKey, "my-cert-key")
	_ = os.Setenv(global.TLSInsecure, "true")
	_ = os.Setenv(global.SaslEnabled, "true")
	_ = os.Setenv(global.SaslUsername, "user")
	_ = os.Setenv(global.SaslPassword, "pass")
	_ = os.Setenv(global.SaslMechanism, "oauth")
	_ = os.Setenv(global.SaslTokenProviderPlugin, "azure")
	_ = os.Setenv(global.SaslTokenProviderOptions, `{"tenantid": "azure-tenant-id", "int-key": 12}`)
	_ = os.Setenv(global.ClientID, "my-client")
	_ = os.Setenv(global.KafkaVersion, "2.0.1")
	_ = os.Setenv(global.AvroSchemaRegistry, "registry:8888")
	_ = os.Setenv(global.AvroJSONCodec, "avro")
	_ = os.Setenv(global.AvroRequestTimeout, "10")
	_ = os.Setenv(global.AvroTLSEnabled, "true")
	_ = os.Setenv(global.AvroTLSCa, "my-avro-ca")
	_ = os.Setenv(global.AvroTLSCert, "my-avro-cert")
	_ = os.Setenv(global.AvroTLSCertKey, "my-avro-cert-key")
	_ = os.Setenv(global.AvroTLSInsecure, "true")
	_ = os.Setenv(global.AvroUsername, "avro-user")
	_ = os.Setenv(global.AvroPassword, "avro-pass")
	_ = os.Setenv(global.ProtobufProtoSetFiles, "/usr/include/protosets/ps1.protoset /usr/lib/ps2.protoset")
	_ = os.Setenv(global.ProtobufImportPaths, "/usr/include/protobuf /usr/lib/protobuf")
	_ = os.Setenv(global.ProtobufProtoFiles, "message.proto other.proto")
	_ = os.Setenv(global.ProducerPartitioner, "hash")
	_ = os.Setenv(global.ProducerRequiredAcks, "WaitForAll")
	_ = os.Setenv(global.ProducerMaxMessageBytes, "1234")

	for _, key := range global.EnvVariables {
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
	testutil.AssertEquals(t, "oauth", viper.GetString("contexts.default.sasl.mechanism"))
	testutil.AssertEquals(t, "azure", viper.GetString("contexts.default.sasl.tokenProvider.plugin"))
	testutil.AssertEquals(t, "azure-tenant-id", viper.GetStringMap("contexts.default.sasl.tokenProvider.options")["tenantid"].(string))
	testutil.AssertEquals(t, "12", fmt.Sprint(viper.GetStringMap("contexts.default.sasl.tokenProvider.options")["int-key"].(float64)))
	testutil.AssertEquals(t, "my-client", viper.GetString("contexts.default.clientID"))
	testutil.AssertEquals(t, "2.0.1", viper.GetString("contexts.default.kafkaVersion"))
	testutil.AssertEquals(t, "registry:8888", viper.GetString("contexts.default.avro.schemaRegistry"))
	testutil.AssertEquals(t, "avro", viper.GetString("contexts.default.avro.jsonCodec"))
	testutil.AssertEquals(t, "10", viper.GetString("contexts.default.avro.requestTimeout"))
	testutil.AssertEquals(t, "true", viper.GetString("contexts.default.avro.tls.enabled"))
	testutil.AssertEquals(t, "my-avro-ca", viper.GetString("contexts.default.avro.tls.ca"))
	testutil.AssertEquals(t, "my-avro-cert", viper.GetString("contexts.default.avro.tls.cert"))
	testutil.AssertEquals(t, "my-avro-cert-key", viper.GetString("contexts.default.avro.tls.certKey"))
	testutil.AssertEquals(t, "true", viper.GetString("contexts.default.avro.tls.insecure"))
	testutil.AssertEquals(t, "avro-user", viper.GetString("contexts.default.avro.username"))
	testutil.AssertEquals(t, "avro-pass", viper.GetString("contexts.default.avro.password"))
	testutil.AssertEquals(t, "/usr/include/protosets/ps1.protoset", viper.GetStringSlice("contexts.default.protobuf.protosetFiles")[0])
	testutil.AssertEquals(t, "/usr/include/protobuf", viper.GetStringSlice("contexts.default.protobuf.importPaths")[0])
	testutil.AssertEquals(t, "message.proto", viper.GetStringSlice("contexts.default.protobuf.protoFiles")[0])
	testutil.AssertEquals(t, "hash", viper.GetString("contexts.default.producer.partitioner"))
	testutil.AssertEquals(t, "WaitForAll", viper.GetString("contexts.default.producer.requiredAcks"))
	testutil.AssertEquals(t, "1234", viper.GetString("contexts.default.producer.maxMessageBytes"))
}
