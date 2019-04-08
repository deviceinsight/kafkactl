package describe

import (
	"github.com/spf13/cobra"
)

var CmdDescribe = &cobra.Command{
	Use:   "describe",
	Short: "describe topics, consumerGroups",
}

func init() {
	CmdDescribe.AddCommand(cmdDescribeTopic)
	CmdDescribe.AddCommand(cmdDescribeConsumerGroup)
}
