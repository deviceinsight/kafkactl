package consume

import (
	"github.com/deviceinsight/kafkactl/operations"
	"github.com/spf13/cobra"
)

var flags operations.ConsumerFlags

var CmdConsume = &cobra.Command{
	Use:   "consume",
	Short: "consume messages from a topic",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cobraCmd *cobra.Command, args []string) {
		(&operations.ConsumerOperation{}).Consume(args[0], flags)
	},
}

func init() {
	CmdConsume.Flags().BoolVarP(&flags.PrintKeys, "print-keys", "k", false, "print message printKeys")
	CmdConsume.Flags().BoolVarP(&flags.PrintTimestamps, "print-timestamps", "t", false, "print message printTimestamps")
	CmdConsume.Flags().StringVarP(&flags.ConsumerGroup, "consumer-group", "g", "", "consumer group to join")
	CmdConsume.Flags().StringArrayVarP(&flags.Offsets, "offset", "f", flags.Offsets, "define offsets for consumer")
	CmdConsume.Flags().StringVarP(&flags.OutputFormat, "output", "o", flags.OutputFormat, "Output format. One of: json|yaml")
}
