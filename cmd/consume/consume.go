package consume

import (
	"github.com/deviceinsight/kafkactl/operations"
	"github.com/deviceinsight/kafkactl/operations/consumer"
	"github.com/deviceinsight/kafkactl/operations/k8s"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/spf13/cobra"
)

func NewConsumeCmd() *cobra.Command {

	var flags consumer.ConsumerFlags

	var cmdConsume = &cobra.Command{
		Use:   "consume TOPIC",
		Short: "consume messages from a topic",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if !(&k8s.K8sOperation{}).TryRun(cmd, args) {
				if err := (&consumer.ConsumerOperation{}).Consume(args[0], flags); err != nil {
					output.Fail(err)
				}
			}
		},
		ValidArgsFunction: operations.CompleteTopicNames,
	}

	cmdConsume.Flags().BoolVarP(&flags.PrintKeys, "print-keys", "k", false, "print message keys")
	cmdConsume.Flags().BoolVarP(&flags.PrintTimestamps, "print-timestamps", "t", false, "print message timestamps")
	cmdConsume.Flags().BoolVarP(&flags.PrintAvroSchema, "print-schema", "a", false, "print details about avro schema used for decoding")
	cmdConsume.Flags().BoolVarP(&flags.PrintHeaders, "print-headers", "", false, "print message headers")
	cmdConsume.Flags().IntVarP(&flags.Tail, "tail", "", -1, "show only the last n messages on the topic")
	cmdConsume.Flags().BoolVarP(&flags.Exit, "exit", "e", flags.Exit, "stop consuming when latest offset is reached")
	cmdConsume.Flags().IntSliceVarP(&flags.Partitions, "partitions", "p", flags.Partitions, "partitions to consume. The default is to consume from all partitions.")
	cmdConsume.Flags().StringVarP(&flags.Separator, "separator", "s", "#", "separator to split key and value")
	cmdConsume.Flags().StringArrayVarP(&flags.Offsets, "offset", "", flags.Offsets, "offsets in format `partition=offset (for partitions not specified, other parameters apply)`")
	cmdConsume.Flags().BoolVarP(&flags.FromBeginning, "from-beginning", "b", false, "set offset for consumer to the oldest offset")
	cmdConsume.Flags().StringVarP(&flags.OutputFormat, "output", "o", flags.OutputFormat, "output format. One of: json|yaml")
	cmdConsume.Flags().StringVarP(&flags.EncodeKey, "key-encoding", "", flags.EncodeKey, "key encoding (auto-detected by default). One of: none|hex|base64")
	cmdConsume.Flags().StringVarP(&flags.EncodeValue, "value-encoding", "", flags.EncodeValue, "value encoding (auto-detected by default). One of: none|hex|base64")

	return cmdConsume
}
