package produce

import (
	"github.com/deviceinsight/kafkactl/operations/producer"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/spf13/cobra"
)

var flags producer.ProducerFlags

func NewProduceCmd() *cobra.Command {

	var cmdProduce = &cobra.Command{
		Use:   "produce",
		Short: "produce messages to a topic",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := (&producer.ProducerOperation{}).Produce(args[0], flags); err != nil {
				output.Fail(err)
			}
		},
	}

	cmdProduce.Flags().Int32VarP(&flags.Partition, "partition", "p", -1, "partition to produce to")
	cmdProduce.Flags().StringVarP(&flags.Partitioner, "partitioner", "P", "", "the partitioning scheme to use. Can be `murmur2`, `hash`, `hash-ref` `manual`, or `random`. (default is murmur2)")
	cmdProduce.Flags().StringVarP(&flags.Key, "key", "k", "", "key to use for all messages")
	cmdProduce.Flags().StringVarP(&flags.Value, "value", "v", "", "value to produce")
	cmdProduce.Flags().StringVarP(&flags.File, "file", "f", "", "file to read input from")
	cmdProduce.Flags().StringVarP(&flags.Separator, "separator", "S", "", "separator to split key and value from stdin or file")
	cmdProduce.Flags().StringVarP(&flags.LineSeparator, "lineSeparator", "L", "\n", "separator to split multiple messages from stdin or file")
	cmdProduce.Flags().IntVarP(&flags.KeySchemaVersion, "key-schema-version", "K", -1, "avro schema version that should be used for key serialization (default is latest)")
	cmdProduce.Flags().IntVarP(&flags.ValueSchemaVersion, "value-schema-version", "i", -1, "avro schema version that should be used for value serialization (default is latest)")
	cmdProduce.Flags().BoolVarP(&flags.Silent, "silent", "s", false, "do not write to standard output")
	cmdProduce.Flags().IntVarP(&flags.RateInSeconds, "rate", "r", -1, "amount of messages per second to produce on the topic")

	return cmdProduce
}
