package clone

import (
	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/consumergroupoffsets"
	"github.com/deviceinsight/kafkactl/v5/internal/consumergroups"
	"github.com/deviceinsight/kafkactl/v5/internal/k8s"
	"github.com/spf13/cobra"
)

func newCloneConsumerGroupCmd() *cobra.Command {

	var cloneConsumerGroupCmd = &cobra.Command{
		Use:     "consumer-group SOURCE_GROUP TARGET_GROUP",
		Aliases: []string{"cg"},
		Short:   "clone existing consumerGroup with all offsets",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if internal.IsKubernetesEnabled() {
				return k8s.NewOperation().Run(cmd, args)
			}
			return (&consumergroupoffsets.ConsumerGroupOffsetOperation{}).CloneConsumerGroup(args[0], args[1])
		},
		ValidArgsFunction: consumergroups.CompleteConsumerGroups,
	}

	return cloneConsumerGroupCmd
}
