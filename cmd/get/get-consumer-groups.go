package get

import (
	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/consumergroups"
	"github.com/deviceinsight/kafkactl/v5/internal/k8s"
	"github.com/deviceinsight/kafkactl/v5/internal/topic"
	"github.com/spf13/cobra"
)

func newGetConsumerGroupsCmd() *cobra.Command {

	var flags consumergroups.GetConsumerGroupFlags

	var cmdGetConsumerGroups = &cobra.Command{
		Use:     "consumer-groups",
		Aliases: []string{"cg"},
		Short:   "list available consumerGroups",
		Args:    cobra.MaximumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			if internal.IsKubernetesEnabled() {
				return k8s.NewOperation().Run(cmd, args)
			}
			return (&consumergroups.ConsumerGroupOperation{}).GetConsumerGroups(flags)
		},
	}

	cmdGetConsumerGroups.Flags().StringVarP(&flags.OutputFormat, "output", "o", flags.OutputFormat, "output format. One of: json|yaml|wide|compact")
	cmdGetConsumerGroups.Flags().StringVarP(&flags.FilterTopic, "topic", "t", "", "show groups for given topic only")

	if err := cmdGetConsumerGroups.RegisterFlagCompletionFunc("topic", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return topic.CompleteTopicNames(cmd, args, toComplete)
	}); err != nil {
		panic(err)
	}

	return cmdGetConsumerGroups
}
