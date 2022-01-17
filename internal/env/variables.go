package env

const (
	RequestTimeout        = "REQUESTTIMEOUT"
	Brokers               = "BROKERS"
	TLSEnabled            = "TLS_ENABLED"
	TLSCa                 = "TLS_CA"
	TLSCert               = "TLS_CERT"
	TLSCertKey            = "TLS_CERTKEY"
	TLSInsecure           = "TLS_INSECURE"
	SaslEnabled           = "SASL_ENABLED"
	SaslUsername          = "SASL_USERNAME"
	SaslPassword          = "SASL_PASSWORD"
	SaslMechanism         = "SASL_MECHANISM"
	ClientID              = "CLIENTID"
	KafkaVersion          = "KAFKAVERSION"
	AvroSchemaRegistry    = "AVRO_SCHEMAREGISTRY"
	ProtobufProtoSetFiles = "PROTOBUF_PROTOSETFILES"
	ProtobufImportPaths   = "PROTOBUF_IMPORTPATHS"
	ProtobufProtoFiles    = "PROTOBUF_PROTOFILES"
	ProducerPartitioner   = "PRODUCER_PARTITIONER"
	ProducerRequiredAcks  = "PRODUCER_REQUIREDACKS"
)

var Variables = []string{
	RequestTimeout,
	Brokers,
	TLSEnabled,
	TLSCa,
	TLSCert,
	TLSCertKey,
	TLSInsecure,
	SaslEnabled,
	SaslUsername,
	SaslPassword,
	SaslMechanism,
	ClientID,
	KafkaVersion,
	AvroSchemaRegistry,
	ProtobufProtoSetFiles,
	ProtobufImportPaths,
	ProtobufProtoFiles,
	ProducerPartitioner,
	ProducerRequiredAcks,
}
