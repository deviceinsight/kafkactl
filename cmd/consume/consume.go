package consume

import (
	"github.com/deviceinsight/kafkactl/operations/consumer"
	"github.com/spf13/cobra"
)

var flags consumer.ConsumerFlags

var CmdConsume = &cobra.Command{
	Use:   "consume TOPIC",
	Short: "consume messages from a topic",
	Args:  cobra.ExactArgs(1),
	Run: func(cobraCmd *cobra.Command, args []string) {
		(&consumer.ConsumerOperation{}).Consume(args[0], flags)
	},
}

func init() {
	CmdConsume.Flags().BoolVarP(&flags.PrintKeys, "print-keys", "k", false, "print message keys")
	CmdConsume.Flags().BoolVarP(&flags.PrintTimestamps, "print-timestamps", "t", false, "print message timestamps")
	CmdConsume.Flags().BoolVarP(&flags.PrintAvroSchema, "print-schema", "a", false, "print details about avro schema used for decoding")
	CmdConsume.Flags().IntSliceP("partitions", "p", flags.Partitions, "partitions to consume. The default is to consume from all partitions.")
	CmdConsume.Flags().StringArrayVarP(&flags.Offsets, "offset", "", flags.Offsets, "offsets in format `partition=offset (for partitions not specified, other parameters apply)`")
	CmdConsume.Flags().BoolVarP(&flags.FromBeginning, "from-beginning", "b", false, "set offset for consumer to the oldest offset")
	CmdConsume.Flags().StringVarP(&flags.OutputFormat, "output", "o", flags.OutputFormat, "Output format. One of: json|yaml")
}
