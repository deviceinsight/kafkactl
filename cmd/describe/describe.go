package describe

import (
	"github.com/spf13/cobra"
)

func NewDescribeCmd() *cobra.Command {

	var cmdDescribe = &cobra.Command{
		Use:   "describe",
		Short: "describe topics, consumerGroups, brokers, users",
	}

	cmdDescribe.AddCommand(newDescribeTopicCmd())
	cmdDescribe.AddCommand(newDescribeConsumerGroupCmd())
	cmdDescribe.AddCommand(newDescribeBrokerCmd())
	cmdDescribe.AddCommand(newDescribeUserCmd())

	return cmdDescribe
}
