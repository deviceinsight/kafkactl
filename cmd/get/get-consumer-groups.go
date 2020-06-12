package get

import (
	"github.com/deviceinsight/kafkactl/operations/consumergroups"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/spf13/cobra"
)

var consumerGroupFlags consumergroups.GetConsumerGroupFlags

func newGetConsumerGroupsCmd() *cobra.Command {

	var cmdGetConsumerGroups = &cobra.Command{
		Use:     "consumer-groups",
		Aliases: []string{"cg"},
		Short:   "list available consumerGroups",
		Args:    cobra.MaximumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := (&consumergroups.ConsumerGroupOperation{}).GetConsumerGroups(consumerGroupFlags); err != nil {
				output.Fail(err)
			}
		},
	}

	cmdGetConsumerGroups.Flags().StringVarP(&consumerGroupFlags.OutputFormat, "output", "o", consumerGroupFlags.OutputFormat, "output format. One of: json|yaml|wide|compact")
	cmdGetConsumerGroups.Flags().StringVarP(&consumerGroupFlags.FilterTopic, "topic", "t", "", "show groups for given topic only")

	return cmdGetConsumerGroups
}
