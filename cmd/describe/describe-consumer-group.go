package describe

import (
	"github.com/deviceinsight/kafkactl/operations/consumergroups"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/spf13/cobra"
)

func newDescribeConsumerGroupCmd() *cobra.Command {

	var flags consumergroups.DescribeConsumerGroupFlags

	var cmdDescribeConsumerGroup = &cobra.Command{
		Use:     "consumer-group GROUP",
		Aliases: []string{"cg"},
		Short:   "describe a consumerGroup",
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := (&consumergroups.ConsumerGroupOperation{}).DescribeConsumerGroup(flags, args[0]); err != nil {
				output.Fail(err)
			}
		},
	}

	cmdDescribeConsumerGroup.Flags().BoolVarP(&flags.OnlyPartitionsWithLag, "only-with-lag", "l", false, "show only partitions that have a lag")
	cmdDescribeConsumerGroup.Flags().StringVarP(&flags.FilterTopic, "topic", "t", "", "show group details for given topic only")
	cmdDescribeConsumerGroup.Flags().StringVarP(&flags.OutputFormat, "output", "o", flags.OutputFormat, "output format. One of: json|yaml|wide")
	cmdDescribeConsumerGroup.Flags().BoolVarP(&flags.PrintTopics, "print-topics", "T", true, "print topic details")
	cmdDescribeConsumerGroup.Flags().BoolVarP(&flags.PrintMembers, "print-members", "m", true, "print group members")

	return cmdDescribeConsumerGroup
}
