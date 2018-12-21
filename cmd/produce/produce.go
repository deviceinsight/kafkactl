package produce

import (
	"github.com/deviceinsight/kafkactl/operations"

	"github.com/spf13/cobra"
)

var flags operations.ProducerFlags

var CmdProduce = &cobra.Command{
	Use:   "produce",
	Short: "produce messages to a topic",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		(&operations.ProducerOperation{}).Produce(args[0], flags)
	},
}

func init() {
	CmdProduce.Flags().Int32VarP(&flags.Partition, "partition", "p", -1, "partition to produce to")
	CmdProduce.Flags().StringVarP(&flags.Partitioner, "partitioner", "P", "", "The partitioning scheme to use. Can be `hash`, `manual`, or `random`")
	CmdProduce.Flags().StringVarP(&flags.Key, "key", "k", "", "key to use for all messages")
	CmdProduce.Flags().StringVarP(&flags.Value, "value", "v", "", "value to produce")
	CmdProduce.Flags().StringVarP(&flags.Separator, "separator", "S", "", "separator to split key and value from stdin")
	CmdProduce.Flags().BoolVarP(&flags.Silent, "silent", "s", false, "do not write to standard output")
}
