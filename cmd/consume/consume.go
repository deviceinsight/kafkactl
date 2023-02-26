package consume

import (
	"github.com/deviceinsight/kafkactl/internal/consume"
	"github.com/deviceinsight/kafkactl/internal/consumergroups"
	"github.com/deviceinsight/kafkactl/internal/k8s"
	"github.com/deviceinsight/kafkactl/internal/topic"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/spf13/cobra"
)

func NewConsumeCmd() *cobra.Command {

	var flags consume.Flags

	var cmdConsume = &cobra.Command{
		Use:   "consume TOPIC",
		Short: "consume messages from a topic",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if !(&k8s.Operation{}).TryRun(cmd, args) {
				if err := (&consume.Operation{}).Consume(args[0], flags); err != nil {
					output.Fail(err)
				}
			}
		},
		ValidArgsFunction: topic.CompleteTopicNames,
	}

	cmdConsume.Flags().BoolVarP(&flags.PrintPartitions, "print-partitions", "", false, "print message partitions")
	cmdConsume.Flags().BoolVarP(&flags.PrintKeys, "print-keys", "k", false, "print message keys")
	cmdConsume.Flags().BoolVarP(&flags.PrintTimestamps, "print-timestamps", "t", false, "print message timestamps")
	cmdConsume.Flags().BoolVarP(&flags.PrintAvroSchema, "print-schema", "a", false, "print details about avro schema used for decoding")
	cmdConsume.Flags().BoolVarP(&flags.PrintHeaders, "print-headers", "", false, "print message headers")
	cmdConsume.Flags().IntVarP(&flags.Tail, "tail", "", -1, "show only the last n messages on the topic")
	cmdConsume.Flags().Int64VarP(&flags.FromTs, "from-timestamp", "", -1, "consume data from offset of given timestamp")
	cmdConsume.Flags().Int64VarP(&flags.EndTs, "to-timestamp", "", -1, "consume data till offset of given timestamp")
	cmdConsume.Flags().Int64VarP(&flags.MaxMessages, "max-messages", "", -1, "stop consuming after n messages have been read")
	cmdConsume.Flags().BoolVarP(&flags.Exit, "exit", "e", flags.Exit, "stop consuming when latest offset is reached")
	cmdConsume.Flags().IntSliceVarP(&flags.Partitions, "partitions", "p", flags.Partitions, "partitions to consume. The default is to consume from all partitions.")
	cmdConsume.Flags().StringVarP(&flags.Separator, "separator", "s", "#", "separator to split key and value")
	cmdConsume.Flags().StringVarP(&flags.Group, "group", "g", "", "consumer group to join")
	cmdConsume.Flags().StringArrayVarP(&flags.Offsets, "offset", "", flags.Offsets, "offsets in format `partition=offset (for partitions not specified, other parameters apply)`")
	cmdConsume.Flags().BoolVarP(&flags.FromBeginning, "from-beginning", "b", false, "set offset for consumer to the oldest offset")
	cmdConsume.Flags().StringVarP(&flags.OutputFormat, "output", "o", flags.OutputFormat, "output format. One of: json|yaml")
	cmdConsume.Flags().StringVarP(&flags.EncodeKey, "key-encoding", "", flags.EncodeKey, "key encoding (auto-detected by default). One of: none|hex|base64")
	cmdConsume.Flags().StringVarP(&flags.EncodeValue, "value-encoding", "", flags.EncodeValue, "value encoding (auto-detected by default). One of: none|hex|base64")
	cmdConsume.Flags().StringSliceVarP(&flags.ProtoFiles, "proto-file", "", flags.ProtoFiles, "additional protobuf description file for searching message description")
	cmdConsume.Flags().StringSliceVarP(&flags.ProtoImportPaths, "proto-import-path", "", flags.ProtoImportPaths, "additional path to search files listed in proto 'import' directive")
	cmdConsume.Flags().StringSliceVarP(&flags.ProtosetFiles, "protoset-file", "", flags.ProtosetFiles, "additional compiled protobuf description file for searching message description")
	cmdConsume.Flags().StringVarP(&flags.KeyProtoType, "key-proto-type", "", flags.KeyProtoType, "key protobuf message type")
	cmdConsume.Flags().StringVarP(&flags.ValueProtoType, "value-proto-type", "", flags.ValueProtoType, "value protobuf message type")

	if err := cmdConsume.RegisterFlagCompletionFunc("group", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return consumergroups.CompleteConsumerGroups(cmd, args, toComplete)
	}); err != nil {
		panic(err)
	}

	return cmdConsume
}
