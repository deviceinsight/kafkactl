package consume

import (
	"github.com/deviceinsight/kafkactl/operations/consumer"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/spf13/cobra"
)

var flags consumer.ConsumerFlags

func NewConsumeCmd() *cobra.Command {

	var cmdConsume = &cobra.Command{
		Use:   "consume TOPIC",
		Short: "consume messages from a topic",
		Args:  cobra.ExactArgs(1),
		Run: func(cobraCmd *cobra.Command, args []string) {
			if err := (&consumer.ConsumerOperation{}).Consume(args[0], flags); err != nil {
				output.Fail(err)
			}
		},
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

	return cmdConsume
}
