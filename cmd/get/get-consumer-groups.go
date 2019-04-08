package get

import (
	"github.com/deviceinsight/kafkactl/operations/consumergroups"
	"github.com/spf13/cobra"
)

var consumerGroupFlags consumergroups.GetConsumerGroupFlags

var cmdGetConsumerGroups = &cobra.Command{
	Use:     "consumer-groups",
	Aliases: []string{"cg"},
	Short:   "list available consumerGroups",
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		topic := ""
		if len(args) > 0 {
			topic = args[0]
		}
		(&consumergroups.ConsumerGroupOperation{}).GetConsumerGroups(consumerGroupFlags, topic)
	},
}

func init() {
	cmdGetConsumerGroups.Flags().StringVarP(&consumerGroupFlags.OutputFormat, "output", "o", consumerGroupFlags.OutputFormat, "output format. One of: json|yaml|wide|compact")
}
