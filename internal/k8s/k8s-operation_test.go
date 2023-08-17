package k8s_test

import (
	"strings"
	"testing"
	"time"

	"github.com/deviceinsight/kafkactl/internal/helpers/avro"

	"github.com/IBM/sarama"
	"github.com/deviceinsight/kafkactl/internal"
	"github.com/deviceinsight/kafkactl/internal/env"
	"github.com/deviceinsight/kafkactl/internal/k8s"
	"github.com/deviceinsight/kafkactl/testutil"
)

func TestAllAvailableEnvironmentVariablesAreParsed(t *testing.T) {

	var context internal.ClientContext
	context.RequestTimeout = 30 * time.Second
	context.Brokers = []string{"broker1:9092", "broker2:9092"}
	context.TLS.Enabled = true
	context.TLS.CA = "my-ca"
	context.TLS.Cert = "my-cert"
	context.TLS.CertKey = "my-cert-key"
	context.TLS.Insecure = true
	context.Sasl.Enabled = true
	context.Sasl.Username = "user"
	context.Sasl.Password = "pass"
	context.Sasl.Mechanism = "scram-sha512"
	context.ClientID = "my-client"
	context.KafkaVersion = sarama.V2_0_1_0
	context.AvroSchemaRegistry = "registry:8888"
	context.AvroJSONCodec = avro.Avro
	context.Protobuf.ProtosetFiles = []string{"/usr/include/protosets/ps1.protoset", "/usr/lib/ps2.protoset"}
	context.Protobuf.ProtoImportPaths = []string{"/usr/include/protobuf", "/usr/lib/protobuf"}
	context.Protobuf.ProtoFiles = []string{"message.proto", "other.proto"}
	context.Producer.Partitioner = "hash"
	context.Producer.RequiredAcks = "WaitForAll"
	context.Producer.MaxMessageBytes = 1234

	environment := k8s.ParsePodEnvironment(context)

	envMap := make(map[string]string)

	for _, envVar := range environment {
		parts := strings.SplitAfterN(envVar, "=", 2)
		envMap[strings.TrimSuffix(parts[0], "=")] = parts[1]
	}

	for _, key := range env.Variables {
		if _, found := envMap[key]; !found {
			t.Fatalf("env variable not found in parsed environment: %s", key)
		}
	}

	testutil.AssertEquals(t, "broker1:9092 broker2:9092", envMap[env.Brokers])
	testutil.AssertEquals(t, "30s", envMap[env.RequestTimeout])
	testutil.AssertEquals(t, "true", envMap[env.TLSEnabled])
	testutil.AssertEquals(t, "my-ca", envMap[env.TLSCa])
	testutil.AssertEquals(t, "my-cert", envMap[env.TLSCert])
	testutil.AssertEquals(t, "my-cert-key", envMap[env.TLSCertKey])
	testutil.AssertEquals(t, "true", envMap[env.TLSInsecure])
	testutil.AssertEquals(t, "true", envMap[env.SaslEnabled])
	testutil.AssertEquals(t, "user", envMap[env.SaslUsername])
	testutil.AssertEquals(t, "pass", envMap[env.SaslPassword])
	testutil.AssertEquals(t, "scram-sha512", envMap[env.SaslMechanism])
	testutil.AssertEquals(t, "my-client", envMap[env.ClientID])
	testutil.AssertEquals(t, "2.0.1", envMap[env.KafkaVersion])
	testutil.AssertEquals(t, "registry:8888", envMap[env.AvroSchemaRegistry])
	testutil.AssertEquals(t, "avro", envMap[env.AvroJSONCodec])
	testutil.AssertEquals(t, "/usr/include/protosets/ps1.protoset /usr/lib/ps2.protoset", envMap[env.ProtobufProtoSetFiles])
	testutil.AssertEquals(t, "/usr/include/protobuf /usr/lib/protobuf", envMap[env.ProtobufImportPaths])
	testutil.AssertEquals(t, "message.proto other.proto", envMap[env.ProtobufProtoFiles])
	testutil.AssertEquals(t, "hash", envMap[env.ProducerPartitioner])
	testutil.AssertEquals(t, "WaitForAll", envMap[env.ProducerRequiredAcks])
	testutil.AssertEquals(t, "1234", envMap[env.ProducerMaxMessageBytes])
}
