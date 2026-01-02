package consume

import (
	"strings"

	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/consume"
	"github.com/deviceinsight/kafkactl/v5/internal/consumergroups"
	"github.com/deviceinsight/kafkactl/v5/internal/helpers/protobuf"
	"github.com/deviceinsight/kafkactl/v5/internal/k8s"
	"github.com/deviceinsight/kafkactl/v5/internal/topic"
	"github.com/spf13/cobra"
)

func NewConsumeCmd() *cobra.Command {
	var flags consume.Flags

	cmdConsume := &cobra.Command{
		Use:   "consume TOPIC",
		Short: "consume messages from a topic",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if internal.IsKubernetesEnabled() {
				return k8s.NewOperation().Run(cmd, args)
			}
			return (&consume.Operation{}).Consume(args[0], flags)
		},
		ValidArgsFunction: topic.CompleteTopicNames,
	}

	cmdConsume.Flags().BoolVarP(&flags.PrintPartitions, "print-partitions", "", false, "print message partitions")
	cmdConsume.Flags().BoolVarP(&flags.PrintKeys, "print-keys", "k", false, "print message keys")
	cmdConsume.Flags().BoolVarP(&flags.PrintTimestamps, "print-timestamps", "t", false, "print message timestamps")
	cmdConsume.Flags().BoolVarP(&flags.PrintSchema, "print-schema", "a", false, "print details about schema used for decoding")
	cmdConsume.Flags().BoolVarP(&flags.PrintHeaders, "print-headers", "", false, "print message headers")
	cmdConsume.Flags().BoolVarP(&flags.PrintAll, "print-all", "", false, "print all messages details")
	cmdConsume.Flags().IntVarP(&flags.Tail, "tail", "", -1, "show only the last n messages on the topic")
	cmdConsume.Flags().StringVarP(&flags.FromTimestamp, "from-timestamp", "", "", "consume data from offset of given timestamp")
	cmdConsume.Flags().StringVarP(&flags.ToTimestamp, "to-timestamp", "", "", "consume data till offset of given timestamp")
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
	cmdConsume.Flags().StringSliceVarP(&flags.ProtoMarshalOptions, "proto-marshal-option", "", flags.ProtoMarshalOptions, "json marshall options to use for protobuf. Format is key=value. Valid keys are "+strings.Join(protobuf.AllMarshalOptions, ","))
	cmdConsume.Flags().StringVarP(&flags.KeyProtoType, "key-proto-type", "", flags.KeyProtoType, "key protobuf message type")
	cmdConsume.Flags().StringVarP(&flags.ValueProtoType, "value-proto-type", "", flags.ValueProtoType, "value protobuf message type")
	cmdConsume.Flags().StringVarP(&flags.FilterKey, "filter-key", "", "", "filter messages keys with glob pattern")
	cmdConsume.Flags().StringVarP(&flags.FilterValue, "filter-value", "", "", "filter messages values with glob pattern")
	cmdConsume.Flags().StringToStringVarP(&flags.FilterHeader, "filter-header", "", map[string]string{}, "filter messages headers with glob pattern")
	cmdConsume.Flags().StringVarP(&flags.IsolationLevel, "isolation-level", "i", "", "isolationLevel to use. One of: ReadUncommitted|ReadCommitted")

	if err := cmdConsume.RegisterFlagCompletionFunc("group", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return consumergroups.CompleteConsumerGroups(cmd, args, toComplete)
	}); err != nil {
		panic(err)
	}

	return cmdConsume
}
