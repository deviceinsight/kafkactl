package describe

import (
	"github.com/deviceinsight/kafkactl/operations"
	"github.com/spf13/cobra"
)

var flags operations.DescribeTopicFlags

var cmdDescribeTopic = &cobra.Command{
	Use:   "topic TOPIC",
	Short: "describe a topic",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		(&operations.TopicOperation{}).DescribeTopic(args[0], flags)
	},
}

func init() {
	cmdDescribeTopic.Flags().StringVarP(&flags.OutputFormat, "output", "o", flags.OutputFormat, "output format. One of: json|yaml|wide")
	cmdDescribeTopic.Flags().BoolVarP(&flags.PrintConfigs, "print-configs", "c", true, "print configs")
}
