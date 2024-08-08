package global

const (
	RequestTimeout           = "REQUESTTIMEOUT"
	Brokers                  = "BROKERS"
	TLSEnabled               = "TLS_ENABLED"
	TLSCa                    = "TLS_CA"
	TLSCert                  = "TLS_CERT"
	TLSCertKey               = "TLS_CERTKEY"
	TLSInsecure              = "TLS_INSECURE"
	SaslEnabled              = "SASL_ENABLED"
	SaslUsername             = "SASL_USERNAME"
	SaslPassword             = "SASL_PASSWORD"
	SaslMechanism            = "SASL_MECHANISM"
	SaslTokenProviderPlugin  = "SASL_TOKENPROVIDER_PLUGIN"
	SaslTokenProviderOptions = "SASL_TOKENPROVIDER_OPTIONS"
	ClientID                 = "CLIENTID"
	KafkaVersion             = "KAFKAVERSION"
	AvroSchemaRegistry       = "AVRO_SCHEMAREGISTRY"
	AvroJSONCodec            = "AVRO_JSONCODEC"
	AvroRequestTimeout       = "AVRO_REQUESTTIMEOUT"
	AvroTLSEnabled           = "AVRO_TLS_ENABLED"
	AvroTLSCa                = "AVRO_TLS_CA"
	AvroTLSCert              = "AVRO_TLS_CERT"
	AvroTLSCertKey           = "AVRO_TLS_CERTKEY"
	AvroTLSInsecure          = "AVRO_TLS_INSECURE"
	AvroUsername             = "AVRO_USERNAME"
	AvroPassword             = "AVRO_PASSWORD"
	ProtobufProtoSetFiles    = "PROTOBUF_PROTOSETFILES"
	ProtobufImportPaths      = "PROTOBUF_IMPORTPATHS"
	ProtobufProtoFiles       = "PROTOBUF_PROTOFILES"
	ProducerPartitioner      = "PRODUCER_PARTITIONER"
	ProducerRequiredAcks     = "PRODUCER_REQUIREDACKS"
	ProducerMaxMessageBytes  = "PRODUCER_MAXMESSAGEBYTES"
)

var EnvVariables = []string{
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
	SaslTokenProviderPlugin,
	SaslTokenProviderOptions,
	ClientID,
	KafkaVersion,
	AvroSchemaRegistry,
	AvroJSONCodec,
	AvroRequestTimeout,
	AvroTLSEnabled,
	AvroTLSCa,
	AvroTLSCert,
	AvroTLSCertKey,
	AvroTLSInsecure,
	AvroUsername,
	AvroPassword,
	ProtobufProtoSetFiles,
	ProtobufImportPaths,
	ProtobufProtoFiles,
	ProducerPartitioner,
	ProducerRequiredAcks,
	ProducerMaxMessageBytes,
}
