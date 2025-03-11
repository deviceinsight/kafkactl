package global

const (
	RequestTimeout               = "REQUESTTIMEOUT"
	Brokers                      = "BROKERS"
	TLSEnabled                   = "TLS_ENABLED"
	TLSCa                        = "TLS_CA"
	TLSCert                      = "TLS_CERT"
	TLSCertKey                   = "TLS_CERTKEY"
	TLSInsecure                  = "TLS_INSECURE"
	SaslEnabled                  = "SASL_ENABLED"
	SaslUsername                 = "SASL_USERNAME"
	SaslPassword                 = "SASL_PASSWORD"
	SaslMechanism                = "SASL_MECHANISM"
	SaslTokenProviderPlugin      = "SASL_TOKENPROVIDER_PLUGIN"
	SaslTokenProviderOptions     = "SASL_TOKENPROVIDER_OPTIONS"
	ClientID                     = "CLIENTID"
	KafkaVersion                 = "KAFKAVERSION"
	AvroJSONCodec                = "AVRO_JSONCODEC"
	SchemaRegistryURL            = "SCHEMAREGISTRY_URL"
	SchemaRegistryRequestTimeout = "SCHEMAREGISTRY_REQUESTTIMEOUT"
	SchemaRegistryTLSEnabled     = "SCHEMAREGISTRY_TLS_ENABLED"
	SchemaRegistryTLSCa          = "SCHEMAREGISTRY_TLS_CA"
	SchemaRegistryTLSCert        = "SCHEMAREGISTRY_TLS_CERT"
	SchemaRegistryTLSCertKey     = "SCHEMAREGISTRY_TLS_CERTKEY"
	SchemaRegistryTLSInsecure    = "SCHEMAREGISTRY_TLS_INSECURE"
	SchemaRegistryUsername       = "SCHEMAREGISTRY_USERNAME"
	SchemaRegistryPassword       = "SCHEMAREGISTRY_PASSWORD"
	ProtobufProtoSetFiles        = "PROTOBUF_PROTOSETFILES"
	ProtobufImportPaths          = "PROTOBUF_IMPORTPATHS"
	ProtobufProtoFiles           = "PROTOBUF_PROTOFILES"
	ProducerPartitioner          = "PRODUCER_PARTITIONER"
	ProducerRequiredAcks         = "PRODUCER_REQUIREDACKS"
	ProducerMaxMessageBytes      = "PRODUCER_MAXMESSAGEBYTES"
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
	AvroJSONCodec,
	SchemaRegistryURL,
	SchemaRegistryRequestTimeout,
	SchemaRegistryTLSEnabled,
	SchemaRegistryTLSCa,
	SchemaRegistryTLSCert,
	SchemaRegistryTLSCertKey,
	SchemaRegistryTLSInsecure,
	SchemaRegistryUsername,
	SchemaRegistryPassword,
	ProtobufProtoSetFiles,
	ProtobufImportPaths,
	ProtobufProtoFiles,
	ProducerPartitioner,
	ProducerRequiredAcks,
	ProducerMaxMessageBytes,
}
