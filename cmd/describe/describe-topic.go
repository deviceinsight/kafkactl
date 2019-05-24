package describe

import (
	"github.com/deviceinsight/kafkactl/operations"
	"github.com/spf13/cobra"
)

var cmdDescribeTopic = &cobra.Command{
	Use:   "topic TOPIC",
	Short: "describe a topic",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		(&operations.TopicOperation{}).DescribeTopic(args[0])
	},
}

func init() {
}
