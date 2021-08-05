package deletion

import (
	"github.com/deviceinsight/kafkactl/operations/consumergroups"
	"github.com/deviceinsight/kafkactl/operations/k8s"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/spf13/cobra"
)

func newDeleteConsumerGroupCmd() *cobra.Command {

	var cmdDeleteConsumerGroup = &cobra.Command{
		Use:     "consumer-group CONSUMER-GROUP",
		Aliases: []string{"consumer-groups"},
		Short:   "delete a consumer-group",
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if !(&k8s.K8sOperation{}).TryRun(cmd, args) {
				if err := (&consumergroups.ConsumerGroupOperation{}).DeleteConsumerGroups(args); err != nil {
					output.Fail(err)
				}
			}
		},
		ValidArgsFunction: consumergroups.CompleteConsumerGroups,
	}

	return cmdDeleteConsumerGroup
}
