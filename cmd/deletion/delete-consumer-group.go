package deletion

import (
	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/consumergroups"
	"github.com/deviceinsight/kafkactl/v5/internal/k8s"
	"github.com/spf13/cobra"
)

func newDeleteConsumerGroupCmd() *cobra.Command {

	var cmdDeleteConsumerGroup = &cobra.Command{
		Use:     "consumer-group CONSUMER-GROUP",
		Aliases: []string{"consumer-groups"},
		Short:   "delete a consumer-group",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if internal.IsKubernetesEnabled() {
				return k8s.NewOperation().Run(cmd, args)
			}
			return (&consumergroups.ConsumerGroupOperation{}).DeleteConsumerGroups(args)
		},
		ValidArgsFunction: consumergroups.CompleteConsumerGroups,
	}

	return cmdDeleteConsumerGroup
}
