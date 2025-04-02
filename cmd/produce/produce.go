package produce

import (
	"fmt"

	"github.com/deviceinsight/kafkactl/v5/internal"

	"github.com/deviceinsight/kafkactl/v5/internal/k8s"
	"github.com/deviceinsight/kafkactl/v5/internal/producer"
	"github.com/deviceinsight/kafkactl/v5/internal/topic"
	"github.com/spf13/cobra"
)

func NewProduceCmd() *cobra.Command {
	var flags producer.Flags

	cmdProduce := &cobra.Command{
		Use:   "produce TOPIC",
		Short: "produce messages to a topic",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if internal.IsKubernetesEnabled() {
				return k8s.NewOperation().Run(cmd, args)
			}
			return (&producer.Operation{}).Produce(args[0], flags)
		},
		ValidArgsFunction: topic.CompleteTopicNames,
	}

	cmdProduce.Flags().Int32VarP(&flags.Partition, "partition", "p", -1, "partition to produce to")
	cmdProduce.Flags().StringVarP(&flags.Partitioner, "partitioner", "P", "", "the partitioning scheme to use. Can be `murmur2`, `hash`, `hash-ref` `manual`, or `random`. (default is murmur2)")
	cmdProduce.Flags().StringVarP(&flags.RequiredAcks, "required-acks", "", "", "required acks. One of `NoResponse`, `WaitForLocal`, `WaitForAll`. (default is WaitForLocal)")
	cmdProduce.Flags().IntVarP(&flags.MaxMessageBytes, "max-message-bytes", "", 0, fmt.Sprintf("the maximum permitted size of a message (defaults to %d)", producer.DefaultMaxMessagesBytes))
	cmdProduce.Flags().StringVarP(&flags.Key, "key", "k", "", "key to use for all messages")
	cmdProduce.Flags().StringVarP(&flags.Value, "value", "v", "", "value to produce")
	cmdProduce.Flags().BoolVarP(&flags.NullValue, "null-value", "", false, "produce a null value (can be used instead of providing a value with --value)")
	cmdProduce.Flags().StringVarP(&flags.File, "file", "f", "", "file to read input from")
	cmdProduce.Flags().StringVarP(&flags.InputFormat, "input-format", "", "", "input format. One of: csv,json (default is csv)")
	cmdProduce.Flags().StringArrayVarP(&flags.Headers, "header", "H", flags.Headers, "headers in format `key:value`")
	cmdProduce.Flags().StringVarP(&flags.Separator, "separator", "S", "", "separator to split key and value from stdin or file")
	cmdProduce.Flags().StringVarP(&flags.LineSeparator, "lineSeparator", "L", "\n", "separator to split multiple messages from stdin or file")
	cmdProduce.Flags().IntVarP(&flags.KeySchemaVersion, "key-schema-version", "K", -1, "avro schema version that should be used for key serialization (default is latest)")
	cmdProduce.Flags().IntVarP(&flags.ValueSchemaVersion, "value-schema-version", "i", -1, "avro schema version that should be used for value serialization (default is latest)")
	cmdProduce.Flags().StringVarP(&flags.KeyEncoding, "key-encoding", "", flags.KeyEncoding, "key encoding (none by default). One of: none|hex|base64")
	cmdProduce.Flags().StringVarP(&flags.ValueEncoding, "value-encoding", "", flags.ValueEncoding, "value encoding (none by default). One of: none|hex|base64")
	cmdProduce.Flags().BoolVarP(&flags.Silent, "silent", "s", false, "do not write to standard output")
	cmdProduce.Flags().IntVarP(&flags.RateInSeconds, "rate", "r", -1, "amount of messages per second to produce on the topic")
	cmdProduce.Flags().StringSliceVarP(&flags.ProtoFiles, "proto-file", "", flags.ProtoFiles, "additional protobuf description file for searching message description")
	cmdProduce.Flags().StringSliceVarP(&flags.ProtoImportPaths, "proto-import-path", "", flags.ProtoImportPaths, "additional path to search files listed in proto 'import' directive")
	cmdProduce.Flags().StringSliceVarP(&flags.ProtosetFiles, "protoset-file", "", flags.ProtosetFiles, "additional compiled protobuf description file for searching message description")
	cmdProduce.Flags().StringVarP(&flags.KeyProtoType, "key-proto-type", "", flags.KeyProtoType, "key protobuf message type")
	cmdProduce.Flags().StringVarP(&flags.ValueProtoType, "value-proto-type", "", flags.ValueProtoType, "value protobuf message type")

	return cmdProduce
}
