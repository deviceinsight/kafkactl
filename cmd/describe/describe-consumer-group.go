package describe

import (
	"github.com/deviceinsight/kafkactl/operations/consumergroups"
	"github.com/spf13/cobra"
)

var consumerGroupFlags consumergroups.DescribeConsumerGroupFlags

var cmdDescribeConsumerGroup = &cobra.Command{
	Use:     "consumer-group GROUP",
	Aliases: []string{"cg"},
	Short:   "describe a consumerGroup",
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		(&consumergroups.ConsumerGroupOperation{}).DescribeConsumerGroup(consumerGroupFlags, args[0])
	},
}

func init() {
	cmdDescribeConsumerGroup.Flags().BoolVarP(&consumerGroupFlags.ShowPartitionDetails, "partitions", "p", false, "show detailed information for each partition")
	cmdDescribeConsumerGroup.Flags().StringVarP(&consumerGroupFlags.FilterTopic, "topic", "t", "", "show group details for given topic only")
}
