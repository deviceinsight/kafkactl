package produce

import (
	"github.com/deviceinsight/kafkactl/operations/producer"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/spf13/cobra"
)

var flags producer.ProducerFlags

var CmdProduce = &cobra.Command{
	Use:   "produce",
	Short: "produce messages to a topic",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := (&producer.ProducerOperation{}).Produce(args[0], flags); err != nil {
			output.Fail(err)
		}
	},
}

func init() {
	CmdProduce.Flags().Int32VarP(&flags.Partition, "partition", "p", -1, "partition to produce to")
	CmdProduce.Flags().StringVarP(&flags.Partitioner, "partitioner", "P", "", "the partitioning scheme to use. Can be `murmur2`, `hash`, `hash-ref` `manual`, or `random`. (default is murmur2)")
	CmdProduce.Flags().StringVarP(&flags.Key, "key", "k", "", "key to use for all messages")
	CmdProduce.Flags().StringVarP(&flags.Value, "value", "v", "", "value to produce")
	CmdProduce.Flags().StringVarP(&flags.File, "file", "f", "", "file to read input from")
	CmdProduce.Flags().StringVarP(&flags.Separator, "separator", "S", "", "separator to split key and value from stdin or file")
	CmdProduce.Flags().StringVarP(&flags.LineSeparator, "lineSeparator", "L", "\n", "separator to split multiple messages from stdin or file")
	CmdProduce.Flags().IntVarP(&flags.KeySchemaVersion, "key-schema-version", "K", -1, "avro schema version that should be used for key serialization (default is latest)")
	CmdProduce.Flags().IntVarP(&flags.ValueSchemaVersion, "value-schema-version", "i", -1, "avro schema version that should be used for value serialization (default is latest)")
	CmdProduce.Flags().BoolVarP(&flags.Silent, "silent", "s", false, "do not write to standard output")
	CmdProduce.Flags().IntVarP(&flags.RateInSeconds, "rate", "r", -1, "amount of messages per second to produce on the topic")
}
