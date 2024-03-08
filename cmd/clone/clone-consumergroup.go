package clone

import (
	"github.com/deviceinsight/kafkactl/v5/internal/consumergroupoffsets"
	"github.com/deviceinsight/kafkactl/v5/internal/consumergroups"
	"github.com/deviceinsight/kafkactl/v5/internal/k8s"
	"github.com/deviceinsight/kafkactl/v5/internal/output"
	"github.com/spf13/cobra"
)

func newCloneConsumerGroupCmd() *cobra.Command {

	var cloneConsumerGroupCmd = &cobra.Command{
		Use:     "consumer-group SOURCE_GROUP TARGET_GROUP",
		Aliases: []string{"cg"},
		Short:   "clone existing consumerGroup with all offsets",
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			if !k8s.NewOperation().TryRun(cmd, args) {
				if err := (&consumergroupoffsets.ConsumerGroupOffsetOperation{}).CloneConsumerGroup(args[0], args[1]); err != nil {
					output.Fail(err)
				}
			}
		},
		ValidArgsFunction: consumergroups.CompleteConsumerGroups,
	}

	return cloneConsumerGroupCmd
}
